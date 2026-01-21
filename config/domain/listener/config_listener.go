package listener

import "context"

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	NamespaceID int64  `json:"namespace_id"` // 命名空间ID
	ConfigKey   string `json:"config_key"`   // 配置键
	ConfigID    int64  `json:"config_id"`    // 配置ID
	Action      string `json:"action"`       // 操作类型: create, update, delete
}

// ConfigListener 配置变更监听器接口
type ConfigListener interface {
	// Subscribe 订阅配置变更
	// ctx: 上下文
	// 返回: 配置变更事件通道
	Subscribe(ctx context.Context) (<-chan *ConfigChangeEvent, error)

	// Publish 发布配置变更事件
	// ctx: 上下文
	// event: 配置变更事件
	Publish(ctx context.Context, event *ConfigChangeEvent) error

	// Close 关闭监听器
	Close() error
}
