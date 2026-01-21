package repository

import (
	"context"
	"time"

	"config-client/config/domain/entity"
	"config-client/config/domain/repository"
	"config-client/config/infrastructure/converter"
	infraEntity "config-client/config/infrastructure/entity"

	"gorm.io/gorm"
)

// SubscriptionRepositoryImpl 订阅仓储实现
type SubscriptionRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.SubscriptionConverter
}

// NewSubscriptionRepository 创建订阅仓储实例
func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &SubscriptionRepositoryImpl{
		db:        db,
		converter: converter.NewSubscriptionConverter(),
	}
}

// Create 创建订阅
func (r *SubscriptionRepositoryImpl) Create(ctx context.Context, subscription *entity.Subscription) error {
	po := r.converter.ToPO(subscription)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	subscription.ID = po.ID
	return nil
}

// Update 更新订阅
func (r *SubscriptionRepositoryImpl) Update(ctx context.Context, subscription *entity.Subscription) error {
	po := r.converter.ToPO(subscription)
	return r.db.WithContext(ctx).Save(po).Error
}

// GetByID 根据ID获取订阅
func (r *SubscriptionRepositoryImpl) GetByID(ctx context.Context, id int) (*entity.Subscription, error) {
	var po infraEntity.SubscriptionPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToEntity(&po), nil
}

// GetByClientAndNamespace 根据客户端ID和命名空间获取订阅
func (r *SubscriptionRepositoryImpl) GetByClientAndNamespace(
	ctx context.Context,
	clientID string,
	namespaceID int,
	environment string,
) (*entity.Subscription, error) {
	var po infraEntity.SubscriptionPO
	err := r.db.WithContext(ctx).
		Where("client_id = ? AND namespace_id = ? AND environment = ?", clientID, namespaceID, environment).
		First(&po).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.converter.ToEntity(&po), nil
}

// FindActiveSubscriptions 查询活跃订阅
func (r *SubscriptionRepositoryImpl) FindActiveSubscriptions(
	ctx context.Context,
	namespaceID int,
	environment string,
) ([]*entity.Subscription, error) {
	var pos []*infraEntity.SubscriptionPO
	err := r.db.WithContext(ctx).
		Where("namespace_id = ? AND environment = ? AND is_active = ?", namespaceID, environment, true).
		Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToEntityList(pos), nil
}

// FindAllActiveSubscriptions 查询所有活跃订阅
func (r *SubscriptionRepositoryImpl) FindAllActiveSubscriptions(ctx context.Context) ([]*entity.Subscription, error) {
	var pos []*infraEntity.SubscriptionPO
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToEntityList(pos), nil
}

// UpdateHeartbeat 更新心跳时间
func (r *SubscriptionRepositoryImpl) UpdateHeartbeat(ctx context.Context, id int) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&infraEntity.SubscriptionPO{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_heartbeat_at": now,
			"heartbeat_count":   gorm.Expr("heartbeat_count + 1"),
			"updated_at":        now,
		}).Error
}

// IncrementPollCount 增加轮询计数
func (r *SubscriptionRepositoryImpl) IncrementPollCount(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).
		Model(&infraEntity.SubscriptionPO{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"poll_count": gorm.Expr("poll_count + 1"),
			"updated_at": time.Now(),
		}).Error
}

// IncrementChangeCount 增加变更计数
func (r *SubscriptionRepositoryImpl) IncrementChangeCount(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).
		Model(&infraEntity.SubscriptionPO{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"change_count": gorm.Expr("change_count + 1"),
			"updated_at":   time.Now(),
		}).Error
}

// Deactivate 停用订阅
func (r *SubscriptionRepositoryImpl) Deactivate(ctx context.Context, id int) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&infraEntity.SubscriptionPO{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":       false,
			"unsubscribed_at": now,
			"updated_at":      now,
		}).Error
}

// CleanExpiredSubscriptions 清理过期订阅
func (r *SubscriptionRepositoryImpl) CleanExpiredSubscriptions(ctx context.Context, expireTime time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&infraEntity.SubscriptionPO{}).
		Where("is_active = ? AND last_heartbeat_at < ?", true, expireTime).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// Delete 删除订阅 (物理删除)
func (r *SubscriptionRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&infraEntity.SubscriptionPO{}, id).Error
}
