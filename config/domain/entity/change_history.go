package entity

import (
	"time"
)

// Operation 操作类型
type Operation string

const (
	OperationCreate   Operation = "CREATE"   // 创建
	OperationUpdate   Operation = "UPDATE"   // 更新
	OperationDelete   Operation = "DELETE"   // 删除
	OperationRollback Operation = "ROLLBACK" // 回滚
)

// ChangeHistory 配置变更历史领域实体
// 记录配置的所有变更操作，用于审计和回滚
type ChangeHistory struct {
	ID           int       `json:"id"`
	ConfigID     int       `json:"config_id"`     // 变更的配置ID
	NamespaceID  int       `json:"namespace_id"`  // 所属命名空间ID
	ConfigKey    string    `json:"config_key"`    // 配置键
	Environment  string    `json:"environment"`   // 环境
	Operation    Operation `json:"operation"`     // 操作类型
	OldValue     string    `json:"old_value"`     // 变更前的值
	NewValue     string    `json:"new_value"`     // 变更后的值
	OldVersion   int       `json:"old_version"`   // 变更前版本号
	NewVersion   int       `json:"new_version"`   // 变更后版本号
	Operator     string    `json:"operator"`      // 操作人
	OperatorIP   string    `json:"operator_ip"`   // 操作人IP
	ChangeReason string    `json:"change_reason"` // 变更原因
	CreatedAt    time.Time `json:"created_at"`    // 变更时间
	Metadata     string    `json:"metadata"`      // 扩展元数据（JSON）
}

// ==================== 领域行为方法 ====================

// CanRollback 是否可以回滚到该版本
// CREATE 操作无法回滚（没有旧值）
// DELETE 操作无法回滚（需要重新创建）
func (h *ChangeHistory) CanRollback() bool {
	return h.Operation == OperationUpdate || h.Operation == OperationRollback
}

// IsRollbackOperation 是否是回滚操作
func (h *ChangeHistory) IsRollbackOperation() bool {
	return h.Operation == OperationRollback
}

// GetSnapshot 获取该版本的配置快照
// 返回该版本的配置值和版本号
func (h *ChangeHistory) GetSnapshot() (value string, version int) {
	if h.Operation == OperationDelete {
		return "", h.OldVersion
	}
	return h.NewValue, h.NewVersion
}

// HasValueChanged 值是否发生变化
func (h *ChangeHistory) HasValueChanged() bool {
	return h.OldValue != h.NewValue
}

// GetChangeSummary 获取变更摘要
func (h *ChangeHistory) GetChangeSummary() string {
	switch h.Operation {
	case OperationCreate:
		return "创建配置"
	case OperationUpdate:
		if h.HasValueChanged() {
			return "更新配置值"
		}
		return "更新配置元数据"
	case OperationDelete:
		return "删除配置"
	case OperationRollback:
		return "回滚配置"
	default:
		return "未知操作"
	}
}

// ChangeRecord 变更记录参数
// 用于在配置变更时传递必要信息
type ChangeRecord struct {
	ConfigID     int
	NamespaceID  int
	ConfigKey    string
	Environment  string
	Operation    Operation
	OldValue     string
	NewValue     string
	OldVersion   int
	NewVersion   int
	Operator     string
	OperatorIP   string
	ChangeReason string
	Metadata     string
}

// ToEntity 转换为领域实体
func (r *ChangeRecord) ToEntity() *ChangeHistory {
	return &ChangeHistory{
		ConfigID:     r.ConfigID,
		NamespaceID:  r.NamespaceID,
		ConfigKey:    r.ConfigKey,
		Environment:  r.Environment,
		Operation:    r.Operation,
		OldValue:     r.OldValue,
		NewValue:     r.NewValue,
		OldVersion:   r.OldVersion,
		NewVersion:   r.NewVersion,
		Operator:     r.Operator,
		OperatorIP:   r.OperatorIP,
		ChangeReason: r.ChangeReason,
		Metadata:     r.Metadata,
	}
}

// RollbackRecord 回滚记录参数
type RollbackRecord struct {
	ConfigID        int
	TargetHistoryID int // 要回滚到的历史记录ID
	Operator        string
	OperatorIP      string
	ChangeReason    string
}
