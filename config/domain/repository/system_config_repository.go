package repository

import (
	"context"

	"config-client/config/domain/entity"
	"config-client/share/repository"
)

// SystemConfigRepository 系统配置仓储接口
// 定义系统配置的数据访问方法
type SystemConfigRepository interface {
	// 继承基础仓储接口，提供通用的 CRUD 操作
	repository.BaseRepository[entity.SystemConfig, int]

	// ==================== 自定义查询方法 ====================

	// FindByKey 根据配置键查询系统配置
	// 参数：key - 配置键
	// 返回：配置实体，如果不存在返回 nil
	FindByKey(ctx context.Context, key string) (*entity.SystemConfig, error)

	// FindAllActive 查询所有启用的系统配置
	// 返回：所有已启用的配置列表
	FindAllActive(ctx context.Context) ([]*entity.SystemConfig, error)

	// FindAll 查询所有系统配置（包括禁用的）
	// 返回：所有配置列表
	FindAll(ctx context.Context) ([]*entity.SystemConfig, error)

	// ExistsByKey 判断指定配置键是否存在
	// 参数：key - 配置键
	// 返回：true - 存在，false - 不存在
	ExistsByKey(ctx context.Context, key string) (bool, error)

	// UpdateValue 更新配置值
	// 参数：key - 配置键，value - 新的配置值
	// 返回：错误信息
	UpdateValue(ctx context.Context, key string, value string) error

	// UpdateActive 更新配置的启用状态
	// 参数：key - 配置键，isActive - 是否启用
	// 返回：错误信息
	UpdateActive(ctx context.Context, key string, isActive bool) error
}
