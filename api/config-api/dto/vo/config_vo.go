package vo

import (
	"time"
)

// ConfigVO 配置视图对象（响应DTO）
type ConfigVO struct {
	ID                   int       `json:"id"`                               // 配置ID
	NamespaceID          int       `json:"namespace_id"`                     // 命名空间ID
	Key                  string    `json:"key"`                              // 配置键
	Value                string    `json:"value"`                            // 配置值
	GroupName            string    `json:"group_name"`                       // 配置分组
	ValueType            string    `json:"value_type"`                       // 值类型
	Environment          string    `json:"environment"`                      // 环境
	Version              int       `json:"version"`                          // 版本号
	IsReleased           bool      `json:"is_released"`                      // 是否已发布
	IsActive             bool      `json:"is_active"`                        // 是否激活
	Description          string    `json:"description,omitempty"`            // 配置描述
	Metadata             string    `json:"metadata,omitempty"`               // 扩展元数据
	ContentHash          string    `json:"content_hash,omitempty"`           // 内容哈希
	ContentHashAlgorithm string    `json:"content_hash_algorithm,omitempty"` // 哈希算法
	CreatedBy            string    `json:"created_by"`                       // 创建人
	UpdatedBy            string    `json:"updated_by"`                       // 更新人
	CreatedAt            time.Time `json:"created_at"`                       // 创建时间
	UpdatedAt            time.Time `json:"updated_at"`                       // 更新时间
}

// ConfigListVO 配置列表视图对象（分页响应）
type ConfigListVO struct {
	Total      int64       `json:"total"`       // 总数
	Page       int         `json:"page"`        // 当前页码
	Size       int         `json:"size"`        // 每页数量
	TotalPages int         `json:"total_pages"` // 总页数
	Items      []*ConfigVO `json:"items"`       // 配置列表
}
