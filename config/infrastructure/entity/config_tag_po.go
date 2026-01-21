package entity

import "time"

// ConfigTagPO 配置标签持久化对象
// 对应数据库表 t_config_tags
type ConfigTagPO struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ConfigID  int       `gorm:"column:config_id;not null;index" json:"config_id"`
	TagKey    string    `gorm:"column:tag_key;type:varchar(100);not null;index" json:"tag_key"`
	TagValue  string    `gorm:"column:tag_value;type:varchar(255);not null;index" json:"tag_value"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (ConfigTagPO) TableName() string {
	return "t_config_tags"
}
