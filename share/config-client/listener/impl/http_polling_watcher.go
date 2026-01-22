package impl

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"config-client/share/config-client/listener"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// HTTPPollingWatcher HTTP长轮询配置监听器
type HTTPPollingWatcher struct {
	serverURL      string                                   // 配置中心服务地址
	httpClient     *http.Client                             // HTTP客户端
	timeout        time.Duration                            // 长轮询超时时间
	clientID       string                                   // 客户端唯一标识
	clientIP       string                                   // 客户端IP地址
	clientHostname string                                   // 客户端主机名
	mu             sync.RWMutex                             // 读写锁
	watchKeys      map[string]*listener.WatchKey            // key -> WatchKey (key格式: "namespaceID:configKey")
	callbacks      map[string]listener.ConfigChangeCallback // key -> callback
	running        bool                                     // 是否正在运行
	ctx            context.Context                          // 上下文
	cancel         context.CancelFunc                       // 取消函数
	wg             sync.WaitGroup                           // 等待组
}

// HTTPPollingRequest 长轮询请求
type HTTPPollingRequest struct {
	ClientID       string             `json:"client_id"`       // 客户端唯一标识
	ClientIP       string             `json:"client_ip"`       // 客户端IP地址
	ClientHostname string             `json:"client_hostname"` // 客户端主机名
	ConfigKeys     []ConfigKeyVersion `json:"config_keys"`     // 配置键列表
}

// ConfigKeyVersion 配置键及其版本
type ConfigKeyVersion struct {
	NamespaceID int    `json:"namespace_id"` // 命名空间ID
	Environment string `json:"environment"`  // 环境
	ConfigKey   string `json:"config_key"`   // 配置键
	Version     string `json:"version"`      // 配置版本
}

// HTTPPollingResponse 长轮询响应
type HTTPPollingResponse struct {
	Changed    bool                 `json:"changed"`
	ConfigKeys []string             `json:"config_keys"`
	Configs    []ConfigChangeDetail `json:"configs"`
}

// ConfigChangeDetail 配置变更详情
type ConfigChangeDetail struct {
	NamespaceID int    `json:"namespace_id"`
	ConfigKey   string `json:"config_key"`
	Version     string `json:"version"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type"`
}

// NewHTTPPollingWatcher 创建HTTP长轮询监听器
func NewHTTPPollingWatcher(serverURL string, timeout time.Duration) *HTTPPollingWatcher {
	// 生成唯一的客户端ID
	clientID := generateClientID()

	// 获取主机名
	hostname, _ := os.Hostname()

	return &HTTPPollingWatcher{
		serverURL:      serverURL,
		clientID:       clientID,
		clientIP:       "", // IP地址由服务端自动获取
		clientHostname: hostname,
		httpClient: &http.Client{
			Timeout: timeout + 10*time.Second, // 比长轮询超时时间多10秒
		},
		timeout:   timeout,
		watchKeys: make(map[string]*listener.WatchKey),
		callbacks: make(map[string]listener.ConfigChangeCallback),
		running:   false,
	}
}

// generateClientID 生成唯一的客户端ID
func generateClientID() string {
	// 使用主机名+随机字符串
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	// 生成8字节随机数
	b := make([]byte, 8)
	rand.Read(b)
	randomStr := hex.EncodeToString(b)

	// 格式: hostname-timestamp-random
	return fmt.Sprintf("%s-%d-%s", hostname, time.Now().Unix(), randomStr)
}

// Start 启动监听器
func (w *HTTPPollingWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("监听器已在运行中")
	}
	w.running = true
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	w.ctx = ctx
	w.cancel = cancel

	// 启动长轮询循环
	w.wg.Add(1)
	go w.pollingLoop()

	return nil
}

// Stop 停止监听器
func (w *HTTPPollingWatcher) Stop() error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = false
	w.mu.Unlock()

	if w.cancel != nil {
		w.cancel()
	}

	w.wg.Wait()
	return nil
}

// Watch 添加监听配置
func (w *HTTPPollingWatcher) Watch(keys []*listener.WatchKey, callback listener.ConfigChangeCallback) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, key := range keys {
		k := w.formatKey(key.NamespaceID, key.Key)
		// 如果已经存在监听,保留原有的版本号(避免重复触发)
		if existingKey, exists := w.watchKeys[k]; exists && key.Version == "" {
			key.Version = existingKey.Version
		}
		w.watchKeys[k] = key
		w.callbacks[k] = callback
	}

	return nil
}

// Unwatch 取消监听配置
func (w *HTTPPollingWatcher) Unwatch(keys []*listener.WatchKey) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, key := range keys {
		k := w.formatKey(key.NamespaceID, key.Key)
		delete(w.watchKeys, k)
		delete(w.callbacks, k)
	}

	return nil
}

// UnwatchAll 取消所有监听
func (w *HTTPPollingWatcher) UnwatchAll() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.watchKeys = make(map[string]*listener.WatchKey)
	w.callbacks = make(map[string]listener.ConfigChangeCallback)
	return nil
}

