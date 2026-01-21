package converter

import (
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/constants"
	"config-client/config/domain/entity"
	domainService "config-client/config/domain/service"
)

// ConfigConverter API层配置转换器，负责领域实体和视图对象之间的转换
type ConfigConverter struct {
	maskingSvc   *domainService.MaskingService   // 脱敏服务
	tagSvc       *domainService.ConfigTagService // 标签服务
	tagConverter *ConfigTagConverter             // 标签转换器
}

// NewConfigConverter 创建配置转换器实例
func NewConfigConverter(
	maskingSvc *domainService.MaskingService,
	tagSvc *domainService.ConfigTagService,
) *ConfigConverter {
	return &ConfigConverter{
		maskingSvc:   maskingSvc,
		tagSvc:       tagSvc,
		tagConverter: NewConfigTagConverter(),
	}
}

// ToVO 将领域实体转换为视图对象（DO -> VO）
// 注意：此方法会自动处理敏感配置的脱敏
func (c *ConfigConverter) ToVO(do *entity.Config) *vo.ConfigVO {
	if do == nil {
		return nil
	}

	// 判断是否为敏感配置
	isSensitive := false
	if c.maskingSvc != nil {
		isSensitive = c.maskingSvc.IsSensitiveKey(do.Key) || do.ValueType == constants.ValueTypeEncrypted
	}

	// 对敏感配置进行脱敏
	displayValue := do.Value
	isMasked := false
	if isSensitive && c.maskingSvc != nil {
		displayValue = c.maskingSvc.MaskValue(do.Value)
		isMasked = true
	}

	return &vo.ConfigVO{
		ID:                   do.ID,
		NamespaceID:          do.NamespaceID,
		Key:                  do.Key,
		Value:                displayValue, // 使用脱敏后的值
		GroupName:            do.GroupName,
		ValueType:            do.ValueType,
		Environment:          do.Environment,
		Version:              do.Version,
		IsReleased:           do.IsReleased,
		IsActive:             do.IsActive,
		IsSensitive:          isSensitive,
		IsMasked:             isMasked,
		Description:          do.Description,
		Metadata:             do.Metadata,
		ContentHash:          do.ContentHash,
		ContentHashAlgorithm: do.ContentHashAlgorithm,
		Tags:                 nil, // 标签需要单独查询和填充
		CreatedBy:            do.CreatedBy,
		UpdatedBy:            do.UpdatedBy,
		CreatedAt:            do.CreatedAt,
		UpdatedAt:            do.UpdatedAt,
	}
}

// ToVOList 批量转换为视图对象列表
func (c *ConfigConverter) ToVOList(dos []*entity.Config) []*vo.ConfigVO {
	if len(dos) == 0 {
		return []*vo.ConfigVO{}
	}

	vos := make([]*vo.ConfigVO, len(dos))
	for i, do := range dos {
		vos[i] = c.ToVO(do)
	}
	return vos
}

// ToListVO 转换为分页列表视图对象
func (c *ConfigConverter) ToListVO(dos []*entity.Config, total int64, page, size int) *vo.ConfigListVO {
	totalPages := int(total) / size
	if int(total)%size != 0 {
		totalPages++
	}

	return &vo.ConfigListVO{
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
		Items:      c.ToVOList(dos),
	}
}

// ToVOWithTags 将领域实体转换为视图对象（包含标签）
// 此方法需要传入标签列表，由调用方负责查询标签
func (c *ConfigConverter) ToVOWithTags(do *entity.Config, tags []*entity.ConfigTag) *vo.ConfigVO {
	configVO := c.ToVO(do)
	if configVO != nil && len(tags) > 0 {
		configVO.Tags = c.tagConverter.ToVOList(tags)
	}
	return configVO
}
