package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"config-client/config/domain/entity"
	"config-client/config/domain/listener"
	"config-client/config/domain/repository"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	ClientID       string            // 客户端唯一标识
	ClientIP       string            // 客户端IP地址
	ClientHostname string            // 客户端主机名
	NamespaceID    int               // 命名空间ID
	Environment    string            // 环境
	ConfigKeys     []string          // 关注的配置键列表 (格式: "namespaceID:configKey")
	Versions       map[string]string // 当前版本映射 (configKey -> MD5)
}

// ChangeNotification 变更通知
type ChangeNotification struct {
	NamespaceID int       // 命名空间ID
	ConfigKey   string    // 变更的配置键
	NewVersion  string    // 新版本MD5
	Timestamp   time.Time // 变更时间
}

// ActiveSubscriber 活跃订阅者 (内存中的长轮询连接)
type ActiveSubscriber struct {
	ClientID        string                   // 客户端ID
	NamespaceID     int                      // 命名空间ID
	Environment     string                   // 环境
	ConfigKeys      []string                 // 关注的配置键列表
	CurrentVersions map[string]string        // 当前版本
	NotifyChan      chan *ChangeNotification // 通知通道
	RegisteredAt    time.Time                // 注册时间
	SubscriptionID  int                      // 数据库订阅记录ID
}

// SubscriptionManager 订阅管理器
// 负责管理客户端订阅关系和配置变更通知
type SubscriptionManager struct {
	// 持久化仓储
	subscriptionRepo repository.SubscriptionRepository
	configRepo       repository.ConfigRepository

	// 配置监听器 (接收变更事件)
	listener listener.ConfigListener

	// 发布管理服务 (用于灰度发布判断)
	releaseSvc *ReleaseService

	// 活跃订阅者 (内存)
	// key: "namespaceID:environment:clientID"
	activeSubscribers map[string]*ActiveSubscriber
	mu                sync.RWMutex

	// 配置键到订阅者的映射 (用于快速查找)
	// key: "namespaceID:configKey"
	// value: 订阅该配置的 subscriber keys
	configSubscribers map[string][]string

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// 配置
	heartbeatTimeout time.Duration // 心跳超时时间
	cleanInterval    time.Duration // 清理过期订阅的间隔
}

// NewSubscriptionManager 创建订阅管理器
func NewSubscriptionManager(
	subscriptionRepo repository.SubscriptionRepository,
	configRepo repository.ConfigRepository,
	listener listener.ConfigListener,
	heartbeatTimeout time.Duration,
) *SubscriptionManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &SubscriptionManager{
		subscriptionRepo:  subscriptionRepo,
		configRepo:        configRepo,
		listener:          listener,
		releaseSvc:        nil, // 通过 SetReleaseService 延迟注入,避免循环依赖
		activeSubscribers: make(map[string]*ActiveSubscriber),
		configSubscribers: make(map[string][]string),
		ctx:               ctx,
		cancel:            cancel,
		heartbeatTimeout:  heartbeatTimeout,
		cleanInterval:     5 * time.Minute, // 默认5分钟清理一次
	}
}

// SetReleaseService 设置发布管理服务（用于支持灰度发布）
func (m *SubscriptionManager) SetReleaseService(releaseSvc *ReleaseService) {
	m.releaseSvc = releaseSvc
}

// Start 启动订阅管理器
func (m *SubscriptionManager) Start() error {
	// 订阅配置变更事件
	eventChan, err := m.listener.Subscribe(m.ctx)
	if err != nil {
		return fmt.Errorf("订阅配置变更事件失败: %w", err)
	}

	// 启动事件处理
	go m.handleConfigChangeEvents(eventChan)

	// 启动定期清理任务
	go m.startCleanupTask()

	hlog.Info("订阅管理器已启动")
	return nil
}

// Stop 停止订阅管理器
func (m *SubscriptionManager) Stop() error {
	m.cancel()
	return m.listener.Close()
}

