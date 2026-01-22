package service

import (
	"context"
	"fmt"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/dto/vo"
	"config-client/config/domain/repository"
	domainService "config-client/config/domain/service"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// LongPollingAppService 长轮询应用服务
type LongPollingAppService struct {
	longPollingService *domainService.LongPollingService
	configRepo         repository.ConfigRepository
}

// NewLongPollingAppService 创建长轮询应用服务
func NewLongPollingAppService(
	longPollingService *domainService.LongPollingService,
	configRepo repository.ConfigRepository,
) *LongPollingAppService {
	return &LongPollingAppService{
		longPollingService: longPollingService,
		configRepo:         configRepo,
	}
}

// WaitForChanges 等待配置变更
func (s *LongPollingAppService) WaitForChanges(ctx context.Context, req *request.LongPollingRequest) (*vo.LongPollingResponse, error) {
	// 1. 转换请求参数
	configKeys := make([]string, len(req.ConfigKeys))
	versions := make(map[string]string)
	var namespaceID int
	var environment string

	for i, item := range req.ConfigKeys {
		configKey := fmt.Sprintf("%d:%s", item.NamespaceID, item.ConfigKey)
		configKeys[i] = configKey
		versions[configKey] = item.Version

		// 记录第一个配置的命名空间和环境 (假设同一批配置在同一命名空间)
		if i == 0 {
			namespaceID = item.NamespaceID
			environment = item.Environment
		}
	}

	// 2. 构建等待请求
	waitReq := &domainService.WaitRequest{
		ClientID:       req.ClientID,
		ClientIP:       req.ClientIP,
		ClientHostname: req.ClientHostname,
		NamespaceID:    namespaceID,
		Environment:    environment,
		ConfigKeys:     configKeys,
		Versions:       versions,
	}

	// 3. 调用领域服务等待变更（传递 context）
	result, err := s.longPollingService.Wait(ctx, waitReq)
	if err != nil {
		return nil, err
	}

	// 4. 如果没有变更，返回未变更响应
	if !result.Changed {
		return &vo.LongPollingResponse{
			Changed:    false,
			ConfigKeys: []string{},
			Configs:    []vo.ConfigChangeDetail{},
		}, nil
	}

	// 5. 如果有变更，获取最新的配置详情
	configs, err := s.getConfigDetails(req.ConfigKeys, result.Versions)
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
func (s *LongPollingAppService) getConfigDetails(
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

		// 从数据库获取完整配置信息
		// 使用配置项中的environment，如果为空则使用默认值
		environment := item.Environment
		if environment == "" {
			environment = "default"
		}

		config, err := s.configRepo.FindByNamespaceAndKey(context.Background(), item.NamespaceID, item.ConfigKey, environment)
		if err != nil {
			hlog.Errorf("获取配置详情失败: namespaceID=%d, key=%s, environment=%s, error=%v", item.NamespaceID, item.ConfigKey, environment, err)
			// 即使获取失败，也返回基础信息
			details = append(details, vo.ConfigChangeDetail{
				NamespaceID: item.NamespaceID,
				ConfigKey:   item.ConfigKey,
				Version:     latestVersion,
				Value:       "",
				ValueType:   "",
			})
			continue
		}

		if config == nil {
			hlog.Warnf("配置不存在: namespaceID=%d, key=%s", item.NamespaceID, item.ConfigKey)
			continue
		}

		details = append(details, vo.ConfigChangeDetail{
			NamespaceID: item.NamespaceID,
			ConfigKey:   item.ConfigKey,
			Version:     latestVersion,
			Value:       config.Value,
			ValueType:   config.ValueType,
		})

		hlog.Infof("配置变更: namespaceID=%d, key=%s, newVersion=%s", item.NamespaceID, item.ConfigKey, latestVersion)
	}

	return details, nil
}
