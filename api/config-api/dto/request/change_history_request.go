package request

// QueryHistoryRequest 查询变更历史请求 DTO
type QueryHistoryRequest struct {
	ConfigID    *int    `json:"config_id" form:"config_id"`               // 配置ID
	NamespaceID *int    `json:"namespace_id" form:"namespace_id"`         // 命名空间ID
	ConfigKey   *string `json:"config_key" form:"config_key"`             // 配置键（支持模糊查询）
	Operation   *string `json:"operation" form:"operation"`               // 操作类型：CREATE/UPDATE/DELETE/ROLLBACK
	StartTime   *string `json:"start_time" form:"start_time"`             // 开始时间，格式：2006-01-02 15:04:05
	EndTime     *string `json:"end_time" form:"end_time"`                 // 结束时间
	Operator    *string `json:"operator" form:"operator"`                 // 操作人（支持模糊查询）
	Page        int     `json:"page" form:"page" binding:"min=1"`         // 页码，默认1
	Size        int     `json:"size" form:"size" binding:"min=1,max=100"` // 每页数量，默认20，最大100
}

// SetDefaults 设置默认值
func (q *QueryHistoryRequest) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Size == 0 {
		q.Size = 20
	}
}

// GetHistoryByIDRequest 根据ID查询变更历史请求 DTO
type GetHistoryByIDRequest struct {
	HistoryID int `json:"history_id" binding:"required,min=1"` // 历史记录ID
}

// RollbackRequest 回滚配置请求 DTO
type RollbackRequest struct {
	HistoryID    int    `json:"history_id" binding:"required,min=1"` // 要回滚到的历史记录ID
	ChangeReason string `json:"change_reason"`                       // 回滚原因
}

// CompareVersionsRequest 版本��比请求 DTO
type CompareVersionsRequest struct {
	HistoryID1 int `json:"history_id1" binding:"required,min=1"` // 源版本历史ID
	HistoryID2 int `json:"history_id2" binding:"required,min=1"` // 目标版本历史ID
}

// GetConfigHistoryRequest 获取配置变更历史请求 DTO
type GetConfigHistoryRequest struct {
	ConfigID int `json:"config_id" binding:"required,min=1"` // 配置ID
	Limit    int `json:"limit" form:"limit"`                 // 返回数量，默认50，最大100
}

// SetDefaults 设置默认值
func (g *GetConfigHistoryRequest) SetDefaults() {
	if g.Limit == 0 {
		g.Limit = 50
	}
	if g.Limit > 100 {
		g.Limit = 100
	}
}
