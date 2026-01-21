package service

import (
	"context"

	"config-client/api/config-api/converter"
	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
	"config-client/config/domain/repository"
	domainService "config-client/config/domain/service"
)

// ReleaseAppService 发布管理应用服务
// 负责协调领域服务和数据转换
type ReleaseAppService struct {
	releaseDomainService *domainService.ReleaseService
	converter            *converter.ReleaseConverter
}

// NewReleaseAppService 创建发布管理应用服务
func NewReleaseAppService(
	releaseDomainService *domainService.ReleaseService,
	converter *converter.ReleaseConverter,
) *ReleaseAppService {
	return &ReleaseAppService{
		releaseDomainService: releaseDomainService,
		converter:            converter,
	}
}

// CreateRelease 创建发布版本
func (s *ReleaseAppService) CreateRelease(ctx context.Context, req *request.CreateReleaseRequest) (*vo.ReleaseVO, error) {
	// 1. 转换请求为领域服务请求
	domainReq := &domainService.CreateReleaseRequest{
		NamespaceID: req.NamespaceID,
		Environment: req.Environment,
		VersionName: req.VersionName,
		ReleaseType: entity.ReleaseType(req.ReleaseType),
		CreatedBy:   req.CreatedBy,
	}

	// 2. 调用领域服务创建发布版本
	release, err := s.releaseDomainService.CreateRelease(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// 3. 转换为VO返回（包含配置快照）
	return s.converter.ToVO(release, true), nil
}

// PublishFull 全量发布
func (s *ReleaseAppService) PublishFull(ctx context.Context, req *request.PublishFullRequest) error {
	// 转换请求并调用领域服务
	domainReq := &domainService.PublishRequest{
		ReleaseID:   req.ReleaseID,
		PublishedBy: req.PublishedBy,
	}

	return s.releaseDomainService.PublishFull(ctx, domainReq)
}

// PublishCanary 灰度发布
func (s *ReleaseAppService) PublishCanary(ctx context.Context, req *request.PublishCanaryRequest) error {
	// 构建灰度规则
	canaryRule := &entity.CanaryRule{
		ClientIDs:  req.ClientIDs,
		IPRanges:   req.IPRanges,
		Percentage: req.CanaryPercentage,
	}

	// 转换请求并调用领域服务
	domainReq := &domainService.PublishCanaryRequest{
		ReleaseID:   req.ReleaseID,
		CanaryRule:  canaryRule,
		PublishedBy: req.PublishedBy,
	}

	return s.releaseDomainService.PublishCanary(ctx, domainReq)
}

// Rollback 回滚到指定版本
func (s *ReleaseAppService) Rollback(ctx context.Context, req *request.ReleaseRollbackRequest) error {
	// 转换请求并调用领域服务
	domainReq := &domainService.RollbackRequest{
		CurrentReleaseID: req.CurrentReleaseID,
		TargetReleaseID:  req.TargetReleaseID,
		RollbackBy:       req.RollbackBy,
		Reason:           req.Reason,
	}

	return s.releaseDomainService.Rollback(ctx, domainReq)
}

// GetReleaseByID 根据ID查询发布版本
func (s *ReleaseAppService) GetReleaseByID(ctx context.Context, id int, includeSnapshot bool) (*vo.ReleaseVO, error) {
	release, err := s.releaseDomainService.GetReleaseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.converter.ToVO(release, includeSnapshot), nil
}

// GetLatestPublishedRelease 获取最新已发布版本
func (s *ReleaseAppService) GetLatestPublishedRelease(ctx context.Context, namespaceID int, environment string) (*vo.ReleaseVO, error) {
	release, err := s.releaseDomainService.GetLatestPublishedRelease(ctx, namespaceID, environment)
	if err != nil {
		return nil, err
	}

	if release == nil {
		return nil, nil
	}

	return s.converter.ToVO(release, false), nil
}

// ListReleasesByNamespace 查询命名空间下的所有发布版本
func (s *ReleaseAppService) ListReleasesByNamespace(ctx context.Context, namespaceID int, environment string) ([]*vo.ReleaseVO, error) {
	releases, err := s.releaseDomainService.ListReleasesByNamespace(ctx, namespaceID, environment)
	if err != nil {
		return nil, err
	}

	return s.converter.ToVOList(releases), nil
}

// QueryReleases 分页查询发布版本
func (s *ReleaseAppService) QueryReleases(ctx context.Context, req *request.QueryReleaseRequest) (*vo.ReleaseListVO, error) {
	// 1. 设置默认值
	req.SetDefaults()

	// 2. 构建领域查询参数
	params := &repository.ReleaseQueryParams{
		NamespaceID: req.NamespaceID,
		Environment: req.Environment,
		Status:      req.Status,
		ReleaseType: req.ReleaseType,
		VersionName: req.VersionName,
		Page:        req.Page,
		Size:        req.Size,
		OrderBy:     req.OrderBy,
	}

	// 3. 调用领域服务查询
	result, err := s.releaseDomainService.QueryReleases(ctx, params)
	if err != nil {
		return nil, err
	}

	// 4. 转换为VO返回
	return &vo.ReleaseListVO{
		Items: s.converter.ToVOList(result.Items),
		Total: result.Total,
		Page:  result.Page,
		Size:  result.Size,
	}, nil
}

// CompareReleases 对比两个版本
func (s *ReleaseAppService) CompareReleases(ctx context.Context, req *request.CompareReleasesRequest) (*vo.ReleaseCompareVO, error) {
	// 调用领域服务对比版本
	result, err := s.releaseDomainService.CompareReleases(ctx, req.FromReleaseID, req.ToReleaseID)
	if err != nil {
		return nil, err
	}

	// 转换为VO返回
	return s.converter.ToCompareVO(result), nil
}
