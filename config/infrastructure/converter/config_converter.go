package converter

import (
	domainEntity "config-client/config/domain/entity"
	infraEntity "config-client/config/infrastructure/entity"
)

// ConfigConverter 配置转换器，负责领域实体和持久化对象之间的转换
type ConfigConverter struct{}

// NewConfigConverter 创建配置转换器实例
func NewConfigConverter() *ConfigConverter {
	return &ConfigConverter{}
}

// 将持久化对象转换为领域实体（PO -> DO）
func (c *ConfigConverter) ToDO(po *infraEntity.ConfigPO) *domainEntity.Config {
	if po == nil {
		return nil
	}

	config := &domainEntity.Config{
		NamespaceID:          po.NamespaceID,
		Key:                  po.Key,
		Value:                po.Value,
		GroupName:            po.GroupName,
		ValueType:            po.ValueType,
		Environment:          po.Environment,
		IsReleased:           po.IsReleased,
		IsActive:             po.IsActive,
		Description:          po.Description,
		Metadata:             po.Metadata,
		ContentHash:          po.ContentHash,
		ContentHashAlgorithm: po.ContentHashAlgorithm,
	}

	// 设置 BaseEntity 字段
	config.ID = po.ID
	config.Version = po.Version
	config.CreatedBy = po.CreatedBy
	config.UpdatedBy = po.UpdatedBy
	config.CreatedAt = po.CreatedAt
	config.UpdatedAt = po.UpdatedAt
	config.DeletedAt = po.DeletedAt

	return config
}

// ToPO 将领域实体转换为持久化对象（DO -> PO）
func (c *ConfigConverter) ToPO(do *domainEntity.Config) *infraEntity.ConfigPO {
	if do == nil {
		return nil
	}

	po := &infraEntity.ConfigPO{
		// BaseEntity 字段
		ID:        do.ID,
		Version:   do.Version,
		CreatedBy: do.CreatedBy,
		UpdatedBy: do.UpdatedBy,
		CreatedAt: do.CreatedAt,
		UpdatedAt: do.UpdatedAt,
		DeletedAt: do.DeletedAt,

		// 业务字段
		NamespaceID:          do.NamespaceID,
		Key:                  do.Key,
		Value:                do.Value,
		GroupName:            do.GroupName,
		ValueType:            do.ValueType,
		Environment:          do.Environment,
		IsReleased:           do.IsReleased,
		IsActive:             do.IsActive,
		Description:          do.Description,
		Metadata:             do.Metadata,
		ContentHash:          do.ContentHash,
		ContentHashAlgorithm: do.ContentHashAlgorithm,
	}

	// 同步软删除状态：如果 DeletedAt 有效，设置 IsDeleted = true
	if do.DeletedAt.Valid {
		po.IsDeleted = true
	}

	return po
}

// ToDOList 批量转换为领域实体列表
func (c *ConfigConverter) ToDOList(pos []*infraEntity.ConfigPO) []*domainEntity.Config {
	if len(pos) == 0 {
		return []*domainEntity.Config{}
	}

	dos := make([]*domainEntity.Config, len(pos))
	for i, po := range pos {
		dos[i] = c.ToDO(po)
	}
	return dos
}

// ToPOList 批量转换为持久化对象列表
func (c *ConfigConverter) ToPOList(dos []*domainEntity.Config) []*infraEntity.ConfigPO {
	if len(dos) == 0 {
		return []*infraEntity.ConfigPO{}
	}

	pos := make([]*infraEntity.ConfigPO, len(dos))
	for i, do := range dos {
		pos[i] = c.ToPO(do)
	}
	return pos
}
