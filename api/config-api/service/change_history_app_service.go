package service

import (
	"context"

	"config-client/api/config-api/converter"
	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
	"config-client/config/domain/repository"
	domainService "config-client/config/domain/service"
	"config-client/share/constants"
)

// ChangeHistoryAppService 变更历史应用服务
type ChangeHistoryAppService struct {
	changeHistoryService *domainService.ChangeHistoryService
	converter            *converter.ChangeHistoryConverter
}

// NewChangeHistoryAppService 创建变更历史应用服务实例
func NewChangeHistoryAppService(
	changeHistoryService *domainService.ChangeHistoryService,
) *ChangeHistoryAppService {
	return &ChangeHistoryAppService{
		changeHistoryService: changeHistoryService,
		converter:            converter.NewChangeHistoryConverter(),
	}
}

// QueryHistory 分页查询变更历史
func (s *ChangeHistoryAppService) QueryHistory(ctx context.Context, req *request.QueryHistoryRequest) (*vo.ChangeHistoryListVO, error) {
	req.SetDefaults()

	params := &repository.ChangeHistoryQueryParams{
		ConfigID:    req.ConfigID,
		NamespaceID: req.NamespaceID,
		ConfigKey:   req.ConfigKey,
		Operation:   req.Operation,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Operator:    req.Operator,
		Page:        req.Page,
		Size:        req.Size,
	}

	pageResult, err := s.changeHistoryService.QueryHistory(ctx, params)
	if err != nil {
		return nil, err
	}

	return s.converter.ToListVO(pageResult), nil
}

// GetHistoryByID 根据ID查询变更记录
func (s *ChangeHistoryAppService) GetHistoryByID(ctx context.Context, historyID int) (*vo.ChangeHistoryVO, error) {
	history, err := s.changeHistoryService.GetHistoryByID(ctx, historyID)
	if err != nil {
		return nil, err
	}

	return s.converter.ToVO(history), nil
}

// GetConfigHistory 获取指定配置的变更历史
func (s *ChangeHistoryAppService) GetConfigHistory(ctx context.Context, req *request.GetConfigHistoryRequest) (*vo.ChangeHistoryListVO, error) {
	req.SetDefaults()

	histories, err := s.changeHistoryService.GetHistoryByConfigID(ctx, req.ConfigID, req.Limit)
	if err != nil {
		return nil, err
	}

	items := s.converter.ToVOList(histories)

	return &vo.ChangeHistoryListVO{
		Total:      int64(len(items)),
		Page:       1,
		Size:       len(items),
		TotalPages: 1,
		Items:      items,
	}, nil
}

// CompareVersions 对比两个版本
func (s *ChangeHistoryAppService) CompareVersions(ctx context.Context, req *request.CompareVersionsRequest) (*vo.VersionCompareVO, error) {
	result, err := s.changeHistoryService.CompareVersions(ctx, req.HistoryID1, req.HistoryID2)
	if err != nil {
		return nil, err
	}

	return &vo.VersionCompareVO{
		FromHistoryID: result.FromHistoryID,
		ToHistoryID:   result.ToHistoryID,
		FromVersion:   result.FromVersion,
		ToVersion:     result.ToVersion,
		FromValue:     result.FromValue,
		ToValue:       result.ToValue,
		ValueChanged:  result.ValueChanged,
		FromOperation: result.FromOperation,
		ToOperation:   result.ToOperation,
		FromChangedAt: result.FromChangedAt,
		ToChangedAt:   result.ToChangedAt,
	}, nil
}

// Rollback 回滚配置到指定版本
func (s *ChangeHistoryAppService) Rollback(ctx context.Context, req *request.RollbackRequest) error {
	// 从 context 中获取操作人信息
	operator := s.getOperator(ctx)
	operatorIP := s.getOperatorIP(ctx)

	rollbackReq := &entity.RollbackRecord{
		ConfigID:        0, // 在服务内部从历史记录获取
		TargetHistoryID: req.HistoryID,
		Operator:        operator,
		OperatorIP:      operatorIP,
		ChangeReason:    req.ChangeReason,
	}

	return s.changeHistoryService.RollbackToHistory(ctx, rollbackReq)
}

// GetStatistics 获取变更统计信息
func (s *ChangeHistoryAppService) GetStatistics(ctx context.Context) (*vo.ChangeStatisticsVO, error) {
	stats, err := s.changeHistoryService.GetChangeStatistics(ctx)
	if err != nil {
		return nil, err
	}

	return &vo.ChangeStatisticsVO{
		TotalChanges:  stats.TotalChanges,
		CreateCount:   stats.CreateCount,
		UpdateCount:   stats.UpdateCount,
		DeleteCount:   stats.DeleteCount,
		RollbackCount: stats.RollbackCount,
	}, nil
}

// ==================== 辅助方法 ====================

// getOperator 从 context 中获取操作人
func (s *ChangeHistoryAppService) getOperator(ctx context.Context) string {
	if operator, ok := ctx.Value(constants.OperatorKey).(string); ok {
		return operator
	}
	return "system"
}

// getOperatorIP 从 context 中获取操作人IP
func (s *ChangeHistoryAppService) getOperatorIP(ctx context.Context) string {
	if ip, ok := ctx.Value(constants.OperatorIPKey).(string); ok {
		return ip
	}
	return ""
}
