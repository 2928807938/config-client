package configclient

import (
	"fmt"

	"config-client/share/config-client/listener"
	"config-client/share/config-client/listener/impl"

	"github.com/redis/go-redis/v9"
)

// WatcherConfig 监听器配置
type WatcherConfig struct {
	Type      WatcherType    // 监听器类型
	ServerURL string         // HTTP服务地址
	RedisOpt  *redis.Options // Redis配置选项
}

// NewWatcher 创建监听器（工厂方法）
func NewWatcher(cfg *WatcherConfig) (listener.Watcher, error) {
	switch cfg.Type {
	case WatcherTypeHTTP:
		if cfg.ServerURL == "" {
			return nil, fmt.Errorf("HTTP模式需要配置ServerURL")
		}
		return impl.NewHTTPPollingWatcher(cfg.ServerURL, 0), nil

	case WatcherTypeRedis:
		if cfg.RedisOpt == nil {
			return nil, fmt.Errorf("Redis模式需要配置RedisOpt")
		}
		client := redis.NewClient(cfg.RedisOpt)
		return impl.NewRedisWatcher(client), nil

	default:
		return nil, fmt.Errorf("不支持的监听器类型: %s", cfg.Type)
	}
}
