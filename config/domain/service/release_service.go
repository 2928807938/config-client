package service

import (
	"context"
	"fmt"

	"config-client/config/domain/entity"
	"config-client/config/domain/listener"
	"config-client/config/domain/repository"
	shareRepo "config-client/share/repository"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// ReleaseService 发布管理领域服务
// 负责配置版本的发布、灰度、回滚等核心业务逻辑
type ReleaseService struct {
	releaseRepo  repository.ReleaseRepository
	configRepo   repository.ConfigRepository
	configSvc    *ConfigService
	listener     listener.ConfigListener
	canaryEngine *CanaryRuleEngine
}

// NewReleaseService 创建发布管理服务
func NewReleaseService(
	releaseRepo repository.ReleaseRepository,
	configRepo repository.ConfigRepository,
	configSvc *ConfigService,
	listener listener.ConfigListener,
	canaryEngine *CanaryRuleEngine,
) *ReleaseService {
	return &ReleaseService{
		releaseRepo:  releaseRepo,
		configRepo:   configRepo,
		configSvc:    configSvc,
		listener:     listener,
		canaryEngine: canaryEngine,
	}
}

// ==================== 版本创建 ====================

// CreateReleaseRequest 创建发布版本请求
type CreateReleaseRequest struct {
	NamespaceID int
	Environment string
	VersionName string
	ReleaseType entity.ReleaseType
	CreatedBy   string
}

// CreateRelease 创建发布版本
// 将当前命名空间下的所有配置打快照,创建发布版本
func (s *ReleaseService) CreateRelease(ctx context.Context, req *CreateReleaseRequest) (*entity.Release, error) {
	// 1. 查询该命名空间下的所有配置
	configs, err := s.configRepo.FindReleasedConfigs(ctx, req.NamespaceID, req.Environment)
	if err != nil {
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("没有已发布的配置可以创建版本")
	}

	// 2. 构建配置快照
	snapshot := make([]entity.ConfigSnapshotItem, 0, len(configs))
	for _, config := range configs {
		snapshot = append(snapshot, entity.ConfigSnapshotItem{
			ConfigID:             config.ID,
			Key:                  config.Key,
			Value:                config.Value,
			ValueType:            config.ValueType,
			GroupName:            config.GroupName,
			ContentHash:          config.ContentHash,
			ContentHashAlgorithm: config.ContentHashAlgorithm,
			Description:          config.Description,
			Version:              config.Version,
		})
	}

	// 3. 获取下一个版本号
	nextVersion, err := s.releaseRepo.GetNextVersion(ctx, req.NamespaceID, req.Environment)
	if err != nil {
		return nil, fmt.Errorf("获取版本号失败: %w", err)
	}

	// 4. 创建发布版本实体
	release := &entity.Release{
		NamespaceID: req.NamespaceID,
		Environment: req.Environment,
		Version:     nextVersion,
		VersionName: req.VersionName,
		Status:      entity.ReleaseStatusTesting,
		ReleaseType: req.ReleaseType,
	}
	release.CreatedBy = req.CreatedBy

	// 设置配置快照
	if err := release.SetConfigSnapshot(snapshot); err != nil {
		return nil, fmt.Errorf("设置配置快照失败: %w", err)
	}

	// 5. 保存发布版本
	if err := s.releaseRepo.Create(ctx, release); err != nil {
		return nil, fmt.Errorf("保存发布版本失败: %w", err)
	}

	hlog.Infof("创建发布版本成功: namespace=%d, env=%s, version=%d, versionName=%s",
		req.NamespaceID, req.Environment, release.Version, release.VersionName)

	return release, nil
}

// ==================== 发布操作 ====================

// PublishRequest 发布请求
type PublishRequest struct {
	ReleaseID   int
	PublishedBy string
}

// PublishFull 全量发布
// 将发布版本标记为已发布,并通知所有订阅者
func (s *ReleaseService) PublishFull(ctx context.Context, req *PublishRequest) error {
	// 1. 查询发布版本
	release, err := s.releaseRepo.GetByID(ctx, req.ReleaseID)
	if err != nil {
		return fmt.Errorf("查询发布版本失败: %w", err)
	}
	if release == nil {
		return fmt.Errorf("发布版本不存在: id=%d", req.ReleaseID)
	}

	// 2. 检查是否可以发布
	if !release.CanPublish() {
		return fmt.Errorf("发布版本状态不允许发布: status=%s", release.Status)
	}

	// 3. 标记为已发布
	release.Publish(req.PublishedBy)

	// 4. 保存更新
	if err := s.releaseRepo.Update(ctx, release); err != nil {
		return fmt.Errorf("更新发布版本失败: %w", err)
	}

	// 5. 发布配置变更事件,通知所有订阅者
	snapshot, err := release.GetConfigSnapshot()
	if err != nil {
		hlog.Errorf("获取配置快照失败: %v", err)
	} else {
		for _, item := range snapshot {
			s.publishConfigChangeEvent(ctx, &listener.ConfigChangeEvent{
				NamespaceID: release.NamespaceID,
				ConfigKey:   item.Key,
				ConfigID:    item.ConfigID,
				Action:      "release",
			})
		}
	}

	hlog.Infof("全量发布成功: releaseID=%d, version=%d, configCount=%d",
		release.ID, release.Version, release.ConfigCount)

	return nil
}

// PublishCanaryRequest 灰度发布请求
type PublishCanaryRequest struct {
	ReleaseID   int
	CanaryRule  *entity.CanaryRule
	PublishedBy string
}

// PublishCanary 灰度发布
// 设置灰度规则,只对匹配规则的客户端生效
func (s *ReleaseService) PublishCanary(ctx context.Context, req *PublishCanaryRequest) error {
	// 1. 验证灰度规则
	if err := s.canaryEngine.ValidateRule(req.CanaryRule); err != nil {
		return fmt.Errorf("灰度规则验证失败: %w", err)
	}

	// 2. 查询发布版本
	release, err := s.releaseRepo.GetByID(ctx, req.ReleaseID)
	if err != nil {
		return fmt.Errorf("查询发布版本失败: %w", err)
	}
	if release == nil {
		return fmt.Errorf("发布版本不存在: id=%d", req.ReleaseID)
	}

	// 3. 检查是否可以发布
	if !release.CanPublish() {
		return fmt.Errorf("发布版本状态不允许发布: status=%s", release.Status)
	}

	// 4. 设置灰度规则
	if err := release.SetCanaryRule(req.CanaryRule); err != nil {
		return fmt.Errorf("设置灰度规则失败: %w", err)
	}

	// 5. 标记为已发布
	release.Publish(req.PublishedBy)
	release.ReleaseType = entity.ReleaseTypeCanary

	// 6. 保存更新
	if err := s.releaseRepo.Update(ctx, release); err != nil {
		return fmt.Errorf("更新发布版本失败: %w", err)
	}

	// 7. 发布配置变更事件（订阅管理器会根据灰度规则过滤）
	snapshot, err := release.GetConfigSnapshot()
	if err != nil {
		hlog.Errorf("获取配置快照失败: %v", err)
	} else {
		for _, item := range snapshot {
			s.publishConfigChangeEvent(ctx, &listener.ConfigChangeEvent{
				NamespaceID: release.NamespaceID,
				ConfigKey:   item.Key,
				ConfigID:    item.ConfigID,
				Action:      "canary_release",
			})
		}
	}

	hlog.Infof("灰度发布成功: releaseID=%d, version=%d, percentage=%d",
		release.ID, release.Version, req.CanaryRule.Percentage)

	return nil
}

// ==================== 回滚操作 ====================

// RollbackRequest 回滚请求
type RollbackRequest struct {
	CurrentReleaseID int
	TargetReleaseID  int
	RollbackBy       string
	Reason           string
}

// Rollback 回滚到指定版本
// 1. 标记当前版本为已回滚
// 2. 恢复目标版本的配置快照
// 3. 通知所有订阅者
func (s *ReleaseService) Rollback(ctx context.Context, req *RollbackRequest) error {
	// 1. 查询当前版本
	currentRelease, err := s.releaseRepo.GetByID(ctx, req.CurrentReleaseID)
	if err != nil {
		return fmt.Errorf("查询当前版本失败: %w", err)
	}
	if currentRelease == nil {
		return fmt.Errorf("当前版本不存在: id=%d", req.CurrentReleaseID)
	}

	// 2. 检查是否可以回滚
	if !currentRelease.CanRollback() {
		return fmt.Errorf("当前版本状态不允许回滚: status=%s", currentRelease.Status)
	}

	// 3. 查询目标版本
	targetRelease, err := s.releaseRepo.GetByID(ctx, req.TargetReleaseID)
	if err != nil {
		return fmt.Errorf("查询目标版本失败: %w", err)
	}
	if targetRelease == nil {
		return fmt.Errorf("目标版本不存在: id=%d", req.TargetReleaseID)
	}

	// 4. 验证版本一致性
	if currentRelease.NamespaceID != targetRelease.NamespaceID ||
		currentRelease.Environment != targetRelease.Environment {
		return fmt.Errorf("版本不在同一命名空间或环境")
	}

	// 5. 恢复目标版本的配置快照
	targetSnapshot, err := targetRelease.GetConfigSnapshot()
	if err != nil {
		return fmt.Errorf("获取目标版本快照失败: %w", err)
	}

	for _, item := range targetSnapshot {
		config, err := s.configRepo.GetByID(ctx, item.ConfigID)
		if err != nil {
			hlog.Errorf("查询配置失败: configID=%d, error=%v", item.ConfigID, err)
			continue
		}
		if config == nil {
			hlog.Warnf("配置不存在，跳过回滚: configID=%d", item.ConfigID)
			continue
		}

		// 更新配置值
		hash, _ := s.configSvc.ComputeContentHash(item.Value, item.ContentHashAlgorithm)
		config.UpdateValue(item.Value, hash)
		if err := s.configRepo.Update(ctx, config); err != nil {
			hlog.Errorf("更新配置失败: configID=%d, error=%v", item.ConfigID, err)
			continue
		}
	}

	// 6. 标记当前版本为已回滚
	currentRelease.Rollback(req.RollbackBy, req.Reason)
	currentRelease.RollbackFromVersion = targetRelease.Version
	if err := s.releaseRepo.Update(ctx, currentRelease); err != nil {
		return fmt.Errorf("更新当前版本状态失败: %w", err)
	}

	// 7. 发布配置变更事件
	for _, item := range targetSnapshot {
		s.publishConfigChangeEvent(ctx, &listener.ConfigChangeEvent{
			NamespaceID: currentRelease.NamespaceID,
			ConfigKey:   item.Key,
			ConfigID:    item.ConfigID,
			Action:      "rollback",
		})
	}

	hlog.Infof("回滚成功: 从版本%d回滚到版本%d, namespace=%d",
		currentRelease.Version, targetRelease.Version, currentRelease.NamespaceID)

	return nil
}

// ==================== 查询操作 ====================

// GetReleaseByID 根据ID查询发布版本
func (s *ReleaseService) GetReleaseByID(ctx context.Context, id int) (*entity.Release, error) {
	release, err := s.releaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, fmt.Errorf("发布版本不存在: id=%d", id)
	}
	return release, nil
}