// IsRunning 是否正在运行
func (w *HTTPPollingWatcher) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// pollingLoop 长轮询循环
func (w *HTTPPollingWatcher) pollingLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		// 如果没有监听的配置，等待一段时间
		w.mu.RLock()
		hasKeys := len(w.watchKeys) > 0
		w.mu.RUnlock()

		if !hasKeys {
			time.Sleep(time.Second)
			continue
		}

		// 执行长轮询请求
		if err := w.doPolling(); err != nil {
			hlog.Errorf("长轮询请求失败: %v", err)
			// 出错后等待一段时间再重试
			select {
			case <-w.ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		// 成功后短暂间隔再发起下一次请求，避免服务器压力过大
		select {
		case <-w.ctx.Done():
			return
		case <-time.After(200 * time.Millisecond):
		}
	}
}

// doPolling 执行一次长轮询请求
func (w *HTTPPollingWatcher) doPolling() error {
	// 构建请求
	w.mu.RLock()
	keys := make([]*listener.WatchKey, 0, len(w.watchKeys))
	for _, key := range w.watchKeys {
		keys = append(keys, key)
	}
	w.mu.RUnlock()

	if len(keys) == 0 {
		return nil
	}

	// 构建请求体
	configKeys := make([]ConfigKeyVersion, len(keys))
	for i, key := range keys {
		configKeys[i] = ConfigKeyVersion{
			NamespaceID: key.NamespaceID,
			Environment: "default", // 使用默认环境
			ConfigKey:   key.Key,
			Version:     key.Version,
		}
		hlog.Infof("[客户端发送] 配置键: %d:%s, 版本号: %s", key.NamespaceID, key.Key, key.Version)
	}

	reqBody := HTTPPollingRequest{
		ClientID:       w.clientID,
		ClientIP:       w.clientIP,
		ClientHostname: w.clientHostname,
		ConfigKeys:     configKeys,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s/api/v1/configs/watch", w.serverURL)
	req, err := http.NewRequestWithContext(w.ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 解析响应 - 先读取完整响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 尝试解析为标准响应格式（包含外层 Response 结构）
	var standardResp struct {
		Code    int                 `json:"code"`
		Message string              `json:"message"`
		Data    HTTPPollingResponse `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &standardResp); err != nil {
		// 如果失败，尝试直接解析为 HTTPPollingResponse
		var pollingResp HTTPPollingResponse
		if err2 := json.Unmarshal(bodyBytes, &pollingResp); err2 != nil {
			return fmt.Errorf("解析响应失败: %w (原始错误: %v)", err2, err)
		}
		// 直接解析成功
		if pollingResp.Changed {
			w.handleConfigChanges(&pollingResp)
		}
		return nil
	}

	// 标准格式解析成功
	if standardResp.Data.Changed {
		w.handleConfigChanges(&standardResp.Data)
	}

	return nil
}

// handleConfigChanges 处理配置变更
func (w *HTTPPollingWatcher) handleConfigChanges(resp *HTTPPollingResponse) {
	w.mu.RLock()
	// 先读取需要的数据
	eventsToSend := make([]struct {
		event    *listener.ConfigChangeEvent
		callback listener.ConfigChangeCallback
	}, 0, len(resp.Configs))

	for _, config := range resp.Configs {
		key := w.formatKey(config.NamespaceID, config.ConfigKey)
		callback, exists := w.callbacks[key]
		if !exists {
			continue
		}

		// 获取原始 WatchKey 信息
		watchKey, exists := w.watchKeys[key]
		if !exists {
			continue
		}

		// 构建事件
		event := &listener.ConfigChangeEvent{
			NamespaceID: config.NamespaceID,
			ConfigKey:   config.ConfigKey,
			Action:      listener.EventTypeUpdate,
			Value:       config.Value,
			Version:     config.Version,
			Timestamp:   time.Now(),
		}

		if watchKey.Namespace != "" {
			event.Namespace = watchKey.Namespace
		}

		eventsToSend = append(eventsToSend, struct {
			event    *listener.ConfigChangeEvent
			callback listener.ConfigChangeCallback
		}{event: event, callback: callback})
	}
	w.mu.RUnlock()

	// 更新版本号（需要写锁）
	if len(eventsToSend) > 0 {
		w.mu.Lock()
		for _, config := range resp.Configs {
			key := w.formatKey(config.NamespaceID, config.ConfigKey)
			if watchKey, exists := w.watchKeys[key]; exists {
				watchKey.Version = config.Version
			}
		}
		w.mu.Unlock()
	}

	// 异步调用回调，避免阻塞
	for _, item := range eventsToSend {
		go item.callback(item.event)
	}
}

// formatKey 格式化配置键
func (w *HTTPPollingWatcher) formatKey(namespaceID int, configKey string) string {
	return fmt.Sprintf("%d:%s", namespaceID, configKey)
}
