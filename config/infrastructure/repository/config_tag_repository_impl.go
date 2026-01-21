package repository

import (
	"context"

	"config-client/config/domain/entity"
	"config-client/config/domain/repository"
	"config-client/config/infrastructure/converter"
	infraEntity "config-client/config/infrastructure/entity"
	"config-client/share/repository/queryutil"

	"gorm.io/gorm"
)

// configTagRepositoryImpl 配置标签仓储实现
type configTagRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.ConfigTagConverter
	fields    *queryutil.EntityFields[infraEntity.ConfigTagPO] // Lambda 字段查询构建器
	model     infraEntity.ConfigTagPO                          // 用于类型安全的字段引用
}

// NewConfigTagRepository 创建配置标签仓储实例
func NewConfigTagRepository(db *gorm.DB) repository.ConfigTagRepository {
	return &configTagRepositoryImpl{
		db:        db,
		converter: converter.NewConfigTagConverter(),
		fields:    queryutil.Lambda[infraEntity.ConfigTagPO](), // 初始化 Lambda 构建器
	}
}

// Create 创建标签
func (r *configTagRepositoryImpl) Create(ctx context.Context, tag *entity.ConfigTag) error {
	po := r.converter.ToPO(tag)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}

	// 回写自增ID
	tag.ID = po.ID
	tag.CreatedAt = po.CreatedAt
	return nil
}

// BatchCreate 批量创建标签
func (r *configTagRepositoryImpl) BatchCreate(ctx context.Context, tags []*entity.ConfigTag) error {
	if len(tags) == 0 {
		return nil
	}

	poList := r.converter.ToPOList(tags)
	if err := r.db.WithContext(ctx).Create(&poList).Error; err != nil {
		return err
	}

	// 回写自增ID
	for i, po := range poList {
		tags[i].ID = po.ID
		tags[i].CreatedAt = po.CreatedAt
	}
	return nil
}

// Delete 删除标签
func (r *configTagRepositoryImpl) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&infraEntity.ConfigTagPO{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteByConfigIDAndTagKey 根据配置ID和标签键删除标签
func (r *configTagRepositoryImpl) DeleteByConfigIDAndTagKey(ctx context.Context, configID int, tagKey string) error {
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("ConfigID").GetColumnName(), configID)
	db = queryutil.WhereEq(db, r.fields.Get("TagKey").GetColumnName(), tagKey)
	result := db.Delete(&infraEntity.ConfigTagPO{})

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteByConfigID 删除某个配置的所有标签
func (r *configTagRepositoryImpl) DeleteByConfigID(ctx context.Context, configID int) error {
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("ConfigID").GetColumnName(), configID)
	result := db.Delete(&infraEntity.ConfigTagPO{})

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByConfigID 查询某个配置的所有标签
func (r *configTagRepositoryImpl) FindByConfigID(ctx context.Context, configID int) ([]*entity.ConfigTag, error) {
	var poList []*infraEntity.ConfigTagPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("ConfigID").GetColumnName(), configID)
	db = queryutil.OrderBy(db, r.fields.Get("TagKey").GetColumnName())
	db = queryutil.OrderBy(db, r.fields.Get("TagValue").GetColumnName())
	err := db.Find(&poList).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDomainList(poList), nil
}

// FindByTagKey 根据标签键查询标签
func (r *configTagRepositoryImpl) FindByTagKey(ctx context.Context, tagKey string) ([]*entity.ConfigTag, error) {
	var poList []*infraEntity.ConfigTagPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("TagKey").GetColumnName(), tagKey)
	err := db.Find(&poList).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDomainList(poList), nil
}

// FindByTagKeyValue 根据标签键值查询标签
func (r *configTagRepositoryImpl) FindByTagKeyValue(ctx context.Context, tagKey, tagValue string) ([]*entity.ConfigTag, error) {
	var poList []*infraEntity.ConfigTagPO
	db := r.db.WithContext(ctx)
	db = queryutil.WhereEq(db, r.fields.Get("TagKey").GetColumnName(), tagKey)
	db = queryutil.WhereEq(db, r.fields.Get("TagValue").GetColumnName(), tagValue)
	err := db.Find(&poList).Error

	if err != nil {
		return nil, err
	}

	return r.converter.ToDomainList(poList), nil
}

// ExistsByConfigIDAndTag 检查某个配置是否已存在指定标签
func (r *configTagRepositoryImpl) ExistsByConfigIDAndTag(ctx context.Context, configID int, tagKey, tagValue string) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&infraEntity.ConfigTagPO{})
	db = queryutil.WhereEq(db, r.fields.Get("ConfigID").GetColumnName(), configID)
	db = queryutil.WhereEq(db, r.fields.Get("TagKey").GetColumnName(), tagKey)
	db = queryutil.WhereEq(db, r.fields.Get("TagValue").GetColumnName(), tagValue)
	err := db.Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// FindConfigIDsByTags 根据标签查询配置ID列表(支持多个标签的AND查询)
func (r *configTagRepositoryImpl) FindConfigIDsByTags(ctx context.Context, tags []entity.TagInput) ([]int, error) {
	if len(tags) == 0 {
		return []int{}, nil
	}

	// 使用子查询实现多标签AND查询
	// SELECT config_id FROM t_config_tags WHERE tag_key = ? AND tag_value = ?
	// INTERSECT SELECT config_id FROM t_config_tags WHERE tag_key = ? AND tag_value = ?
	// ...

	var configIDs []int

	// 第一个标签作为基础查询
	db := r.db.WithContext(ctx).Model(&infraEntity.ConfigTagPO{}).Select("config_id")
	db = queryutil.WhereEq(db, r.fields.Get("TagKey").GetColumnName(), tags[0].TagKey)
	db = queryutil.WhereEq(db, r.fields.Get("TagValue").GetColumnName(), tags[0].TagValue)

	// 后续标签作为 INTERSECT 查询
	for i := 1; i < len(tags); i++ {
		subQuery := r.db.Model(&infraEntity.ConfigTagPO{}).Select("config_id")
		subQuery = queryutil.WhereEq(subQuery, r.fields.Get("TagKey").GetColumnName(), tags[i].TagKey)
		subQuery = queryutil.WhereEq(subQuery, r.fields.Get("TagValue").GetColumnName(), tags[i].TagValue)

		// GORM 不直接支持 INTERSECT,使用 IN 子查询实现
		db = queryutil.WhereIn(db, "config_id", subQuery)
	}

	err := db.Pluck("config_id", &configIDs).Error
	if err != nil {
		return nil, err
	}

	return configIDs, nil
}
