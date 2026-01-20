package entity

import (
	baseGorm "config-client/share/repository/gorm"
)

// Namespace 命名空间领域实体（聚合根）
// 用于隔离不同应用的配置，例如：user-service、order-service、payment-app
// 纯粹的领域模型，不包含持久化相关的标签
type Namespace struct {
	baseGorm.BaseEntity        // 组合通用审计字段
	Name                string `json:"name"`         // 命名空间名称，唯一标识
	DisplayName         string `json:"display_name"` // 显示名称
	Description         string `json:"description"`  // 描述信息
	IsActive            bool   `json:"is_active"`    // 是否启用
	Metadata            string `json:"metadata"`     // 扩展元数据（JSON格式）
}

// ==================== 领域行为方法 ====================

// Activate 激活命名空间
func (n *Namespace) Activate() {
	n.IsActive = true
	n.UpdatedAt = n.UpdatedAt // 触发 GORM 的 autoUpdateTime
}

// Deactivate 停用命名空间
func (n *Namespace) Deactivate() {
	n.IsActive = false
	n.UpdatedAt = n.UpdatedAt
}

// UpdateInfo 更新命名空间基本信息
func (n *Namespace) UpdateInfo(displayName, description, metadata string) {
	n.DisplayName = displayName
	n.Description = description
	n.Metadata = metadata
	n.UpdatedAt = n.UpdatedAt
}

// ==================== 查询方法 ====================

// IsActiveStatus 判断是否激活
func (n *Namespace) IsActiveStatus() bool {
	return n.IsActive
}

// GetIdentifier 获取唯一标识符
func (n *Namespace) GetIdentifier() string {
	return n.Name
}
