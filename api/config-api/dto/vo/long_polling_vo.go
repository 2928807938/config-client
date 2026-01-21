package vo

// LongPollingResponse 长轮询响应
type LongPollingResponse struct {
	Changed    bool                 `json:"changed"`     // 是否有配置变更
	ConfigKeys []string             `json:"config_keys"` // 变更的配置键列表（格式: "namespaceID:configKey"）
	Configs    []ConfigChangeDetail `json:"configs"`     // 变更的配置详情
}

// ConfigChangeDetail 配置变更详情
type ConfigChangeDetail struct {
	NamespaceID int    `json:"namespace_id"` // 命名空间ID
	ConfigKey   string `json:"config_key"`   // 配置键
	Version     string `json:"version"`      // 最新版本号（MD5）
	Value       string `json:"value"`        // 配置值
	ValueType   string `json:"value_type"`   // 值类型
}
