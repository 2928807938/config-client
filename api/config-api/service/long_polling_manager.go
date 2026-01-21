package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"config-client/config/domain/listener"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// WaitRequest 等待请求
type WaitRequest struct {
	ConfigKeys []string          // 配置键列表
	Versions   map[string]string // 配置键 -> 版本号映射
	ResultChan chan *WaitResult  // 结果通道
}

// WaitResult 等待结果
type WaitResult struct {
	Changed    bool              // 是否有变更
	ConfigKeys []string          // 变更的配置键列表
	Versions   map[string]string // 最新版本号映射
}

// LongPollingManager 长轮询管理器
type LongPollingManager struct {
	listener      listener.ConfigListener      // 配置监听器
	waitRequests  map[string][]*WaitRequest    // 配置键 -> 等待请求列表
	mu            sync.RWMutex                 // 读写锁
	timeout       time.Duration                // 超时时间
	ctx           context.Context              // 上下文
	cancel        context.CancelFunc           // 取消函数
	versionGetter func(string) (string, error) // 版本获取函数
}

// NewLongPollingManager 创建长轮询管理器
func NewLongPollingManager(
	listener listener.ConfigListener,
	timeout time.Duration,
	versionGetter func(string) (string, error),
) *LongPollingManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &LongPollingManager{
		listener:      listener,
		waitRequests:  make(map[string][]*WaitRequest),
		timeout:       timeout,
		ctx:           ctx,
		cancel:        cancel,
		versionGetter: versionGetter,
	}
}

// Start 启动长轮询管理器
func (m *LongPollingManager) Start() error {
	// 订阅配置变更事件
	eventChan, err := m.listener.Subscribe(m.ctx)
	if err != nil {
		return fmt.Errorf("订阅配置变更失败: %w", err)
	}

	// 启动事件监听
	go m.handleEvents(eventChan)

	hlog.Info("长轮询管理器已启动")
	return nil
}

// Stop 停止长轮询管理器
func (m *LongPollingManager) Stop() error {
	m.cancel()
	return m.listener.Close()
}

// Wait 等待配置变更
// configKeys: 配置键列表（格式: "namespaceID:configKey"）
// versions: 当前版本号映射
// 返回: 变更结果或超时
func (m *LongPollingManager) Wait(configKeys []string, versions map[string]string) (*WaitResult, error) {
	// 先检查是否有配置已经变更
	changed, changedKeys := m.checkVersions(configKeys, versions)
	if changed {
		// 获取最新版本号
		latestVersions, err := m.getLatestVersions(configKeys)
		if err != nil {
			return nil, err
		}
		return &WaitResult{
			Changed:    true,
			ConfigKeys: changedKeys,
			Versions:   latestVersions,
		}, nil
	}

	// 创建等待请求
	req := &WaitRequest{
		ConfigKeys: configKeys,
		Versions:   versions,
		ResultChan: make(chan *WaitResult, 1),
	}

	// 注册等待请求
	m.registerWaitRequest(req)
	defer m.unregisterWaitRequest(req)

	// 等待结果或超时
	select {
	case result := <-req.ResultChan:
		return result, nil
	case <-time.After(m.timeout):
		// 超时，返回未变更
		return &WaitResult{
			Changed:    false,
			ConfigKeys: []string{},
			Versions:   versions,
		}, nil
	case <-m.ctx.Done():
		return nil, fmt.Errorf("长轮询管理器已关闭")
	}
}

// PublishChange 发布配置变更
func (m *LongPollingManager) PublishChange(ctx context.Context, event *listener.ConfigChangeEvent) error {
	return m.listener.Publish(ctx, event)
}

// handleEvents 处理配置变更事件
func (m *LongPollingManager) handleEvents(eventChan <-chan *listener.ConfigChangeEvent) {
	for {
		select {
		case <-m.ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			// 处理配置变更事件
			m.notifyWaiters(event)
		}
	}
}

// notifyWaiters 通知等待的客户端
func (m *LongPollingManager) notifyWaiters(event *listener.ConfigChangeEvent) {
	configKey := fmt.Sprintf("%d:%s", event.NamespaceID, event.ConfigKey)

	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取等待该配置的请求列表
	requests, exists := m.waitRequests[configKey]
	if !exists || len(requests) == 0 {
		return
	}

	hlog.Infof("配置变更通知: %s, 等待数: %d", configKey, len(requests))

	// 通知所有等待的请求
	for _, req := range requests {
		// 检查该请求是否关注这个配置
		if m.containsConfigKey(req.ConfigKeys, configKey) {
			// 获取最新版本号
			latestVersions, err := m.getLatestVersions(req.ConfigKeys)
			if err != nil {
				hlog.Errorf("获取最新版本失败: %v", err)
				continue
			}

			// 发送结果（非阻塞）
			select {
			case req.ResultChan <- &WaitResult{
				Changed:    true,
				ConfigKeys: []string{configKey},
				Versions:   latestVersions,
			}:
			default:
			}
		}
	}

	// 清空该配置的等待列表
	delete(m.waitRequests, configKey)
}

// registerWaitRequest 注册等待请求
func (m *LongPollingManager) registerWaitRequest(req *WaitRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, configKey := range req.ConfigKeys {
		m.waitRequests[configKey] = append(m.waitRequests[configKey], req)
	}
}

// unregisterWaitRequest 注销等待请求
func (m *LongPollingManager) unregisterWaitRequest(req *WaitRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, configKey := range req.ConfigKeys {
		requests := m.waitRequests[configKey]
		for i, r := range requests {
			if r == req {
				// 删除该请求
				m.waitRequests[configKey] = append(requests[:i], requests[i+1:]...)
				break
			}
		}
		// 如果该配置的等待列表为空，删除key
		if len(m.waitRequests[configKey]) == 0 {
			delete(m.waitRequests, configKey)
		}
	}
}

// checkVersions 检查版本是否有变更
func (m *LongPollingManager) checkVersions(configKeys []string, clientVersions map[string]string) (bool, []string) {
	var changedKeys []string
	for _, configKey := range configKeys {
		clientVersion := clientVersions[configKey]
		serverVersion, err := m.versionGetter(configKey)
		if err != nil {
			hlog.Errorf("获取配置版本失败: %s, error: %v", configKey, err)
			continue
		}
		if clientVersion != serverVersion {
			changedKeys = append(changedKeys, configKey)
		}
	}
	return len(changedKeys) > 0, changedKeys
}

// getLatestVersions 获取最新版本号
func (m *LongPollingManager) getLatestVersions(configKeys []string) (map[string]string, error) {
	versions := make(map[string]string)
	for _, configKey := range configKeys {
		version, err := m.versionGetter(configKey)
		if err != nil {
			return nil, fmt.Errorf("获取配置版本失败: %s, error: %w", configKey, err)
		}
		versions[configKey] = version
	}
	return versions, nil
}

// containsConfigKey 检查配置键是否在列表中
func (m *LongPollingManager) containsConfigKey(keys []string, target string) bool {
	for _, key := range keys {
		if key == target {
			return true
		}
	}
	return false
}

// ComputeConfigVersion 计算配置版本（使用MD5）
func ComputeConfigVersion(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}