// GetLatestPublishedRelease 获取最新已发布版本
func (s *ReleaseService) GetLatestPublishedRelease(ctx context.Context, namespaceID int, environment string) (*entity.Release, error) {
	return s.releaseRepo.FindLatestPublishedRelease(ctx, namespaceID, environment)
}

// ListReleasesByNamespace 查询命名空间下的所有发布版本
func (s *ReleaseService) ListReleasesByNamespace(ctx context.Context, namespaceID int, environment string) ([]*entity.Release, error) {
	return s.releaseRepo.FindByNamespace(ctx, namespaceID, environment)
}

// QueryReleases 分页查询发布版本
func (s *ReleaseService) QueryReleases(ctx context.Context, params *repository.ReleaseQueryParams) (*shareRepo.PageResult[*entity.Release], error) {
	return s.releaseRepo.QueryByParams(ctx, params)
}

// ==================== 灰度判断 ====================

// ShouldUseCanaryRelease 判断客户端是否应该使用灰度版本
func (s *ReleaseService) ShouldUseCanaryRelease(ctx context.Context, namespaceID int, environment string, clientID string, clientIP string) (*entity.Release, bool, error) {
	// 1. 查询最新发布版本
	release, err := s.releaseRepo.FindLatestPublishedRelease(ctx, namespaceID, environment)
	if err != nil {
		return nil, false, err
	}
	if release == nil {
		return nil, false, nil
	}

	// 2. 如果不是灰度发布,直接返回
	if !release.IsCanaryRelease() {
		return release, false, nil
	}

	// 3. 获取灰度规则
	rule, err := release.GetCanaryRule()
	if err != nil {
		hlog.Errorf("获取灰度规则失败: %v", err)
		return release, false, nil
	}

	// 4. 判断是否匹配灰度规则
	matched := s.canaryEngine.Match(rule, clientID, clientIP)
	return release, matched, nil
}

