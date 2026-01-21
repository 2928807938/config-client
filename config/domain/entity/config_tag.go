package entity

import "time"

// ConfigTag 配置标签领域实体
// 用于为配置打标签，支持分组和筛选
type ConfigTag struct {
	ID        int       `json:"id"`         // 主键ID
	ConfigID  int       `json:"config_id"`  // 配置ID
	TagKey    string    `json:"tag_key"`    // 标签键
	TagValue  string    `json:"tag_value"`  // 标签值
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// ==================== 领域行为方法 ====================

// GetFullTag 获取完整的标签表示（格式: key:value）
func (t *ConfigTag) GetFullTag() string {
	return t.TagKey + ":" + t.TagValue
}

// IsSensitiveTag 判断是否为敏感标签
func (t *ConfigTag) IsSensitiveTag() bool {
	return t.TagKey == "sensitive" && t.TagValue == "true"
}

// IsImportanceTag 判断是否为重要性标签
func (t *ConfigTag) IsImportanceTag() bool {
	return t.TagKey == "importance"
}

// IsCategoryTag 判断是否为分类标签
func (t *ConfigTag) IsCategoryTag() bool {
	return t.TagKey == "category"
}

// ==================== 辅助类型 ====================

// TagInput 标签输入结构（用于创建标签）
type TagInput struct {
	TagKey   string `json:"tag_key"`   // 标签键
	TagValue string `json:"tag_value"` // 标签值
}

// Validate 验证标签输入是否有效
func (t *TagInput) Validate() error {
	if t.TagKey == "" {
		return &ValidationError{Field: "tag_key", Message: "标签键不能为空"}
	}
	if t.TagValue == "" {
		return &ValidationError{Field: "tag_value", Message: "标签值不能为空"}
	}
	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
