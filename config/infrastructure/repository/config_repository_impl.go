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
	gormRepo "config-client/share/repository/gorm"
	"config-client/share/repository/queryutil"
)

// ConfigRepositoryImpl 配置仓储实现
type ConfigRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.ConfigConverter
	fields    *queryutil.EntityFields[infraEntity.ConfigPO] // Lambda 字段查询构建器
	model     infraEntity.ConfigPO                          // 用于类型安全的字段引用
}

// NewConfigRepository 创建配置仓储实例
func NewConfigRepository(db *gorm.DB) repository.ConfigRepository {
	return &ConfigRepositoryImpl{
		db:        db,
		converter: converter.NewConfigConverter(),
		fields:    queryutil.Lambda[infraEntity.ConfigPO](), // 初始化 Lambda 构建器
	}
}

// ==================== 基础 CRUD 实现 ====================

// Create 创建配置
func (r *ConfigRepositoryImpl) Create(ctx context.Context, entity *domainEntity.Config) error {
	po := r.converter.ToPO(entity)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	// 更新实体ID
	entity.ID = po.ID
	return nil
}

// CreateBatch 批量创建配置
func (r *ConfigRepositoryImpl) CreateBatch(ctx context.Context, entities []*domainEntity.Config) error {
	if len(entities) == 0 {
		return nil
	}
	pos := r.converter.ToPOList(entities)
	return r.db.WithContext(ctx).Create(pos).Error
}

// GetByID 根据ID查询配置
func (r *ConfigRepositoryImpl) GetByID(ctx context.Context, id int) (*domainEntity.Config, error) {
	var po infraEntity.ConfigPO
	err := r.db.WithContext(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToDO(&po), nil
}

// Update 更新配置
func (r *ConfigRepositoryImpl) Update(ctx context.Context, entity *domainEntity.Config) error {
	po := r.converter.ToPO(entity)
	return r.db.WithContext(ctx).Save(po).Error
}

// Delete 删除配置（软删除）
func (r *ConfigRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&infraEntity.ConfigPO{}, id).Error
}

// List 查询全部列表
func (r *ConfigRepositoryImpl) List(ctx context.Context) ([]*domainEntity.Config, error) {
	var pos []*infraEntity.ConfigPO
	err := r.db.WithContext(ctx).Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return r.converter.ToDOList(pos), nil
}

// Page 分页查询
func (r *ConfigRepositoryImpl) Page(ctx context.Context, request *shareRepo.PageRequest) (*shareRepo.PageResult[*domainEntity.Config], error) {
	db := r.db.WithContext(ctx)

	// 统计总数
	var total int64
	var po infraEntity.ConfigPO
	if err := db.Model(&po).Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用排序
	db = gormRepo.ApplyOrderBys(db, request.OrderBy)

	// 应用分页
	db = db.Offset(request.Offset()).Limit(request.Size)

	// 查询数据
	var pos []*infraEntity.ConfigPO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, request.Page, request.Size), nil
}

// ==================== 自定义查询方法实现 ====================

// FindByNamespaceAndKey 根据命名空间ID、配置键和环境查询配置
func (r *ConfigRepositoryImpl) FindByNamespaceAndKey(ctx context.Context, namespaceID int, key string, environment string) (*domainEntity.Config, error) {
	var po infraEntity.ConfigPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Key).GetColumnName(), key)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	err := db.First(&po).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.converter.ToDO(&po), nil
}

// FindByNamespace 根据命名空间ID查询该命名空间下的所有配置
func (r *ConfigRepositoryImpl) FindByNamespace(ctx context.Context, namespaceID int) ([]*domainEntity.Config, error) {
	var pos []*infraEntity.ConfigPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.OrderBy(db, r.fields.Of(&r.model.GroupName).GetColumnName())
	db = queryutil.OrderBy(db, r.fields.Of(&r.model.Key).GetColumnName())
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDOList(pos), nil
}

// FindByEnvironment 根据环境查询配置列表
func (r *ConfigRepositoryImpl) FindByEnvironment(ctx context.Context, environment string) ([]*domainEntity.Config, error) {
	var pos []*infraEntity.ConfigPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	db = queryutil.OrderBy(db, r.fields.Of(&r.model.NamespaceID).GetColumnName())
	db = queryutil.OrderBy(db, r.fields.Of(&r.model.GroupName).GetColumnName())
	db = queryutil.OrderBy(db, r.fields.Of(&r.model.Key).GetColumnName())
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDOList(pos), nil
}

