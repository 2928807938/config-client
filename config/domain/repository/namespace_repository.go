package repository

import (
	"context"

	"config-client/config/domain/entity"
	"config-client/share/repository"
)

// NamespaceQueryParams 命名空间查询参数
type NamespaceQueryParams struct {
	Name     *string // 命名空间名称（模糊查询）
	IsActive *bool   // 是否激活
	Page     int     // 页码
	Size     int     // 每页数量
	OrderBy  string  // 排序字段
}

// NamespaceRepository 命名空间仓储接口
// 定义命名空间的数据访问操作
type NamespaceRepository interface {
	// 继承基础仓储接口，提供通用的 CRUD 操作
	repository.BaseRepository[entity.Namespace, int]

	// ==================== 自定义查询方法 ====================

	// Query 根据查询参数查询命名空间（分页）
	Query(ctx context.Context, params *NamespaceQueryParams) (*repository.PageResult[*entity.Namespace], error)

	// FindByName 根据名称查询命名空间
	FindByName(ctx context.Context, name string) (*entity.Namespace, error)

	// FindAllActive 查询所有激活的命名空间
	FindAllActive(ctx context.Context) ([]*entity.Namespace, error)

	// PageWithConditions 根据多条件分页查询命名空间
	PageWithConditions(ctx context.Context, req *repository.PageRequest, conditions ...*repository.Condition) (*repository.PageResult[*entity.Namespace], error)

	// ==================== 存在性检查 ====================

	// ExistsByName 检查指定名称的命名空间是否存在
	ExistsByName(ctx context.Context, name string) (bool, error)

	// ==================== 统计查询 ====================

	// CountActive 统计激活的命名空间数量
	CountActive(ctx context.Context) (int64, error)
}
