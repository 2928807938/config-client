package converter

import (
	domainEntity "config-client/config/domain/entity"
	infraEntity "config-client/config/infrastructure/entity"
)

// SystemConfigConverter 系统配置转换器，负责领域实体和持久化对象之间的转换
type SystemConfigConverter struct{}

// NewSystemConfigConverter 创建系统配置转换器实例
func NewSystemConfigConverter() *SystemConfigConverter {
	return &SystemConfigConverter{}
}

// ToDO 将持久化对象转换为领域实体（PO -> DO）
func (c *SystemConfigConverter) ToDO(po *infraEntity.SystemConfigPO) *domainEntity.SystemConfig {
	if po == nil {
		return nil
	}

	config := &domainEntity.SystemConfig{
		ID:          po.ID,
		ConfigKey:   po.ConfigKey,
		ConfigValue: po.ConfigValue,
		Description: po.Description,
		IsActive:    po.IsActive,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}

	return config
}

// ToPO 将领域实体转换为持久化对象（DO -> PO）
func (c *SystemConfigConverter) ToPO(do *domainEntity.SystemConfig) *infraEntity.SystemConfigPO {
	if do == nil {
		return nil
	}

	po := &infraEntity.SystemConfigPO{
		ID:          do.ID,
		ConfigKey:   do.ConfigKey,
		ConfigValue: do.ConfigValue,
		Description: do.Description,
		IsActive:    do.IsActive,
		CreatedAt:   do.CreatedAt,
		UpdatedAt:   do.UpdatedAt,
	}

	return po
}

// ToDOList 批量转换 PO -> DO
func (c *SystemConfigConverter) ToDOList(pos []*infraEntity.SystemConfigPO) []*domainEntity.SystemConfig {
	if len(pos) == 0 {
		return []*domainEntity.SystemConfig{}
	}

	dos := make([]*domainEntity.SystemConfig, 0, len(pos))
	for _, po := range pos {
		dos = append(dos, c.ToDO(po))
	}
	return dos
}

// ToPOList 批量转换 DO -> PO
func (c *SystemConfigConverter) ToPOList(dos []*domainEntity.SystemConfig) []*infraEntity.SystemConfigPO {
	if len(dos) == 0 {
		return []*infraEntity.SystemConfigPO{}
	}

	pos := make([]*infraEntity.SystemConfigPO, 0, len(dos))
	for _, do := range dos {
		pos = append(pos, c.ToPO(do))
	}
	return pos
}
