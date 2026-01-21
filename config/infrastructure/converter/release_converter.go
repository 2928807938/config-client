package converter

import (
	domainEntity "config-client/config/domain/entity"
	infraEntity "config-client/config/infrastructure/entity"
)

// ReleaseConverter 发布版本转换器，负责领域实体和持久化对象之间的转换
type ReleaseConverter struct{}

// NewReleaseConverter 创建发布版本转换器实例
func NewReleaseConverter() *ReleaseConverter {
	return &ReleaseConverter{}
}

// ToDO 将持久化对象转换为领域实体（PO -> DO）
func (c *ReleaseConverter) ToDO(po *infraEntity.ReleasePO) *domainEntity.Release {
	if po == nil {
		return nil
	}

	release := &domainEntity.Release{
		NamespaceID:         po.NamespaceID,
		Environment:         po.Environment,
		Version:             po.Version,
		VersionName:         po.VersionName,
		ConfigSnapshot:      po.ConfigSnapshot,
		ConfigCount:         po.ConfigCount,
		Status:              domainEntity.ReleaseStatus(po.Status),
		ReleaseType:         domainEntity.ReleaseType(po.ReleaseType),
		CanaryRule:          po.CanaryRule,
		CanaryPercentage:    po.CanaryPercentage,
		ReleasedBy:          po.ReleasedBy,
		ReleasedAt:          po.ReleasedAt,
		RollbackFromVersion: po.RollbackFromVersion,
		RollbackBy:          po.RollbackBy,
		RollbackAt:          po.RollbackAt,
		RollbackReason:      po.RollbackReason,
	}

	// 设置 BaseEntity 字段
	release.ID = po.ID
	release.CreatedBy = po.CreatedBy
	release.CreatedAt = po.CreatedAt
	release.UpdatedAt = po.UpdatedAt
	release.DeletedAt = po.DeletedAt

	return release
}

// ToPO 将领域实体转换为持久化对象（DO -> PO）
func (c *ReleaseConverter) ToPO(do *domainEntity.Release) *infraEntity.ReleasePO {
	if do == nil {
		return nil
	}

	po := &infraEntity.ReleasePO{
		// BaseEntity 字段
		ID:        do.ID,
		CreatedBy: do.CreatedBy,
		CreatedAt: do.CreatedAt,
		UpdatedAt: do.UpdatedAt,
		DeletedAt: do.DeletedAt,

		// 业务字段
		NamespaceID:         do.NamespaceID,
		Environment:         do.Environment,
		Version:             do.Version,
		VersionName:         do.VersionName,
		ConfigSnapshot:      do.ConfigSnapshot,
		ConfigCount:         do.ConfigCount,
		Status:              string(do.Status),
		ReleaseType:         string(do.ReleaseType),
		CanaryRule:          do.CanaryRule,
		CanaryPercentage:    do.CanaryPercentage,
		ReleasedBy:          do.ReleasedBy,
		ReleasedAt:          do.ReleasedAt,
		RollbackFromVersion: do.RollbackFromVersion,
		RollbackBy:          do.RollbackBy,
		RollbackAt:          do.RollbackAt,
		RollbackReason:      do.RollbackReason,
	}

	return po
}

// ToDOList 批量转换 PO -> DO
func (c *ReleaseConverter) ToDOList(pos []*infraEntity.ReleasePO) []*domainEntity.Release {
	if len(pos) == 0 {
		return []*domainEntity.Release{}
	}

	dos := make([]*domainEntity.Release, 0, len(pos))
	for _, po := range pos {
		dos = append(dos, c.ToDO(po))
	}
	return dos
}

// ToPOList 批量转换 DO -> PO
func (c *ReleaseConverter) ToPOList(dos []*domainEntity.Release) []*infraEntity.ReleasePO {
	if len(dos) == 0 {
		return []*infraEntity.ReleasePO{}
	}

	pos := make([]*infraEntity.ReleasePO, 0, len(dos))
	for _, do := range dos {
		pos = append(pos, c.ToPO(do))
	}
	return pos
}
