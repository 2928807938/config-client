package request

// CreateReleaseRequest 创建发布版本请求
type CreateReleaseRequest struct {
	NamespaceID int    `json:"namespace_id" binding:"required"`                               // 命名空间ID
	Environment string `json:"environment" binding:"required"`                                // 发布环境
	VersionName string `json:"version_name" binding:"required"`                               // 版本名称
	ReleaseType string `json:"release_type" binding:"required,oneof=full incremental canary"` // 发布类型
	CreatedBy   string `json:"created_by" binding:"required"`                                 // 创建人
}

// PublishFullRequest 全量发布请求
type PublishFullRequest struct {
	ReleaseID   int    `json:"release_id" binding:"required"`   // 发布版本ID
	PublishedBy string `json:"published_by" binding:"required"` // 发布人
}

// PublishCanaryRequest 灰度发布请求
type PublishCanaryRequest struct {
	ReleaseID        int      `json:"release_id" binding:"required"`             // 发布版本ID
	ClientIDs        []string `json:"client_ids"`                                // 客户端ID白名单
	IPRanges         []string `json:"ip_ranges"`                                 // IP段白名单
	CanaryPercentage int      `json:"canary_percentage" binding:"min=0,max=100"` // 灰度百分比
	PublishedBy      string   `json:"published_by" binding:"required"`           // 发布人
}

// ReleaseRollbackRequest 版本回滚请求
type ReleaseRollbackRequest struct {
	CurrentReleaseID int    `json:"current_release_id" binding:"required"` // 当前版本ID
	TargetReleaseID  int    `json:"target_release_id" binding:"required"`  // 目标版本ID
	RollbackBy       string `json:"rollback_by" binding:"required"`        // 回滚人
	Reason           string `json:"reason" binding:"required"`             // 回滚原因
}

// QueryReleaseRequest 查询发布版本请求
type QueryReleaseRequest struct {
	NamespaceID *int    `json:"namespace_id" form:"namespace_id"` // 命名空间ID
	Environment *string `json:"environment" form:"environment"`   // 环境
	Status      *string `json:"status" form:"status"`             // 状态
	ReleaseType *string `json:"release_type" form:"release_type"` // 发布类型
	VersionName *string `json:"version_name" form:"version_name"` // 版本名称
	Page        int     `json:"page" form:"page"`                 // 页码
	Size        int     `json:"size" form:"size"`                 // 每页数量
	OrderBy     string  `json:"order_by" form:"order_by"`         // 排序字段
}

// SetDefaults 设置默认值
func (r *QueryReleaseRequest) SetDefaults() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Size <= 0 {
		r.Size = 20
	}
	if r.Size > 100 {
		r.Size = 100
	}
	if r.OrderBy == "" {
		r.OrderBy = "version DESC"
	}
}

// CompareReleasesRequest 对比版本请求
type CompareReleasesRequest struct {
	FromReleaseID int `json:"from_release_id" binding:"required"` // 源版本ID
	ToReleaseID   int `json:"to_release_id" binding:"required"`   // 目标版本ID
}
