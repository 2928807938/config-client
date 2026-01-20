package vo

import "time"

// NamespaceVO 命名空间视图对象
type NamespaceVO struct {
	ID          int       `json:"id"`           // ID
	Name        string    `json:"name"`         // 命名空间名称
	DisplayName string    `json:"display_name"` // 显示名称
	Description string    `json:"description"`  // 描述信息
	IsActive    bool      `json:"is_active"`    // 是否激活
	Metadata    string    `json:"metadata"`     // 扩展元数据
	CreatedBy   string    `json:"created_by"`   // 创建人
	UpdatedBy   string    `json:"updated_by"`   // 更新人
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
}

// NamespaceListVO 命名空间列表视图对象
type NamespaceListVO struct {
	Total      int64          `json:"total"`      // 总数
	Page       int            `json:"page"`       // 当前页
	PageSize   int            `json:"page_size"`  // 每页数量
	Namespaces []*NamespaceVO `json:"namespaces"` // 命名空间列表
}
