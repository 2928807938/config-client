package converter

import (
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
)

// SubscriptionConverter 订阅转换器
type SubscriptionConverter struct{}

// NewSubscriptionConverter 创建订阅转换器
func NewSubscriptionConverter() *SubscriptionConverter {
	return &SubscriptionConverter{}
}

// ToVO 将领域实体转换为VO
func (c *SubscriptionConverter) ToVO(subscription *entity.Subscription) *vo.SubscriptionVO {
	if subscription == nil {
		return nil
	}

	return &vo.SubscriptionVO{
		ID:                 subscription.ID,
		NamespaceID:        subscription.NamespaceID,
		ClientID:           subscription.ClientID,
		ClientIP:           subscription.ClientIP,
		ClientHostname:     subscription.ClientHostname,
		Environment:        subscription.Environment,
		LastVersion:        subscription.LastVersion,
		ConfigSnapshotHash: subscription.ConfigSnapshotHash,
		IsActive:           subscription.IsActive,
		LastHeartbeatAt:    subscription.LastHeartbeatAt,
		HeartbeatCount:     subscription.HeartbeatCount,
		PollCount:          subscription.PollCount,
		ChangeCount:        subscription.ChangeCount,
		SubscribedAt:       subscription.SubscribedAt,
		UnsubscribedAt:     subscription.UnsubscribedAt,
		CreatedAt:          subscription.CreatedAt,
		UpdatedAt:          subscription.UpdatedAt,
	}
}

// ToVOList 将领域实体列表转换为VO列表
func (c *SubscriptionConverter) ToVOList(subscriptions []*entity.Subscription) []*vo.SubscriptionVO {
	if subscriptions == nil {
		return nil
	}

	result := make([]*vo.SubscriptionVO, 0, len(subscriptions))
	for _, item := range subscriptions {
		result = append(result, c.ToVO(item))
	}
	return result
}
