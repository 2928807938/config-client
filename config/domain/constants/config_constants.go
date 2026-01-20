package constants

// ==================== 哈希算法常量 ====================

const (
	// HashAlgorithmMD5 MD5哈希算法
	HashAlgorithmMD5 = "md5"

	// HashAlgorithmSHA256 SHA256哈希算法（保留，未来可能使用）
	HashAlgorithmSHA256 = "sha256"
)

// ==================== 环境常量 ====================

const (
	// EnvDev 开发环境
	EnvDev = "dev"

	// EnvTest 测试环境
	EnvTest = "test"

	// EnvUAT UAT环境
	EnvUAT = "uat"

	// EnvProd 生产环境
	EnvProd = "prod"

	// EnvLocal 本地环境
	EnvLocal = "local"
)

// ValidEnvironments 有效的环境列表
var ValidEnvironments = []string{
	EnvDev,
	EnvTest,
	EnvUAT,
	EnvProd,
	EnvLocal,
}

// ==================== 值类型常量 ====================

const (
	// ValueTypeString 字符串类型
	ValueTypeString = "string"

	// ValueTypeInt 整数类型
	ValueTypeInt = "int"

	// ValueTypeBool 布尔类型
	ValueTypeBool = "bool"

	// ValueTypeFloat 浮点数类型
	ValueTypeFloat = "float"

	// ValueTypeJSON JSON格式
	ValueTypeJSON = "json"

	// ValueTypeYAML YAML格式
	ValueTypeYAML = "yaml"
)

// ValidValueTypes 有效的值类型列表
var ValidValueTypes = []string{
	ValueTypeString,
	ValueTypeInt,
	ValueTypeBool,
	ValueTypeFloat,
	ValueTypeJSON,
	ValueTypeYAML,
}

// ==================== 默认值常量 ====================

const (
	// DefaultGroupName 默认配置分组名称
	DefaultGroupName = "default"

	// DefaultValueType 默认配置值类型
	DefaultValueType = ValueTypeString
)
