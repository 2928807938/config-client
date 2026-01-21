package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ChangeHistoryPO 配置变更历史持久化对象，与数据库表 t_change_history 对应
type ChangeHistoryPO struct {
	// 主键
	ID int `gorm:"primaryKey;autoIncrement" json:"id"`

	// 关联信息
	ConfigID    int `gorm:"column:config_id;not null;index:idx_config_id" json:"config_id"`
	NamespaceID int `gorm:"column:namespace_id;not null;index:idx_namespace_id" json:"namespace_id"`

	// 配置快照
	ConfigKey   string `gorm:"column:config_key;type:varchar(500);not null;index:idx_config_key" json:"config_key"`
	Environment string `gorm:"column:environment;type:varchar(50);default:'default'" json:"environment"`

	// 变更信息
	Operation string `gorm:"column:operation;type:varchar(20);not null;index:idx_operation" json:"operation"`
	OldValue  string `gorm:"column:old_value;type:text" json:"old_value"`
	NewValue  string `gorm:"column:new_value;type:text" json:"new_value"`

	// 版本信息
	OldVersion int `gorm:"column:old_version" json:"old_version"`
	NewVersion int `gorm:"column:new_version" json:"new_version"`

	// 操作人信息
	Operator   string `gorm:"column:operator;type:varchar(100);not null" json:"operator"`
	OperatorIP string `gorm:"column:operator_ip;type:varchar(50)" json:"operator_ip"`

	// 变更原因
	ChangeReason string `gorm:"column:change_reason;type:text" json:"change_reason"`

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_created_at" json:"created_at"`

	// 扩展字段
	Metadata JSONB `gorm:"column:metadata;type:jsonb;default:'{}'" json:"metadata"`
}

// TableName 指定表名
func (ChangeHistoryPO) TableName() string {
	return "t_change_history"
}

// GetID 获取主键ID
func (h *ChangeHistoryPO) GetID() int {
	return h.ID
}

// JSONB 自定义类型，用于处理 PostgreSQL 的 JSONB 类型
type JSONB map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}
