package converter

import (
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
	shareRepo "config-client/share/repository"
)

// ChangeHistoryConverter API层变更历史转换器，负责领域实体和视图对象之间的转换
type ChangeHistoryConverter struct{}

// NewChangeHistoryConverter 创建变更历史转换器实例
func NewChangeHistoryConverter() *ChangeHistoryConverter {
	return &ChangeHistoryConverter{}
}

// ToVO 将领域实体转换为视图对象（DO -> VO）
func (c *ChangeHistoryConverter) ToVO(history *entity.ChangeHistory) *vo.ChangeHistoryVO {
	if history == nil {
		return nil
	}

	return &vo.ChangeHistoryVO{
		ID:            history.ID,
		ConfigID:      history.ConfigID,
		NamespaceID:   history.NamespaceID,
		ConfigKey:     history.ConfigKey,
		Environment:   history.Environment,
		Operation:     string(history.Operation),
		OldValue:      history.OldValue,
		NewValue:      history.NewValue,
		OldVersion:    history.OldVersion,
		NewVersion:    history.NewVersion,
		Operator:      history.Operator,
		OperatorIP:    history.OperatorIP,
		ChangeReason:  history.ChangeReason,
		CreatedAt:     history.CreatedAt,
		Metadata:      history.Metadata,
		CanRollback:   history.CanRollback(),
		ChangeSummary: history.GetChangeSummary(),
		ValueChanged:  history.HasValueChanged(),
	}
}

// ToVOList 批量转换为视图对象列表
func (c *ChangeHistoryConverter) ToVOList(histories []*entity.ChangeHistory) []*vo.ChangeHistoryVO {
	if len(histories) == 0 {
		return []*vo.ChangeHistoryVO{}
	}

	vos := make([]*vo.ChangeHistoryVO, len(histories))
	for i, history := range histories {
		vos[i] = c.ToVO(history)
	}

	return vos
}

// ToListVO 将分页结果转换为列表VO
func (c *ChangeHistoryConverter) ToListVO(result *shareRepo.PageResult[*entity.ChangeHistory]) *vo.ChangeHistoryListVO {
	if result == nil {
		return &vo.ChangeHistoryListVO{
			Total:      0,
			Page:       1,
			Size:       0,
			TotalPages: 0,
			Items:      []*vo.ChangeHistoryVO{},
		}
	}

	items := c.ToVOList(result.Items)

	return &vo.ChangeHistoryListVO{
		Total:      result.Total,
		Page:       result.Page,
		Size:       result.Size,
		TotalPages: int((result.Total + int64(result.Size) - 1) / int64(result.Size)),
		Items:      items,
	}
}
