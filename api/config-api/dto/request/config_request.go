package request

// CreateConfigRequest 创建配置请求 DTO
type CreateConfigRequest struct {
	NamespaceID int    `json:"namespace_id" binding:"required,min=1"` // 命名空间ID
	Key         string `json:"key" binding:"required,max=500"`        // 配置键
	Value       string `json:"value" binding:"required"`              // 配置值
	GroupName   string `json:"group_name" binding:"max=255"`          // 配置分组，默认"default"
	ValueType   string `json:"value_type" binding:"max=50"`           // 值类型，默认"string"
	Environment string `json:"environment" binding:"max=50"`          // 环境，默认"default"
	Description string `json:"description"`                           // 配置描述
	Metadata    string `json:"metadata"`                              // 扩展元数据（JSON格式）
	CreatedBy   string `json:"created_by" binding:"max=100"`          // 创建人
}

// UpdateConfigRequest 更新配置请求 DTO
type UpdateConfigRequest struct {
	ID          int    `json:"id" binding:"required,min=1"`  // 配置ID
	Value       string `json:"value" binding:"required"`     // 配置值
	GroupName   string `json:"group_name" binding:"max=255"` // 配置分组
	ValueType   string `json:"value_type" binding:"max=50"`  // 值类型
	Description string `json:"description"`                  // 配置描述
	Metadata    string `json:"metadata"`                     // 扩展元数据
	IsActive    *bool  `json:"is_active"`                    // 是否激活（指针类型，允许null）
	IsReleased  *bool  `json:"is_released"`                  // 是否已发布（指针类型，允许null）
	UpdatedBy   string `json:"updated_by" binding:"max=100"` // 更新人
}

// QueryConfigRequest 查询配置请求 DTO（多条件查询）
type QueryConfigRequest struct {
	NamespaceID *int    `json:"namespace_id" form:"namespace_id"`         // 命名空间ID（指针类型，允许null）
	Key         *string `json:"key" form:"key"`                           // 配置键（支持模糊查询）
	GroupName   *string `json:"group_name" form:"group_name"`             // 配置分组
	Environment *string `json:"environment" form:"environment"`           // 环境
	IsActive    *bool   `json:"is_active" form:"is_active"`               // 是否激活
	IsReleased  *bool   `json:"is_released" form:"is_released"`           // 是否已发布
	ValueType   *string `json:"value_type" form:"value_type"`             // 值类型
	Page        int     `json:"page" form:"page" binding:"min=1"`         // 页码，默认1
	Size        int     `json:"size" form:"size" binding:"min=1,max=100"` // 每页数量，默认10，最大100
	OrderBy     string  `json:"order_by" form:"order_by"`                 // 排序字段，例如：created_at desc
}

// SetDefaults 设置默认值
func (q *QueryConfigRequest) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Size == 0 {
		q.Size = 10
	}
}

// GetConfigByIDRequest 根据ID获取配置请求 DTO
type GetConfigByIDRequest struct {
	ID int `json:"id" binding:"required,min=1"` // 配置ID
}

// DeleteConfigRequest 删除配置请求 DTO
type DeleteConfigRequest struct {
	ID int `json:"id" binding:"required,min=1"` // 配置ID
}

// ReleaseConfigRequest 发布配置请求 DTO
type ReleaseConfigRequest struct {
	IsReleased bool   `json:"is_released" binding:"required"` // 是否发布
	UpdatedBy  string `json:"updated_by" binding:"max=100"`   // 操作人
}
