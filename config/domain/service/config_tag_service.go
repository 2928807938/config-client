package service

import (
	"context"
	"strings"

	"config-client/config/domain/entity"
	"config-client/config/domain/repository"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// ConfigTagService 配置标签服务
// 负责处理配置标签相关的业务逻辑
type ConfigTagService struct {
	tagRepo    repository.ConfigTagRepository
	maskingSvc *MaskingService
}

// NewConfigTagService 创建配置标签服务实例
func NewConfigTagService(
	tagRepo repository.ConfigTagRepository,
	maskingSvc *MaskingService,
) *ConfigTagService {
	return &ConfigTagService{
		tagRepo:    tagRepo,
		maskingSvc: maskingSvc,
	}
}

// ==================== 标签管理 ====================

// AddTags 为配置添加标签
func (s *ConfigTagService) AddTags(ctx context.Context, configID int, inputs []entity.TagInput) error {
	if len(inputs) == 0 {
		return nil
	}

	// 验证标签输入
	for _, input := range inputs {
		if err := input.Validate(); err != nil {
			return err
		}
	}

	// 去重：检查标签是否已存在
	tags := make([]*entity.ConfigTag, 0, len(inputs))
	for _, input := range inputs {
		exists, err := s.tagRepo.ExistsByConfigIDAndTag(ctx, configID, input.TagKey, input.TagValue)
		if err != nil {
			return err
		}

		// 如果标签已存在，跳过
		if exists {
			hlog.CtxWarnf(ctx, "标签已存在，跳过: configID=%d, tag=%s:%s", configID, input.TagKey, input.TagValue)
			continue
		}

		tags = append(tags, &entity.ConfigTag{
			ConfigID: configID,
			TagKey:   input.TagKey,
			TagValue: input.TagValue,
		})
	}

	// 批量创建标签
	if len(tags) > 0 {
		return s.tagRepo.BatchCreate(ctx, tags)
	}

	return nil
}

// RemoveTags 删除配置的标签（根据标签键）
func (s *ConfigTagService) RemoveTags(ctx context.Context, configID int, tagKeys []string) error {
	if len(tagKeys) == 0 {
		return nil
	}

	for _, tagKey := range tagKeys {
		if err := s.tagRepo.DeleteByConfigIDAndTagKey(ctx, configID, tagKey); err != nil {
			hlog.CtxErrorf(ctx, "删除标签失败: configID=%d, tagKey=%s, err=%v", configID, tagKey, err)
			return err
		}
	}

	return nil
}

// RemoveAllTags 删除配置的所有标签
func (s *ConfigTagService) RemoveAllTags(ctx context.Context, configID int) error {
	return s.tagRepo.DeleteByConfigID(ctx, configID)
}

// GetTags 获取配置的所有标签
func (s *ConfigTagService) GetTags(ctx context.Context, configID int) ([]*entity.ConfigTag, error) {
	return s.tagRepo.FindByConfigID(ctx, configID)
}

// UpdateTags 更新配置的标签（全量替换）
func (s *ConfigTagService) UpdateTags(ctx context.Context, configID int, inputs []entity.TagInput) error {
	// 1. 删除现有标签
	if err := s.RemoveAllTags(ctx, configID); err != nil {
		return err
	}

	// 2. 添加新标签
	return s.AddTags(ctx, configID, inputs)
}

// ==================== 自动标签生成 ====================

// AutoGenerateTags 为配置自动生成标签
// 规则:
// 1. 如果是敏感配置，添加 sensitive:true
// 2. 根据 group_name 添加 category 标签
// 3. 根据 value_type 添加类型标签
func (s *ConfigTagService) AutoGenerateTags(ctx context.Context, config *entity.Config) []entity.TagInput {
	tags := make([]entity.TagInput, 0)

	// 1. 敏感配置标签
	if s.maskingSvc.IsSensitiveKey(config.Key) {
		tags = append(tags, entity.TagInput{
			TagKey:   "sensitive",
			TagValue: "true",
		})
	} else {
		tags = append(tags, entity.TagInput{
			TagKey:   "sensitive",
			TagValue: "false",
		})
	}

	// 2. 分类标签（根据group_name）
	if config.GroupName != "" && config.GroupName != "default" {
		tags = append(tags, entity.TagInput{
			TagKey:   "category",
			TagValue: config.GroupName,
		})
	}

	// 3. 类型标签
	if config.ValueType != "" {
		tags = append(tags, entity.TagInput{
			TagKey:   "type",
			TagValue: config.ValueType,
		})
	}

	// 4. 环境标签
	if config.Environment != "" && config.Environment != "default" {
		tags = append(tags, entity.TagInput{
			TagKey:   "environment",
			TagValue: config.Environment,
		})
	}

	// 5. 根据键名推断重要程度
	importance := s.inferImportance(config.Key, config.GroupName)
	if importance != "" {
		tags = append(tags, entity.TagInput{
			TagKey:   "importance",
			TagValue: importance,
		})
	}

	return tags
}

// inferImportance 推断配置的重要程度
func (s *ConfigTagService) inferImportance(key, groupName string) string {
	lowerKey := strings.ToLower(key)
	lowerGroup := strings.ToLower(groupName)

	// 高优先级：数据库、安全相关
	highPriorityKeywords := []string{"database", "db", "security", "auth", "password", "secret", "token"}
	for _, keyword := range highPriorityKeywords {
		if strings.Contains(lowerKey, keyword) || strings.Contains(lowerGroup, keyword) {
			return "high"
		}
	}

	// 中优先级：缓存、API相关
	mediumPriorityKeywords := []string{"cache", "redis", "api", "service", "host", "port"}
	for _, keyword := range mediumPriorityKeywords {
		if strings.Contains(lowerKey, keyword) || strings.Contains(lowerGroup, keyword) {
			return "medium"
		}
	}

	// 低优先级：其他
	return "low"
}

// ==================== 按标签查询配置 ====================

// QueryConfigIDsByTags 根据标签查询配置ID列表
func (s *ConfigTagService) QueryConfigIDsByTags(ctx context.Context, tags []entity.TagInput) ([]int, error) {
	if len(tags) == 0 {
		return []int{}, nil
	}

	return s.tagRepo.FindConfigIDsByTags(ctx, tags)
}

// FindByTagKeyValue 根据标签键值查询标签
func (s *ConfigTagService) FindByTagKeyValue(ctx context.Context, tagKey, tagValue string) ([]*entity.ConfigTag, error) {
	return s.tagRepo.FindByTagKeyValue(ctx, tagKey, tagValue)
}

// ==================== 标签验证 ====================

// HasTag 检查配置是否包含指定标签
func (s *ConfigTagService) HasTag(ctx context.Context, configID int, tagKey, tagValue string) (bool, error) {
	return s.tagRepo.ExistsByConfigIDAndTag(ctx, configID, tagKey, tagValue)
}

// IsSensitiveConfig 检查配置是否为敏感配置（根据标签）
func (s *ConfigTagService) IsSensitiveConfig(ctx context.Context, configID int) (bool, error) {
	return s.HasTag(ctx, configID, "sensitive", "true")
}

// GetImportance 获取配置的重要程度
func (s *ConfigTagService) GetImportance(ctx context.Context, configID int) (string, error) {
	tags, err := s.GetTags(ctx, configID)
	if err != nil {
		return "", err
	}

	for _, tag := range tags {
		if tag.IsImportanceTag() {
			return tag.TagValue, nil
		}
	}

	return "unknown", nil
}

// GetCategory 获取配置的分类
func (s *ConfigTagService) GetCategory(ctx context.Context, configID int) (string, error) {
	tags, err := s.GetTags(ctx, configID)
	if err != nil {
		return "", err
	}

	for _, tag := range tags {
		if tag.IsCategoryTag() {
			return tag.TagValue, nil
		}
	}

	return "unknown", nil
}
