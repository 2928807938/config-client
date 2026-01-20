module config-client/bom

go 1.24.11

require (
	github.com/bytedance/sonic v1.12.6
	// Hertz HTTP 框架
	github.com/cloudwego/hertz v0.9.3

	// Kitex RPC 框架
	github.com/cloudwego/kitex v0.11.3

	// 日志（hlog 已包含在 hertz 中）

	// 验证器
	github.com/go-playground/validator/v10 v10.23.0

	// 通用工具
	github.com/google/uuid v1.6.0

	// 配置管理
	github.com/spf13/viper v1.19.0

	// 数据库
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.25.12
)
