package errors

import "config-client/share/errors"

// ==================== 配置领域错误码 ====================
// 错误码分段: 20000-20999
// 错误码规则: xxxYY，其中 YY 表示 HTTP 状态码类型
//   - xx01: bad_request (400)
//   - xx02: unauthorized (401)
//   - xx03: forbidden (403)
//   - xx04: not_found (404)
//   - xx05: conflict (409)
const (
	// 配置相关错误码 20000-20099
	ConfigNotFound           = 20004 // 配置不存在 (404)
	ConfigAlreadyExists      = 20005 // 配置已存在 (409)
	ConfigNotReleased        = 20104 // 配置未发布 (404)
	ConfigAlreadyReleased    = 20105 // 配置已发布 (409)
	ConfigNotActive          = 20204 // 配置未激活 (404)
	ConfigKeyInvalid         = 20301 // 配置键无效 (400)
	ConfigValueInvalid       = 20401 // 配置值无效 (400)
	ConfigValueTypeInvalid   = 20402 // 配置值类型无效 (400)
	ConfigHashMismatch       = 20500 // 配置哈希不匹配 (500)
	ConfigVersionConflict    = 20605 // 配置版本冲突 (409)
	ConfigGroupNotFound      = 20704 // 配置分组不存在 (404)
	ConfigEnvironmentInvalid = 20801 // 环境参数无效 (400)
	ConfigCannotDelete       = 20903 // 配置无法删除 (403)

	// 命名空间相关错误码 20100-20199
	NamespaceNotFound       = 21004 // 命名空间不存在 (404)
	NamespaceAlreadyExists  = 21005 // 命名空间已存在 (409)
	NamespaceNotActive      = 21104 // 命名空间未激活 (404)
	NamespaceNameInvalid    = 21201 // 命名空间名称无效 (400)
	NamespaceCannotDelete   = 21303 // 命名空间无法删除 (403)
	NamespaceMustDeactivate = 21401 // 命名空间必须先停用 (400)
)

// ==================== 配置领域业务异常 ====================

// ErrConfigNotFound 配置不存在
func ErrConfigNotFound(key string, environment string) *errors.AppError {
	return errors.New(ConfigNotFound, "配置不存在: key="+key+", env="+environment)
}

// ErrConfigAlreadyExists 配置已存在
func ErrConfigAlreadyExists(key string, environment string) *errors.AppError {
	return errors.New(ConfigAlreadyExists, "配置已存在: key="+key+", env="+environment)
}

// ErrConfigNotReleased 配置未发布
func ErrConfigNotReleased(key string) *errors.AppError {
	return errors.New(ConfigNotReleased, "配置未发布，无法使用: key="+key)
}

// ErrConfigAlreadyReleased 配置已发布
func ErrConfigAlreadyReleased(key string) *errors.AppError {
	return errors.New(ConfigAlreadyReleased, "配置已发布，无法修改: key="+key)
}

// ErrConfigNotActive 配置未激活
func ErrConfigNotActive(key string) *errors.AppError {
	return errors.New(ConfigNotActive, "配置未激活: key="+key)
}

// ErrConfigKeyInvalid 配置键无效
func ErrConfigKeyInvalid(key string, reason string) *errors.AppError {
	return errors.New(ConfigKeyInvalid, "配置键无效: key="+key+", 原因: "+reason)
}

// ErrConfigKeyEmpty 配置键不能为空
func ErrConfigKeyEmpty() *errors.AppError {
	return errors.New(ConfigKeyInvalid, "配置键不能为空")
}

// ErrConfigKeyFormatInvalid 配置键格式无效
func ErrConfigKeyFormatInvalid(key string) *errors.AppError {
	return errors.New(ConfigKeyInvalid, "配置键格式无效: key="+key+", 只能包含字母、数字、下划线、中划线、点号")
}

// ErrConfigValueInvalid 配置值无效
func ErrConfigValueInvalid(key string, reason string) *errors.AppError {
	return errors.New(ConfigValueInvalid, "配置值无效: key="+key+", 原因: "+reason)
}

// ErrConfigValueEmpty 配置值不能为空
func ErrConfigValueEmpty(key string) *errors.AppError {
	return errors.New(ConfigValueInvalid, "配置值不能为空: key="+key)
}

// ErrConfigValueTypeInvalid 配置值类型无效
func ErrConfigValueTypeInvalid(valueType string, reason string) *errors.AppError {
	return errors.New(ConfigValueTypeInvalid, "配置值类型验证失败: type="+valueType+", 原因: "+reason)
}

