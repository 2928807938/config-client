package request

// TagInput 标签输入结构
type TagInput struct {
	TagKey   string `json:"tag_key" binding:"required,max=100"`   // 标签键
	TagValue string `json:"tag_value" binding:"required,max=255"` // 标签值
}

// AddTagsRequest 添加标签请求
type AddTagsRequest struct {
	ConfigID int        `json:"config_id" binding:"required,min=1"` // 配置ID
	Tags     []TagInput `json:"tags" binding:"required,min=1,dive"` // 标签列表
}

// RemoveTagsRequest 删除标签请求
type RemoveTagsRequest struct {
	ConfigID int      `json:"config_id" binding:"required,min=1"` // 配置ID
	TagKeys  []string `json:"tag_keys" binding:"required,min=1"`  // 要删除的标签键列表
}

// QueryByTagsRequest 根据标签查询配置请求
type QueryByTagsRequest struct {
	Tags []TagInput `json:"tags" form:"tags" binding:"required,min=1,dive"` // 标签列表（AND查询）
}
