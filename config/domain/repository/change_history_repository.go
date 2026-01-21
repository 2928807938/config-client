package repository

import (
	"context"

	"config-client/config/domain/entity"
	"config-client/share/repository"
)

// ChangeHistoryQueryParams 变更历史查询参数
type ChangeHistoryQueryParams struct {
	ConfigID    *int    // 配置ID
	NamespaceID *int    // 命名空间ID
	ConfigKey   *string // 配置键
	Operation   *string // 操作类型
	StartTime   *string // 开始时间
	EndTime     *string // 结束时间
	Operator    *string // 操作人
	Page        int     // 页码（从1开始）
	Size        int     // 每页数量
}

// ChangeHistoryRepository 变更历史仓储接口
type ChangeHistoryRepository interface {
	// 继承基础仓储接口
	repository.BaseRepository[entity.ChangeHistory, int]

	// ==================== 写操作 ====================

	// Save 保存变更记录
	Save(ctx context.Context, history *entity.ChangeHistory) error

	// BatchSave 批量保存变更记录
	BatchSave(ctx context.Context, histories []*entity.ChangeHistory) error

	// ==================== 读操作 ====================

	// FindByConfigID 查询指定配置的所有变更历���（按时间倒序）
	FindByConfigID(ctx context.Context, configID int, limit int) ([]*entity.ChangeHistory, error)

	// FindByNamespaceAndKey 查询指定命名空间和配置键的变更历史
	FindByNamespaceAndKey(ctx context.Context, namespaceID int, configKey string, limit int) ([]*entity.ChangeHistory, error)

	// FindByID 根据ID查询变更记录
	FindByID(ctx context.Context, id int) (*entity.ChangeHistory, error)

	// FindLatestByConfigID 查询指定配置的最新变更记录
	FindLatestByConfigID(ctx context.Context, configID int) (*entity.ChangeHistory, error)

	// FindByOperator 查询指定操作人的变更记录
	FindByOperator(ctx context.Context, operator string, page, size int) (*repository.PageResult[*entity.ChangeHistory], error)

	// QueryByParams 根据查询参数分页查询变更历史
	QueryByParams(ctx context.Context, params *ChangeHistoryQueryParams) (*repository.PageResult[*entity.ChangeHistory], error)

	// ==================== 统计操作 ====================

	// CountByConfigID 统计指定配置的变更次数
	CountByConfigID(ctx context.Context, configID int) (int64, error)

	// CountByOperation 统计指定操作类型的变更次数
	CountByOperation(ctx context.Context, operation string) (int64, error)

	// CountByTimeRange 统计时间范围内的变更次数
	CountByTimeRange(ctx context.Context, startTime, endTime string) (int64, error)
}