// FindByGroup 根据分组查询配置列表
func (r *ConfigRepositoryImpl) FindByGroup(ctx context.Context, namespaceID int, groupName string) ([]*domainEntity.Config, error) {
	var pos []*infraEntity.ConfigPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.GroupName).GetColumnName(), groupName)
	db = queryutil.OrderBy(db, r.fields.Of(&r.model.Key).GetColumnName())
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDOList(pos), nil
}

// FindReleasedConfigs 查询已发布的配置列表
func (r *ConfigRepositoryImpl) FindReleasedConfigs(ctx context.Context, namespaceID int, environment string) ([]*domainEntity.Config, error) {
	var pos []*infraEntity.ConfigPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.IsReleased).GetColumnName(), true)
	db = queryutil.OrderBy(db, r.fields.Of(&r.model.GroupName).GetColumnName())
	db = queryutil.OrderBy(db, r.fields.Of(&r.model.Key).GetColumnName())
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDOList(pos), nil
}

// PageWithConditions 根据多条件分页查询配置
func (r *ConfigRepositoryImpl) PageWithConditions(ctx context.Context, req *shareRepo.PageRequest, conditions ...*shareRepo.Condition) (*shareRepo.PageResult[*domainEntity.Config], error) {
	db := r.db.WithContext(ctx)

	// 应用查询条件
	if len(conditions) > 0 {
		for _, cond := range conditions {
			db = gormRepo.ApplyCondition(db, cond)
		}
	}

	// 统计总数
	var total int64
	var po infraEntity.ConfigPO
	if err := db.Model(&po).Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用排序
	db = gormRepo.ApplyOrderBys(db, req.OrderBy)

	// 应用分页
	db = db.Offset(req.Offset()).Limit(req.Size)

	// 查询数据
	var pos []*infraEntity.ConfigPO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, req.Page, req.Size), nil
}

// ExistsByNamespaceAndKey 判断指定命名空间和键的配置是否存在
func (r *ConfigRepositoryImpl) ExistsByNamespaceAndKey(ctx context.Context, namespaceID int, key string, environment string) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.ConfigPO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Key).GetColumnName(), key)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	err := db.Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CountByNamespace 统计指定命名空间的配置数量
func (r *ConfigRepositoryImpl) CountByNamespace(ctx context.Context, namespaceID int) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.ConfigPO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	err := db.Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// QueryByParams 根据查询参数分页查询配置
// 封装了查询条件的构建逻辑和字段映射
func (r *ConfigRepositoryImpl) QueryByParams(ctx context.Context, params *repository.ConfigQueryParams) (*shareRepo.PageResult[*domainEntity.Config], error) {
	db := r.db.WithContext(ctx).Model(&infraEntity.ConfigPO{})

	// 构建查询条件(字段映射在这里处理)
	if params.NamespaceID != nil {
		db = queryutil.WhereEq(db, r.fields.Get("NamespaceID").GetColumnName(), *params.NamespaceID)
	}

	if params.Key != nil && *params.Key != "" {
		// 支持模糊查询
		db = queryutil.WhereLike(db, r.fields.Get("Key").GetColumnName(), "%"+*params.Key+"%")
	}

	if params.GroupName != nil && *params.GroupName != "" {
		db = queryutil.WhereEq(db, r.fields.Get("GroupName").GetColumnName(), *params.GroupName)
	}

	if params.Environment != nil && *params.Environment != "" {
		db = queryutil.WhereEq(db, r.fields.Get("Environment").GetColumnName(), *params.Environment)
	}

	if params.IsActive != nil {
		db = queryutil.WhereEq(db, r.fields.Get("IsActive").GetColumnName(), *params.IsActive)
	}

	if params.IsReleased != nil {
		db = queryutil.WhereEq(db, r.fields.Get("IsReleased").GetColumnName(), *params.IsReleased)
	}

	if params.ValueType != nil && *params.ValueType != "" {
		db = queryutil.WhereEq(db, r.fields.Get("ValueType").GetColumnName(), *params.ValueType)
	}

	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 处理排序
	if params.OrderBy != "" {
		// 这里需要将排序字段名转换为数据库字段名
		// 简化处理，直接使用传入的字段名
		db = db.Order(params.OrderBy)
	} else {
		// 默认按创建时间倒序排列
		db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())
	}

	// 应用分页
	offset := (params.Page - 1) * params.Size
	db = db.Offset(offset).Limit(params.Size)

	// 查询数据
	var pos []*infraEntity.ConfigPO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, params.Page, params.Size), nil
}

// 确保实现了接口
var _ repository.ConfigRepository = (*ConfigRepositoryImpl)(nil)
