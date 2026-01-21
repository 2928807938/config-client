package entity

import (
	"time"

	"gorm.io/gorm"
)

// ReleasePO 发布版本持久化对象，与数据库表 t_release_versions 对应
type ReleasePO struct {
	// 主键
	ID int `gorm:"primaryKey;autoIncrement" json:"id"`

	// 关联信息
	NamespaceID int    `gorm:"column:namespace_id;not null;index" json:"namespace_id"`
	Environment string `gorm:"column:environment;type:varchar(50);not null" json:"environment"`

	// 版本信息
	Version     int    `gorm:"column:version;not null" json:"version"`
	VersionName string `gorm:"column:version_name;type:varchar(255);not null" json:"version_name"`

	// 发布范围
	ConfigSnapshot string `gorm:"column:config_snapshot;type:jsonb;not null" json:"config_snapshot"`
	ConfigCount    int    `gorm:"column:config_count;default:0" json:"config_count"`

	// 发布状态
	Status      string `gorm:"column:status;type:varchar(20);default:'testing'" json:"status"`
	ReleaseType string `gorm:"column:release_type;type:varchar(20);default:'full'" json:"release_type"`

	// 灰度发布
	CanaryRule       string `gorm:"column:canary_rule;type:jsonb" json:"canary_rule"`
	CanaryPercentage int    `gorm:"column:canary_percentage;default:0" json:"canary_percentage"`

	// 发布信息
	ReleasedBy string     `gorm:"column:released_by;type:varchar(100)" json:"released_by"`
	ReleasedAt *time.Time `gorm:"column:released_at" json:"released_at"`

	// 回滚信息
	RollbackFromVersion int        `gorm:"column:rollback_from_version" json:"rollback_from_version"`
	RollbackBy          string     `gorm:"column:rollback_by;type:varchar(100)" json:"rollback_by"`
	RollbackAt          *time.Time `gorm:"column:rollback_at" json:"rollback_at"`
	RollbackReason      string     `gorm:"column:rollback_reason;type:text" json:"rollback_reason"`

	// 审计字段
	CreatedBy string         `gorm:"column:created_by;type:varchar(100);default:'system'" json:"created_by"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (ReleasePO) TableName() string {
	return "t_release_versions"
}
