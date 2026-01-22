package service

import (
	"context"

	"config-client/api/config-api/converter"
	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/entity"
	"config-client/config/domain/repository"
	domainService "config-client/config/domain/service"
)

// ConfigAppService 配置应用服务
// 负责协调领域服务和数据转换，不包含业务逻辑和异常处理
// 异常由领域服务捕获并向上传递，最终由统一异常处理器处理
type ConfigAppService struct {
	configDomainService *domainService.ConfigService
	converter           *converter.ConfigConverter
}

// NewConfigAppService 创建配置应用服务实例
func NewConfigAppService(
	configDomainService *domainService.ConfigService,
	converter *converter.ConfigConverter,
) *ConfigAppService {
	return &ConfigAppService{
		configDomainService: configDomainService,
		converter:           converter,
	}
}

// CreateConfig 创建配置
func (s *ConfigAppService) CreateConfig(ctx context.Context, req *request.CreateConfigRequest) (*vo.ConfigVO, error) {
	// 1. 将请求DTO转换为领域实体
	// 处理 Metadata：如果为空字符串，使用默认值 "{}"
	metadata := req.Metadata
	if metadata == "" {
		metadata = "{}"
	}

	config := &entity.Config{
		NamespaceID: req.NamespaceID,
		Key:         req.Key,
		Value:       req.Value,
		GroupName:   req.GroupName,
		ValueType:   req.ValueType,
		Environment: req.Environment,
		Description: req.Description,
		Metadata:    metadata,
	}
	// 设置审计字段
	config.CreatedBy = req.CreatedBy
	config.UpdatedBy = req.CreatedBy // 创建时，更新人同创建人

	// 2. 调用领域服务创建配置（错误直接向上传递）
	if err := s.configDomainService.CreateConfig(ctx, config); err != nil {
		return nil, err
	}

	// 3. 将领域实体转换为VO返回
	return s.converter.ToVO(config), nil
}

// UpdateConfig 更新配置
func (s *ConfigAppService) UpdateConfig(ctx context.Context, configID int, req *request.UpdateConfigRequest) (*vo.ConfigVO, error) {
	// 1. 先查询现有配置以获取完整信息
	existingConfig, err := s.configDomainService.GetByID(ctx, configID)
	if err != nil {
		return nil, err
	}

	// 2. 构建更新实体（保留原有的关键字段）
	// 处理 Metadata：如果为空字符串，保留原值或使用默认值 "{}"
	metadata := req.Metadata
	if metadata == "" {
		if existingConfig.Metadata != "" {
			metadata = existingConfig.Metadata
		} else {
			metadata = "{}"
		}
	}

	config := &entity.Config{
		NamespaceID: existingConfig.NamespaceID,
		Key:         existingConfig.Key,
		Environment: existingConfig.Environment,
		Value:       req.Value,
		GroupName:   req.GroupName,
		ValueType:   req.ValueType,
		Description: req.Description,
		Metadata:    metadata,
		IsActive:    boolValue(req.IsActive, existingConfig.IsActive),
		IsReleased:  boolValue(req.IsReleased, existingConfig.IsReleased),
	}
	config.ID = configID
	config.UpdatedBy = req.UpdatedBy

	// 3. 调用领域服务更新配置（错误直接向上传递）
	if err := s.configDomainService.UpdateConfig(ctx, config); err != nil {
		return nil, err
	}

	// 4. 重新查询最新配置并返回
	updatedConfig, err := s.configDomainService.GetByID(ctx, configID)
	if err != nil {
		return nil, err
	}

	return s.converter.ToVO(updatedConfig), nil
}

// QueryConfigs 分页查询配置
func (s *ConfigAppService) QueryConfigs(ctx context.Context, req *request.QueryConfigRequest) (*vo.ConfigListVO, error) {
	// 1. 设置默认值
	req.SetDefaults()

	// 2. 将 DTO 转换为仓储层查询参数
	params := &repository.ConfigQueryParams{
		NamespaceID: req.NamespaceID,
		Key:         req.Key,
		GroupName:   req.GroupName,
		Environment: req.Environment,
		IsActive:    req.IsActive,
		IsReleased:  req.IsReleased,
		ValueType:   req.ValueType,
		Page:        req.Page,
		Size:        req.Size,
		OrderBy:     req.OrderBy,
	}

	// 3. 调用领域服务查询配置（错误直接向上传递）
	pageResult, err := s.configDomainService.QueryConfigs(ctx, params)
	if err != nil {
		return nil, err
	}

	// 4. 转换为VO返回
	return s.converter.ToListVO(pageResult.Items, pageResult.Total, pageResult.Page, pageResult.Size), nil
}

// GetConfigByID 根据ID获取配置
func (s *ConfigAppService) GetConfigByID(ctx context.Context, configID int) (*vo.ConfigVO, error) {
	// 1. 调用领域服务获取配置（错误直接向上传递）
	config, err := s.configDomainService.GetByID(ctx, configID)
	if err != nil {
		return nil, err
	}

	// 2. 转换为VO返回
	return s.converter.ToVO(config), nil
}

// DeleteConfig 删除配置（逻辑删除）
func (s *ConfigAppService) DeleteConfig(ctx context.Context, configID int) error {
	// 直接调用领域服务删除配置（错误直接向上传递）
	return s.configDomainService.DeleteConfig(ctx, configID)
}

// ==================== 辅助函数 ====================

// boolValue 获取布尔指针的值，如果为nil则返回默认值
func boolValue(ptr *bool, defaultValue bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
