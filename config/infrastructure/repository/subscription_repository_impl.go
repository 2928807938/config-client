package repository

import (
	"context"
	"time"

	"config-client/config/domain/entity"
	"config-client/config/domain/repository"
	"config-client/config/infrastructure/converter"
	infraEntity "config-client/config/infrastructure/entity"
	"config-client/share/repository/queryutil"

	"gorm.io/gorm"
)

// SubscriptionRepositoryImpl 订阅仓储实现
type SubscriptionRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.SubscriptionConverter
	fields    *queryutil.EntityFields[infraEntity.SubscriptionPO] // Lambda 字段查询构建器
	model     infraEntity.SubscriptionPO                          // 用于类型安全的字段引用
}

// NewSubscriptionRepository 创建订阅仓储实例
func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &SubscriptionRepositoryImpl{
		db:        db,
		converter: converter.NewSubscriptionConverter(),
		fields:    queryutil.Lambda[infraEntity.SubscriptionPO](), // 初始化 Lambda 构建器
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
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.ClientID).GetColumnName(), clientID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	err := db.First(&po).Error

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
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.IsActive).GetColumnName(), true)
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToEntityList(pos), nil
}

// FindAllActiveSubscriptions 查询所有活跃订阅
func (r *SubscriptionRepositoryImpl) FindAllActiveSubscriptions(ctx context.Context) ([]*entity.Subscription, error) {
	var pos []*infraEntity.SubscriptionPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.IsActive).GetColumnName(), true)
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToEntityList(pos), nil
}

// UpdateHeartbeat 更新心跳时间
func (r *SubscriptionRepositoryImpl) UpdateHeartbeat(ctx context.Context, id int) error {
	now := time.Now()
	db := r.db.WithContext(ctx).Model(&infraEntity.SubscriptionPO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.ID).GetColumnName(), id)
	return db.Updates(map[string]interface{}{
		r.fields.Of(&r.model.LastHeartbeatAt).GetColumnName(): now,
		r.fields.Of(&r.model.HeartbeatCount).GetColumnName():  gorm.Expr(r.fields.Of(&r.model.HeartbeatCount).GetColumnName() + " + 1"),
		r.fields.Of(&r.model.UpdatedAt).GetColumnName():       now,
	}).Error
}

// IncrementPollCount 增加轮询计数
func (r *SubscriptionRepositoryImpl) IncrementPollCount(ctx context.Context, id int) error {
	db := r.db.WithContext(ctx).Model(&infraEntity.SubscriptionPO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.ID).GetColumnName(), id)
	return db.Updates(map[string]interface{}{
		r.fields.Of(&r.model.PollCount).GetColumnName(): gorm.Expr(r.fields.Of(&r.model.PollCount).GetColumnName() + " + 1"),
		r.fields.Of(&r.model.UpdatedAt).GetColumnName(): time.Now(),
	}).Error
}

// IncrementChangeCount 增加变更计数
func (r *SubscriptionRepositoryImpl) IncrementChangeCount(ctx context.Context, id int) error {
	db := r.db.WithContext(ctx).Model(&infraEntity.SubscriptionPO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.ID).GetColumnName(), id)
	return db.Updates(map[string]interface{}{
		r.fields.Of(&r.model.ChangeCount).GetColumnName(): gorm.Expr(r.fields.Of(&r.model.ChangeCount).GetColumnName() + " + 1"),
		r.fields.Of(&r.model.UpdatedAt).GetColumnName():   time.Now(),
	}).Error
}

// Deactivate 停用订阅
func (r *SubscriptionRepositoryImpl) Deactivate(ctx context.Context, id int) error {
	now := time.Now()
	db := r.db.WithContext(ctx).Model(&infraEntity.SubscriptionPO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.ID).GetColumnName(), id)
	return db.Updates(map[string]interface{}{
		r.fields.Of(&r.model.IsActive).GetColumnName():       false,
		r.fields.Of(&r.model.UnsubscribedAt).GetColumnName(): now,
		r.fields.Of(&r.model.UpdatedAt).GetColumnName():      now,
	}).Error
}

// CleanExpiredSubscriptions 清理过期订阅
func (r *SubscriptionRepositoryImpl) CleanExpiredSubscriptions(ctx context.Context, expireTime time.Time) (int64, error) {
	db := r.db.WithContext(ctx).Model(&infraEntity.SubscriptionPO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.IsActive).GetColumnName(), true)
	db = queryutil.WhereLt(db, r.fields.Of(&r.model.LastHeartbeatAt).GetColumnName(), expireTime)
	result := db.Updates(map[string]interface{}{
		r.fields.Of(&r.model.IsActive).GetColumnName():  false,
		r.fields.Of(&r.model.UpdatedAt).GetColumnName(): time.Now(),
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
