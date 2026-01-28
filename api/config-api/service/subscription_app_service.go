package service

import (
	"context"
	"time"

	"config-client/api/config-api/converter"
	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/repository"
	domainService "config-client/config/domain/service"
	"config-client/share/errors"
)

// SubscriptionAppService 订阅应用服务
// 负责订阅管理的查询与协调，不包含业务逻辑
type SubscriptionAppService struct {
	subscriptionRepo repository.SubscriptionRepository
	subscriptionMgr  *domainService.SubscriptionManager
	systemConfigSvc  *domainService.SystemConfigService
	converter        *converter.SubscriptionConverter
}

// NewSubscriptionAppService 创建订阅应用服务实例
func NewSubscriptionAppService(
	subscriptionRepo repository.SubscriptionRepository,
	subscriptionMgr *domainService.SubscriptionManager,
	systemConfigSvc *domainService.SystemConfigService,
	converter *converter.SubscriptionConverter,
) *SubscriptionAppService {
	return &SubscriptionAppService{
		subscriptionRepo: subscriptionRepo,
		subscriptionMgr:  subscriptionMgr,
		systemConfigSvc:  systemConfigSvc,
		converter:        converter,
	}
}

// QuerySubscriptions 分页查询订阅
func (s *SubscriptionAppService) QuerySubscriptions(ctx context.Context, req *request.QuerySubscriptionRequest) (*vo.SubscriptionListVO, error) {
	req.SetDefaults()

	params := &repository.SubscriptionQueryParams{
		NamespaceID: req.NamespaceID,
		Environment: req.Environment,
		ClientID:    req.ClientID,
		IsActive:    req.IsActive,
		Page:        req.Page,
		Size:        req.Size,
		OrderBy:     req.OrderBy,
	}

	pageResult, err := s.subscriptionRepo.Query(ctx, params)
	if err != nil {
		return nil, err
	}

	return &vo.SubscriptionListVO{
		Total:         pageResult.Total,
		Page:          pageResult.Page,
		PageSize:      pageResult.Size,
		Subscriptions: s.converter.ToVOList(pageResult.Items),
	}, nil
}

// DeactivateSubscription 停用订阅
func (s *SubscriptionAppService) DeactivateSubscription(ctx context.Context, id int) (*vo.SubscriptionVO, error) {
	subscription, err := s.subscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, errors.ErrNotFound("订阅不存在")
	}

	if err := s.subscriptionRepo.Deactivate(ctx, id); err != nil {
		return nil, err
	}

	if s.subscriptionMgr != nil {
		_ = s.subscriptionMgr.Unsubscribe(subscription.ClientID, subscription.NamespaceID, subscription.Environment)
	}

	updated, err := s.subscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.converter.ToVO(updated), nil
}

// GetStatistics 获取订阅统计信息
func (s *SubscriptionAppService) GetStatistics(ctx context.Context) (*vo.SubscriptionStatisticsVO, error) {
	total, err := s.subscriptionRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	active, err := s.subscriptionRepo.CountByActive(ctx, true)
	if err != nil {
		return nil, err
	}

	heartbeatTimeout := domainService.DefaultHeartbeatTimeout
	if s.systemConfigSvc != nil {
		heartbeatTimeout = s.systemConfigSvc.GetHeartbeatTimeout()
	}

	expireTime := time.Now().Add(-time.Duration(heartbeatTimeout) * time.Second)
	expired, err := s.subscriptionRepo.CountExpired(ctx, expireTime)
	if err != nil {
		return nil, err
	}

	activeInMemory := 0
	if s.subscriptionMgr != nil {
		activeInMemory = s.subscriptionMgr.GetActiveSubscriberCount()
	}

	return &vo.SubscriptionStatisticsVO{
		Total:                   total,
		Active:                  active,
		Inactive:                total - active,
		Expired:                 expired,
		ActiveInMemory:          activeInMemory,
		HeartbeatTimeoutSeconds: heartbeatTimeout,
	}, nil
}
