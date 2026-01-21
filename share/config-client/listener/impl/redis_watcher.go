package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"config-client/share/config-client/listener"

	"github.com/redis/go-redis/v9"
)

const (
	// ConfigChangeChannel Redis Pub/Sub 通道名称
	ConfigChangeChannel = "config:change"
)

// RedisWatcher Redis直连配置监听器
// 通过订阅Redis Pub/Sub频道实时接收配置变更事件
type RedisWatcher struct {
	client    *redis.Client                              // Redis客户端
	mu        sync.RWMutex                               // 读写锁
	watchKeys map[string]*listener.WatchKey              // key -> WatchKey (key格式: "namespaceID:configKey")
	callbacks map[string][]listener.ConfigChangeCallback // key -> callbacks (支持多个回调)
	running   bool                                       // 是否正在运行
	ctx       context.Context                            // 上下文
	cancel    context.CancelFunc                         // 取消函数
	wg        sync.WaitGroup                             // 等待组
	pubsub    *redis.PubSub                              // Pub/Sub实例
}

// RedisConfigEvent Redis中的配置变更事件
type RedisConfigEvent struct {
	NamespaceID int    `json:"namespace_id"`
	ConfigKey   string `json:"config_key"`
	ConfigID    int    `json:"config_id"`
	Action      string `json:"action"`
}

// NewRedisWatcher 创建Redis监听器
func NewRedisWatcher(client *redis.Client) *RedisWatcher {
	return &RedisWatcher{
		client:    client,
		watchKeys: make(map[string]*listener.WatchKey),
		callbacks: make(map[string][]listener.ConfigChangeCallback),
		running:   false,
	}
}

// Start 启动监听器
func (w *RedisWatcher) Start(ctx context.Context) error {
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

	// 订阅Redis频道
	w.pubsub = w.client.Subscribe(ctx, ConfigChangeChannel)

	// 等待订阅确认
	_, err := w.pubsub.Receive(ctx)
	if err != nil {
		cancel()
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
		return fmt.Errorf("订阅Redis频道失败: %w", err)
	}

	// 启动事件监听循环
	w.wg.Add(1)
	go w.receiveLoop()

	return nil
}

// Stop 停止监听器
func (w *RedisWatcher) Stop() error {
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

	if w.pubsub != nil {
		if err := w.pubsub.Close(); err != nil {
			return fmt.Errorf("关闭Redis Pub/Sub失败: %w", err)
		}
	}

	w.wg.Wait()
	return nil
}

// Watch 添加监听配置
func (w *RedisWatcher) Watch(keys []*listener.WatchKey, callback listener.ConfigChangeCallback) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, key := range keys {
		k := w.formatKey(key.NamespaceID, key.Key)
		w.watchKeys[k] = key
		w.callbacks[k] = append(w.callbacks[k], callback)
	}

	return nil
}

// Unwatch 取消监听配置
func (w *RedisWatcher) Unwatch(keys []*listener.WatchKey) error {
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
func (w *RedisWatcher) UnwatchAll() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.watchKeys = make(map[string]*listener.WatchKey)
	w.callbacks = make(map[string][]listener.ConfigChangeCallback)
	return nil
}

// IsRunning 是否正在运行
func (w *RedisWatcher) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// receiveLoop 接收Redis消息循环
func (w *RedisWatcher) receiveLoop() {
	defer w.wg.Done()

	ch := w.pubsub.Channel()
	for {
		select {
		case <-w.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			w.handleMessage(msg)
		}
	}
}

// handleMessage 处理Redis消息
func (w *RedisWatcher) handleMessage(msg *redis.Message) {
	// 解析事件
	var event RedisConfigEvent
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		return
	}

	key := w.formatKey(event.NamespaceID, event.ConfigKey)

	w.mu.RLock()
	callbacks, exists := w.callbacks[key]
	watchKey, hasKey := w.watchKeys[key]
	w.mu.RUnlock()

	if !exists || len(callbacks) == 0 {
		return
	}

	// 构建变更事件
	changeEvent := &listener.ConfigChangeEvent{
		NamespaceID: event.NamespaceID,
		ConfigKey:   event.ConfigKey,
		ConfigID:    event.ConfigID,
		Action:      listener.ConfigEventType(event.Action),
		Timestamp:   time.Now(),
	}

	if hasKey && watchKey.Namespace != "" {
		changeEvent.Namespace = watchKey.Namespace
	}

	// 异步调用所有回调
	for _, callback := range callbacks {
		go callback(changeEvent)
	}
}

// formatKey 格式化配置键
func (w *RedisWatcher) formatKey(namespaceID int, configKey string) string {
	return fmt.Sprintf("%d:%s", namespaceID, configKey)
}

// ParseKey 解析配置键
// 格式: "namespaceID:configKey" -> (namespaceID, configKey)
func ParseKey(key string) (int, string, error) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("无效的配置键格式: %s", key)
	}

	var namespaceID int
	_, err := fmt.Sscanf(parts[0], "%d", &namespaceID)
	if err != nil {
		return 0, "", fmt.Errorf("无效的命名空间ID: %s", parts[0])
	}

	return namespaceID, parts[1], nil
}
