package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"config-client/config/domain/errors"
	"config-client/config/domain/listener"
	"config-client/config/domain/repository"

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

// LongPollingService 长轮询领域服务
type LongPollingService struct {
	listener     listener.ConfigListener     // 配置监听器
	configRepo   repository.ConfigRepository // 配置仓储
	waitRequests map[string][]*WaitRequest   // 配置键 -> 等待请求列表
	mu           sync.RWMutex                // 读写锁
	timeout      time.Duration               // 超时时间
	ctx          context.Context             // 上下文
	cancel       context.CancelFunc          // 取消函数
}

// NewLongPollingService 创建长轮询领域服务
func NewLongPollingService(
	listener listener.ConfigListener,
	configRepo repository.ConfigRepository,
	timeout time.Duration,
) *LongPollingService {
	ctx, cancel := context.WithCancel(context.Background())
	return &LongPollingService{
		listener:     listener,
		configRepo:   configRepo,
		waitRequests: make(map[string][]*WaitRequest),
		timeout:      timeout,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start 启动长轮询服务
func (s *LongPollingService) Start() error {
	// 订阅配置变更事件
	eventChan, err := s.listener.Subscribe(s.ctx)
	if err != nil {
		return errors.ErrLongPollingSubscribeFailed(err)
	}

	// 启动事件监听
	go s.handleEvents(eventChan)

	hlog.Info("长轮询领域服务已启动")
	return nil
}

// Stop 停止长轮询服务
func (s *LongPollingService) Stop() error {
	s.cancel()
	return s.listener.Close()
}

// Wait 等待配置变更
// configKeys: 配置键列表（格式: "namespaceID:configKey"）
// versions: 当前版本号映射
// 返回: 变更结果或超时
func (s *LongPollingService) Wait(configKeys []string, versions map[string]string) (*WaitResult, error) {
	// 先检查是否有配置已经变更
	changed, changedKeys := s.checkVersions(configKeys, versions)
	if changed {
		// 获取最新版本号
		latestVersions, err := s.getLatestVersions(configKeys)
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
	s.registerWaitRequest(req)
	defer s.unregisterWaitRequest(req)

	// 等待结果或超时
	select {
	case result := <-req.ResultChan:
		return result, nil
	case <-time.After(s.timeout):
		// 超时，返回未变更
		return &WaitResult{
			Changed:    false,
			ConfigKeys: []string{},
			Versions:   versions,
		}, nil
	case <-s.ctx.Done():
		return nil, errors.ErrLongPollingServiceStopped()
	}
}

// PublishChange 发布配置变更
func (s *LongPollingService) PublishChange(ctx context.Context, event *listener.ConfigChangeEvent) error {
	return s.listener.Publish(ctx, event)
}

// GetConfigVersion 获取配置版本号
func (s *LongPollingService) GetConfigVersion(namespaceID int, configKey string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := s.configRepo.FindByNamespaceAndKey(ctx, namespaceID, configKey, "")
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", errors.ErrConfigNotFound(configKey, "")
	}

	return ComputeVersion(config.Value), nil
}

// handleEvents 处理配置变更事件
func (s *LongPollingService) handleEvents(eventChan <-chan *listener.ConfigChangeEvent) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			// 处理配置变更事件
			s.notifyWaiters(event)
		}
	}
}

// notifyWaiters 通知等待的客户端
func (s *LongPollingService) notifyWaiters(event *listener.ConfigChangeEvent) {
	configKey := fmt.Sprintf("%d:%s", event.NamespaceID, event.ConfigKey)

	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取等待该配置的请求列表
	requests, exists := s.waitRequests[configKey]
	if !exists || len(requests) == 0 {
		return
	}

	hlog.Infof("配置变更通知: %s, 等待数: %d", configKey, len(requests))

	// 通知所有等待的请求
	for _, req := range requests {
		// 检查该请求是否关注这个配置
		if s.containsConfigKey(req.ConfigKeys, configKey) {
			// 获取最新版本号
			latestVersions, err := s.getLatestVersions(req.ConfigKeys)
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
	delete(s.waitRequests, configKey)
}

// registerWaitRequest 注册等待请求
func (s *LongPollingService) registerWaitRequest(req *WaitRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, configKey := range req.ConfigKeys {
		s.waitRequests[configKey] = append(s.waitRequests[configKey], req)
	}
}

// unregisterWaitRequest 注销等待请求
func (s *LongPollingService) unregisterWaitRequest(req *WaitRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, configKey := range req.ConfigKeys {
		requests := s.waitRequests[configKey]
		for i, r := range requests {
			if r == req {
				// 删除该请求
				s.waitRequests[configKey] = append(requests[:i], requests[i+1:]...)
				break
			}
		}
		// 如果该配置的等待列表为空，删除key
		if len(s.waitRequests[configKey]) == 0 {
			delete(s.waitRequests, configKey)
		}
	}
}

// checkVersions 检查版本是否有变更
func (s *LongPollingService) checkVersions(configKeys []string, clientVersions map[string]string) (bool, []string) {
	var changedKeys []string
	for _, configKey := range configKeys {
		clientVersion := clientVersions[configKey]
		serverVersion, err := s.getVersionFromConfigKey(configKey)
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
func (s *LongPollingService) getLatestVersions(configKeys []string) (map[string]string, error) {
	versions := make(map[string]string)
	for _, configKey := range configKeys {
		version, err := s.getVersionFromConfigKey(configKey)
		if err != nil {
			return nil, errors.ErrLongPollingGetVersionFailed(configKey, err)
		}
		versions[configKey] = version
	}
	return versions, nil
}

// getVersionFromConfigKey 从配置键获取版本号
func (s *LongPollingService) getVersionFromConfigKey(configKey string) (string, error) {
	// 解析configKey (格式: "namespaceID:key")
	var namespaceID int
	var key string
	_, err := fmt.Sscanf(configKey, "%d:%s", &namespaceID, &key)
	if err != nil {
		return "", errors.ErrLongPollingInvalidConfigKey(configKey)
	}

	return s.GetConfigVersion(namespaceID, key)
}

// containsConfigKey 检查配置键是否在列表中
func (s *LongPollingService) containsConfigKey(keys []string, target string) bool {
	for _, key := range keys {
		if key == target {
			return true
		}
	}
	return false
}

// ComputeVersion 计算配置版本（使用MD5）
func ComputeVersion(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}