// ErrUnsupportedHashAlgorithm 不支持的哈希算法
func ErrUnsupportedHashAlgorithm(algorithm string) *errors.AppError {
	return errors.New(ConfigValueInvalid, "不支持的哈希算法: "+algorithm)
}

// ErrConfigHashMismatch 配置哈希不匹配
func ErrConfigHashMismatch(key string, expected string, actual string) *errors.AppError {
	return errors.New(ConfigHashMismatch, "配置哈希不匹配: key="+key)
}

// ErrConfigVersionConflict 配置版本冲突
func ErrConfigVersionConflict(key string, expectedVersion int, actualVersion int) *errors.AppError {
	return errors.New(ConfigVersionConflict, "配置版本冲突: key="+key)
}

// ErrConfigGroupNotFound 配置分组不存在
func ErrConfigGroupNotFound(groupName string) *errors.AppError {
	return errors.New(ConfigGroupNotFound, "配置分组不存在: group="+groupName)
}

// ErrConfigEnvironmentInvalid 环境参数无效
func ErrConfigEnvironmentInvalid(environment string) *errors.AppError {
	return errors.New(ConfigEnvironmentInvalid, "环境参数无效: env="+environment)
}

// ErrConfigCannotDelete 配置无法删除（已发布的配置不能删除）
func ErrConfigCannotDelete(key string) *errors.AppError {
	return errors.New(ConfigCannotDelete, "配置已发布，无法删除，请先取消发布: key="+key)
}

// ==================== 命名空间领域业务异常 ====================

// ErrNamespaceNotFound 命名空间不存在
func ErrNamespaceNotFound(name string) *errors.AppError {
	if name == "" {
		return errors.New(NamespaceNotFound, "命名空间不存在")
	}
	return errors.New(NamespaceNotFound, "命名空间不存在: name="+name)
}

// ErrNamespaceAlreadyExists 命名空间已存在
func ErrNamespaceAlreadyExists(name string) *errors.AppError {
	return errors.New(NamespaceAlreadyExists, "命名空间已存在: name="+name)
}

// ErrNamespaceNotActive 命名空间未激活
func ErrNamespaceNotActive(name string) *errors.AppError {
	return errors.New(NamespaceNotActive, "命名空间未激活: name="+name)
}

// ErrNamespaceNameInvalid 命名空间名称无效
func ErrNamespaceNameInvalid(name string, reason string) *errors.AppError {
	return errors.New(NamespaceNameInvalid, "命名空间名称无效: name="+name+", 原因: "+reason)
}

// ErrNamespaceNameEmpty 命名空间名称不能为空
func ErrNamespaceNameEmpty() *errors.AppError {
	return errors.New(NamespaceNameInvalid, "命名空间名称不能为空")
}

// ErrNamespaceNameLengthInvalid 命名空间名称长度无效
func ErrNamespaceNameLengthInvalid(name string) *errors.AppError {
	return errors.New(NamespaceNameInvalid, "命名空间名称长度必须在 2-255 之间: name="+name)
}

// ErrNamespaceNameFormatInvalid 命名空间名称格式无效
func ErrNamespaceNameFormatInvalid(name string) *errors.AppError {
	return errors.New(NamespaceNameInvalid, "命名空间名称只能包含小写字母、数字、下划线、中划线: name="+name)
}

// ErrNamespaceDisplayNameTooLong 命名空间显示名称过长
func ErrNamespaceDisplayNameTooLong() *errors.AppError {
	return errors.New(NamespaceNameInvalid, "显示名称长度不能超过 255")
}

// ErrNamespaceDisplayNameEmpty 命名空间显示名称不能为空
func ErrNamespaceDisplayNameEmpty() *errors.AppError {
	return errors.New(NamespaceNameInvalid, "显示名称不能为空")
}

// ErrNamespaceCannotDelete 命名空间无法删除（存在关联配置）
func ErrNamespaceCannotDelete(name string, configCount int64) *errors.AppError {
	return errors.New(NamespaceCannotDelete, "命名空间无法删除，存在关联配置")
}

// ErrNamespaceMustDeactivate 命名空间必须先停用才能删除
func ErrNamespaceMustDeactivate(name string) *errors.AppError {
	return errors.New(NamespaceMustDeactivate, "命名空间必须先停用才能删除: name="+name)
}
