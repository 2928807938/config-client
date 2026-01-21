package converter

import (
	"encoding/json"

	domainEntity "config-client/config/domain/entity"
	infraEntity "config-client/config/infrastructure/entity"
)

// ChangeHistoryConverter 变更历史转换器
type ChangeHistoryConverter struct{}

// NewChangeHistoryConverter 创建变更历史转换器实例
func NewChangeHistoryConverter() *ChangeHistoryConverter {
	return &ChangeHistoryConverter{}
}

// ToDO 将持久化对象转换为领域实体（PO -> DO）
func (c *ChangeHistoryConverter) ToDO(po *infraEntity.ChangeHistoryPO) *domainEntity.ChangeHistory {
	if po == nil {
		return nil
	}

	// 将 JSONB 转换为 JSON 字符串
	metadata := "{}"
	if po.Metadata != nil && len(po.Metadata) > 0 {
		// 将 map 序列化为 JSON 字符串
		if bytes, err := json.Marshal(po.Metadata); err == nil {
			metadata = string(bytes)
		}
	}

	return &domainEntity.ChangeHistory{
		ID:           po.ID,
		ConfigID:     po.ConfigID,
		NamespaceID:  po.NamespaceID,
		ConfigKey:    po.ConfigKey,
		Environment:  po.Environment,
		Operation:    domainEntity.Operation(po.Operation),
		OldValue:     po.OldValue,
		NewValue:     po.NewValue,
		OldVersion:   po.OldVersion,
		NewVersion:   po.NewVersion,
		Operator:     po.Operator,
		OperatorIP:   po.OperatorIP,
		ChangeReason: po.ChangeReason,
		CreatedAt:    po.CreatedAt,
		Metadata:     metadata,
	}
}

// ToPO 将领域实体转换为持久化对象（DO -> PO）
func (c *ChangeHistoryConverter) ToPO(do *domainEntity.ChangeHistory) *infraEntity.ChangeHistoryPO {
	if do == nil {
		return nil
	}

	// 解析 JSON 字符串为 JSONB
	var metadata infraEntity.JSONB
	if do.Metadata != "" && do.Metadata != "{}" {
		// 简单处理，实际使用时可以更精细地解析
		metadata = infraEntity.JSONB{}
	}

	return &infraEntity.ChangeHistoryPO{
		ID:           do.ID,
		ConfigID:     do.ConfigID,
		NamespaceID:  do.NamespaceID,
		ConfigKey:    do.ConfigKey,
		Environment:  do.Environment,
		Operation:    string(do.Operation),
		OldValue:     do.OldValue,
		NewValue:     do.NewValue,
		OldVersion:   do.OldVersion,
		NewVersion:   do.NewVersion,
		Operator:     do.Operator,
		OperatorIP:   do.OperatorIP,
		ChangeReason: do.ChangeReason,
		CreatedAt:    do.CreatedAt,
		Metadata:     metadata,
	}
}

// ToDOList 批量转换为领域实体列表
func (c *ChangeHistoryConverter) ToDOList(pos []*infraEntity.ChangeHistoryPO) []*domainEntity.ChangeHistory {
	if len(pos) == 0 {
		return []*domainEntity.ChangeHistory{}
	}

	dos := make([]*domainEntity.ChangeHistory, len(pos))
	for i, po := range pos {
		dos[i] = c.ToDO(po)
	}
	return dos
}

// ToPOList 批量转换为持久化对象列表
func (c *ChangeHistoryConverter) ToPOList(dos []*domainEntity.ChangeHistory) []*infraEntity.ChangeHistoryPO {
	if len(dos) == 0 {
		return []*infraEntity.ChangeHistoryPO{}
	}

	pos := make([]*infraEntity.ChangeHistoryPO, len(dos))
	for i, do := range dos {
		pos[i] = c.ToPO(do)
	}
	return pos
}
