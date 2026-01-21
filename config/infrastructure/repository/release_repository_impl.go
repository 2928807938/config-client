package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	domainEntity "config-client/config/domain/entity"
	"config-client/config/domain/repository"
	"config-client/config/infrastructure/converter"
	infraEntity "config-client/config/infrastructure/entity"
	shareRepo "config-client/share/repository"
	gormRepo "config-client/share/repository/gorm"
	"config-client/share/repository/queryutil"
)

// ReleaseRepositoryImpl 发布版本仓储实现
type ReleaseRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.ReleaseConverter
	fields    *queryutil.EntityFields[infraEntity.ReleasePO]
	model     infraEntity.ReleasePO
}

// NewReleaseRepository 创建发布版本仓储实例
func NewReleaseRepository(db *gorm.DB) repository.ReleaseRepository {
	return &ReleaseRepositoryImpl{
		db:        db,
		converter: converter.NewReleaseConverter(),
		fields:    queryutil.Lambda[infraEntity.ReleasePO](),
	}
}

// ==================== 基础 CRUD 实现 ====================

// Create 创建发布版本
func (r *ReleaseRepositoryImpl) Create(ctx context.Context, entity *domainEntity.Release) error {
	po := r.converter.ToPO(entity)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	entity.ID = po.ID
	return nil
}

// CreateBatch 批量创建发布版本
func (r *ReleaseRepositoryImpl) CreateBatch(ctx context.Context, entities []*domainEntity.Release) error {
	if len(entities) == 0 {
		return nil
	}
	pos := r.converter.ToPOList(entities)
	return r.db.WithContext(ctx).Create(pos).Error
}

// GetByID 根据ID查询发布版本
func (r *ReleaseRepositoryImpl) GetByID(ctx context.Context, id int) (*domainEntity.Release, error) {
	var po infraEntity.ReleasePO
	err := r.db.WithContext(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToDO(&po), nil
}

// Update 更新发布版本
func (r *ReleaseRepositoryImpl) Update(ctx context.Context, entity *domainEntity.Release) error {
	po := r.converter.ToPO(entity)
	return r.db.WithContext(ctx).Save(po).Error
}

// Delete 删除发布版本（软删除）
func (r *ReleaseRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&infraEntity.ReleasePO{}, id).Error
}

// List 查询全部发布版本
func (r *ReleaseRepositoryImpl) List(ctx context.Context) ([]*domainEntity.Release, error) {
	var pos []*infraEntity.ReleasePO
	err := r.db.WithContext(ctx).Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return r.converter.ToDOList(pos), nil
}

// Page 分页查询发布版本
func (r *ReleaseRepositoryImpl) Page(ctx context.Context, request *shareRepo.PageRequest) (*shareRepo.PageResult[*domainEntity.Release], error) {
	db := r.db.WithContext(ctx)

	// 统计总数
	var total int64
	var po infraEntity.ReleasePO
	if err := db.Model(&po).Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用排序
	db = gormRepo.ApplyOrderBys(db, request.OrderBy)

	// 应用分页
	db = db.Offset(request.Offset()).Limit(request.Size)

	// 查询数据
	var pos []*infraEntity.ReleasePO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, request.Page, request.Size), nil
}

// ==================== 自定义查询方法实现 ====================

// FindByNamespaceAndVersion 根据命名空间和版本号查询发布版本
func (r *ReleaseRepositoryImpl) FindByNamespaceAndVersion(ctx context.Context, namespaceID int, version int, environment string) (*domainEntity.Release, error) {
	var po infraEntity.ReleasePO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Version).GetColumnName(), version)
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

// FindByVersionName 根据版本名称查询发布版本
func (r *ReleaseRepositoryImpl) FindByVersionName(ctx context.Context, namespaceID int, versionName string, environment string) (*domainEntity.Release, error) {
	var po infraEntity.ReleasePO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.VersionName).GetColumnName(), versionName)
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

// FindLatestPublishedRelease 查询最新的已发布版本
func (r *ReleaseRepositoryImpl) FindLatestPublishedRelease(ctx context.Context, namespaceID int, environment string) (*domainEntity.Release, error) {
	var po infraEntity.ReleasePO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Status).GetColumnName(), domainEntity.ReleaseStatusPublished)
	db = queryutil.OrderByDesc(db, r.fields.Of(&r.model.Version).GetColumnName())
	err := db.First(&po).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToDO(&po), nil
}

