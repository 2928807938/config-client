package configsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ChangeCallback 配置变更回调（简化版）
type ChangeCallback func(key, value string)

// ConfigItem 配置缓存项
type ConfigItem struct {
	Value     string
	Version   string
	UpdatedAt time.Time
}

// ConfigCache 配置缓存
type ConfigCache struct {
	mu    sync.RWMutex
	items map[string]*ConfigItem
}

// NewConfigCache 创建配置缓存
func NewConfigCache() *ConfigCache {
	return &ConfigCache{
		items: make(map[string]*ConfigItem),
	}
}

func (c *ConfigCache) Set(key, value, version string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = &ConfigItem{
		Value:     value,
		Version:   version,
		UpdatedAt: time.Now(),
	}
}

func (c *ConfigCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.items[key]
	if !exists {
		return "", false
	}
	return item.Value, true
}

func (c *ConfigCache) GetAll() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]string, len(c.items))
	for key, item := range c.items {
		result[key] = item.Value
	}
	return result
}

func (c *ConfigCache) GetByPrefix(prefix string) map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]string)
	for key, item := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			result[key] = item.Value
		}
	}
	return result
}

func (c *ConfigCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.items[key]
	return exists
}

func (c *ConfigCache) SetBatch(configs map[string]struct {
	Value   string
	Version string
}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, config := range configs {
		c.items[key] = &ConfigItem{
			Value:     config.Value,
			Version:   config.Version,
			UpdatedAt: time.Now(),
		}
	}
}

