package entity

import (
	"strconv"
	"strings"
	"time"
)

// SystemConfig 系统配置领域实体
// 用于存储配置中心自身的运行时参数，如长轮询超时、最大订阅数等
type SystemConfig struct {
	ID          int       `json:"id"`           // 主键ID
	ConfigKey   string    `json:"config_key"`   // 配置键（唯一）
	ConfigValue string    `json:"config_value"` // 配置值
	Description string    `json:"description"`  // 描述信息
	IsActive    bool      `json:"is_active"`    // 是否启用
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
}

// ==================== 领域行为方法 ====================

// Activate 启用配置
func (s *SystemConfig) Activate() {
	s.IsActive = true
	s.UpdatedAt = s.UpdatedAt
}

// Deactivate 禁用配置
func (s *SystemConfig) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = s.UpdatedAt
}

// UpdateValue 更新配置值
func (s *SystemConfig) UpdateValue(value string) {
	s.ConfigValue = value
	s.UpdatedAt = s.UpdatedAt
}

// ==================== 查询方法 ====================

// IsActiveStatus 判断是否启用
func (s *SystemConfig) IsActiveStatus() bool {
	return s.IsActive
}

// GetKey 获取配置键
func (s *SystemConfig) GetKey() string {
	return s.ConfigKey
}

// GetValue 获取配置值
func (s *SystemConfig) GetValue() string {
	return s.ConfigValue
}

// ==================== 类型转换方法 ====================

// GetIntValue 获取整数类型的配置值
// 如果转换失败，返回默认值
func (s *SystemConfig) GetIntValue(defaultValue int) int {
	if s.ConfigValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strings.TrimSpace(s.ConfigValue))
	if err != nil {
		return defaultValue
	}

	return value
}

// GetInt64Value 获取 int64 类型的配置值
// 如果转换失败，返回默认值
func (s *SystemConfig) GetInt64Value(defaultValue int64) int64 {
	if s.ConfigValue == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(strings.TrimSpace(s.ConfigValue), 10, 64)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetBoolValue 获取布尔类型的配置值
// 支持：true/false, 1/0, yes/no, on/off
// 如果转换失败，返回默认值
func (s *SystemConfig) GetBoolValue(defaultValue bool) bool {
	if s.ConfigValue == "" {
		return defaultValue
	}

	normalized := strings.ToLower(strings.TrimSpace(s.ConfigValue))
	switch normalized {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// GetFloatValue 获取浮点数类型的配置值
// 如果转换失败，返回默认值
func (s *SystemConfig) GetFloatValue(defaultValue float64) float64 {
	if s.ConfigValue == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(s.ConfigValue), 64)
	if err != nil {
		return defaultValue
	}

	return value
}