// FindByNamespace 查询指定命名空间的所有发布版本
func (r *ReleaseRepositoryImpl) FindByNamespace(ctx context.Context, namespaceID int, environment string) ([]*domainEntity.Release, error) {
	var pos []*infraEntity.ReleasePO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	db = queryutil.OrderByDesc(db, r.fields.Of(&r.model.Version).GetColumnName())
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}
	return r.converter.ToDOList(pos), nil
}

// FindByStatus 根据状态查询发布版本列表
func (r *ReleaseRepositoryImpl) FindByStatus(ctx context.Context, namespaceID int, environment string, status domainEntity.ReleaseStatus) ([]*domainEntity.Release, error) {
	var pos []*infraEntity.ReleasePO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Status).GetColumnName(), status)
	db = queryutil.OrderByDesc(db, r.fields.Of(&r.model.Version).GetColumnName())
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}
	return r.converter.ToDOList(pos), nil
}

// QueryByParams 根据查询参数分页查询发布版本
func (r *ReleaseRepositoryImpl) QueryByParams(ctx context.Context, params *repository.ReleaseQueryParams) (*shareRepo.PageResult[*domainEntity.Release], error) {
	db := r.db.WithContext(ctx).Model(&infraEntity.ReleasePO{})

	// 构建查询条件
	if params.NamespaceID != nil {
		db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), *params.NamespaceID)
	}
	if params.Environment != nil {
		db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), *params.Environment)
	}
	if params.Status != nil {
		db = queryutil.WhereEq(db, r.fields.Of(&r.model.Status).GetColumnName(), *params.Status)
	}
	if params.ReleaseType != nil {
		db = queryutil.WhereEq(db, r.fields.Of(&r.model.ReleaseType).GetColumnName(), *params.ReleaseType)
	}
	if params.VersionName != nil && *params.VersionName != "" {
		db = queryutil.WhereLike(db, r.fields.Of(&r.model.VersionName).GetColumnName(), "%"+*params.VersionName+"%")
	}

	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Size <= 0 {
		params.Size = 20
	}

	// 应用排序
	if params.OrderBy != "" {
		db = db.Order(params.OrderBy)
	} else {
		db = queryutil.OrderByDesc(db, r.fields.Of(&r.model.Version).GetColumnName())
	}

	// 应用分页
	offset := (params.Page - 1) * params.Size
	db = db.Offset(offset).Limit(params.Size)

	// 查询数据
	var pos []*infraEntity.ReleasePO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, params.Page, params.Size), nil
}

// GetNextVersion 获取下一个版本号
func (r *ReleaseRepositoryImpl) GetNextVersion(ctx context.Context, namespaceID int, environment string) (int, error) {
	var maxVersion int
	db := r.db.WithContext(ctx).Model(&infraEntity.ReleasePO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	err := db.Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error

	if err != nil {
		return 0, err
	}
	return maxVersion + 1, nil
}

// CountByNamespace 统计指定命名空间的发布版本数量
func (r *ReleaseRepositoryImpl) CountByNamespace(ctx context.Context, namespaceID int, environment string) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.ReleasePO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	err := db.Count(&count).Error
	return count, err
}

// FindReleasesInTimeRange 查询指定时间范围内的发布版本
func (r *ReleaseRepositoryImpl) FindReleasesInTimeRange(ctx context.Context, namespaceID int, environment string, startTime, endTime time.Time) ([]*domainEntity.Release, error) {
	var pos []*infraEntity.ReleasePO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	db = queryutil.WhereBetween(db, r.fields.Of(&r.model.ReleasedAt).GetColumnName(), startTime, endTime)
	db = queryutil.OrderByDesc(db, r.fields.Of(&r.model.ReleasedAt).GetColumnName())
	err := db.Find(&pos).Error

	if err != nil {
		return nil, err
	}
	return r.converter.ToDOList(pos), nil
}

// ExistsByVersion 判断指定版本是否存在
func (r *ReleaseRepositoryImpl) ExistsByVersion(ctx context.Context, namespaceID int, version int, environment string) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.ReleasePO{})
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.NamespaceID).GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Version).GetColumnName(), version)
	db = queryutil.WhereEq(db, r.fields.Of(&r.model.Environment).GetColumnName(), environment)
	err := db.Count(&count).Error
	return count > 0, err
}
