module config-client/cmd/api

go 1.24.11

require (
	config-client/bom v0.0.0
	config-client/share v0.0.0

	// Hertz HTTP 框架
	github.com/cloudwego/hertz v0.9.3

	// 数据库
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.25.12

	// 配置文件解析
	gopkg.in/yaml.v3 v3.0.1
)

replace (
	config-client/bom => ../../bom
	config-client/share => ../../share
)