// Subscribe 订阅配置变更
// 返回: 通知通道, 订阅ID, 错误
func (m *SubscriptionManager) Subscribe(ctx context.Context, req *SubscribeRequest) (<-chan *ChangeNotification, int, error) {
	// 1. 检查或创建数据库订阅记录
	subscription, err := m.getOrCreateSubscription(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("获取或创建订阅失败: %w", err)
	}

	// 2. 增加轮询计数
	if err := m.subscriptionRepo.IncrementPollCount(ctx, subscription.ID); err != nil {
		hlog.Errorf("增加轮询计数失败: %v", err)
	}

	// 3. 检查灰度发布,判断该客户端应该使用哪个版本的配置
	versionToUse := m.determineVersionForClient(ctx, req)

	// 4. 检查配置是否已有变更
	changed, changedKey, newVersion := m.checkVersionChanges(req.ConfigKeys, req.Versions, versionToUse, req.Environment)
	if changed {
		// 配置已变更，立即返回
		hlog.Infof("配置已变更: %s, 立即返回", changedKey)

		// 创建一个带缓冲的通道，立即发送通知
		notifyChan := make(chan *ChangeNotification, 1)
		notifyChan <- &ChangeNotification{
			NamespaceID: req.NamespaceID,
			ConfigKey:   changedKey,
			NewVersion:  newVersion,
			Timestamp:   time.Now(),
		}

		// 增加变更计数
		if err := m.subscriptionRepo.IncrementChangeCount(ctx, subscription.ID); err != nil {
			hlog.Errorf("增加变更计数失败: %v", err)
		}

		return notifyChan, subscription.ID, nil
	}

	// 5. 注册活跃订阅者 (内存)
	subscriberKey := m.makeSubscriberKey(req.NamespaceID, req.Environment, req.ClientID)
	notifyChan := make(chan *ChangeNotification, 1)

	subscriber := &ActiveSubscriber{
		ClientID:        req.ClientID,
		NamespaceID:     req.NamespaceID,
		Environment:     req.Environment,
		ConfigKeys:      req.ConfigKeys,
		CurrentVersions: req.Versions,
		NotifyChan:      notifyChan,
		RegisteredAt:    time.Now(),
		SubscriptionID:  subscription.ID,
	}

	m.registerActiveSubscriber(subscriberKey, subscriber)

	hlog.Infof("注册活跃订阅者: clientID=%s, namespace=%d, env=%s, configKeys=%v",
		req.ClientID, req.NamespaceID, req.Environment, req.ConfigKeys)

	return notifyChan, subscription.ID, nil
}

// determineVersionForClient 判断客户端应该使用哪个版本的配置
// 如果有灰度发布且客户端匹配灰度规则,返回灰度版本;否则返回nil(使用当前版本)
func (m *SubscriptionManager) determineVersionForClient(ctx context.Context, req *SubscribeRequest) map[string]string {
	if m.releaseSvc == nil {
		return nil
	}

	// 判断是否应该使用灰度版本
	release, matched, err := m.releaseSvc.ShouldUseCanaryRelease(
		ctx,
		req.NamespaceID,
		req.Environment,
		req.ClientID,
		req.ClientIP,
	)

	if err != nil {
		hlog.Errorf("判断灰度发布失败: %v", err)
		return nil
	}

	if !matched || release == nil {
		return nil
	}

	// 获取灰度版本的配置快照
	snapshot, err := release.GetConfigSnapshot()
	if err != nil {
		hlog.Errorf("获取灰度版本快照失败: %v", err)
		return nil
	}

	// 构建灰度版本的配置版本映射
	canaryVersions := make(map[string]string)
	for _, item := range snapshot {
		configKey := fmt.Sprintf("%d:%s", req.NamespaceID, item.Key)
		canaryVersions[configKey] = ComputeVersion(item.Value)
	}

	hlog.Infof("客户端匹配灰度规则: clientID=%s, releaseID=%d, version=%d",
		req.ClientID, release.ID, release.Version)

	return canaryVersions
}

// Unsubscribe 取消订阅 (移除内存中的活跃订阅)
func (m *SubscriptionManager) Unsubscribe(clientID string, namespaceID int, environment string) error {
	subscriberKey := m.makeSubscriberKey(namespaceID, environment, clientID)
	m.unregisterActiveSubscriber(subscriberKey)
	hlog.Infof("取消订阅: clientID=%s, namespace=%d, env=%s", clientID, namespaceID, environment)
	return nil
}

// UpdateHeartbeat 更新心跳
func (m *SubscriptionManager) UpdateHeartbeat(ctx context.Context, clientID string, namespaceID int, environment string) error {
	// 获取订阅记录
	subscription, err := m.subscriptionRepo.GetByClientAndNamespace(ctx, clientID, namespaceID, environment)
	if err != nil {
		return err
	}
	if subscription == nil {
		return fmt.Errorf("订阅不存在")
	}

	// 更新心跳
	return m.subscriptionRepo.UpdateHeartbeat(ctx, subscription.ID)
}

