package vo

import (
	"time"
)

// ChangeHistoryVO 变更历史视图对象
type ChangeHistoryVO struct {
	ID           int       `json:"id"`                      // 历史记录ID
	ConfigID     int       `json:"config_id"`               // 配置ID
	NamespaceID  int       `json:"namespace_id"`            // 命名空间ID
	ConfigKey    string    `json:"config_key"`              // 配置键
	Environment  string    `json:"environment"`             // 环境
	Operation    string    `json:"operation"`               // 操作类型：CREATE/UPDATE/DELETE/ROLLBACK
	OldValue     string    `json:"old_value,omitempty"`     // 变更前的值
	NewValue     string    `json:"new_value,omitempty"`     // 变更后的值
	OldVersion   int       `json:"old_version"`             // 变更前版本号
	NewVersion   int       `json:"new_version"`             // 变更后版本号
	Operator     string    `json:"operator"`                // 操作人
	OperatorIP   string    `json:"operator_ip,omitempty"`   // 操作人IP
	ChangeReason string    `json:"change_reason,omitempty"` // 变更原因
	CreatedAt    time.Time `json:"created_at"`              // 变更时间
	Metadata     string    `json:"metadata,omitempty"`      // 扩展元数据

	// 扩展字段
	CanRollback   bool   `json:"can_rollback"`   // 是否可以回滚到该版本
	ChangeSummary string `json:"change_summary"` // 变更摘要
	ValueChanged  bool   `json:"value_changed"`  // 值是否发生变化
}

// ChangeHistoryListVO 变更历史列表视图对象（分页响应）
type ChangeHistoryListVO struct {
	Total      int64              `json:"total"`       // 总数
	Page       int                `json:"page"`        // 当前页码
	Size       int                `json:"size"`        // 每页数量
	TotalPages int                `json:"total_pages"` // 总页数
	Items      []*ChangeHistoryVO `json:"items"`       // 变更历史列表
}

// VersionCompareVO 版本对比视图对象
type VersionCompareVO struct {
	FromHistoryID int       `json:"from_history_id"` // 源版本历史ID
	ToHistoryID   int       `json:"to_history_id"`   // 目标版本历史ID
	FromVersion   int       `json:"from_version"`    // 源版本号
	ToVersion     int       `json:"to_version"`      // 目标版本号
	FromValue     string    `json:"from_value"`      // 源版本值
	ToValue       string    `json:"to_value"`        // 目标版本值
	ValueChanged  bool      `json:"value_changed"`   // 值是否变化
	FromOperation string    `json:"from_operation"`  // 源操作类型
	ToOperation   string    `json:"to_operation"`    // 目标操作类型
	FromChangedAt time.Time `json:"from_changed_at"` // 源变更时间
	ToChangedAt   time.Time `json:"to_changed_at"`   // 目标变更时间
}

// ChangeStatisticsVO 变更统计视图对象
type ChangeStatisticsVO struct {
	TotalChanges  int64 `json:"total_changes"`  // 总变更次数
	CreateCount   int64 `json:"create_count"`   // 创建次数
	UpdateCount   int64 `json:"update_count"`   // 更新次数
	DeleteCount   int64 `json:"delete_count"`   // 删除次数
	RollbackCount int64 `json:"rollback_count"` // 回滚次数
}
