package configsdk

import (
	"fmt"
	"strconv"
	"strings"
)

// GetString 获取字符串类型配置
func (c *Client) GetString(key string) string {
	value, _ := c.Get(key)
	return value
}

// GetInt 获取整数类型配置
func (c *Client) GetInt(key string) (int, error) {
	value, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	intVal, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("配置 %s 不是有效的整数: %w", key, err)
	}

	return intVal, nil
}

// GetIntOrDefault 获取整数配置，失败时返回默认值
func (c *Client) GetIntOrDefault(key string, defaultValue int) int {
	val, err := c.GetInt(key)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetInt64 获取 int64 类型配置
func (c *Client) GetInt64(key string) (int64, error) {
	value, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	int64Val, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("配置 %s 不是有效的 int64: %w", key, err)
	}

	return int64Val, nil
}

// GetFloat64 获取浮点数类型配置
func (c *Client) GetFloat64(key string) (float64, error) {
	value, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	floatVal, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("配置 %s 不是有效的浮点数: %w", key, err)
	}

	return floatVal, nil
}

// GetFloat64OrDefault 获取浮点数配置，失败时返回默认值
func (c *Client) GetFloat64OrDefault(key string, defaultValue float64) float64 {
	val, err := c.GetFloat64(key)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetBool 获取布尔类型配置
func (c *Client) GetBool(key string) (bool, error) {
	value, err := c.Get(key)
	if err != nil {
		return false, err
	}

	value = strings.ToLower(strings.TrimSpace(value))

	switch value {
	case "true", "1", "yes", "on", "enabled":
		return true, nil
	case "false", "0", "no", "off", "disabled", "":
		return false, nil
	default:
		return false, fmt.Errorf("配置 %s 不是有效的布尔值: %s", key, value)
	}
}

// GetBoolOrDefault 获取布尔配置，失败时返回默认值
func (c *Client) GetBoolOrDefault(key string, defaultValue bool) bool {
	val, err := c.GetBool(key)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetStringSlice 获取字符串切片（逗号分隔）
func (c *Client) GetStringSlice(key string) ([]string, error) {
	value, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	if value == "" {
		return []string{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

// GetStringSliceOrDefault 获取字符串切片，失败时返回默认值
func (c *Client) GetStringSliceOrDefault(key string, defaultValue []string) []string {
	val, err := c.GetStringSlice(key)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetOrDefault 获取配置，失败时返回默认值
func (c *Client) GetOrDefault(key, defaultValue string) string {
	value, err := c.Get(key)
	if err != nil {
		return defaultValue
	}
	return value
}
