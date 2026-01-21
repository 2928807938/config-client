package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"time"

	"config-client/config/domain/errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// WaitRequest 等待请求
type WaitRequest struct {
	ClientID       string            // 客户端ID
	ClientIP       string            // 客户端IP
	ClientHostname string            // 客户端主机名
	NamespaceID    int               // 命名空间ID
	Environment    string            // 环境
	ConfigKeys     []string          // 配置键列表 (格式: "namespaceID:configKey")
	Versions       map[string]string // 配置键 -> 版本号映射
}

// WaitResult 等待结果
type WaitResult struct {
	Changed    bool              // 是否有变更
	ConfigKeys []string          // 变更的配置键列表
	Versions   map[string]string // 最新版本号映射
}

// LongPollingService 长轮询领域服务
// 重构后的服务，委托订阅管理器处理订阅逻辑
type LongPollingService struct {
	subscriptionMgr *SubscriptionManager // 订阅管理器
	systemConfigSvc *SystemConfigService // 系统配置服务（可选）
	defaultTimeout  time.Duration        // 默认长轮询超时时间（用于向后兼容）
}

// NewLongPollingService 创建长轮询领域服务
// 如果传入 systemConfigSvc，则从系统配置读取超时时间
// 否则使用传入的 timeout 参数
func NewLongPollingService(
	subscriptionMgr *SubscriptionManager,
	timeout time.Duration,
	systemConfigSvc *SystemConfigService,
) *LongPollingService {
	return &LongPollingService{
		subscriptionMgr: subscriptionMgr,
		systemConfigSvc: systemConfigSvc,
		defaultTimeout:  timeout,
	}
}

// Start 启动长轮询服务
// 注意: 订阅管理器需要单独启动
func (s *LongPollingService) Start() error {
	hlog.Info("长轮询服务已启动")
	return nil
}

// Stop 停止长轮询服务
// 注意: 订阅管理器需要单独停止
func (s *LongPollingService) Stop() error {
	hlog.Info("长轮询服务已停止")
	return nil
}

// getTimeout 获取长轮询超时时间
// 优先从系统配置读取，如果系统配置服务未注入或配置不存在，则使用默认值
func (s *LongPollingService) getTimeout() time.Duration {
	if s.systemConfigSvc != nil {
		return s.systemConfigSvc.GetLongPollingTimeoutDuration()
	}
	return s.defaultTimeout
}

// Wait 等待配置变更
// ctx: 请求上下文（用于取消等待）
// req: 等待请求
// 返回: 变更结果或超时
func (s *LongPollingService) Wait(ctx context.Context, req *WaitRequest) (*WaitResult, error) {
	// 1. 订阅配置变更
	notifyChan, subscriptionID, err := s.subscriptionMgr.Subscribe(ctx, &SubscribeRequest{
		ClientID:       req.ClientID,
		ClientIP:       req.ClientIP,
		ClientHostname: req.ClientHostname,
		NamespaceID:    req.NamespaceID,
		Environment:    req.Environment,
		ConfigKeys:     req.ConfigKeys,
		Versions:       req.Versions,
	})
	if err != nil {
		return nil, errors.ErrLongPollingSubscribeFailed(err)
	}

	hlog.Infof("客户端开始长轮询: clientID=%s, namespace=%d, env=%s, subscriptionID=%d",
		req.ClientID, req.NamespaceID, req.Environment, subscriptionID)

	// 2. 延迟取消订阅
	defer func() {
		if err := s.subscriptionMgr.Unsubscribe(req.ClientID, req.NamespaceID, req.Environment); err != nil {
			hlog.Errorf("取消订阅失败: %v", err)
		}
	}()

	// 3. 获取超时时间（从系统配置读取）
	timeout := s.getTimeout()

	// 4. 等待通知或超时
	select {
	case notification, ok := <-notifyChan:
		if !ok {
			// 通道已关闭
			return &WaitResult{
				Changed:    false,
				ConfigKeys: []string{},
				Versions:   req.Versions,
			}, nil
		}

		// 收到变更通知
		hlog.Infof("配置变更通知: clientID=%s, configKey=%s, newVersion=%s",
			req.ClientID, notification.ConfigKey, notification.NewVersion)

		return &WaitResult{
			Changed:    true,
			ConfigKeys: []string{notification.ConfigKey},
			Versions: map[string]string{
				notification.ConfigKey: notification.NewVersion,
			},
		}, nil

	case <-time.After(timeout):
		// 超时，返回未变更
		hlog.Infof("长轮询超时: clientID=%s, namespace=%d, timeout=%v", req.ClientID, req.NamespaceID, timeout)
		return &WaitResult{
			Changed:    false,
			ConfigKeys: []string{},
			Versions:   req.Versions,
		}, nil

	case <-ctx.Done():
		// 客户端取消请求
		hlog.Infof("客户端取消请求: clientID=%s, error=%v", req.ClientID, ctx.Err())
		return nil, ctx.Err()
	}
}

// ComputeVersion 计算配置版本（使用MD5）
func ComputeVersion(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}
