package entity

import (
	"time"

	"gorm.io/gorm"
)

// NamespacePO 命名空间持久化对象，与数据库表 t_namespaces 对应
type NamespacePO struct {
	// 主键
	ID int `gorm:"primaryKey;autoIncrement" json:"id"`

	// 基本信息
	Name        string `gorm:"column:name;type:varchar(255);uniqueIndex;not null" json:"name"`
	DisplayName string `gorm:"column:display_name;type:varchar(255)" json:"display_name"`
	Description string `gorm:"column:description;type:text" json:"description"`

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
	Metadata string `gorm:"column:metadata;type:jsonb;default:'{}'" json:"metadata"`
}

// TableName 指定表名
func (NamespacePO) TableName() string {
	return "t_namespaces"
}

// GetID 获取主键ID
func (n *NamespacePO) GetID() int {
	return n.ID
}