// handleConfigChangeEvents 处理配置变更事件
func (m *SubscriptionManager) handleConfigChangeEvents(eventChan <-chan *listener.ConfigChangeEvent) {
	for {
		select {
		case <-m.ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			m.handleConfigChangeEvent(event)
		}
	}
}

// handleConfigChangeEvent 处理单个配置变更事件
func (m *SubscriptionManager) handleConfigChangeEvent(event *listener.ConfigChangeEvent) {
	configKey := fmt.Sprintf("%d:%s", event.NamespaceID, event.ConfigKey)

	hlog.Infof("收到配置变更事件: %s, action=%s", configKey, event.Action)

	// 获取订阅该配置的活跃订阅者
	subscribers := m.findSubscribersByConfigKey(configKey)
	if len(subscribers) == 0 {
		hlog.Infof("没有活跃订阅者关注配置: %s", configKey)
		return
	}

	// 获取第一个订阅者的环境（同一配置的所有订阅者应该在相同环境）
	environment := "default"
	if len(subscribers) > 0 && subscribers[0].Environment != "" {
		environment = subscribers[0].Environment
	}

	// 获取最新版本
	newVersion, err := m.getConfigVersion(event.NamespaceID, event.ConfigKey, environment)
	if err != nil {
		hlog.Errorf("获取配置版本失败: %s, error: %v", configKey, err)
		return
	}

	// 通知所有订阅者
	notification := &ChangeNotification{
		NamespaceID: event.NamespaceID,
		ConfigKey:   configKey,
		NewVersion:  newVersion,
		Timestamp:   time.Now(),
	}

	for _, subscriber := range subscribers {
		// 非阻塞发送通知
		select {
		case subscriber.NotifyChan <- notification:
			hlog.Infof("通知订阅者: clientID=%s, configKey=%s, newVersion=%s",
				subscriber.ClientID, configKey, newVersion)

			// 增加变更计数
			if err := m.subscriptionRepo.IncrementChangeCount(context.Background(), subscriber.SubscriptionID); err != nil {
				hlog.Errorf("增加变更计数失败: %v", err)
			}
		default:
			hlog.Warnf("订阅者通道已满，跳过通知: clientID=%s", subscriber.ClientID)
		}
	}
}

// startCleanupTask 启动定期清理任务
func (m *SubscriptionManager) startCleanupTask() {
	ticker := time.NewTicker(m.cleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanExpiredSubscriptions()
		}
	}
}

// cleanExpiredSubscriptions 清理过期订阅
func (m *SubscriptionManager) cleanExpiredSubscriptions() {
	expireTime := time.Now().Add(-m.heartbeatTimeout)
	count, err := m.subscriptionRepo.CleanExpiredSubscriptions(context.Background(), expireTime)
	if err != nil {
		hlog.Errorf("清理过期订阅失败: %v", err)
		return
	}

	if count > 0 {
		hlog.Infof("清理过期订阅: 数量=%d", count)
	}
}

// getOrCreateSubscription 获取或创建订阅记录
func (m *SubscriptionManager) getOrCreateSubscription(ctx context.Context, req *SubscribeRequest) (*entity.Subscription, error) {
	// 查询是否已存在订阅
	subscription, err := m.subscriptionRepo.GetByClientAndNamespace(ctx, req.ClientID, req.NamespaceID, req.Environment)
	if err != nil {
		return nil, err
	}

	if subscription != nil {
		// 已存在，更新心跳
		subscription.UpdateHeartbeat()
		if err := m.subscriptionRepo.Update(ctx, subscription); err != nil {
			hlog.Errorf("更新订阅心跳失败: %v", err)
		}
		return subscription, nil
	}

	// 不存在，创建新订阅
	now := time.Now()
	subscription = &entity.Subscription{
		NamespaceID:     req.NamespaceID,
		ClientID:        req.ClientID,
		ClientIP:        req.ClientIP,
		ClientHostname:  req.ClientHostname,
		Environment:     req.Environment,
		IsActive:        true,
		LastHeartbeatAt: &now,
		HeartbeatCount:  1,
		PollCount:       0,
		ChangeCount:     0,
		SubscribedAt:    now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := m.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, err
	}

	hlog.Infof("创建新订阅: clientID=%s, namespace=%d, env=%s", req.ClientID, req.NamespaceID, req.Environment)
	return subscription, nil
}

