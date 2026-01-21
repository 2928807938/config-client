package converter

import (
	domainEntity "config-client/config/domain/entity"
	infraEntity "config-client/config/infrastructure/entity"
)

// NamespaceConverter 命名空间转换器，负责领域实体和持久化对象之间的转换
type NamespaceConverter struct{}

// NewNamespaceConverter 创建命名空间转换器实例
func NewNamespaceConverter() *NamespaceConverter {
	return &NamespaceConverter{}
}

// 将持久化对象转换为领域实体（PO -> DO）
func (c *NamespaceConverter) ToDO(po *infraEntity.NamespacePO) *domainEntity.Namespace {
	if po == nil {
		return nil
	}

	namespace := &domainEntity.Namespace{
		Name:        po.Name,
		DisplayName: po.DisplayName,
		Description: po.Description,
		IsActive:    po.IsActive,
		Metadata:    po.Metadata,
	}

	// 设置 BaseEntity 字段
	namespace.ID = po.ID
	namespace.CreatedBy = po.CreatedBy
	namespace.UpdatedBy = po.UpdatedBy
	namespace.CreatedAt = po.CreatedAt
	namespace.UpdatedAt = po.UpdatedAt
	namespace.DeletedAt = po.DeletedAt

	return namespace
}

// ToPO 将领域实体转换为持久化对象（DO -> PO）
func (c *NamespaceConverter) ToPO(do *domainEntity.Namespace) *infraEntity.NamespacePO {
	if do == nil {
		return nil
	}

	po := &infraEntity.NamespacePO{
		// BaseEntity 字段
		ID:        do.ID,
		CreatedBy: do.CreatedBy,
		UpdatedBy: do.UpdatedBy,
		CreatedAt: do.CreatedAt,
		UpdatedAt: do.UpdatedAt,
		DeletedAt: do.DeletedAt,

		// 业务字段
		Name:        do.Name,
		DisplayName: do.DisplayName,
		Description: do.Description,
		IsActive:    do.IsActive,
		Metadata:    do.Metadata,
	}

	// 同步软删除状态：如果 DeletedAt 有效，设置 IsDeleted = true
	if do.DeletedAt.Valid {
		po.IsDeleted = true
	}

	return po
}

// ToDOList 批量转换为领域实体列表
func (c *NamespaceConverter) ToDOList(pos []*infraEntity.NamespacePO) []*domainEntity.Namespace {
	if len(pos) == 0 {
		return []*domainEntity.Namespace{}
	}

	dos := make([]*domainEntity.Namespace, len(pos))
	for i, po := range pos {
		dos[i] = c.ToDO(po)
	}
	return dos
}

// ToPOList 批量转换为持久化对象列表
func (c *NamespaceConverter) ToPOList(dos []*domainEntity.Namespace) []*infraEntity.NamespacePO {
	if len(dos) == 0 {
		return []*infraEntity.NamespacePO{}
	}

	pos := make([]*infraEntity.NamespacePO, len(dos))
	for i, do := range dos {
		pos[i] = c.ToPO(do)
	}
	return pos
}
