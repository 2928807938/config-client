package converter

import (
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
	domainService "config-client/config/domain/service"
)

// ReleaseConverter 发布版本转换器
type ReleaseConverter struct{}

// NewReleaseConverter 创建发布版本转换器
func NewReleaseConverter() *ReleaseConverter {
	return &ReleaseConverter{}
}

// ToVO 将领域实体转换为VO
func (c *ReleaseConverter) ToVO(release *entity.Release, includeSnapshot bool) *vo.ReleaseVO {
	if release == nil {
		return nil
	}

	releaseVO := &vo.ReleaseVO{
		ID:                  release.ID,
		NamespaceID:         release.NamespaceID,
		Environment:         release.Environment,
		Version:             release.Version,
		VersionName:         release.VersionName,
		ConfigCount:         release.ConfigCount,
		Status:              string(release.Status),
		ReleaseType:         string(release.ReleaseType),
		CanaryPercentage:    release.CanaryPercentage,
		ReleasedBy:          release.ReleasedBy,
		ReleasedAt:          release.ReleasedAt,
		RollbackFromVersion: release.RollbackFromVersion,
		RollbackBy:          release.RollbackBy,
		RollbackAt:          release.RollbackAt,
		RollbackReason:      release.RollbackReason,
		CreatedBy:           release.CreatedBy,
		CreatedAt:           release.CreatedAt,
		UpdatedAt:           release.UpdatedAt,
	}

	// 转换灰度规则
	if release.IsCanaryRelease() {
		canaryRule, err := release.GetCanaryRule()
		if err == nil && canaryRule != nil {
			releaseVO.CanaryRule = &vo.CanaryRuleVO{
				ClientIDs:  canaryRule.ClientIDs,
				IPRanges:   canaryRule.IPRanges,
				Percentage: canaryRule.Percentage,
			}
		}
	}

	// 转换配置快照 (可选)
	if includeSnapshot {
		snapshot, err := release.GetConfigSnapshot()
		if err == nil && len(snapshot) > 0 {
			releaseVO.ConfigSnapshot = make([]vo.ConfigSnapshotItemVO, 0, len(snapshot))
			for _, item := range snapshot {
				releaseVO.ConfigSnapshot = append(releaseVO.ConfigSnapshot, vo.ConfigSnapshotItemVO{
					ConfigID:             item.ConfigID,
					Key:                  item.Key,
					Value:                item.Value,
					ValueType:            item.ValueType,
					GroupName:            item.GroupName,
					ContentHash:          item.ContentHash,
					ContentHashAlgorithm: item.ContentHashAlgorithm,
					Description:          item.Description,
					Version:              item.Version,
				})
			}
		}
	}

	return releaseVO
}

// ToVOList 批量转换为VO列表
func (c *ReleaseConverter) ToVOList(releases []*entity.Release) []*vo.ReleaseVO {
	if len(releases) == 0 {
		return []*vo.ReleaseVO{}
	}

	vos := make([]*vo.ReleaseVO, 0, len(releases))
	for _, release := range releases {
		vos = append(vos, c.ToVO(release, false)) // 列表不包含快照
	}
	return vos
}

// ToCompareVO 转换版本对比结果
func (c *ReleaseConverter) ToCompareVO(result *domainService.ReleaseCompareResult) *vo.ReleaseCompareVO {
	if result == nil {
		return nil
	}

	compareVO := &vo.ReleaseCompareVO{
		FromReleaseID: result.FromReleaseID,
		ToReleaseID:   result.ToReleaseID,
		FromVersion:   result.FromVersion,
		ToVersion:     result.ToVersion,
		Added:         make([]vo.ConfigSnapshotItemVO, 0),
		Deleted:       make([]vo.ConfigSnapshotItemVO, 0),
		Modified:      make([]vo.ConfigDiffVO, 0),
	}

	// 转换新增的配置
	for _, item := range result.Added {
		compareVO.Added = append(compareVO.Added, vo.ConfigSnapshotItemVO{
			ConfigID:             item.ConfigID,
			Key:                  item.Key,
			Value:                item.Value,
			ValueType:            item.ValueType,
			GroupName:            item.GroupName,
			ContentHash:          item.ContentHash,
			ContentHashAlgorithm: item.ContentHashAlgorithm,
			Description:          item.Description,
			Version:              item.Version,
		})
	}

	// 转换删除的配置
	for _, item := range result.Deleted {
		compareVO.Deleted = append(compareVO.Deleted, vo.ConfigSnapshotItemVO{
			ConfigID:             item.ConfigID,
			Key:                  item.Key,
			Value:                item.Value,
			ValueType:            item.ValueType,
			GroupName:            item.GroupName,
			ContentHash:          item.ContentHash,
			ContentHashAlgorithm: item.ContentHashAlgorithm,
			Description:          item.Description,
			Version:              item.Version,
		})
	}

	// 转换修改的配置
	for _, diff := range result.Modified {
		compareVO.Modified = append(compareVO.Modified, vo.ConfigDiffVO{
			Key:      diff.Key,
			OldValue: diff.OldValue,
			NewValue: diff.NewValue,
		})
	}

	return compareVO
}
