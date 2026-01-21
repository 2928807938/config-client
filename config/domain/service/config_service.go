package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"config-client/config/domain/constants"
	"config-client/config/domain/entity"
	domainErrors "config-client/config/domain/errors"
	"config-client/config/domain/listener"
	"config-client/config/domain/repository"
	shareRepo "config-client/share/repository"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gopkg.in/yaml.v3"
)

// ConfigService 配置领域服务
// 负责处理配置相关的复杂业务逻辑，包括：
// - 配置的创建、更新、发布等业务流程
// - 配置的验证和校验
// - 配置的版本管理
// - 配置的哈希计算和验证
type ConfigService struct {
	configRepo repository.ConfigRepository
	listener   listener.ConfigListener // 配置变更监听器（可选）
}

// NewConfigService 创建配置领域服务实例
func NewConfigService(configRepo repository.ConfigRepository, listener listener.ConfigListener) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
		listener:   listener,
	}
}

// CreateConfig 创建配置
// 业务规则：
// 1. 配置键不能为空，且必须符合命名规范
// 2. 同一命名空间+环境下，配置键必须唯一
// 3. 自动计算并设置内容哈希
func (s *ConfigService) CreateConfig(ctx context.Context, config *entity.Config) error {
	// 1. 验证配置有效性
	if err := s.ValidateConfig(ctx, config); err != nil {
		return err
	}

	// 2. 检查配置是否已存在
	exists, err := s.configRepo.ExistsByNamespaceAndKey(ctx, config.NamespaceID, config.Key, config.Environment)
	if err != nil {
		return err
	}
	if exists {
		return domainErrors.ErrConfigAlreadyExists(config.Key, config.Environment)
	}

	// 3. 计算内容哈希
	hash, err := s.ComputeContentHash(config.Value, constants.HashAlgorithmMD5)
	if err != nil {
		return err
	}
	config.ContentHash = hash
	config.ContentHashAlgorithm = constants.HashAlgorithmMD5

	// 4. 设置默认值
	if config.GroupName == "" {
		config.GroupName = constants.DefaultGroupName
	}
	config.IsReleased = false // 新创建的配置默认未发布
	config.IsActive = true    // 新创建的配置默认激活

	// 5. 保存配置
	if err := s.configRepo.Create(ctx, config); err != nil {
		return err
	}

	// 6. 发布配置变更事件
	s.publishConfigChangeEvent(ctx, &listener.ConfigChangeEvent{
		NamespaceID: config.NamespaceID,
		ConfigKey:   config.Key,
		ConfigID:    int64(config.ID),
		Action:      "create",
	})

	return nil
}

// UpdateConfig 更新配置
// 业务规则：
// 1. 配置必须存在
// 2. 已发布的配置不能直接修改，需要先取消发布
// 3. 更新时自动计算新的内容哈希
// 4. 自动增加版本号
func (s *ConfigService) UpdateConfig(ctx context.Context, config *entity.Config) error {
	// 1. 检查配置是否存在
	existingConfig, err := s.configRepo.GetByID(ctx, config.ID)
	if err != nil {
		return err
	}
	if existingConfig == nil {
		return domainErrors.ErrConfigNotFound(config.Key, config.Environment)
	}

	// 2. 已发布的配置不能直接修改
	if existingConfig.IsReleased {
		return domainErrors.ErrConfigAlreadyReleased(config.Key)
	}

	// 3. 验证配置有效性
	if err := s.ValidateConfig(ctx, config); err != nil {
		return err
	}

	// 4. 重新计算内容哈希
	hash, err := s.ComputeContentHash(config.Value, constants.HashAlgorithmMD5)
	if err != nil {
		return err
	}

	// 5. 使用领域实体的方法更新配置值
	existingConfig.UpdateValue(config.Value, hash)
	existingConfig.Description = config.Description
	existingConfig.Metadata = config.Metadata
	existingConfig.GroupName = config.GroupName
	existingConfig.ValueType = config.ValueType

	// 6. 保存更新
	if err := s.configRepo.Update(ctx, existingConfig); err != nil {
		return err
	}

	// 7. 发布配置变更事件
	s.publishConfigChangeEvent(ctx, &listener.ConfigChangeEvent{
		NamespaceID: existingConfig.NamespaceID,
		ConfigKey:   existingConfig.Key,
		ConfigID:    int64(existingConfig.ID),
		Action:      "update",
	})

	return nil
}