// HTTPClient HTTP API 客户端
type HTTPClient struct {
	serverURL  string
	httpClient *http.Client
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient(serverURL string) *HTTPClient {
	return &HTTPClient{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ConfigVO 配置值对象
type ConfigVO struct {
	ID          int       `json:"id"`
	NamespaceID int       `json:"namespace_id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	ValueType   string    `json:"value_type"`
	GroupName   string    `json:"group_name"`
	Environment string    `json:"environment"`
	IsActive    bool      `json:"is_active"`
	IsReleased  bool      `json:"is_released"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Response 标准响应格式
type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// ConfigListVO 配置列表值对象
type ConfigListVO struct {
	Items      []ConfigVO `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	Size       int        `json:"size"`
	TotalPages int        `json:"total_pages"`
}

// GetConfigByKey 根据命名空间和键获取配置
func (c *HTTPClient) GetConfigByKey(namespaceID int, key string) (*ConfigVO, error) {
	url := fmt.Sprintf("%s/api/v1/configs", c.serverURL)
	httpReq, _ := http.NewRequest("GET", url, nil)

	q := httpReq.URL.Query()
	q.Add("namespace_id", fmt.Sprintf("%d", namespaceID))
	q.Add("key", key)
	q.Add("is_active", "true")
	q.Add("is_released", "true")
	q.Add("page", "1")
	q.Add("size", "1")
	httpReq.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("请求失败: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	var configList ConfigListVO
	if err := json.Unmarshal(result.Data, &configList); err != nil {
		return nil, fmt.Errorf("解析配置列表失败: %w", err)
	}

	if len(configList.Items) == 0 {
		return nil, fmt.Errorf("配置不存在: namespace_id=%d, key=%s", namespaceID, key)
	}

	return &configList.Items[0], nil
}

// GetConfigsByNamespace 获取命名空间下的所有配置
func (c *HTTPClient) GetConfigsByNamespace(namespaceID int) ([]ConfigVO, error) {
	url := fmt.Sprintf("%s/api/v1/configs", c.serverURL)
	httpReq, _ := http.NewRequest("GET", url, nil)

	q := httpReq.URL.Query()
	q.Add("namespace_id", fmt.Sprintf("%d", namespaceID))
	q.Add("is_active", "true")
	q.Add("is_released", "true")
	q.Add("page", "1")
	q.Add("size", "1000")
	httpReq.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败: status=%d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	var configList ConfigListVO
	if err := json.Unmarshal(result.Data, &configList); err != nil {
		return nil, fmt.Errorf("解析配置列表失败: %w", err)
	}

	return configList.Items, nil
}

// Client 配置中心 SDK 客户端
type Client struct {
	opts       *Options
	httpClient *HTTPClient
	cache      *ConfigCache
	watcher    Watcher
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	callbacks  map[string][]ChangeCallback
}

// New 创建配置中心客户端
func New(opts ...Option) (*Client, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	if options.ServerURL == "" {
		return nil, fmt.Errorf("ServerURL 不能为空")
	}

	client := &Client{
		opts:       options,
		httpClient: NewHTTPClient(options.ServerURL),
		callbacks:  make(map[string][]ChangeCallback),
	}

	// 启用缓存
	if options.EnableCache {
		client.cache = NewConfigCache()
	}

	// 初始化时拉取所有配置
	if options.FetchOnInit {
		if err := client.fetchAllConfigs(); err != nil {
			return nil, fmt.Errorf("拉取初始配置失败: %w", err)
		}
	}

	// 创建底层监听器
	if err := client.createWatcher(); err != nil {
		return nil, fmt.Errorf("创建监听器失败: %w", err)
	}

	// 自动启动
	if options.AutoStart && client.watcher != nil {
		ctx := context.Background()
		if err := client.Start(ctx); err != nil {
			return nil, fmt.Errorf("启动客户端失败: %w", err)
		}
	}

	return client, nil
}

// createWatcher 创建底层监听器
func (c *Client) createWatcher() error {
	watcher, err := createWatcherFromOptions(c.opts)
	if err != nil {
		return err
	}
	c.watcher = watcher
	return nil
}

// fetchAllConfigs 拉取所有配置
func (c *Client) fetchAllConfigs() error {
	if !c.opts.EnableCache {
		return nil
	}

	configs, err := c.httpClient.GetConfigsByNamespace(c.opts.NamespaceID)
	if err != nil {
		return err
	}

	batch := make(map[string]struct {
		Value   string
		Version string
	})

	for _, cfg := range configs {
		batch[cfg.Key] = struct {
			Value   string
			Version string
		}{
			Value:   cfg.Value,
			Version: strconv.Itoa(cfg.Version),
		}
	}

	c.cache.SetBatch(batch)
	return nil
}

// Start 启动客户端
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ctx != nil {
		return fmt.Errorf("客户端已启动")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	if c.watcher != nil {
		if err := c.watcher.Start(c.ctx); err != nil {
			return err
		}
	}

	return nil
}

// Stop 停止客户端
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	if c.watcher != nil {
		if err := c.watcher.Stop(); err != nil {
			return err
		}
	}

	return nil
}

// Get 获取配置
func (c *Client) Get(key string) (string, error) {
	// 先从缓存获取
	if c.opts.EnableCache {
		if value, ok := c.cache.Get(key); ok {
			return value, nil
		}
	}

	// 缓存未命中，从服务器获取
	config, err := c.httpClient.GetConfigByKey(c.opts.NamespaceID, key)
	if err != nil {
		// 尝试使用降级配置
		if fallbackValue, ok := c.opts.Fallback[key]; ok {
			return fallbackValue, nil
		}
		return "", fmt.Errorf("获取配置失败: %w", err)
	}

	// 更新缓存
	if c.opts.EnableCache {
		c.cache.Set(key, config.Value, strconv.Itoa(config.Version))
	}

	return config.Value, nil
}

// GetAll 获取所有配置
func (c *Client) GetAll() map[string]string {
	if c.opts.EnableCache {
		return c.cache.GetAll()
	}

	configs, err := c.httpClient.GetConfigsByNamespace(c.opts.NamespaceID)
	if err != nil {
		return make(map[string]string)
	}

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Key] = cfg.Value
	}

	return result
}

// GetByPrefix 根据前缀获取配置
func (c *Client) GetByPrefix(prefix string) map[string]string {
	if c.opts.EnableCache {
		return c.cache.GetByPrefix(prefix)
	}

	allConfigs := c.GetAll()
	result := make(map[string]string)
	for key, value := range allConfigs {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			result[key] = value
		}
	}

	return result
}

// Watch 监听配置变更
func (c *Client) Watch(key string, callback ChangeCallback) error {
	c.mu.Lock()
	c.callbacks[key] = append(c.callbacks[key], callback)
	c.mu.Unlock()

	if c.watcher != nil {
		return c.watcher.Watch([]string{key}, c.handleConfigChange)
	}

	return nil
}

// GetAndWatch 获取配置并监听变更
func (c *Client) GetAndWatch(key string, callback ChangeCallback) error {
	value, err := c.Get(key)
	if err != nil {
		value = ""
	}

	callback(key, value)

	return c.Watch(key, callback)
}

// Unwatch 取消监听
func (c *Client) Unwatch(key string) error {
	c.mu.Lock()
	delete(c.callbacks, key)
	c.mu.Unlock()

	if c.watcher != nil {
		return c.watcher.Unwatch([]string{key})
	}

	return nil
}

// handleConfigChange 处理配置变更事件
func (c *Client) handleConfigChange(key, value, version string) {
	if c.opts.EnableCache {
		c.cache.Set(key, value, version)
	}

	c.mu.RLock()
	callbacks := c.callbacks[key]
	c.mu.RUnlock()

	for _, callback := range callbacks {
		go callback(key, value)
	}
}

// Refresh 刷新指定配置
func (c *Client) Refresh(key string) error {
	config, err := c.httpClient.GetConfigByKey(c.opts.NamespaceID, key)
	if err != nil {
		return err
	}

	if c.opts.EnableCache {
		c.cache.Set(key, config.Value, strconv.Itoa(config.Version))
	}

	return nil
}

// RefreshAll 刷新所有配置
func (c *Client) RefreshAll() error {
	return c.fetchAllConfigs()
}

// Has 检查配置是否存在
func (c *Client) Has(key string) bool {
	if c.opts.EnableCache {
		return c.cache.Has(key)
	}

	_, err := c.Get(key)
	return err == nil
}

// IsRunning 是否正在运行
func (c *Client) IsRunning() bool {
	if c.watcher != nil {
		return c.watcher.IsRunning()
	}
	return false
}

// GetOptions 获取配置选项
func (c *Client) GetOptions() *Options {
	return c.opts
}
