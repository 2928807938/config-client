package entity

import (
	baseGorm "config-client/share/repository/gorm"
)

// Config 配置领域实体（聚合根）
// 纯粹的领域模型，不包含持久化相关的标签
type Config struct {
	baseGorm.BaseEntity         // 组合通用审计字段
	NamespaceID          int    `json:"namespace_id"`           // 命名空间ID
	Key                  string `json:"key"`                    // 配置键
	Value                string `json:"value"`                  // 配置值
	GroupName            string `json:"group_name"`             // 配置分组
	ValueType            string `json:"value_type"`             // 值类型
	Environment          string `json:"environment"`            // 环境
	IsReleased           bool   `json:"is_released"`            // 是否已发布
	IsActive             bool   `json:"is_active"`              // 是否激活
	Description          string `json:"description"`            // 描述
	Metadata             string `json:"metadata"`               // 元数据
	ContentHash          string `json:"content_hash"`           // 内容哈希
	ContentHashAlgorithm string `json:"content_hash_algorithm"` // 哈希算法
}

// ==================== 领域行为方法 ====================

// Release 发布配置
func (c *Config) Release() {
	c.IsReleased = true
	c.UpdatedAt = c.UpdatedAt // 触发 GORM 的 autoUpdateTime
}

// Unrelease 取消发布
func (c *Config) Unrelease() {
	c.IsReleased = false
	c.UpdatedAt = c.UpdatedAt
}

// Activate 激活配置
func (c *Config) Activate() {
	c.IsActive = true
	c.UpdatedAt = c.UpdatedAt
}

// Deactivate 停用配置
func (c *Config) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = c.UpdatedAt
}

// UpdateValue 更新配置值
func (c *Config) UpdateValue(value string, contentHash string) {
	c.Value = value
	c.ContentHash = contentHash
	c.IncrementVersion() // 继承自 BaseEntity
}

// ==================== 查询方法 ====================

// IsActiveStatus 判断是否激活
func (c *Config) IsActiveStatus() bool {
	return c.IsActive
}

// IsReleasedStatus 判断是否已发布
func (c *Config) IsReleasedStatus() bool {
	return c.IsReleased
}

// GetFullKey 获取完整的配置键（包含命名空间和环境）
func (c *Config) GetFullKey() string {
	return c.Key + ":" + c.Environment
}