// ReleaseConfig 发布配置
// 业务规则：
// 1. 配置必须存在且已激活
// 2. 配置值必须有效
// 3. 发布后配置不可修改（需先取消发布）
func (s *ConfigService) ReleaseConfig(ctx context.Context, configID int) error {
	// 1. 查询配置
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return err
	}
	if config == nil {
		return domainErrors.ErrConfigNotFound("", "")
	}

	// 2. 检查配置是否已激活
	if !config.IsActive {
		return domainErrors.ErrConfigNotActive(config.Key)
	}

	// 3. 验证配置有效性
	if err := s.ValidateConfig(ctx, config); err != nil {
		return err
	}

	// 4. 验证内容哈希
	if err := s.VerifyContentHash(ctx, config); err != nil {
		return err
	}

	// 5. 发布配置（使用领域实体的方法）
	config.Release()

	// 6. 保存更新
	return s.configRepo.Update(ctx, config)
}

// UnreleaseConfig 取消发布配置
// 业务规则：
// 1. 配置必须存在
// 2. 配置必须是已发布状态
func (s *ConfigService) UnreleaseConfig(ctx context.Context, configID int) error {
	// 1. 查询配置
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return err
	}
	if config == nil {
		return domainErrors.ErrConfigNotFound("", "")
	}

	// 2. 检查配置是否已发布
	if !config.IsReleased {
		return domainErrors.ErrConfigNotReleased(config.Key)
	}

	// 3. 取消发布（使用领域实体的方法）
	config.Unrelease()

	// 4. 保存更新
	return s.configRepo.Update(ctx, config)
}

// DeleteConfig 删除配置（软删除）
// 业务规则：
// 1. 配置必须存在
// 2. 已发布的配置不能删除，需要先取消发布
// 3. 执行软删除
func (s *ConfigService) DeleteConfig(ctx context.Context, configID int) error {
	// 1. 检查配置是否存在
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return err
	}
	if config == nil {
		return domainErrors.ErrConfigNotFound("", "")
	}

	// 2. 检查配置是否已发布
	if config.IsReleased {
		return domainErrors.ErrConfigCannotDelete(config.Key)
	}

	// 3. 执行软删除
	if err := s.configRepo.Delete(ctx, configID); err != nil {
		return err
	}

	// 4. 发布配置变更事件
	s.publishConfigChangeEvent(ctx, &listener.ConfigChangeEvent{
		NamespaceID: config.NamespaceID,
		ConfigKey:   config.Key,
		ConfigID:    int64(config.ID),
		Action:      "delete",
	})

	return nil
}

// ValidateConfig 验证配置的有效性
// 验证规则：
// 1. 配置键符合命名规范（字母、数字、下划线、中划线、点号）
// 2. 配置值根据 ValueType 进行类型验证
// 3. 环境参数必须有效（dev/test/uat/prod）
func (s *ConfigService) ValidateConfig(ctx context.Context, config *entity.Config) error {
	// 1. 验证配置键
	if config.Key == "" {
		return domainErrors.ErrConfigKeyEmpty()
	}

	// 配置键只能包含字母、数字、下划线、中划线、点号
	if !isValidConfigKey(config.Key) {
		return domainErrors.ErrConfigKeyFormatInvalid(config.Key)
	}

	// 2. 验证环境参数
	if !contains(constants.ValidEnvironments, config.Environment) {
		return domainErrors.ErrConfigEnvironmentInvalid(config.Environment)
	}

	// 3. 验证 ValueType 是否有效
	if config.ValueType == "" {
		config.ValueType = constants.DefaultValueType // 如果未指定，使用默认类型
	}

	// 4. 验证配置值（基础检查）
	if config.Value == "" && config.ValueType != constants.ValueTypeString {
		return domainErrors.ErrConfigValueEmpty(config.Key)
	}

	// 5. 根据 ValueType 进行详细的类型验证
	if err := validateValueByType(config.Value, config.ValueType); err != nil {
		return err
	}

	return nil
}

