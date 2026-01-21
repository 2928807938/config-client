package entity

import "time"

// Subscription 订阅实体
// 记录客户端订阅信息，用于长轮询和配置变更推送
type Subscription struct {
	// 主键
	ID int

	// 关联信息
	NamespaceID int // 订阅的命名空间ID

	// 客户端信息
	ClientID       string // 客户端唯一标识
	ClientIP       string // 客户端IP地址
	ClientHostname string // 客户端主机名

	// 订阅范围
	Environment string // 订阅的环境

	// 版本跟踪
	LastVersion int // 客户端当前版本号

	// 配置快照哈希（用于批量比对多个配置的变化）
	ConfigSnapshotHash string // 客户端当前配置快照的MD5

	// 状态管理
	IsActive bool // 订阅是否激活

	// 心跳检测
	LastHeartbeatAt *time.Time // 最后心跳时间
	HeartbeatCount  int        // 心跳次数

	// 统计信息
	PollCount   int // 长轮询次数
	ChangeCount int // 接收到的变更次数

	// 审计字段
	SubscribedAt   time.Time  // 订阅时间
	UnsubscribedAt *time.Time // 取消订阅时间
	CreatedAt      time.Time  // 创建时间
	UpdatedAt      time.Time  // 更新时间
}

// UpdateHeartbeat 更新心跳
func (s *Subscription) UpdateHeartbeat() {
	now := time.Now()
	s.LastHeartbeatAt = &now
	s.HeartbeatCount++
	s.UpdatedAt = now
}

// IncrementPollCount 增加轮询计数
func (s *Subscription) IncrementPollCount() {
	s.PollCount++
	s.UpdatedAt = time.Now()
}

// IncrementChangeCount 增加变更计数
func (s *Subscription) IncrementChangeCount() {
	s.ChangeCount++
	s.UpdatedAt = time.Now()
}

// Deactivate 停用订阅
func (s *Subscription) Deactivate() {
	s.IsActive = false
	now := time.Now()
	s.UnsubscribedAt = &now
	s.UpdatedAt = now
}

// Activate 激活订阅
func (s *Subscription) Activate() {
	s.IsActive = true
	s.UnsubscribedAt = nil
	s.UpdatedAt = time.Now()
}

// IsExpired 判断订阅是否过期
// timeout: 超时时长
// 返回: true-已过期, false-未过期
func (s *Subscription) IsExpired(timeout time.Duration) bool {
	if s.LastHeartbeatAt == nil {
		// 如果没有心跳记录，检查创建时间
		return time.Since(s.CreatedAt) > timeout
	}
	return time.Since(*s.LastHeartbeatAt) > timeout
}

// UpdateSnapshotHash 更新配置快照哈希
func (s *Subscription) UpdateSnapshotHash(hash string) {
	s.ConfigSnapshotHash = hash
	s.UpdatedAt = time.Now()
}
