package request

// LongPollingRequest 长轮询请求
type LongPollingRequest struct {
	ClientID       string             `json:"client_id" binding:"required"`              // 客户端唯一标识
	ClientIP       string             `json:"client_ip"`                                 // 客户端IP地址 (可选,服务端可自动获取)
	ClientHostname string             `json:"client_hostname"`                           // 客户端主机名 (可选)
	ConfigKeys     []ConfigKeyVersion `json:"config_keys" binding:"required,min=1,dive"` // 配置键列表
}

// ConfigKeyVersion 配置键及其版本
type ConfigKeyVersion struct {
	NamespaceID int    `json:"namespace_id" binding:"required,min=1"` // 命名空间ID
	ConfigKey   string `json:"config_key" binding:"required"`         // 配置键
	Version     string `json:"version" binding:"required"`            // 当前客户端持有的版本号（MD5）
	Environment string `json:"environment" binding:"required"`        // 环境
}
