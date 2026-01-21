package request

// LongPollingRequest 长轮询请求
type LongPollingRequest struct {
	ConfigKeys []ConfigKeyVersion `json:"config_keys" binding:"required,min=1,dive"` // 配置键列表
}

// ConfigKeyVersion 配置键及其版本
type ConfigKeyVersion struct {
	NamespaceID int64  `json:"namespace_id" binding:"required,min=1"` // 命名空间ID
	ConfigKey   string `json:"config_key" binding:"required"`         // 配置键
	Version     string `json:"version" binding:"required"`            // 当前客户端持有的版本号（MD5）
}
