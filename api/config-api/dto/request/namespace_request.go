package request

// CreateNamespaceRequest 创建命名空间请求
type CreateNamespaceRequest struct {
	Name        string `json:"name" binding:"required"`         // 命名空间名称（必填）
	DisplayName string `json:"display_name" binding:"required"` // 显示名称（必填）
	Description string `json:"description"`                     // 描述信息
	Metadata    string `json:"metadata"`                        // 扩展元数据（JSON格式）
}

// UpdateNamespaceRequest 更新命名空间请求
type UpdateNamespaceRequest struct {
	ID          int    `json:"id" binding:"required,min=1"`     // 命名空间ID
	DisplayName string `json:"display_name" binding:"required"` // 显示名称（必填）
	Description string `json:"description"`                     // 描述信息
	Metadata    string `json:"metadata"`                        // 扩展元数据（JSON格式）
}

// QueryNamespaceRequest 查询命名空间请求
type QueryNamespaceRequest struct {
	Name     string `json:"name" form:"name"`           // 命名空间名称（模糊查询）
	IsActive *bool  `json:"is_active" form:"is_active"` // 是否激活
	Page     int    `json:"page" form:"page"`           // 页码
	PageSize int    `json:"page_size" form:"page_size"` // 每页数量
}

// GetNamespaceByIDRequest 根据ID获取命名空间请求
type GetNamespaceByIDRequest struct {
	ID int `json:"id" binding:"required,min=1"` // 命名空间ID
}

// DeleteNamespaceRequest 删除命名空间请求
type DeleteNamespaceRequest struct {
	ID int `json:"id" binding:"required,min=1"` // 命名空间ID
}

// ActivateNamespaceRequest 激活命名空间请求
type ActivateNamespaceRequest struct {
	ID int `json:"id" binding:"required,min=1"` // 命名空间ID
}

// DeactivateNamespaceRequest 停用命名空间请求
type DeactivateNamespaceRequest struct {
	ID int `json:"id" binding:"required,min=1"` // 命名空间ID
}
