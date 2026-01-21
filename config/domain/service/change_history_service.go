package service

import (
	"context"
	"fmt"
	"time"

	"config-client/config/domain/entity"
	domainErrors "config-client/config/domain/errors"
	"config-client/config/domain/repository"
	shareRepo "config-client/share/repository"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// ChangeHistoryService 变更历史领域服务
// 负责配置变更历史的记录、查询、对比和回滚
type ChangeHistoryService struct {
	historyRepo repository.ChangeHistoryRepository // 变更历史仓储
	configRepo  repository.ConfigRepository        // 配置仓储
	configSrv   *ConfigService                     // 配置服务（用于回滚时更新配置）
}

// NewChangeHistoryService 创建变更历史服务实例
func NewChangeHistoryService(
	historyRepo repository.ChangeHistoryRepository,
	configRepo repository.ConfigRepository,
	configSrv *ConfigService,
) *ChangeHistoryService {
	return &ChangeHistoryService{
		historyRepo: historyRepo,
		configRepo:  configRepo,
		configSrv:   configSrv,
	}
}

// ==================== 变更记录 ====================

// RecordChange 记录配置变更
// 在配置创建、更新、删除时调用
func (s *ChangeHistoryService) RecordChange(ctx context.Context, record *entity.ChangeRecord) error {
	if record == nil {
		return nil
	}

	history := record.ToEntity()

	// 异步保存，不阻塞主流程
	go func() {
		if err := s.historyRepo.Save(context.Background(), history); err != nil {
			hlog.Errorf("保存变更历史失败: %v, record: %+v", err, record)
		}
	}()

	return nil
}

// RecordChangeWithTx 在事务中记录配置变更（同步保存）
func (s *ChangeHistoryService) RecordChangeWithTx(ctx context.Context, record *entity.ChangeRecord) error {
	if record == nil {
		return nil
	}

	history := record.ToEntity()
	return s.historyRepo.Save(ctx, history)
}

// ==================== 变更查询 ====================

// GetHistoryByConfigID 查询指定配置的变更历史
func (s *ChangeHistoryService) GetHistoryByConfigID(ctx context.Context, configID int, limit int) ([]*entity.ChangeHistory, error) {
	if limit <= 0 {
		limit = 50 // 默认返回最近50条
	}
	return s.historyRepo.FindByConfigID(ctx, configID, limit)
}

// GetHistoryByID 根据ID查询变更记录
func (s *ChangeHistoryService) GetHistoryByID(ctx context.Context, historyID int) (*entity.ChangeHistory, error) {
	history, err := s.historyRepo.FindByID(ctx, historyID)
	if err != nil {
		return nil, err
	}
	if history == nil {
		return nil, domainErrors.ErrConfigNotFound("", fmt.Sprintf("history_id:%d", historyID))
	}
	return history, nil
}

// QueryHistory 分页查询变更历史
func (s *ChangeHistoryService) QueryHistory(ctx context.Context, params *repository.ChangeHistoryQueryParams) (*shareRepo.PageResult[*entity.ChangeHistory], error) {
	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Size <= 0 {
		params.Size = 20
	}
	if params.Size > 100 {
		params.Size = 100 // 限制最大每页数量
	}

	return s.historyRepo.QueryByParams(ctx, params)
}

// ==================== 变更对比 ====================

// CompareVersions 对比两个版本
// 返回差异信息
func (s *ChangeHistoryService) CompareVersions(ctx context.Context, historyID1, historyID2 int) (*VersionCompareResult, error) {
	history1, err := s.historyRepo.FindByID(ctx, historyID1)
	if err != nil {
		return nil, err
	}
	if history1 == nil {
		return nil, fmt.Errorf("变更记录 %d 不存在", historyID1)
	}

	history2, err := s.historyRepo.FindByID(ctx, historyID2)
	if err != nil {
		return nil, err
	}
	if history2 == nil {
		return nil, fmt.Errorf("变更记录 %d 不存在", historyID2)
	}

	// 确保 history1 是较早的版本
	if history1.CreatedAt.After(history2.CreatedAt) {
		history1, history2 = history2, history1
		historyID1, historyID2 = historyID2, historyID1
	}

	return &VersionCompareResult{
		FromHistoryID: historyID1,
		ToHistoryID:   historyID2,
		FromVersion:   history1.NewVersion,
		ToVersion:     history2.NewVersion,
		FromValue:     history1.NewValue,
		ToValue:       history2.NewValue,
		ValueChanged:  history1.NewValue != history2.NewValue,
		FromOperation: string(history1.Operation),
		ToOperation:   string(history2.Operation),
		FromChangedAt: history1.CreatedAt,
		ToChangedAt:   history2.CreatedAt,
	}, nil
}

// VersionCompareResult 版本对比结果
type VersionCompareResult struct {
	FromHistoryID int       // 源版本历史ID
	ToHistoryID   int       // 目标版本历史ID
	FromVersion   int       // 源版本号
	ToVersion     int       // 目标版本号
	FromValue     string    // 源版本值
	ToValue       string    // 目标版本值
	ValueChanged  bool      // 值是否变化
	FromOperation string    // 源操作类型
	ToOperation   string    // 目标操作类型
	FromChangedAt time.Time // 源变更时间
	ToChangedAt   time.Time // 目标变更时间
}

// ==================== 变更回滚 ====================

// RollbackToHistory 回滚到指定的历史版本
// 流程：
// 1. 查询目标历史记录
// 2. 验证是否可以回滚
// 3. 恢复配置值
// 4. 记录回滚操作
func (s *ChangeHistoryService) RollbackToHistory(ctx context.Context, req *entity.RollbackRecord) error {
	// 1. 查询目标历史记录
	targetHistory, err := s.historyRepo.FindByID(ctx, req.TargetHistoryID)
	if err != nil {
		return err
	}
	if targetHistory == nil {
		return fmt.Errorf("变更记录 %d 不存在", req.TargetHistoryID)
	}

	// 2. 验证是否可以回滚
	if !targetHistory.CanRollback() {
		return fmt.Errorf("该记录不支持回滚，操作类型: %s", targetHistory.Operation)
	}

	// 3. 查询当前配置
	config, err := s.configRepo.GetByID(ctx, targetHistory.ConfigID)
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("配置 %d 不存在", targetHistory.ConfigID)
	}

	// 4. 获取目标版本的值
	targetValue, _ := targetHistory.GetSnapshot()

	// 5. 记录当前状态（用于回滚前的历史记录）
	oldValue := config.Value
	oldVersion := config.Version

	// 6. 更新配置值
	// 重新计算内容哈希
	newHash, err := s.configSrv.ComputeContentHash(targetValue, "md5")
	if err != nil {
		return fmt.Errorf("计算内容哈希失败: %w", err)
	}
	config.UpdateValue(targetValue, newHash)

	// 7. 保存配置更新
	if err := s.configRepo.Update(ctx, config); err != nil {
		return fmt.Errorf("回滚配置失败: %w", err)
	}

	// 8. 记录回滚操作的变更历史
	rollbackRecord := &entity.ChangeRecord{
		ConfigID:     config.ID,
		NamespaceID:  config.NamespaceID,
		ConfigKey:    config.Key,
		Environment:  config.Environment,
		Operation:    entity.OperationRollback,
		OldValue:     oldValue,
		NewValue:     targetValue,
		OldVersion:   oldVersion,
		NewVersion:   config.Version,
		Operator:     req.Operator,
		OperatorIP:   req.OperatorIP,
		ChangeReason: fmt.Sprintf("回滚到版本 %d (历史记录ID: %d). %s", targetHistory.NewVersion, req.TargetHistoryID, req.ChangeReason),
	}

	if err := s.RecordChangeWithTx(ctx, rollbackRecord); err != nil {
		hlog.Errorf("记录回滚变更历史失败: %v", err)
	}

	hlog.Infof("配置回滚成功: configID=%d, key=%s, 回滚到版本=%d", config.ID, config.Key, targetHistory.NewVersion)

	return nil
}

