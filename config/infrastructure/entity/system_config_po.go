package entity

import (
	"time"
)

// SystemConfigPO 系统配置持久化对象，与数据库表 t_system_configs 对应
type SystemConfigPO struct {
	// 主键
	ID int `gorm:"primaryKey;autoIncrement" json:"id"`

	// 配置信息
	ConfigKey   string `gorm:"column:config_key;type:varchar(255);uniqueIndex;not null" json:"config_key"`
	ConfigValue string `gorm:"column:config_value;type:text" json:"config_value"`
	Description string `gorm:"column:description;type:text" json:"description"`

	// 状态管理
	IsActive bool `gorm:"column:is_active;default:true" json:"is_active"`

	// 审计字段
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (SystemConfigPO) TableName() string {
	return "t_system_configs"
}

// GetID 获取主键ID
func (s *SystemConfigPO) GetID() int {
	return s.ID
}