// ==================== 辅助方法 ====================

// publishConfigChangeEvent 发布配置变更事件
func (s *ReleaseService) publishConfigChangeEvent(ctx context.Context, event *listener.ConfigChangeEvent) {
	if s.listener == nil {
		return
	}

	go func() {
		if err := s.listener.Publish(context.Background(), event); err != nil {
			hlog.Errorf("发布配置变更事件失败: %v, event: %+v", err, event)
		}
	}()
}

// CompareReleases 对比两个版本的差异
func (s *ReleaseService) CompareReleases(ctx context.Context, fromReleaseID, toReleaseID int) (*ReleaseCompareResult, error) {
	// 1. 查询两个版本
	fromRelease, err := s.releaseRepo.GetByID(ctx, fromReleaseID)
	if err != nil {
		return nil, err
	}
	if fromRelease == nil {
		return nil, fmt.Errorf("源版本不存在: id=%d", fromReleaseID)
	}

	toRelease, err := s.releaseRepo.GetByID(ctx, toReleaseID)
	if err != nil {
		return nil, err
	}
	if toRelease == nil {
		return nil, fmt.Errorf("目标版本不存在: id=%d", toReleaseID)
	}

	// 2. 获取配置快照
	fromSnapshot, _ := fromRelease.GetConfigSnapshot()
	toSnapshot, _ := toRelease.GetConfigSnapshot()

	// 3. 构建配置映射
	fromMap := make(map[string]*entity.ConfigSnapshotItem)
	for i := range fromSnapshot {
		fromMap[fromSnapshot[i].Key] = &fromSnapshot[i]
	}

	toMap := make(map[string]*entity.ConfigSnapshotItem)
	for i := range toSnapshot {
		toMap[toSnapshot[i].Key] = &toSnapshot[i]
	}

	// 4. 对比差异
	result := &ReleaseCompareResult{
		FromReleaseID: fromReleaseID,
		ToReleaseID:   toReleaseID,
		FromVersion:   fromRelease.Version,
		ToVersion:     toRelease.Version,
	}

	// 新增的配置
	for key, item := range toMap {
		if _, exists := fromMap[key]; !exists {
			result.Added = append(result.Added, item)
		}
	}

	// 删除的配置
	for key, item := range fromMap {
		if _, exists := toMap[key]; !exists {
			result.Deleted = append(result.Deleted, item)
		}
	}

	// 修改的配置
	for key, fromItem := range fromMap {
		if toItem, exists := toMap[key]; exists {
			if fromItem.Value != toItem.Value {
				result.Modified = append(result.Modified, &ConfigDiff{
					Key:      key,
					OldValue: fromItem.Value,
					NewValue: toItem.Value,
				})
			}
		}
	}

	return result, nil
}

// ReleaseCompareResult 版本对比结果
type ReleaseCompareResult struct {
	FromReleaseID int
	ToReleaseID   int
	FromVersion   int
	ToVersion     int
	Added         []*entity.ConfigSnapshotItem
	Deleted       []*entity.ConfigSnapshotItem
	Modified      []*ConfigDiff
}

// ConfigDiff 配置差异
type ConfigDiff struct {
	Key      string
	OldValue string
	NewValue string
}
