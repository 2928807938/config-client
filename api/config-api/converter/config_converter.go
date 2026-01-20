package converter

import (
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
)

// ConfigConverter API层配置转换器，负责领域实体和视图对象之间的转换
type ConfigConverter struct{}

// NewConfigConverter 创建配置转换器实例
func NewConfigConverter() *ConfigConverter {
	return &ConfigConverter{}
}

// ToVO 将领域实体转换为视图对象（DO -> VO）
func (c *ConfigConverter) ToVO(do *entity.Config) *vo.ConfigVO {
	if do == nil {
		return nil
	}

	return &vo.ConfigVO{
		ID:          do.ID,
		NamespaceID: do.NamespaceID,
		Key:         do.Key,
		Value:       do.Value,
		GroupName:   do.GroupName,
		ValueType:   do.ValueType,
		Environment: do.Environment,
		Version:     do.Version,
		IsReleased:  do.IsReleased,
		IsActive:    do.IsActive,
		Description: do.Description,
		Metadata:    do.Metadata,
		CreatedBy:   do.CreatedBy,
		UpdatedBy:   do.UpdatedBy,
		CreatedAt:   do.CreatedAt,
		UpdatedAt:   do.UpdatedAt,
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