// ComputeContentHash 计算配置内容的哈希值
// 用于检测配置内容是否被篡改
func (s *ConfigService) ComputeContentHash(value string, algorithm string) (string, error) {
	switch algorithm {
	case constants.HashAlgorithmMD5:
		hash := md5.Sum([]byte(value))
		return hex.EncodeToString(hash[:]), nil
	default:
		return "", domainErrors.ErrUnsupportedHashAlgorithm(algorithm)
	}
}

// VerifyContentHash 验证配置内容哈希
// 比较配置的实际哈希值与存储的哈希值是否一致
func (s *ConfigService) VerifyContentHash(ctx context.Context, config *entity.Config) error {
	// 重新计算哈希
	actualHash, err := s.ComputeContentHash(config.Value, config.ContentHashAlgorithm)
	if err != nil {
		return err
	}

	// 比较哈希值
	if actualHash != config.ContentHash {
		return domainErrors.ErrConfigHashMismatch(config.Key, config.ContentHash, actualHash)
	}

	return nil
}

// BatchReleaseConfigs 批量发布配置
// 用于批量发布某个命名空间或分组下的多个配置
func (s *ConfigService) BatchReleaseConfigs(ctx context.Context, namespaceID int, environment string, groupName string) ([]int, error) {
	// 1. 查询符合条件的配置
	var configs []*entity.Config
	var err error

	if groupName != "" {
		// 按分组查询
		configs, err = s.configRepo.FindByGroup(ctx, namespaceID, groupName)
	} else {
		// 查询整个命名空间
		configs, err = s.configRepo.FindByNamespace(ctx, namespaceID)
	}

	if err != nil {
		return nil, err
	}

	// 2. 过滤环境
	var targetConfigs []*entity.Config
	for _, config := range configs {
		if config.Environment == environment && config.IsActive && !config.IsReleased {
			targetConfigs = append(targetConfigs, config)
		}
	}

	// 3. 批量发布
	var releasedIDs []int
	for _, config := range targetConfigs {
		if err := s.ReleaseConfig(ctx, config.ID); err == nil {
			releasedIDs = append(releasedIDs, config.ID)
		}
		// 忽略单个配置发布失败的错误，继续处理其他配置
	}

	return releasedIDs, nil
}

// GetActiveConfig 获取激活的配置
// 业务规则：
// 1. 配置必须存在
// 2. 配置必须已发布且已激活
func (s *ConfigService) GetActiveConfig(ctx context.Context, namespaceID int, key string, environment string) (*entity.Config, error) {
	// 1. 查询配置
	config, err := s.configRepo.FindByNamespaceAndKey(ctx, namespaceID, key, environment)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domainErrors.ErrConfigNotFound(key, environment)
	}

	// 2. 检查配置是否已发布
	if !config.IsReleased {
		return nil, domainErrors.ErrConfigNotReleased(key)
	}

	// 3. 检查配置是否已激活
	if !config.IsActive {
		return nil, domainErrors.ErrConfigNotActive(key)
	}

	return config, nil
}

// GetByID 根据ID获取配置
// 简单的查询方法，不做业务验证
func (s *ConfigService) GetByID(ctx context.Context, id int) (*entity.Config, error) {
	config, err := s.configRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domainErrors.ErrConfigNotFound("", "")
	}
	return config, nil
}

// QueryConfigs 根据查询参数分页查询配置
// 直接委托给仓储层处理字段映射和查询逻辑
func (s *ConfigService) QueryConfigs(ctx context.Context, params *repository.ConfigQueryParams) (*shareRepo.PageResult[*entity.Config], error) {
	return s.configRepo.QueryByParams(ctx, params)
}

// ==================== 辅助函数 ====================

// isValidConfigKey 验证配置键是否符合命名规范
func isValidConfigKey(key string) bool {
	if key == "" {
		return false
	}

	// 只允许字母、数字、下划线、中划线、点号
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_' || c == '-' || c == '.') {
			return false
		}
	}

	return true
}

