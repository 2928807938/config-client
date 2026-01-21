package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/dto/vo"
	domainService "config-client/config/domain/service"
	"config-client/config/infrastructure/repository"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// LongPollingService 长轮询应用服务
type LongPollingService struct {
	manager       *LongPollingManager
	configService *domainService.ConfigService
	configRepo    repository.ConfigRepository
}

// NewLongPollingService 创建长轮询应用服务
func NewLongPollingService(
	manager *LongPollingManager,
	configService *domainService.ConfigService,
	configRepo repository.ConfigRepository,
) *LongPollingService {
	return &LongPollingService{
		manager:       manager,
		configService: configService,
		configRepo:    configRepo,
	}
}

// WaitForChanges 等待配置变更
func (s *LongPollingService) WaitForChanges(ctx context.Context, req *request.LongPollingRequest) (*vo.LongPollingResponse, error) {
	// 1. 转换请求参数
	configKeys := make([]string, len(req.ConfigKeys))
	versions := make(map[string]string)
	for i, item := range req.ConfigKeys {
		configKey := fmt.Sprintf("%d:%s", item.NamespaceID, item.ConfigKey)
		configKeys[i] = configKey
		versions[configKey] = item.Version
	}

	// 2. 调用长轮询管理器等待变更
	result, err := s.manager.Wait(configKeys, versions)
	if err != nil {
		return nil, err
	}

	// 3. 如果没有变更，返回304
	if !result.Changed {
		return &vo.LongPollingResponse{
			Changed:    false,
			ConfigKeys: []string{},
			Configs:    []vo.ConfigChangeDetail{},
		}, nil
	}

	// 4. 如果有变更，获取最新的配置详情
	configs, err := s.getConfigDetails(ctx, req.ConfigKeys, result.Versions)
	if err != nil {
		return nil, err
	}

	return &vo.LongPollingResponse{
		Changed:    true,
		ConfigKeys: result.ConfigKeys,
		Configs:    configs,
	}, nil
}

// getConfigDetails 获取配置详情
func (s *LongPollingService) getConfigDetails(
	ctx context.Context,
	requestKeys []request.ConfigKeyVersion,
	latestVersions map[string]string,
) ([]vo.ConfigChangeDetail, error) {
	var details []vo.ConfigChangeDetail

	for _, item := range requestKeys {
		configKey := fmt.Sprintf("%d:%s", item.NamespaceID, item.ConfigKey)

		// 检查这个配置是否有变更
		latestVersion := latestVersions[configKey]
		if latestVersion == item.Version {
			// 没有变更，跳过
			continue
		}

		// 获取配置详情
		config, err := s.configRepo.FindByNamespaceAndKey(ctx, int(item.NamespaceID), item.ConfigKey, "")
		if err != nil {
			hlog.Errorf("获取配置详情失败: namespaceID=%d, key=%s, error=%v", item.NamespaceID, item.ConfigKey, err)
			continue
		}
		if config == nil {
			hlog.Warnf("配置不存在: namespaceID=%d, key=%s", item.NamespaceID, item.ConfigKey)
			continue
		}

		// 添加到结果列表
		details = append(details, vo.ConfigChangeDetail{
			NamespaceID: item.NamespaceID,
			ConfigKey:   item.ConfigKey,
			Version:     latestVersion,
			Value:       config.Value,
			ValueType:   config.ValueType,
		})
	}

	return details, nil
}

// GetConfigVersion 获取配置的当前版本号
// 这个方法供LongPollingManager调用
func GetConfigVersionFunc(configRepo repository.ConfigRepository) func(string) (string, error) {
	return func(configKey string) (string, error) {
		// 解析configKey (格式: "namespaceID:key")
		parts := strings.Split(configKey, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("无效的配置键格式: %s", configKey)
		}

		namespaceID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return "", fmt.Errorf("解析命名空间ID失败: %s", parts[0])
		}

		key := parts[1]

		// 查询配置
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		config, err := configRepo.FindByNamespaceAndKey(ctx, int(namespaceID), key, "")
		if err != nil {
			return "", err
		}
		if config == nil {
			return "", fmt.Errorf("配置不存在: %s", configKey)
		}

		// 计算版本号（使用MD5）
		return ComputeConfigVersion(config.Value), nil
	}
}
