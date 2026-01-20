package converter

import (
	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
)

// NamespaceConverter 命名空间转换器，负责 DTO 和领域实体之间的转换
type NamespaceConverter struct{}

// NewNamespaceConverter 创建命名空间转换器实例
func NewNamespaceConverter() *NamespaceConverter {
	return &NamespaceConverter{}
}

// ToEntity 将创建请求转换为领域实体（Request -> DO）
func (c *NamespaceConverter) ToEntity(req *request.CreateNamespaceRequest) *entity.Namespace {
	if req == nil {
		return nil
	}

	namespace := &entity.Namespace{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Metadata:    req.Metadata,
	}

	// 如果 Metadata 为空，设置默认值
	if namespace.Metadata == "" {
		namespace.Metadata = "{}"
	}

	return namespace
}

// ToVO 将领域实体转换为视图对象（DO -> VO）
func (c *NamespaceConverter) ToVO(do *entity.Namespace) *vo.NamespaceVO {
	if do == nil {
		return nil
	}

	return &vo.NamespaceVO{
		ID:          do.ID,
		Name:        do.Name,
		DisplayName: do.DisplayName,
		Description: do.Description,
		IsActive:    do.IsActive,
		Metadata:    do.Metadata,
		CreatedBy:   do.CreatedBy,
		UpdatedBy:   do.UpdatedBy,
		CreatedAt:   do.CreatedAt,
		UpdatedAt:   do.UpdatedAt,
	}
}

// ToVOList 批量转换为视图对象列表
func (c *NamespaceConverter) ToVOList(dos []*entity.Namespace) []*vo.NamespaceVO {
	if len(dos) == 0 {
		return []*vo.NamespaceVO{}
	}

	vos := make([]*vo.NamespaceVO, len(dos))
	for i, do := range dos {
		vos[i] = c.ToVO(do)
	}
	return vos
}

// UpdateEntityFromRequest 从更新请求更新领域实体
func (c *NamespaceConverter) UpdateEntityFromRequest(entity *entity.Namespace, req *request.UpdateNamespaceRequest) {
	if entity == nil || req == nil {
		return
	}

	entity.DisplayName = req.DisplayName
	entity.Description = req.Description
	entity.Metadata = req.Metadata

	// 如果 Metadata 为空，设置默认值
	if entity.Metadata == "" {
		entity.Metadata = "{}"
	}
}