// contains 检查字符串切片是否包含指定元素
func contains(slice []string, item string) bool {
	item = strings.ToLower(strings.TrimSpace(item))
	for _, s := range slice {
		if strings.ToLower(strings.TrimSpace(s)) == item {
			return true
		}
	}
	return false
}

// ==================== 值类型验证函数 ====================

// validateValueByType 根据类型验证配置值
func validateValueByType(value string, valueType string) error {
	switch valueType {
	case constants.ValueTypeString:
		return validateStringValue(value)
	case constants.ValueTypeInt:
		return validateIntValue(value)
	case constants.ValueTypeBool:
		return validateBoolValue(value)
	case constants.ValueTypeFloat:
		return validateFloatValue(value)
	case constants.ValueTypeJSON:
		return validateJSONValue(value)
	case constants.ValueTypeYAML:
		return validateYAMLValue(value)
	default:
		// 如果没有指定类型或类型不在预定义列表中，默认按 string 处理
		return nil
	}
}

// validateStringValue 验证字符串类型的值
// 字符串类型接受任何值（包括空字符串）
func validateStringValue(value string) error {
	// 字符串类型无特殊验证要求
	return nil
}

// validateIntValue 验证整数类型的值
func validateIntValue(value string) error {
	if value == "" {
		return domainErrors.ErrConfigValueTypeInvalid("int", "值不能为空")
	}

	// 尝试解析为整数
	_, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return domainErrors.ErrConfigValueTypeInvalid("int", "无法解析为整数: "+err.Error())
	}

	return nil
}

// validateBoolValue 验证布尔类型的值
func validateBoolValue(value string) error {
	if value == "" {
		return domainErrors.ErrConfigValueTypeInvalid("bool", "值不能为空")
	}

	// 支持的布尔值：true, false, 1, 0, yes, no, on, off（不区分大小写）
	normalized := strings.ToLower(strings.TrimSpace(value))
	validBools := []string{"true", "false", "1", "0", "yes", "no", "on", "off"}

	for _, valid := range validBools {
		if normalized == valid {
			return nil
		}
	}

	return domainErrors.ErrConfigValueTypeInvalid("bool", "值必须是 true/false、1/0、yes/no 或 on/off 之一")
}

// validateFloatValue 验证浮点数类型的值
func validateFloatValue(value string) error {
	if value == "" {
		return domainErrors.ErrConfigValueTypeInvalid("float", "值不能为空")
	}

	// 尝试解析为浮点数
	_, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return domainErrors.ErrConfigValueTypeInvalid("float", "无法解析为浮点数: "+err.Error())
	}

	return nil
}

// validateJSONValue 验证JSON格式的值
func validateJSONValue(value string) error {
	if value == "" {
		return domainErrors.ErrConfigValueTypeInvalid("json", "值不能为空")
	}

	// 尝试解析为 JSON
	var js interface{}
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		return domainErrors.ErrConfigValueTypeInvalid("json", "无效的JSON格式: "+err.Error())
	}

	return nil
}

// validateYAMLValue 验证YAML格式的值
func validateYAMLValue(value string) error {
	if value == "" {
		return domainErrors.ErrConfigValueTypeInvalid("yaml", "值不能为空")
	}

	// 尝试解析为 YAML
	var ym interface{}
	if err := yaml.Unmarshal([]byte(value), &ym); err != nil {
		return domainErrors.ErrConfigValueTypeInvalid("yaml", "无效的YAML格式: "+err.Error())
	}

	return nil
}

// publishConfigChangeEvent 发布配置变更事件
func (s *ConfigService) publishConfigChangeEvent(ctx context.Context, event *listener.ConfigChangeEvent) {
	if s.listener == nil {
		return
	}

	// 异步发布事件，不阻塞主流程
	go func() {
		if err := s.listener.Publish(ctx, event); err != nil {
			hlog.Errorf("发布配置变更事件失败: %v, event: %+v", err, event)
		}
	}()
}
