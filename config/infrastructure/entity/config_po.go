package entity

import (
	"time"

	"gorm.io/gorm"
)

// ConfigPO 配置持久化对象，与数据库表 t_configs 对应
type ConfigPO struct {
	// 主键
	ID int `gorm:"primaryKey;autoIncrement" json:"id"`

	// 关联信息
	NamespaceID int `gorm:"column:namespace_id;not null;index" json:"namespace_id"`

	// 配置标识
	Key       string `gorm:"column:key;type:varchar(500);not null" json:"key"`
	GroupName string `gorm:"column:group_name;type:varchar(255);default:'default'" json:"group_name"`

	// 配置值
	Value     string `gorm:"column:value;type:text" json:"value"`
	ValueType string `gorm:"column:value_type;type:varchar(50);default:'string'" json:"value_type"`

	// 配置哈希
	ContentHash          string `gorm:"column:content_hash;type:varchar(32)" json:"content_hash"`
	ContentHashAlgorithm string `gorm:"column:content_hash_algorithm;type:varchar(20);default:'md5'" json:"content_hash_algorithm"`

	// 环境隔离
	Environment string `gorm:"column:environment;type:varchar(50);default:'default'" json:"environment"`

	// 版本控制
	Version    int  `gorm:"column:version;default:1" json:"version"`
	IsReleased bool `gorm:"column:is_released;default:false" json:"is_released"`

	// 状态管理
	IsActive  bool `gorm:"column:is_active;default:true" json:"is_active"`
	IsDeleted bool `gorm:"column:is_deleted;default:false" json:"is_deleted"`

	// 审计字段
	CreatedBy string         `gorm:"column:created_by;type:varchar(100);default:'system'" json:"created_by"`
	UpdatedBy string         `gorm:"column:updated_by;type:varchar(100);default:'system'" json:"updated_by"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`

	// 扩展字段
	Description string `gorm:"column:description;type:text" json:"description"`
	Metadata    string `gorm:"column:metadata;type:jsonb;default:'{}'" json:"metadata"`
}

// TableName 指定表名
func (ConfigPO) TableName() string {
	return "t_configs"
}

// GetID 获取主键ID
func (c *ConfigPO) GetID() int {
	return c.ID
}
