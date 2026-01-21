package service

import (
	"context"
	"sync"
	"time"

	"config-client/config/domain/entity"
	"config-client/config/domain/repository"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// 预定义的系统配置键常量
const (
	// 长轮询相关配置
	ConfigKeyLongPollingTimeout = "long.polling.timeout"  // 长轮询超时时间（秒）
	ConfigKeyLongPollingMaxWait = "long.polling.max.wait" // 长轮询最大等待时间（秒）

	// 订阅相关配置
	ConfigKeyMaxSubscriptions = "max.subscriptions" // 最大订阅数

	// 心跳相关配置
	ConfigKeyHeartbeatInterval = "heartbeat.interval" // 心跳间隔（秒）
	ConfigKeyHeartbeatTimeout  = "heartbeat.timeout"  // 心跳超时（秒）

	// 默认值
	DefaultLongPollingTimeout = 30    // 默认长轮询超时 30 秒
	DefaultLongPollingMaxWait = 60    // 默认长轮询最大等待 60 秒
	DefaultMaxSubscriptions   = 10000 // 默认最大订阅数 10000
	DefaultHeartbeatInterval  = 60    // 默认心跳间隔 60 秒
	DefaultHeartbeatTimeout   = 300   // 默认心跳超时 300 秒
)

// SystemConfigService 系统配置服务
// 提供系统配置的读取和管理功能，支持内存缓存
type SystemConfigService struct {
	repo  repository.SystemConfigRepository
	cache map[string]*entity.SystemConfig // 内存缓存
	mu    sync.RWMutex                    // 读写锁保护缓存
}

// NewSystemConfigService 创建系统配置服务实例
func NewSystemConfigService(repo repository.SystemConfigRepository) *SystemConfigService {
	svc := &SystemConfigService{
		repo:  repo,
		cache: make(map[string]*entity.SystemConfig),
	}

	// 启动时加载缓存
	ctx := context.Background()
	if err := svc.loadCache(ctx); err != nil {
		hlog.Errorf("系统配置服务初始化失败，无法加载缓存: %v", err)
	}

	return svc
}

// ==================== 缓存管理 ====================

// loadCache 从数据库加载所有启用的配置到内存缓存
func (s *SystemConfigService) loadCache(ctx context.Context) error {
	configs, err := s.repo.FindAllActive(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 清空旧缓存
	s.cache = make(map[string]*entity.SystemConfig)

	// 加载新缓存
	for _, config := range configs {
		s.cache[config.ConfigKey] = config
	}

	hlog.Infof("系统配置缓存已加载，共 %d 个配置项", len(s.cache))
	return nil
}

// RefreshCache 刷新缓存
// 手动触发缓存刷新，通常在配置更新后调用
func (s *SystemConfigService) RefreshCache(ctx context.Context) error {
	return s.loadCache(ctx)
}

// ==================== 配置读取方法 ====================

// GetConfig 根据配置键获取配置（从缓存读取）
func (s *SystemConfigService) GetConfig(key string) (*entity.SystemConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.cache[key]
	return config, exists
}

// GetConfigValue 获取配置值（字符串类型）
// 如果配置不存在或未启用，返回默认值
func (s *SystemConfigService) GetConfigValue(key string, defaultValue string) string {
	config, exists := s.GetConfig(key)
	if !exists || !config.IsActive {
		return defaultValue
	}
	return config.ConfigValue
}

// GetIntValue 获取整数类型的配置值
// 如果配置不存在、未启用或转换失败，返回默认值
func (s *SystemConfigService) GetIntValue(key string, defaultValue int) int {
	config, exists := s.GetConfig(key)
	if !exists || !config.IsActive {
		return defaultValue
	}
	return config.GetIntValue(defaultValue)
}

// GetInt64Value 获取 int64 类型的配置值
// 如果配置不存在、未启用或转换失败，返回默认值
func (s *SystemConfigService) GetInt64Value(key string, defaultValue int64) int64 {
	config, exists := s.GetConfig(key)
	if !exists || !config.IsActive {
		return defaultValue
	}
	return config.GetInt64Value(defaultValue)
}

// GetBoolValue 获取布尔类型的配置值
// 如果配置不存在、未启用或转换失败，返回默认值
func (s *SystemConfigService) GetBoolValue(key string, defaultValue bool) bool {
	config, exists := s.GetConfig(key)
	if !exists || !config.IsActive {
		return defaultValue
	}
	return config.GetBoolValue(defaultValue)
}

// GetFloatValue 获取浮点数类型的配置值
// 如果配置不存在、未启用或转换失败，返回默认值
func (s *SystemConfigService) GetFloatValue(key string, defaultValue float64) float64 {
	config, exists := s.GetConfig(key)
	if !exists || !config.IsActive {
		return defaultValue
	}
	return config.GetFloatValue(defaultValue)
}

// ==================== 预定义配置项的便捷方法 ====================

// GetLongPollingTimeout 获取长轮询超时时间（秒）
func (s *SystemConfigService) GetLongPollingTimeout() int {
	return s.GetIntValue(ConfigKeyLongPollingTimeout, DefaultLongPollingTimeout)
}

// GetLongPollingTimeoutDuration 获取长轮询超时时间（Duration）
func (s *SystemConfigService) GetLongPollingTimeoutDuration() time.Duration {
	seconds := s.GetLongPollingTimeout()
	return time.Duration(seconds) * time.Second
}

// GetLongPollingMaxWait 获取长轮询最大等待时间（秒）
func (s *SystemConfigService) GetLongPollingMaxWait() int {
	return s.GetIntValue(ConfigKeyLongPollingMaxWait, DefaultLongPollingMaxWait)
}

// GetLongPollingMaxWaitDuration 获取长轮询最大等待时间（Duration）
func (s *SystemConfigService) GetLongPollingMaxWaitDuration() time.Duration {
	seconds := s.GetLongPollingMaxWait()
	return time.Duration(seconds) * time.Second
}

// GetMaxSubscriptions 获取最大订阅数
func (s *SystemConfigService) GetMaxSubscriptions() int {
	return s.GetIntValue(ConfigKeyMaxSubscriptions, DefaultMaxSubscriptions)
}

// GetHeartbeatInterval 获取心跳间隔（秒）
func (s *SystemConfigService) GetHeartbeatInterval() int {
	return s.GetIntValue(ConfigKeyHeartbeatInterval, DefaultHeartbeatInterval)
}

// GetHeartbeatIntervalDuration 获取心跳间隔（Duration）
func (s *SystemConfigService) GetHeartbeatIntervalDuration() time.Duration {
	seconds := s.GetHeartbeatInterval()
	return time.Duration(seconds) * time.Second
}

// GetHeartbeatTimeout 获取心跳超时（秒）
func (s *SystemConfigService) GetHeartbeatTimeout() int {
	return s.GetIntValue(ConfigKeyHeartbeatTimeout, DefaultHeartbeatTimeout)
}

// GetHeartbeatTimeoutDuration 获取心跳超时（Duration）
func (s *SystemConfigService) GetHeartbeatTimeoutDuration() time.Duration {
	seconds := s.GetHeartbeatTimeout()
	return time.Duration(seconds) * time.Second
}

// ==================== 配置管理方法（可选，用于内部管理） ====================

// SetConfig 设置或更新配置
// 如果配置已存在则更新，否则创建新配置
func (s *SystemConfigService) SetConfig(ctx context.Context, key, value, description string) error {
	// 检查配置是否存在
	existingConfig, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return err
	}

	if existingConfig != nil {
		// 更新已有配置
		existingConfig.UpdateValue(value)
		existingConfig.Description = description
		if err := s.repo.Update(ctx, existingConfig); err != nil {
			return err
		}
	} else {
		// 创建新配置
		newConfig := &entity.SystemConfig{
			ConfigKey:   key,
			ConfigValue: value,
			Description: description,
			IsActive:    true,
		}
		if err := s.repo.Create(ctx, newConfig); err != nil {
			return err
		}
	}

	// 刷新缓存
	return s.RefreshCache(ctx)
}

// UpdateConfigValue 更新配置值
func (s *SystemConfigService) UpdateConfigValue(ctx context.Context, key, value string) error {
	if err := s.repo.UpdateValue(ctx, key, value); err != nil {
		return err
	}

	// 刷新缓存
	return s.RefreshCache(ctx)
}

// ActivateConfig 启用配置
func (s *SystemConfigService) ActivateConfig(ctx context.Context, key string) error {
	if err := s.repo.UpdateActive(ctx, key, true); err != nil {
		return err
	}

	// 刷新缓存
	return s.RefreshCache(ctx)
}

// DeactivateConfig 禁用配置
func (s *SystemConfigService) DeactivateConfig(ctx context.Context, key string) error {
	if err := s.repo.UpdateActive(ctx, key, false); err != nil {
		return err
	}

	// 刷新缓存
	return s.RefreshCache(ctx)
}

// GetAllConfigs 获取所有配置（从数据库查询，不使用缓存）
func (s *SystemConfigService) GetAllConfigs(ctx context.Context) ([]*entity.SystemConfig, error) {
	return s.repo.FindAll(ctx)
}
