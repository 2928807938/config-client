package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	domainEntity "config-client/config/domain/entity"
	"config-client/config/domain/repository"
	"config-client/config/infrastructure/converter"
	infraEntity "config-client/config/infrastructure/entity"
	shareRepo "config-client/share/repository"
)

// SystemConfigRepositoryImpl 系统配置仓储实现
type SystemConfigRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.SystemConfigConverter
}

// NewSystemConfigRepository 创建系统配置仓储实例
func NewSystemConfigRepository(db *gorm.DB) repository.SystemConfigRepository {
	return &SystemConfigRepositoryImpl{
		db:        db,
		converter: converter.NewSystemConfigConverter(),
	}
}

// ==================== 基础 CRUD 实现 ====================

// Create 创建系统配置
func (r *SystemConfigRepositoryImpl) Create(ctx context.Context, entity *domainEntity.SystemConfig) error {
	po := r.converter.ToPO(entity)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	// 更新实体ID
	entity.ID = po.ID
	return nil
}

// CreateBatch 批量创建系统配置
func (r *SystemConfigRepositoryImpl) CreateBatch(ctx context.Context, entities []*domainEntity.SystemConfig) error {
	if len(entities) == 0 {
		return nil
	}
	pos := r.converter.ToPOList(entities)
	return r.db.WithContext(ctx).Create(pos).Error
}

// GetByID 根据ID查询系统配置
func (r *SystemConfigRepositoryImpl) GetByID(ctx context.Context, id int) (*domainEntity.SystemConfig, error) {
	var po infraEntity.SystemConfigPO
	err := r.db.WithContext(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToDO(&po), nil
}

// Update 更新系统配置
func (r *SystemConfigRepositoryImpl) Update(ctx context.Context, entity *domainEntity.SystemConfig) error {
	po := r.converter.ToPO(entity)
	return r.db.WithContext(ctx).Save(po).Error
}

// Delete 删除系统配置（软删除）
func (r *SystemConfigRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&infraEntity.SystemConfigPO{}, id).Error
}

// List 查询全部列表
func (r *SystemConfigRepositoryImpl) List(ctx context.Context) ([]*domainEntity.SystemConfig, error) {
	var pos []*infraEntity.SystemConfigPO
	err := r.db.WithContext(ctx).Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return r.converter.ToDOList(pos), nil
}

// Page 分页查询
func (r *SystemConfigRepositoryImpl) Page(ctx context.Context, request *shareRepo.PageRequest) (*shareRepo.PageResult[*domainEntity.SystemConfig], error) {
	db := r.db.WithContext(ctx)

	// 统计总数
	var total int64
	var po infraEntity.SystemConfigPO
	if err := db.Model(&po).Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var pos []*infraEntity.SystemConfigPO
	offset := (request.Page - 1) * request.Size
	err := db.Offset(offset).Limit(request.Size).
		Order(request.OrderBy).
		Find(&pos).Error
	if err != nil {
		return nil, err
	}

	// 转换为领域实体
	dos := r.converter.ToDOList(pos)

	return &shareRepo.PageResult[*domainEntity.SystemConfig]{
		Items: dos,
		Total: total,
		Page:  request.Page,
		Size:  request.Size,
	}, nil
}

// ==================== 自定义查询方法 ====================

// FindByKey 根据配置键查询系统配置
func (r *SystemConfigRepositoryImpl) FindByKey(ctx context.Context, key string) (*domainEntity.SystemConfig, error) {
	var po infraEntity.SystemConfigPO
	err := r.db.WithContext(ctx).
		Where("config_key = ?", key).
		First(&po).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.converter.ToDO(&po), nil
}

// FindAllActive 查询所有启用的系统配置
func (r *SystemConfigRepositoryImpl) FindAllActive(ctx context.Context) ([]*domainEntity.SystemConfig, error) {
	var pos []*infraEntity.SystemConfigPO
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDOList(pos), nil
}

// FindAll 查询所有系统配置（包括禁用的）
func (r *SystemConfigRepositoryImpl) FindAll(ctx context.Context) ([]*domainEntity.SystemConfig, error) {
	return r.List(ctx)
}

// ExistsByKey 判断指定配置键是否存在
func (r *SystemConfigRepositoryImpl) ExistsByKey(ctx context.Context, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&infraEntity.SystemConfigPO{}).
		Where("config_key = ?", key).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// UpdateValue 更新配置值
func (r *SystemConfigRepositoryImpl) UpdateValue(ctx context.Context, key string, value string) error {
	return r.db.WithContext(ctx).
		Model(&infraEntity.SystemConfigPO{}).
		Where("config_key = ?", key).
		Updates(map[string]interface{}{
			"config_value": value,
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

// UpdateActive 更新配置的启用状态
func (r *SystemConfigRepositoryImpl) UpdateActive(ctx context.Context, key string, isActive bool) error {
	return r.db.WithContext(ctx).
		Model(&infraEntity.SystemConfigPO{}).
		Where("config_key = ?", key).
		Updates(map[string]interface{}{
			"is_active":  isActive,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}
