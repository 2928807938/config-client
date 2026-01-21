package converter

import (
	domainEntity "config-client/config/domain/entity"
	infraEntity "config-client/config/infrastructure/entity"
)

// SubscriptionConverter 订阅转换器
type SubscriptionConverter struct{}

// NewSubscriptionConverter 创建订阅转换器
func NewSubscriptionConverter() *SubscriptionConverter {
	return &SubscriptionConverter{}
}

// ToEntity 将PO转换为领域实体
func (c *SubscriptionConverter) ToEntity(po *infraEntity.SubscriptionPO) *domainEntity.Subscription {
	if po == nil {
		return nil
	}

	return &domainEntity.Subscription{
		ID:                 po.ID,
		NamespaceID:        po.NamespaceID,
		ClientID:           po.ClientID,
		ClientIP:           po.ClientIP,
		ClientHostname:     po.ClientHostname,
		Environment:        po.Environment,
		LastVersion:        po.LastVersion,
		ConfigSnapshotHash: po.ConfigSnapshotHash,
		IsActive:           po.IsActive,
		LastHeartbeatAt:    po.LastHeartbeatAt,
		HeartbeatCount:     po.HeartbeatCount,
		PollCount:          po.PollCount,
		ChangeCount:        po.ChangeCount,
		SubscribedAt:       po.SubscribedAt,
		UnsubscribedAt:     po.UnsubscribedAt,
		CreatedAt:          po.CreatedAt,
		UpdatedAt:          po.UpdatedAt,
	}
}

// ToPO 将领域实体转换为PO
func (c *SubscriptionConverter) ToPO(entity *domainEntity.Subscription) *infraEntity.SubscriptionPO {
	if entity == nil {
		return nil
	}

	return &infraEntity.SubscriptionPO{
		ID:                 entity.ID,
		NamespaceID:        entity.NamespaceID,
		ClientID:           entity.ClientID,
		ClientIP:           entity.ClientIP,
		ClientHostname:     entity.ClientHostname,
		Environment:        entity.Environment,
		LastVersion:        entity.LastVersion,
		ConfigSnapshotHash: entity.ConfigSnapshotHash,
		IsActive:           entity.IsActive,
		LastHeartbeatAt:    entity.LastHeartbeatAt,
		HeartbeatCount:     entity.HeartbeatCount,
		PollCount:          entity.PollCount,
		ChangeCount:        entity.ChangeCount,
		SubscribedAt:       entity.SubscribedAt,
		UnsubscribedAt:     entity.UnsubscribedAt,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
	}
}

// ToEntityList 将PO列表转换为领域实体列表
func (c *SubscriptionConverter) ToEntityList(pos []*infraEntity.SubscriptionPO) []*domainEntity.Subscription {
	if pos == nil {
		return nil
	}

	entities := make([]*domainEntity.Subscription, 0, len(pos))
	for _, po := range pos {
		entities = append(entities, c.ToEntity(po))
	}
	return entities
}

// ToPOList 将领域实体列表转换为PO列表
func (c *SubscriptionConverter) ToPOList(entities []*domainEntity.Subscription) []*infraEntity.SubscriptionPO {
	if entities == nil {
		return nil
	}

	pos := make([]*infraEntity.SubscriptionPO, 0, len(entities))
	for _, entity := range entities {
		pos = append(pos, c.ToPO(entity))
	}
	return pos
}
