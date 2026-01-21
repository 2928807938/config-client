package repository

import (
	"context"
	"time"

	"config-client/config/domain/entity"
	"config-client/share/repository"
)

// ReleaseQueryParams 发布版本查询参数
type ReleaseQueryParams struct {
	NamespaceID *int
	Environment *string
	Status      *string // testing/published/rollback
	ReleaseType *string // full/incremental/canary
	VersionName *string
	Page        int
	Size        int
	OrderBy     string
}

// ReleaseRepository 发布版本仓储接口
// 定义发布版本管理的数据访问方法
type ReleaseRepository interface {
	// 继承基础仓储接口，提供通用的 CRUD 操作
	repository.BaseRepository[entity.Release, int]

	// ==================== 自定义查询方法 ====================

	// FindByNamespaceAndVersion 根据命名空间和版本号查询发布版本
	FindByNamespaceAndVersion(ctx context.Context, namespaceID int, version int, environment string) (*entity.Release, error)

	// FindByVersionName 根据版本名称查询发布版本
	FindByVersionName(ctx context.Context, namespaceID int, versionName string, environment string) (*entity.Release, error)

	// FindLatestPublishedRelease 查询最新的已发布版本
	FindLatestPublishedRelease(ctx context.Context, namespaceID int, environment string) (*entity.Release, error)

	// FindByNamespace 查询指定命名空间的所有发布版本
	FindByNamespace(ctx context.Context, namespaceID int, environment string) ([]*entity.Release, error)

	// FindByStatus 根据状态查询发布版本列表
	FindByStatus(ctx context.Context, namespaceID int, environment string, status entity.ReleaseStatus) ([]*entity.Release, error)

	// QueryByParams 根据查询参数分页查询发布版本
	QueryByParams(ctx context.Context, params *ReleaseQueryParams) (*repository.PageResult[*entity.Release], error)

	// GetNextVersion 获取下一个版本号
	GetNextVersion(ctx context.Context, namespaceID int, environment string) (int, error)

	// CountByNamespace 统计指定命名空间的发布版本数量
	CountByNamespace(ctx context.Context, namespaceID int, environment string) (int64, error)

	// FindReleasesInTimeRange 查询指定时间范围内的发布版本
	FindReleasesInTimeRange(ctx context.Context, namespaceID int, environment string, startTime, endTime time.Time) ([]*entity.Release, error)

	// ExistsByVersion 判断指定版本是否存在
	ExistsByVersion(ctx context.Context, namespaceID int, version int, environment string) (bool, error)
}
