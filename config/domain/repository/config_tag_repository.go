package repository

import (
	"context"

	"config-client/config/domain/entity"
)

// ConfigTagRepository 配置标签仓储接口
// 负责配置标签的持久化操作
type ConfigTagRepository interface {
	// Create 创建标签
	Create(ctx context.Context, tag *entity.ConfigTag) error

	// BatchCreate 批量创建标签
	BatchCreate(ctx context.Context, tags []*entity.ConfigTag) error

	// Delete 删除标签
	Delete(ctx context.Context, id int) error

	// DeleteByConfigIDAndTagKey 根据配置ID和标签键删除标签
	DeleteByConfigIDAndTagKey(ctx context.Context, configID int, tagKey string) error

	// DeleteByConfigID 删除某个配置的所有标签
	DeleteByConfigID(ctx context.Context, configID int) error

	// FindByConfigID 查询某个配置的所有标签
	FindByConfigID(ctx context.Context, configID int) ([]*entity.ConfigTag, error)

	// FindByTagKey 根据标签键查询标签
	FindByTagKey(ctx context.Context, tagKey string) ([]*entity.ConfigTag, error)

	// FindByTagKeyValue 根据标签键值查询标签
	FindByTagKeyValue(ctx context.Context, tagKey, tagValue string) ([]*entity.ConfigTag, error)

	// ExistsByConfigIDAndTag 检查某个配置是否已存在指定标签
	ExistsByConfigIDAndTag(ctx context.Context, configID int, tagKey, tagValue string) (bool, error)

	// FindConfigIDsByTags 根据标签查询配置ID列表（支持多个标签的AND查询）
	FindConfigIDsByTags(ctx context.Context, tags []entity.TagInput) ([]int, error)
}
