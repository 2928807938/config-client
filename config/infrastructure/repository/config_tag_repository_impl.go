package repository

import (
	"context"

	"config-client/config/domain/entity"
	"config-client/config/domain/repository"
	"config-client/config/infrastructure/converter"
	infraEntity "config-client/config/infrastructure/entity"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

// configTagRepositoryImpl 配置标签仓储实现
type configTagRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.ConfigTagConverter
}

// NewConfigTagRepository 创建配置标签仓储实例
func NewConfigTagRepository(db *gorm.DB) repository.ConfigTagRepository {
	return &configTagRepositoryImpl{
		db:        db,
		converter: converter.NewConfigTagConverter(),
	}
}

// Create 创建标签
func (r *configTagRepositoryImpl) Create(ctx context.Context, tag *entity.ConfigTag) error {
	po := r.converter.ToPO(tag)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		hlog.CtxErrorf(ctx, "创建配置标签失败: %v, tag: %+v", err, tag)
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
		hlog.CtxErrorf(ctx, "批量创建配置标签失败: %v, count: %d", err, len(tags))
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
		hlog.CtxErrorf(ctx, "删除配置标签失败: %v, id: %d", result.Error, id)
		return result.Error
	}
	return nil
}

// DeleteByConfigIDAndTagKey 根据配置ID和标签键删除标签
func (r *configTagRepositoryImpl) DeleteByConfigIDAndTagKey(ctx context.Context, configID int, tagKey string) error {
	result := r.db.WithContext(ctx).
		Where("config_id = ? AND tag_key = ?", configID, tagKey).
		Delete(&infraEntity.ConfigTagPO{})

	if result.Error != nil {
		hlog.CtxErrorf(ctx, "删除配置标签失败: %v, configID: %d, tagKey: %s", result.Error, configID, tagKey)
		return result.Error
	}
	return nil
}

// DeleteByConfigID 删除某个配置的所有标签
func (r *configTagRepositoryImpl) DeleteByConfigID(ctx context.Context, configID int) error {
	result := r.db.WithContext(ctx).
		Where("config_id = ?", configID).
		Delete(&infraEntity.ConfigTagPO{})

	if result.Error != nil {
		hlog.CtxErrorf(ctx, "删除配置所有标签失败: %v, configID: %d", result.Error, configID)
		return result.Error
	}
	return nil
}

// FindByConfigID 查询某个配置的所有标签
func (r *configTagRepositoryImpl) FindByConfigID(ctx context.Context, configID int) ([]*entity.ConfigTag, error) {
	var poList []*infraEntity.ConfigTagPO
	err := r.db.WithContext(ctx).
		Where("config_id = ?", configID).
		Order("tag_key ASC, tag_value ASC").
		Find(&poList).Error

	if err != nil {
		hlog.CtxErrorf(ctx, "查询配置标签失败: %v, configID: %d", err, configID)
		return nil, err
	}

	return r.converter.ToDomainList(poList), nil
}

// FindByTagKey 根据标签键查询标签
func (r *configTagRepositoryImpl) FindByTagKey(ctx context.Context, tagKey string) ([]*entity.ConfigTag, error) {
	var poList []*infraEntity.ConfigTagPO
	err := r.db.WithContext(ctx).
		Where("tag_key = ?", tagKey).
		Find(&poList).Error

	if err != nil {
		hlog.CtxErrorf(ctx, "根据标签键查询失败: %v, tagKey: %s", err, tagKey)
		return nil, err
	}

	return r.converter.ToDomainList(poList), nil
}

// FindByTagKeyValue 根据标签键值查询标签
func (r *configTagRepositoryImpl) FindByTagKeyValue(ctx context.Context, tagKey, tagValue string) ([]*entity.ConfigTag, error) {
	var poList []*infraEntity.ConfigTagPO
	err := r.db.WithContext(ctx).
		Where("tag_key = ? AND tag_value = ?", tagKey, tagValue).
		Find(&poList).Error

	if err != nil {
		hlog.CtxErrorf(ctx, "根据标签键值查询失败: %v, tagKey: %s, tagValue: %s", err, tagKey, tagValue)
		return nil, err
	}

	return r.converter.ToDomainList(poList), nil
}

// ExistsByConfigIDAndTag 检查某个配置是否已存在指定标签
func (r *configTagRepositoryImpl) ExistsByConfigIDAndTag(ctx context.Context, configID int, tagKey, tagValue string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&infraEntity.ConfigTagPO{}).
		Where("config_id = ? AND tag_key = ? AND tag_value = ?", configID, tagKey, tagValue).
		Count(&count).Error

	if err != nil {
		hlog.CtxErrorf(ctx, "检查标签存在性失败: %v, configID: %d, tag: %s:%s", err, configID, tagKey, tagValue)
		return false, err
	}

	return count > 0, nil
}

// FindConfigIDsByTags 根据标签查询配置ID列表（支持多个标签的AND查询）
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
	query := r.db.WithContext(ctx).
		Model(&infraEntity.ConfigTagPO{}).
		Select("config_id").
		Where("tag_key = ? AND tag_value = ?", tags[0].TagKey, tags[0].TagValue)

	// 后续标签作为 INTERSECT 查询
	for i := 1; i < len(tags); i++ {
		subQuery := r.db.Model(&infraEntity.ConfigTagPO{}).
			Select("config_id").
			Where("tag_key = ? AND tag_value = ?", tags[i].TagKey, tags[i].TagValue)

		// GORM 不直接支持 INTERSECT，使用 IN 子查询实现
		query = query.Where("config_id IN (?)", subQuery)
	}

	err := query.Pluck("config_id", &configIDs).Error
	if err != nil {
		hlog.CtxErrorf(ctx, "根据标签查询配置ID失败: %v, tags: %+v", err, tags)
		return nil, err
	}

	return configIDs, nil
}
