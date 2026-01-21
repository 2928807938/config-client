package listener

import (
	"context"
	"time"
)

// ConfigEventType 配置变更事件类型
type ConfigEventType string

const (
	EventTypeCreate ConfigEventType = "create" // 创建
	EventTypeUpdate ConfigEventType = "update" // 更新
	EventTypeDelete ConfigEventType = "delete" // 删除
)

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	NamespaceID int             `json:"namespace_id"` // 命名空间ID
	Namespace   string          `json:"namespace"`    // 命名空间名称
	ConfigKey   string          `json:"config_key"`   // 配置键
	ConfigID    int             `json:"config_id"`    // 配置ID
	Action      ConfigEventType `json:"action"`       // 操作类型
	Value       string          `json:"value"`        // 配置值
	Version     string          `json:"version"`      // 版本号
	Timestamp   time.Time       `json:"timestamp"`    // 变更时间
}

// ConfigChangeCallback 配置变更回调函数
type ConfigChangeCallback func(event *ConfigChangeEvent)

// WatchKey 监听的配置键
type WatchKey struct {
	NamespaceID int    // 命名空间ID
	Namespace   string // 命名空间名称
	Key         string // 配置键
	Version     string // 当前版本号
}

// Watcher 配置监听器接口
type Watcher interface {
	// Start 启动监听器
	Start(ctx context.Context) error

	// Stop 停止监听器
	Stop() error

	// Watch 添加监听配置
	// keys: 要监听的配置键列表
	// callback: 配置变更时的回调函数
	Watch(keys []*WatchKey, callback ConfigChangeCallback) error

	// Unwatch 取消监听配置
	Unwatch(keys []*WatchKey) error

	// UnwatchAll 取消所有监听
	UnwatchAll() error

	// IsRunning 是否正在运行
	IsRunning() bool
}
