package vo

import "time"

// SubscriptionVO 订阅视图对象
type SubscriptionVO struct {
	ID                 int        `json:"id"`                   // 订阅ID
	NamespaceID        int        `json:"namespace_id"`         // 命名空间ID
	ClientID           string     `json:"client_id"`            // 客户端ID
	ClientIP           string     `json:"client_ip"`            // 客户端IP
	ClientHostname     string     `json:"client_hostname"`      // 客户端主机名
	Environment        string     `json:"environment"`          // 环境
	LastVersion        int        `json:"last_version"`         // 客户端当前版本号
	ConfigSnapshotHash string     `json:"config_snapshot_hash"` // 配置快照哈希
	IsActive           bool       `json:"is_active"`            // 是否激活
	LastHeartbeatAt    *time.Time `json:"last_heartbeat_at"`    // 最后心跳时间
	HeartbeatCount     int        `json:"heartbeat_count"`      // 心跳次数
	PollCount          int        `json:"poll_count"`           // 长轮询次数
	ChangeCount        int        `json:"change_count"`         // 变更次数
	SubscribedAt       time.Time  `json:"subscribed_at"`        // 订阅时间
	UnsubscribedAt     *time.Time `json:"unsubscribed_at"`      // 取消订阅时间
	CreatedAt          time.Time  `json:"created_at"`           // 创建时间
	UpdatedAt          time.Time  `json:"updated_at"`           // 更新时间
}

// SubscriptionListVO 订阅列表视图对象
type SubscriptionListVO struct {
	Total         int64             `json:"total"`         // 总数
	Page          int               `json:"page"`          // 当前页
	PageSize      int               `json:"page_size"`     // 每页数量
	Subscriptions []*SubscriptionVO `json:"subscriptions"` // 订阅列表
}

// SubscriptionStatisticsVO 订阅统计信息
type SubscriptionStatisticsVO struct {
	Total                   int64 `json:"total"`                     // 订阅总数
	Active                  int64 `json:"active"`                    // 激活订阅数
	Inactive                int64 `json:"inactive"`                  // 未激活订阅数
	Expired                 int64 `json:"expired"`                   // 过期订阅数（激活但心跳超时）
	ActiveInMemory          int   `json:"active_in_memory"`          // 内存活跃订阅数
	HeartbeatTimeoutSeconds int   `json:"heartbeat_timeout_seconds"` // 心跳超时阈值（秒）
}
