package repository

import (
	"context"
	"time"

	"config-client/config/domain/entity"
	"config-client/share/repository"
)

// SubscriptionQueryParams 订阅查询参数
type SubscriptionQueryParams struct {
	NamespaceID *int
	Environment *string
	ClientID    *string
	IsActive    *bool
	Page        int
	Size        int
	OrderBy     string
}

// SubscriptionRepository 订阅仓储接口
type SubscriptionRepository interface {
	// Create 创建订阅
	Create(ctx context.Context, subscription *entity.Subscription) error

	// Update 更新订阅
	Update(ctx context.Context, subscription *entity.Subscription) error

	// GetByID 根据ID获取订阅
	GetByID(ctx context.Context, id int) (*entity.Subscription, error)

	// GetByClientAndNamespace 根据客户端ID和命名空间获取订阅
	// clientID: 客户端唯一标识
	// namespaceID: 命名空间ID
	// environment: 环境
	// 返回: 订阅实体, 错误
	GetByClientAndNamespace(ctx context.Context, clientID string, namespaceID int, environment string) (*entity.Subscription, error)

	// FindActiveSubscriptions 查询活跃订阅
	// namespaceID: 命名空间ID
	// environment: 环境
	// 返回: 订阅列表, 错误
	FindActiveSubscriptions(ctx context.Context, namespaceID int, environment string) ([]*entity.Subscription, error)

	// FindAllActiveSubscriptions 查询所有活跃订阅
	FindAllActiveSubscriptions(ctx context.Context) ([]*entity.Subscription, error)

	// Query 根据查询参数分页查询订阅
	Query(ctx context.Context, params *SubscriptionQueryParams) (*repository.PageResult[*entity.Subscription], error)

	// CountAll 统计订阅总数
	CountAll(ctx context.Context) (int64, error)

	// CountByActive 按是否激活统计订阅数量
	CountByActive(ctx context.Context, isActive bool) (int64, error)

	// CountExpired 统计过期订阅数量（仅统计激活状态）
	CountExpired(ctx context.Context, expireTime time.Time) (int64, error)

	// UpdateHeartbeat 更新心跳时间
	// id: 订阅ID
	UpdateHeartbeat(ctx context.Context, id int) error

	// IncrementPollCount 增加轮询计数
	// id: 订阅ID
	IncrementPollCount(ctx context.Context, id int) error

	// IncrementChangeCount 增加变更计数
	// id: 订阅ID
	IncrementChangeCount(ctx context.Context, id int) error

	// Deactivate 停用订阅
	// id: 订阅ID
	Deactivate(ctx context.Context, id int) error

	// CleanExpiredSubscriptions 清理过期订阅
	// expireTime: 过期时间点 (在此时间之前的心跳视为过期)
	// 返回: 清理数量, 错误
	CleanExpiredSubscriptions(ctx context.Context, expireTime time.Time) (int64, error)

	// Delete 删除订阅 (物理删除)
	// id: 订阅ID
	Delete(ctx context.Context, id int) error
}