// checkVersionChanges 检查版本是否有变更
// 返回: 是否有变更, 变更的配置键, 新版本
func (m *SubscriptionManager) checkVersionChanges(configKeys []string, clientVersions map[string]string, canaryVersions map[string]string, environment string) (bool, string, string) {
	for _, configKey := range configKeys {
		clientVersion := clientVersions[configKey]

		// 如果有灰度版本,优先使用灰度版本进行比较
		if canaryVersions != nil {
			if canaryVersion, exists := canaryVersions[configKey]; exists {
				if clientVersion != canaryVersion {
					hlog.Infof("[版本比较] 灰度版本不同: configKey=%s, clientVersion=%s, canaryVersion=%s", configKey, clientVersion, canaryVersion)
					return true, configKey, canaryVersion
				}
				continue
			}
		}

		// 解析 configKey
		var namespaceID int
		var key string
		_, err := fmt.Sscanf(configKey, "%d:%s", &namespaceID, &key)
		if err != nil {
			hlog.Errorf("解析配置键失败: %s, error: %v", configKey, err)
			continue
		}

		// 获取服务端版本
		serverVersion, err := m.getConfigVersion(namespaceID, key, environment)
		if err != nil {
			hlog.Errorf("获取配置版本失败: %s, error: %v", configKey, err)
			continue
		}

		hlog.Infof("[版本比较] configKey=%s, clientVersion=%s, serverVersion=%s, 是否相同=%v", configKey, clientVersion, serverVersion, clientVersion == serverVersion)

		// 比较版本
		if clientVersion != serverVersion {
			return true, configKey, serverVersion
		}
	}

	return false, "", ""
}

// getConfigVersion 获取配置版本
func (m *SubscriptionManager) getConfigVersion(namespaceID int, configKey string, environment string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 如果environment为空，使用默认值
	if environment == "" {
		environment = "default"
	}

	config, err := m.configRepo.FindByNamespaceAndKey(ctx, namespaceID, configKey, environment)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", fmt.Errorf("配置不存在: %s (environment=%s)", configKey, environment)
	}

	return ComputeVersion(config.Value), nil
}

// registerActiveSubscriber 注册活跃订阅者
func (m *SubscriptionManager) registerActiveSubscriber(subscriberKey string, subscriber *ActiveSubscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 注册订阅者
	m.activeSubscribers[subscriberKey] = subscriber

	// 建立配置键到订阅者的映射
	for _, configKey := range subscriber.ConfigKeys {
		m.configSubscribers[configKey] = append(m.configSubscribers[configKey], subscriberKey)
	}
}

// unregisterActiveSubscriber 注销活跃订阅者
func (m *SubscriptionManager) unregisterActiveSubscriber(subscriberKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscriber, exists := m.activeSubscribers[subscriberKey]
	if !exists {
		return
	}

	// 关闭通知通道
	close(subscriber.NotifyChan)

	// 移除配置键映射
	for _, configKey := range subscriber.ConfigKeys {
		subscribers := m.configSubscribers[configKey]
		for i, key := range subscribers {
			if key == subscriberKey {
				m.configSubscribers[configKey] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
		// 如果该配置没有订阅者了，删除key
		if len(m.configSubscribers[configKey]) == 0 {
			delete(m.configSubscribers, configKey)
		}
	}

	// 移除订阅者
	delete(m.activeSubscribers, subscriberKey)
}

// findSubscribersByConfigKey 根据配置键查找订阅者
func (m *SubscriptionManager) findSubscribersByConfigKey(configKey string) []*ActiveSubscriber {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subscriberKeys := m.configSubscribers[configKey]
	if len(subscriberKeys) == 0 {
		return nil
	}

	subscribers := make([]*ActiveSubscriber, 0, len(subscriberKeys))
	for _, key := range subscriberKeys {
		if subscriber, exists := m.activeSubscribers[key]; exists {
			subscribers = append(subscribers, subscriber)
		}
	}

	return subscribers
}

// makeSubscriberKey 生成订阅者唯一键
func (m *SubscriptionManager) makeSubscriberKey(namespaceID int, environment string, clientID string) string {
	return fmt.Sprintf("%d:%s:%s", namespaceID, environment, clientID)
}

// GetSubscriberCount 获取某个配置的订阅者数量
func (m *SubscriptionManager) GetSubscriberCount(configKey string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.configSubscribers[configKey])
}

// GetActiveSubscriberCount 获取活跃订阅者总数
func (m *SubscriptionManager) GetActiveSubscriberCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.activeSubscribers)
}
