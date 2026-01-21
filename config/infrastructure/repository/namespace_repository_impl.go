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

// NamespaceRepositoryImpl 命名空间仓储实现
type NamespaceRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.NamespaceConverter
	fields    *queryutil.EntityFields[infraEntity.NamespacePO] // Lambda 字段查询构建器
	model     infraEntity.NamespacePO                          // 用于类型安全的字段引用
}

// NewNamespaceRepository 创建命名空间仓储实例
func NewNamespaceRepository(db *gorm.DB) repository.NamespaceRepository {
	return &NamespaceRepositoryImpl{
		db:        db,
		converter: converter.NewNamespaceConverter(),
		fields:    queryutil.Lambda[infraEntity.NamespacePO](), // 初始化 Lambda 构建器
	}
}

// ==================== 基础 CRUD 实现 ====================

// Create 创建命名空间
func (r *NamespaceRepositoryImpl) Create(ctx context.Context, entity *domainEntity.Namespace) error {
	po := r.converter.ToPO(entity)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	// 更新实体ID
	entity.ID = po.ID
	return nil
}

// CreateBatch 批量创建命名空间
func (r *NamespaceRepositoryImpl) CreateBatch(ctx context.Context, entities []*domainEntity.Namespace) error {
	if len(entities) == 0 {
		return nil
	}
	pos := r.converter.ToPOList(entities)
	return r.db.WithContext(ctx).Create(pos).Error
}

// GetByID 根据ID查询命名空间
func (r *NamespaceRepositoryImpl) GetByID(ctx context.Context, id int) (*domainEntity.Namespace, error) {
	var po infraEntity.NamespacePO
	err := r.db.WithContext(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToDO(&po), nil
}

// Update 更新命名空间
func (r *NamespaceRepositoryImpl) Update(ctx context.Context, entity *domainEntity.Namespace) error {
	po := r.converter.ToPO(entity)
	return r.db.WithContext(ctx).Save(po).Error
}

// Delete 删除命名空间（软删除）
func (r *NamespaceRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&infraEntity.NamespacePO{}, id).Error
}

// List 查询全部列表
func (r *NamespaceRepositoryImpl) List(ctx context.Context) ([]*domainEntity.Namespace, error) {
	var pos []*infraEntity.NamespacePO
	db := r.db.WithContext(ctx)
	db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())
	err := db.Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return r.converter.ToDOList(pos), nil
}

// Page 分页查询
func (r *NamespaceRepositoryImpl) Page(ctx context.Context, request *shareRepo.PageRequest) (*shareRepo.PageResult[*domainEntity.Namespace], error) {
	db := r.db.WithContext(ctx)

	// 统计总数
	var total int64
	var po infraEntity.NamespacePO
	if err := db.Model(&po).Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用排序
	db = gormRepo.ApplyOrderBys(db, request.OrderBy)

	// 应用分页
	db = db.Offset(request.Offset()).Limit(request.Size)

	// 查询数据
	var pos []*infraEntity.NamespacePO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, request.Page, request.Size), nil
}

// ==================== 自定义查询方法实现 ====================

// Query 根据查询参数查询命名空间（分页）
func (r *NamespaceRepositoryImpl) Query(ctx context.Context, params *repository.NamespaceQueryParams) (*shareRepo.PageResult[*domainEntity.Namespace], error) {
	db := r.db.WithContext(ctx)

	// 应用查询条件
	if params.Name != nil && *params.Name != "" {
		db = queryutil.WhereLike(db, r.fields.Get("Name").GetColumnName(), *params.Name)
	}
	if params.IsActive != nil {
		db = queryutil.WhereEq(db, r.fields.Get("IsActive").GetColumnName(), *params.IsActive)
	}

	// 统计总数
	var total int64
	if err := db.Model(&infraEntity.NamespacePO{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用排序
	if params.OrderBy != "" {
		db = db.Order(params.OrderBy)
	} else {
		// 默认按创建时间降序
		db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())
	}

	// 应用分页
	page := params.Page
	size := params.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size
	db = db.Offset(offset).Limit(size)

	// 查询数据
	var pos []*infraEntity.NamespacePO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, page, size), nil
}

// FindByName 根据名称查询命名空间
func (r *NamespaceRepositoryImpl) FindByName(ctx context.Context, name string) (*domainEntity.Namespace, error) {
	var po infraEntity.NamespacePO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("Name").GetColumnName(), name)
	err := db.First(&po).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.converter.ToDO(&po), nil
}

// FindAllActive 查询所有激活的命名空间
func (r *NamespaceRepositoryImpl) FindAllActive(ctx context.Context) ([]*domainEntity.Namespace, error) {
	var pos []*infraEntity.NamespacePO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("IsActive").GetColumnName(), true)
	db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDOList(pos), nil
}

// PageWithConditions 根据多条件分页查询命名空间
func (r *NamespaceRepositoryImpl) PageWithConditions(ctx context.Context, req *shareRepo.PageRequest, conditions ...*shareRepo.Condition) (*shareRepo.PageResult[*domainEntity.Namespace], error) {
	db := r.db.WithContext(ctx)

	// 应用查询条件
	if len(conditions) > 0 {
		for _, cond := range conditions {
			db = gormRepo.ApplyCondition(db, cond)
		}
	}

	// 统计总数
	var total int64
	var po infraEntity.NamespacePO
	if err := db.Model(&po).Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用排序
	db = gormRepo.ApplyOrderBys(db, req.OrderBy)

	// 应用分页
	db = db.Offset(req.Offset()).Limit(req.Size)

	// 查询数据
	var pos []*infraEntity.NamespacePO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, req.Page, req.Size), nil
}

// ExistsByName 检查指定名称的命名空间是否存在
func (r *NamespaceRepositoryImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.NamespacePO{})
	db = queryutil.WhereEq(db, r.fields.Get("Name").GetColumnName(), name)
	err := db.Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CountActive 统计激活的命名空间数量
func (r *NamespaceRepositoryImpl) CountActive(ctx context.Context) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.NamespacePO{})
	db = queryutil.WhereEq(db, r.fields.Get("IsActive").GetColumnName(), true)
	err := db.Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// 确保实现了接口
var _ repository.NamespaceRepository = (*NamespaceRepositoryImpl)(nil)
