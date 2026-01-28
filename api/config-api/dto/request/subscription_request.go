package request

// QuerySubscriptionRequest 订阅查询请求 DTO
type QuerySubscriptionRequest struct {
	NamespaceID *int    `json:"namespace_id" form:"namespace_id"`         // 命名空间ID
	Environment *string `json:"environment" form:"environment"`           // 环境
	ClientID    *string `json:"client_id" form:"client_id"`               // 客户端ID（模糊查询）
	IsActive    *bool   `json:"is_active" form:"is_active"`               // 是否激活
	Page        int     `json:"page" form:"page" binding:"min=1"`         // 页码，默认1
	Size        int     `json:"size" form:"size" binding:"min=1,max=200"` // 每页数量，默认20
	OrderBy     string  `json:"order_by" form:"order_by"`                 // 排序字段
}

// SetDefaults 设置默认值
func (q *QuerySubscriptionRequest) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Size == 0 {
		q.Size = 20
	}
}

// DeactivateSubscriptionRequest 停用订阅请求 DTO
type DeactivateSubscriptionRequest struct {
	ID int `json:"id" binding:"required,min=1"` // 订阅ID
}
