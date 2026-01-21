package entity

import "time"

// SubscriptionPO 订阅持久化对象
type SubscriptionPO struct {
	// 主键
	ID int `gorm:"primaryKey;column:id"`

	// 关联信息
	NamespaceID int `gorm:"column:namespace_id;not null;index:idx_t_subscriptions_namespace_id"`

	// 客户端信息
	ClientID       string `gorm:"column:client_id;type:varchar(255);not null;index:idx_t_subscriptions_client_id"`
	ClientIP       string `gorm:"column:client_ip;type:varchar(50)"`
	ClientHostname string `gorm:"column:client_hostname;type:varchar(255)"`

	// 订阅范围
	Environment string `gorm:"column:environment;type:varchar(50);default:'default';index:idx_t_subscriptions_env"`

	// 版本跟踪
	LastVersion int `gorm:"column:last_version;default:0"`

	// 配置快照哈希
	ConfigSnapshotHash string `gorm:"column:config_snapshot_hash;type:varchar(32)"`

	// 状态管理
	IsActive bool `gorm:"column:is_active;default:true;index:idx_t_subscriptions_active"`

	// 心跳检测
	LastHeartbeatAt *time.Time `gorm:"column:last_heartbeat_at;index:idx_t_subscriptions_heartbeat"`
	HeartbeatCount  int        `gorm:"column:heartbeat_count;default:0"`

	// 统计信息
	PollCount   int `gorm:"column:poll_count;default:0"`
	ChangeCount int `gorm:"column:change_count;default:0"`

	// 审计字段
	SubscribedAt   time.Time  `gorm:"column:subscribed_at;default:CURRENT_TIMESTAMP"`
	UnsubscribedAt *time.Time `gorm:"column:unsubscribed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

// TableName 指定表名
func (SubscriptionPO) TableName() string {
	return "t_subscriptions"
}
