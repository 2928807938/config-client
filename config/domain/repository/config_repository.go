package repository

import (
	"context"

	"config-client/config/domain/entity"
	"config-client/share/repository"
)

// ConfigQueryParams 配置查询参数
// 定义在仓储接口中，供领域层使用
type ConfigQueryParams struct {
	NamespaceID *int
	Key         *string
	GroupName   *string
	Environment *string
	IsActive    *bool
	IsReleased  *bool
	ValueType   *string
	Page        int
	Size        int
	OrderBy     string
}

// ConfigRepository 配置仓储接口，定义配置管理的数据访问方法
type ConfigRepository interface {
	// 继承基础仓储接口，提供通用的 CRUD 操作
	repository.BaseRepository[entity.Config, int]

	// ==================== 自定义查询方法 ====================

	// FindByNamespaceAndKey 根据命名空间ID和配置键查询配置
	// 参数：namespaceID - 命名空间ID，key - 配置键，environment - 环境
	FindByNamespaceAndKey(ctx context.Context, namespaceID int, key string, environment string) (*entity.Config, error)

	// FindByNamespace 根据命名空间ID查询该命名空间下的所有配置
	FindByNamespace(ctx context.Context, namespaceID int) ([]*entity.Config, error)

	// FindByEnvironment 根据环境查询配置列表
	FindByEnvironment(ctx context.Context, environment string) ([]*entity.Config, error)

	// FindByGroup 根据分组查询配置列表
	FindByGroup(ctx context.Context, namespaceID int, groupName string) ([]*entity.Config, error)

	// FindReleasedConfigs 查询已发布的配置列表
	FindReleasedConfigs(ctx context.Context, namespaceID int, environment string) ([]*entity.Config, error)

	// PageWithConditions 根据多条件分页查询配置
	PageWithConditions(ctx context.Context, req *repository.PageRequest, conditions ...*repository.Condition) (*repository.PageResult[*entity.Config], error)

	// QueryByParams 根据查询参数分页查询配置
	// 封装了查询条件的构建逻辑，由仓储层实现字段映射
	QueryByParams(ctx context.Context, params *ConfigQueryParams) (*repository.PageResult[*entity.Config], error)

	// ExistsByNamespaceAndKey 判断指定命名空间和键的配置是否存在
	ExistsByNamespaceAndKey(ctx context.Context, namespaceID int, key string, environment string) (bool, error)

	// CountByNamespace 统计指定命名空间的配置数量
	CountByNamespace(ctx context.Context, namespaceID int) (int64, error)
}