// CanRollback 判断是否可以回滚到指定历史版本
func (s *ChangeHistoryService) CanRollback(ctx context.Context, historyID int) (bool, error) {
	history, err := s.historyRepo.FindByID(ctx, historyID)
	if err != nil {
		return false, err
	}
	if history == nil {
		return false, nil
	}
	return history.CanRollback(), nil
}

// ==================== 统计信息 ====================

// GetChangeStatistics 获取变更统计信息
func (s *ChangeHistoryService) GetChangeStatistics(ctx context.Context) (*ChangeStatistics, error) {
	stats := &ChangeStatistics{}

	// 统计各操作类型的数量
	createCount, _ := s.historyRepo.CountByOperation(ctx, "CREATE")
	updateCount, _ := s.historyRepo.CountByOperation(ctx, "UPDATE")
	deleteCount, _ := s.historyRepo.CountByOperation(ctx, "DELETE")
	rollbackCount, _ := s.historyRepo.CountByOperation(ctx, "ROLLBACK")

	stats.TotalChanges = createCount + updateCount + deleteCount + rollbackCount
	stats.CreateCount = createCount
	stats.UpdateCount = updateCount
	stats.DeleteCount = deleteCount
	stats.RollbackCount = rollbackCount

	return stats, nil
}

// ChangeStatistics 变更统计信息
type ChangeStatistics struct {
	TotalChanges  int64 // 总变更次数
	CreateCount   int64 // 创建次数
	UpdateCount   int64 // 更新次数
	DeleteCount   int64 // 删除次数
	RollbackCount int64 // 回滚次数
}
