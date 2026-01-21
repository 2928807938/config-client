package configsdk

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// WatcherType 监听器类型
type WatcherType string

const (
	// WatcherTypeHTTP HTTP 长轮询模式
	WatcherTypeHTTP WatcherType = "http"
	// WatcherTypeRedis Redis 订阅模式
	WatcherTypeRedis WatcherType = "redis"
)

// Options SDK 配置选项
type Options struct {
	// ServerURL 配置中心服务地址（必填）
	ServerURL string

	// NamespaceID 命名空间 ID（默认: 1）
	NamespaceID int

	// Namespace 命名空间名称（默认: "default"）
	Namespace string

	// WatcherType 监听器类型（默认: HTTP）
	WatcherType WatcherType

	// RedisClient Redis 客户端（Redis 模式需要）
	RedisClient *redis.Client

	// PollingTimeout 长轮询超时时间（默认: 60s）
	PollingTimeout time.Duration

	// AutoStart 是否自动启动监听器（默认: true）
	AutoStart bool

	// EnableCache 是否启用配置缓存（默认: true）
	EnableCache bool

	// FetchOnInit 初始化时是否拉取所有配置（默认: true）
	FetchOnInit bool

	// Fallback 降级配置（网络故障时使用）
	Fallback map[string]string
}

// Option 配置选项函数
type Option func(*Options)

// DefaultOptions 默认配置
func DefaultOptions() *Options {
	return &Options{
		ServerURL:      "http://localhost:8080",
		NamespaceID:    1,
		Namespace:      "default",
		WatcherType:    WatcherTypeHTTP,
		PollingTimeout: 60 * time.Second,
		AutoStart:      true,
		EnableCache:    true,
		FetchOnInit:    true,
		Fallback:       make(map[string]string),
	}
}

// WithServerURL 设置服务器地址
func WithServerURL(url string) Option {
	return func(o *Options) {
		o.ServerURL = url
	}
}

// WithNamespace 设置命名空间
func WithNamespace(name string) Option {
	return func(o *Options) {
		o.Namespace = name
	}
}

// WithNamespaceID 设置命名空间 ID
func WithNamespaceID(id int) Option {
	return func(o *Options) {
		o.NamespaceID = id
	}
}

// WithHTTPWatcher 使用 HTTP 长轮询监听器
func WithHTTPWatcher(timeout time.Duration) Option {
	return func(o *Options) {
		o.WatcherType = WatcherTypeHTTP
		if timeout > 0 {
			o.PollingTimeout = timeout
		}
	}
}

// WithRedisWatcher 使用 Redis 订阅监听器
func WithRedisWatcher(client *redis.Client) Option {
	return func(o *Options) {
		o.WatcherType = WatcherTypeRedis
		o.RedisClient = client
	}
}

// WithAutoStart 设置是否自动启动
func WithAutoStart(autoStart bool) Option {
	return func(o *Options) {
		o.AutoStart = autoStart
	}
}

// WithCache 设置是否启用缓存
func WithCache(enable bool) Option {
	return func(o *Options) {
		o.EnableCache = enable
	}
}

// WithFetchOnInit 设置初始化时是否拉取配置
func WithFetchOnInit(fetch bool) Option {
	return func(o *Options) {
		o.FetchOnInit = fetch
	}
}

// WithFallback 设置降级配置
func WithFallback(fallback map[string]string) Option {
	return func(o *Options) {
		o.Fallback = fallback
	}
}

// WithRedisOptions 使用 Redis 选项创建监听器
func WithRedisOptions(opt *redis.Options) Option {
	return func(o *Options) {
		o.WatcherType = WatcherTypeRedis
		o.RedisClient = redis.NewClient(opt)
	}
}
