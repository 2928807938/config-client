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
	"config-client/share/repository/queryutil"
)

// ChangeHistoryRepositoryImpl 变更历史仓储实现
type ChangeHistoryRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.ChangeHistoryConverter
	fields    *queryutil.EntityFields[infraEntity.ChangeHistoryPO] // Lambda 字段查询构建器
	model     infraEntity.ChangeHistoryPO                          // 用于类型安全的字段引用
}

// NewChangeHistoryRepository 创建变更历史仓储实例
func NewChangeHistoryRepository(db *gorm.DB) repository.ChangeHistoryRepository {
	return &ChangeHistoryRepositoryImpl{
		db:        db,
		converter: converter.NewChangeHistoryConverter(),
		fields:    queryutil.Lambda[infraEntity.ChangeHistoryPO](), // 初始化 Lambda 构建器
	}
}

// ==================== 写操作实现 ====================

// Save 保存变更记录
func (r *ChangeHistoryRepositoryImpl) Save(ctx context.Context, history *domainEntity.ChangeHistory) error {
	po := r.converter.ToPO(history)
	return r.db.WithContext(ctx).Create(po).Error
}

// BatchSave 批量保存变更记录
func (r *ChangeHistoryRepositoryImpl) BatchSave(ctx context.Context, histories []*domainEntity.ChangeHistory) error {
	if len(histories) == 0 {
		return nil
	}
	pos := r.converter.ToPOList(histories)
	return r.db.WithContext(ctx).Create(pos).Error
}

// ==================== 读操作实现 ====================

// FindByConfigID 查询指定配置的所有变更历史(按时间倒序)
func (r *ChangeHistoryRepositoryImpl) FindByConfigID(ctx context.Context, configID int, limit int) ([]*domainEntity.ChangeHistory, error) {
	var pos []*infraEntity.ChangeHistoryPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("ConfigID").GetColumnName(), configID)
	db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())

	if limit > 0 {
		db = db.Limit(limit)
	}

	err := db.Find(&pos).Error
	if err != nil {
		return nil, err
	}

	return r.converter.ToDOList(pos), nil
}

// FindByNamespaceAndKey 查询指定命名空间和配置键的变更历史
func (r *ChangeHistoryRepositoryImpl) FindByNamespaceAndKey(ctx context.Context, namespaceID int, configKey string, limit int) ([]*domainEntity.ChangeHistory, error) {
	var pos []*infraEntity.ChangeHistoryPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("NamespaceID").GetColumnName(), namespaceID)
	db = queryutil.WhereEq(db, r.fields.Get("ConfigKey").GetColumnName(), configKey)
	db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())

	if limit > 0 {
		db = db.Limit(limit)
	}

	err := db.Find(&pos).Error
	if err != nil {
		return nil, err
	}

	return r.converter.ToDOList(pos), nil
}

// FindByID 根据ID查询变更记录
func (r *ChangeHistoryRepositoryImpl) FindByID(ctx context.Context, id int) (*domainEntity.ChangeHistory, error) {
	var po infraEntity.ChangeHistoryPO
	err := r.db.WithContext(ctx).First(&po, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToDO(&po), nil
}

// FindLatestByConfigID 查询指定配置的最新变更记录
func (r *ChangeHistoryRepositoryImpl) FindLatestByConfigID(ctx context.Context, configID int) (*domainEntity.ChangeHistory, error) {
	var po infraEntity.ChangeHistoryPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("ConfigID").GetColumnName(), configID)
	db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())
	err := db.First(&po).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToDO(&po), nil
}

// FindByOperator 查询指定操作人的变更记录
func (r *ChangeHistoryRepositoryImpl) FindByOperator(ctx context.Context, operator string, page, size int) (*shareRepo.PageResult[*domainEntity.ChangeHistory], error) {
	db := r.db.WithContext(ctx).Model(&infraEntity.ChangeHistoryPO{})
	db = queryutil.WhereEq(db, r.fields.Get("Operator").GetColumnName(), operator)

	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用分页
	offset := (page - 1) * size
	db = db.Offset(offset).Limit(size)
	db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())

	// 查询数据
	var pos []*infraEntity.ChangeHistoryPO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, page, size), nil
}

