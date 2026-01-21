package converter

import (
	domainEntity "config-client/config/domain/entity"
	"config-client/config/infrastructure/entity"
)

// ConfigTagConverter 配置标签转换器
// 负责领域实体和持久化对象之间的转换
type ConfigTagConverter struct{}

// NewConfigTagConverter 创建配置标签转换器实例
func NewConfigTagConverter() *ConfigTagConverter {
	return &ConfigTagConverter{}
}

// ToPO 将领域实体转换为持久化对象
func (c *ConfigTagConverter) ToPO(tag *domainEntity.ConfigTag) *entity.ConfigTagPO {
	if tag == nil {
		return nil
	}

	return &entity.ConfigTagPO{
		ID:        tag.ID,
		ConfigID:  tag.ConfigID,
		TagKey:    tag.TagKey,
		TagValue:  tag.TagValue,
		CreatedAt: tag.CreatedAt,
	}
}

// ToDomain 将持久化对象转换为领域实体
func (c *ConfigTagConverter) ToDomain(po *entity.ConfigTagPO) *domainEntity.ConfigTag {
	if po == nil {
		return nil
	}

	return &domainEntity.ConfigTag{
		ID:        po.ID,
		ConfigID:  po.ConfigID,
		TagKey:    po.TagKey,
		TagValue:  po.TagValue,
		CreatedAt: po.CreatedAt,
	}
}

// ToDomainList 将持久化对象列表转换为领域实体列表
func (c *ConfigTagConverter) ToDomainList(poList []*entity.ConfigTagPO) []*domainEntity.ConfigTag {
	if len(poList) == 0 {
		return []*domainEntity.ConfigTag{}
	}

	result := make([]*domainEntity.ConfigTag, 0, len(poList))
	for _, po := range poList {
		result = append(result, c.ToDomain(po))
	}
	return result
}

// ToPOList 将领域实体列表转换为持久化对象列表
func (c *ConfigTagConverter) ToPOList(tagList []*domainEntity.ConfigTag) []*entity.ConfigTagPO {
	if len(tagList) == 0 {
		return []*entity.ConfigTagPO{}
	}

	result := make([]*entity.ConfigTagPO, 0, len(tagList))
	for _, tag := range tagList {
		result = append(result, c.ToPO(tag))
	}
	return result
}
