package vo

import "time"

// ConfigTagVO 配置标签视图对象
type ConfigTagVO struct {
	ID        int       `json:"id"`         // 标签ID
	ConfigID  int       `json:"config_id"`  // 配置ID
	TagKey    string    `json:"tag_key"`    // 标签键
	TagValue  string    `json:"tag_value"`  // 标签值
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// ConfigTagListVO 标签列表视图对象
type ConfigTagListVO struct {
	ConfigID int            `json:"config_id"` // 配置ID
	Tags     []*ConfigTagVO `json:"tags"`      // 标签列表
}
