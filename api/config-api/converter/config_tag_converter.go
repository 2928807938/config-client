package converter

import (
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
)

// ConfigTagConverter API层标签转换器
type ConfigTagConverter struct{}

// NewConfigTagConverter 创建标签转换器实例
func NewConfigTagConverter() *ConfigTagConverter {
	return &ConfigTagConverter{}
}

// ToVO 将领域实体转换为VO
func (c *ConfigTagConverter) ToVO(tag *entity.ConfigTag) *vo.ConfigTagVO {
	if tag == nil {
		return nil
	}

	return &vo.ConfigTagVO{
		ID:        tag.ID,
		ConfigID:  tag.ConfigID,
		TagKey:    tag.TagKey,
		TagValue:  tag.TagValue,
		CreatedAt: tag.CreatedAt,
	}
}

// ToVOList 将领域实体列表转换为VO列表
func (c *ConfigTagConverter) ToVOList(tags []*entity.ConfigTag) []*vo.ConfigTagVO {
	if len(tags) == 0 {
		return []*vo.ConfigTagVO{}
	}

	result := make([]*vo.ConfigTagVO, 0, len(tags))
	for _, tag := range tags {
		result = append(result, c.ToVO(tag))
	}
	return result
}

// ToTagListVO 将领域实体列表转换为标签列表VO
func (c *ConfigTagConverter) ToTagListVO(configID int, tags []*entity.ConfigTag) *vo.ConfigTagListVO {
	return &vo.ConfigTagListVO{
		ConfigID: configID,
		Tags:     c.ToVOList(tags),
	}
}

// ToDomainInputs 将请求DTO转换为领域实体输入
func (c *ConfigTagConverter) ToDomainInputs(inputs []entity.TagInput) []entity.TagInput {
	return inputs // TagInput在domain和api层使用相同结构
}
