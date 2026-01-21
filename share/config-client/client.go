package configclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"config-client/share/config-client/listener"
	"config-client/share/config-client/listener/impl"

	"github.com/redis/go-redis/v9"
)

// WatcherType 监听器类型
type WatcherType string

const (
	WatcherTypeHTTP  WatcherType = "http"  // HTTP长轮询
	WatcherTypeRedis WatcherType = "redis" // Redis直连
)

// Config 客户端配置
type Config struct {
	// ServerURL 配置中心服务地址（HTTP模式使用）
	ServerURL string

	// NamespaceID 默认命名空间ID
	NamespaceID int

	// Namespace 默认命名空间名称
	Namespace string

	// WatcherType 监听器类型（http/redis）
	WatcherType WatcherType

	// RedisClient Redis客户端（Redis模式使用）
	RedisClient *redis.Client

	// PollingTimeout 长轮询超时时间（HTTP模式使用）
	PollingTimeout time.Duration

	// AutoStart 是否自动启动监听器
	AutoStart bool
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		ServerURL:      "http://localhost:8080",
		NamespaceID:    1,
		Namespace:      "default",
		WatcherType:    WatcherTypeHTTP,
		PollingTimeout: 60 * time.Second,
		AutoStart:      true,
	}
}

// Client 配置中心客户端
type Client struct {
	cfg     *Config
	watcher listener.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
	once    sync.Once
	started bool
}

// NewClient 创建配置中心客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	client := &Client{
		cfg: cfg,
	}

	// 创建监听器
	watcher, err := createWatcher(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建监听器失败: %w", err)
	}
	client.watcher = watcher

	// 自动启动
	if cfg.AutoStart {
		ctx := context.Background()
		if err := client.Start(ctx); err != nil {
			return nil, fmt.Errorf("启动客户端失败: %w", err)
		}
	}

	return client, nil
}

// createWatcher 创建监听器
func createWatcher(cfg *Config) (listener.Watcher, error) {
	switch cfg.WatcherType {
	case WatcherTypeHTTP:
		if cfg.ServerURL == "" {
			return nil, fmt.Errorf("HTTP模式需要配置ServerURL")
		}
		return impl.NewHTTPPollingWatcher(cfg.ServerURL, cfg.PollingTimeout), nil

	case WatcherTypeRedis:
		if cfg.RedisClient == nil {
			return nil, fmt.Errorf("Redis模式需要配置RedisClient")
		}
		return impl.NewRedisWatcher(cfg.RedisClient), nil

	default:
		return nil, fmt.Errorf("不支持的监听器类型: %s", cfg.WatcherType)
	}
}

// Start 启动客户端
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("客户端已启动")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	if err := c.watcher.Start(c.ctx); err != nil {
		return fmt.Errorf("启动监听器失败: %w", err)
	}

	c.started = true
	return nil
}

// Stop 停止客户端
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	if c.cancel != nil {
		c.cancel()
	}

	if err := c.watcher.Stop(); err != nil {
		return fmt.Errorf("停止监听器失败: %w", err)
	}

	c.started = false
	return nil
}

// Watch 监听配置变更
// keys: 配置键列表
// callback: 配置变更回调
func (c *Client) Watch(keys []string, callback listener.ConfigChangeCallback) error {
	watchKeys := make([]*listener.WatchKey, len(keys))
	for i, key := range keys {
		watchKeys[i] = &listener.WatchKey{
			NamespaceID: c.cfg.NamespaceID,
			Namespace:   c.cfg.Namespace,
			Key:         key,
			Version:     "", // 初次监听，版本为空
		}
	}

	return c.watcher.Watch(watchKeys, callback)
}

// WatchWithVersion 带版本号监听配置变更
func (c *Client) WatchWithVersion(keys map[string]string, callback listener.ConfigChangeCallback) error {
	watchKeys := make([]*listener.WatchKey, 0, len(keys))
	for key, version := range keys {
		watchKeys = append(watchKeys, &listener.WatchKey{
			NamespaceID: c.cfg.NamespaceID,
			Namespace:   c.cfg.Namespace,
			Key:         key,
			Version:     version,
		})
	}

	return c.watcher.Watch(watchKeys, callback)
}

// WatchNamespace 监听指定命名空间的配置变更
func (c *Client) WatchNamespace(namespaceID int, namespace string, keys []string, callback listener.ConfigChangeCallback) error {
	watchKeys := make([]*listener.WatchKey, len(keys))
	for i, key := range keys {
		watchKeys[i] = &listener.WatchKey{
			NamespaceID: namespaceID,
			Namespace:   namespace,
			Key:         key,
			Version:     "",
		}
	}

	return c.watcher.Watch(watchKeys, callback)
}

// Unwatch 取消监听配置
func (c *Client) Unwatch(keys []string) error {
	watchKeys := make([]*listener.WatchKey, len(keys))
	for i, key := range keys {
		watchKeys[i] = &listener.WatchKey{
			NamespaceID: c.cfg.NamespaceID,
			Key:         key,
		}
	}

	return c.watcher.Unwatch(watchKeys)
}

// UnwatchAll 取消所有监听
func (c *Client) UnwatchAll() error {
	return c.watcher.UnwatchAll()
}

// IsRunning 是否正在运行
func (c *Client) IsRunning() bool {
	return c.watcher.IsRunning()
}

// GetConfig 获取客户端配置
func (c *Client) GetConfig() *Config {
	return c.cfg
}
