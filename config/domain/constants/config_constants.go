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
	// EnvDefault 默认环境
	EnvDefault = "default"

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
	EnvDefault,
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

	// ValueTypeEncrypted 加密类型（敏感配置）
	ValueTypeEncrypted = "encrypted"
)

// ValidValueTypes 有效的值类型列表
var ValidValueTypes = []string{
	ValueTypeString,
	ValueTypeInt,
	ValueTypeBool,
	ValueTypeFloat,
	ValueTypeJSON,
	ValueTypeYAML,
	ValueTypeEncrypted,
}

// ==================== 默认值常量 ====================

const (
	// DefaultGroupName 默认配置分组名称
	DefaultGroupName = "default"

	// DefaultValueType 默认配置值类型
	DefaultValueType = ValueTypeString
)

// ==================== 标签相关常量 ====================

const (
	// TagKeySensitive 敏感标签键
	TagKeySensitive = "sensitive"

	// TagKeyCategory 分类标签键
	TagKeyCategory = "category"

	// TagKeyImportance 重要性标签键
	TagKeyImportance = "importance"

	// TagKeyType 类型标签键
	TagKeyType = "type"

	// TagKeyEnvironment 环境标签键
	TagKeyEnvironment = "environment"

	// TagKeyModule 模块标签键
	TagKeyModule = "module"
)

// 标签值常量
const (
	// TagValueTrue 布尔标签true值
	TagValueTrue = "true"

	// TagValueFalse 布尔标签false值
	TagValueFalse = "false"

	// TagValueHigh 高重要性
	TagValueHigh = "high"

	// TagValueMedium 中等重要性
	TagValueMedium = "medium"

	// TagValueLow 低重要性
	TagValueLow = "low"
)

// ValidImportanceLevels 有效的重要性级别列表
var ValidImportanceLevels = []string{
	TagValueHigh,
	TagValueMedium,
	TagValueLow,
}