// QueryByParams 根据查询参数分页查询变更历史
func (r *ChangeHistoryRepositoryImpl) QueryByParams(ctx context.Context, params *repository.ChangeHistoryQueryParams) (*shareRepo.PageResult[*domainEntity.ChangeHistory], error) {
	db := r.db.WithContext(ctx).Model(&infraEntity.ChangeHistoryPO{})

	// 构建查询条件
	if params.ConfigID != nil {
		db = queryutil.WhereEq(db, r.fields.Get("ConfigID").GetColumnName(), *params.ConfigID)
	}
	if params.NamespaceID != nil {
		db = queryutil.WhereEq(db, r.fields.Get("NamespaceID").GetColumnName(), *params.NamespaceID)
	}
	if params.ConfigKey != nil && *params.ConfigKey != "" {
		db = queryutil.WhereLike(db, r.fields.Get("ConfigKey").GetColumnName(), "%"+*params.ConfigKey+"%")
	}
	if params.Operation != nil && *params.Operation != "" {
		db = queryutil.WhereEq(db, r.fields.Get("Operation").GetColumnName(), *params.Operation)
	}
	if params.Operator != nil && *params.Operator != "" {
		db = queryutil.WhereLike(db, r.fields.Get("Operator").GetColumnName(), "%"+*params.Operator+"%")
	}
	if params.StartTime != nil {
		startTime, err := time.Parse("2006-01-02 15:04:05", *params.StartTime)
		if err == nil {
			db = queryutil.WhereGte(db, r.fields.Get("CreatedAt").GetColumnName(), startTime)
		}
	}
	if params.EndTime != nil {
		endTime, err := time.Parse("2006-01-02 15:04:05", *params.EndTime)
		if err == nil {
			db = queryutil.WhereLte(db, r.fields.Get("CreatedAt").GetColumnName(), endTime)
		}
	}

	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 默认按时间倒序
	db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())

	// 应用分页
	offset := (params.Page - 1) * params.Size
	db = db.Offset(offset).Limit(params.Size)

	// 查询数据
	var pos []*infraEntity.ChangeHistoryPO
	if err := db.Find(&pos).Error; err != nil {
		return nil, err
	}

	dos := r.converter.ToDOList(pos)
	return shareRepo.NewPageResult(dos, total, params.Page, params.Size), nil
}

// ==================== 基础 CRUD 实现 ====================

// Create 创建变更记录
func (r *ChangeHistoryRepositoryImpl) Create(ctx context.Context, entity *domainEntity.ChangeHistory) error {
	return r.Save(ctx, entity)
}

// CreateBatch 批量创建变更记录
func (r *ChangeHistoryRepositoryImpl) CreateBatch(ctx context.Context, entities []*domainEntity.ChangeHistory) error {
	return r.BatchSave(ctx, entities)
}

// GetByID 根据ID查询变更记录
func (r *ChangeHistoryRepositoryImpl) GetByID(ctx context.Context, id int) (*domainEntity.ChangeHistory, error) {
	return r.FindByID(ctx, id)
}

// Update 更新变更记录（历史记录不允许修改）
func (r *ChangeHistoryRepositoryImpl) Update(ctx context.Context, entity *domainEntity.ChangeHistory) error {
	// 变更历史是只读的，不允许更新
	return errors.New("change history is read-only")
}

// Delete 删除变更记录（通常不允许删除）
func (r *ChangeHistoryRepositoryImpl) Delete(ctx context.Context, id int) error {
	// 变更历史通常不允许删除
	return errors.New("change history cannot be deleted")
}

// List 查询全部列表
func (r *ChangeHistoryRepositoryImpl) List(ctx context.Context) ([]*domainEntity.ChangeHistory, error) {
	var pos []*infraEntity.ChangeHistoryPO
	db := r.db.WithContext(ctx)
	db = queryutil.OrderByDesc(db, r.fields.Get("CreatedAt").GetColumnName())
	err := db.Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return r.converter.ToDOList(pos), nil
}

// Page 分页查询
func (r *ChangeHistoryRepositoryImpl) Page(ctx context.Context, request *shareRepo.PageRequest) (*shareRepo.PageResult[*domainEntity.ChangeHistory], error) {
	params := &repository.ChangeHistoryQueryParams{
		Page: request.Page,
		Size: request.Size,
	}
	return r.QueryByParams(ctx, params)
}

// ==================== 统计操作实现 ====================

// CountByConfigID 统计指定配置的变更次数
func (r *ChangeHistoryRepositoryImpl) CountByConfigID(ctx context.Context, configID int) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.ChangeHistoryPO{})
	db = queryutil.WhereEq(db, r.fields.Get("ConfigID").GetColumnName(), configID)
	err := db.Count(&count).Error
	return count, err
}

// CountByOperation 统计指定操作类型的变更次数
func (r *ChangeHistoryRepositoryImpl) CountByOperation(ctx context.Context, operation string) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.ChangeHistoryPO{})
	db = queryutil.WhereEq(db, r.fields.Get("Operation").GetColumnName(), operation)
	err := db.Count(&count).Error
	return count, err
}

// CountByTimeRange 统计时间范围内的变更次数
func (r *ChangeHistoryRepositoryImpl) CountByTimeRange(ctx context.Context, startTime, endTime string) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.ChangeHistoryPO{})

	if startTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			db = queryutil.WhereGte(db, r.fields.Get("CreatedAt").GetColumnName(), t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			db = queryutil.WhereLte(db, r.fields.Get("CreatedAt").GetColumnName(), t)
		}
	}

	err := db.Count(&count).Error
	return count, err
}

// 确保实现了接口
var _ repository.ChangeHistoryRepository = (*ChangeHistoryRepositoryImpl)(nil)
