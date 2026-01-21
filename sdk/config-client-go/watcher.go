package configsdk

import (
	"config-client/share/config-client/listener"
	"config-client/share/config-client/listener/impl"
	"context"
	"fmt"
)

// Watcher 监听器接口（SDK 简化版）
type Watcher interface {
	Start(ctx context.Context) error
	Stop() error
	Watch(keys []string, callback func(key, value, version string)) error
	Unwatch(keys []string) error
	IsRunning() bool
}

// httpWatcher HTTP 长轮询监听器包装
type httpWatcher struct {
	underlying  *impl.HTTPPollingWatcher // 底层 HTTP 长轮询监听器
	namespaceID int
	namespace   string
}

// redisWatcher Redis 订阅监听器包装
type redisWatcher struct {
	underlying  *impl.RedisWatcher // 底层 Redis 监听器
	namespaceID int
	namespace   string
}

// createWatcherFromOptions 根据选项创建监听器
func createWatcherFromOptions(opts *Options) (Watcher, error) {
	switch opts.WatcherType {
	case WatcherTypeHTTP:
		if opts.ServerURL == "" {
			return nil, fmt.Errorf("HTTP 模式需要配置 ServerURL")
		}
		// 创建底层 HTTP 长轮询监听器
		underlying := impl.NewHTTPPollingWatcher(opts.ServerURL, opts.PollingTimeout)
		return &httpWatcher{
			underlying:  underlying,
			namespaceID: opts.NamespaceID,
			namespace:   opts.Namespace,
		}, nil

	case WatcherTypeRedis:
		if opts.RedisClient == nil {
			return nil, fmt.Errorf("Redis 模式需要配置 RedisClient")
		}
		// 创建底层 Redis 监听器
		underlying := impl.NewRedisWatcher(opts.RedisClient)
		return &redisWatcher{
			underlying:  underlying,
			namespaceID: opts.NamespaceID,
			namespace:   opts.Namespace,
		}, nil

	default:
		return nil, fmt.Errorf("不支持的监听器类型: %s", opts.WatcherType)
	}
}

// Start 启动 HTTP 监听器
func (w *httpWatcher) Start(ctx context.Context) error {
	return w.underlying.Start(ctx)
}

// Stop 停止 HTTP 监听器
func (w *httpWatcher) Stop() error {
	return w.underlying.Stop()
}

// Watch 监听配置
func (w *httpWatcher) Watch(keys []string, callback func(key, value, version string)) error {
	// 将简化的 key 列表转换为底层的 WatchKey 列表
	watchKeys := make([]*listener.WatchKey, len(keys))
	for i, key := range keys {
		watchKeys[i] = &listener.WatchKey{
			NamespaceID: w.namespaceID,
			Namespace:   w.namespace,
			Key:         key,
			Version:     "", // 初始版本为空
		}
	}

	// 将简化的回调转换为底层的回调
	underlyingCallback := func(event *listener.ConfigChangeEvent) {
		callback(event.ConfigKey, event.Value, event.Version)
	}

	return w.underlying.Watch(watchKeys, underlyingCallback)
}

// Unwatch 取消监听
func (w *httpWatcher) Unwatch(keys []string) error {
	watchKeys := make([]*listener.WatchKey, len(keys))
	for i, key := range keys {
		watchKeys[i] = &listener.WatchKey{
			NamespaceID: w.namespaceID,
			Namespace:   w.namespace,
			Key:         key,
		}
	}
	return w.underlying.Unwatch(watchKeys)
}

// IsRunning 是否运行中
func (w *httpWatcher) IsRunning() bool {
	return w.underlying.IsRunning()
}

// Start 启动 Redis 监听器
func (w *redisWatcher) Start(ctx context.Context) error {
	return w.underlying.Start(ctx)
}

// Stop 停止 Redis 监听器
func (w *redisWatcher) Stop() error {
	return w.underlying.Stop()
}

// Watch 监听配置
func (w *redisWatcher) Watch(keys []string, callback func(key, value, version string)) error {
	// 将简化的 key 列表转换为底层的 WatchKey 列表
	watchKeys := make([]*listener.WatchKey, len(keys))
	for i, key := range keys {
		watchKeys[i] = &listener.WatchKey{
			NamespaceID: w.namespaceID,
			Namespace:   w.namespace,
			Key:         key,
			Version:     "", // 初始版本为空
		}
	}

	// 将简化的回调转换为底层的回调
	underlyingCallback := func(event *listener.ConfigChangeEvent) {
		callback(event.ConfigKey, event.Value, event.Version)
	}

	return w.underlying.Watch(watchKeys, underlyingCallback)
}

// Unwatch 取消监听
func (w *redisWatcher) Unwatch(keys []string) error {
	watchKeys := make([]*listener.WatchKey, len(keys))
	for i, key := range keys {
		watchKeys[i] = &listener.WatchKey{
			NamespaceID: w.namespaceID,
			Namespace:   w.namespace,
			Key:         key,
		}
	}
	return w.underlying.Unwatch(watchKeys)
}

// IsRunning 是否运行中
func (w *redisWatcher) IsRunning() bool {
	return w.underlying.IsRunning()
}
