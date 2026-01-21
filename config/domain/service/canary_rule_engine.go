package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"

	"config-client/config/domain/entity"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// CanaryRuleEngine 灰度规则引擎
// 负责判断客户端是否符合灰度发布规则
type CanaryRuleEngine struct{}

// NewCanaryRuleEngine 创建灰度规则引擎
func NewCanaryRuleEngine() *CanaryRuleEngine {
	return &CanaryRuleEngine{}
}

// Match 判断客户端是否匹配灰度规则
// 返回: true - 匹配(使用灰度版本), false - 不匹配(使用生产版本)
func (e *CanaryRuleEngine) Match(rule *entity.CanaryRule, clientID string, clientIP string) bool {
	if rule == nil {
		return false
	}

	// 1. 检查客户端ID白名单
	if len(rule.ClientIDs) > 0 && e.matchClientID(rule.ClientIDs, clientID) {
		hlog.Infof("灰度匹配: 客户端ID白名单命中, clientID=%s", clientID)
		return true
	}

	// 2. 检查IP段白名单
	if len(rule.IPRanges) > 0 && e.matchIPRange(rule.IPRanges, clientIP) {
		hlog.Infof("灰度匹配: IP段白名单命中, clientIP=%s", clientIP)
		return true
	}

	// 3. 检查灰度百分比（哈希分流）
	if rule.Percentage > 0 && e.matchPercentage(rule.Percentage, clientID) {
		hlog.Infof("灰度匹配: 百分比命中, clientID=%s, percentage=%d", clientID, rule.Percentage)
		return true
	}

	return false
}

// matchClientID 检查客户端ID是否在白名单中
func (e *CanaryRuleEngine) matchClientID(whitelist []string, clientID string) bool {
	for _, id := range whitelist {
		if id == clientID {
			return true
		}
		// 支持通配符匹配
		if e.matchWildcard(id, clientID) {
			return true
		}
	}
	return false
}

// matchIPRange 检查IP是否在IP段白名单中
func (e *CanaryRuleEngine) matchIPRange(ipRanges []string, clientIP string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		hlog.Warnf("无效的IP地址: %s", clientIP)
		return false
	}

	for _, ipRange := range ipRanges {
		// 支持CIDR格式: 192.168.1.0/24
		if strings.Contains(ipRange, "/") {
			_, ipNet, err := net.ParseCIDR(ipRange)
			if err != nil {
				hlog.Warnf("无效的CIDR格式: %s, error: %v", ipRange, err)
				continue
			}
			if ipNet.Contains(ip) {
				return true
			}
		} else {
			// 支持单个IP匹配
			if clientIP == ipRange {
				return true
			}
		}
	}
	return false
}

// matchPercentage 基于客户端ID哈希的百分比分流
// 使用一致性哈希算法,确保同一客户端的分流结果稳定
func (e *CanaryRuleEngine) matchPercentage(percentage int, clientID string) bool {
	if percentage <= 0 || percentage >= 100 {
		return false
	}

	// 计算客户端ID的MD5哈希值
	hash := md5.Sum([]byte(clientID))
	hashStr := hex.EncodeToString(hash[:])

	// 取哈希值的前8位作为整数
	hashInt, err := strconv.ParseUint(hashStr[:8], 16, 32)
	if err != nil {
		hlog.Warnf("哈希值转换失败: %v", err)
		return false
	}

	// 计算模100的余数,判断是否小于百分比
	return int(hashInt%100) < percentage
}

// matchWildcard 通配符匹配
// 支持 * 匹配任意字符
func (e *CanaryRuleEngine) matchWildcard(pattern, str string) bool {
	// 简单的通配符实现
	if !strings.Contains(pattern, "*") {
		return pattern == str
	}

	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == str
	}

	// 检查前缀
	if !strings.HasPrefix(str, parts[0]) {
		return false
	}
	str = strings.TrimPrefix(str, parts[0])

	// 检查后缀
	if len(parts) > 1 {
		suffix := parts[len(parts)-1]
		if !strings.HasSuffix(str, suffix) {
			return false
		}
		str = strings.TrimSuffix(str, suffix)
	}

	// 检查中间部分
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(str, parts[i])
		if idx == -1 {
			return false
		}
		str = str[idx+len(parts[i]):]
	}

	return true
}

// ValidateRule 验证灰度规则的有效性
func (e *CanaryRuleEngine) ValidateRule(rule *entity.CanaryRule) error {
	if rule == nil {
		return fmt.Errorf("灰度规则不能为空")
	}

	// 验证百分比范围
	if rule.Percentage < 0 || rule.Percentage > 100 {
		return fmt.Errorf("灰度百分比必须在0-100之间: %d", rule.Percentage)
	}

	// 验证IP段格式
	for _, ipRange := range rule.IPRanges {
		if strings.Contains(ipRange, "/") {
			_, _, err := net.ParseCIDR(ipRange)
			if err != nil {
				return fmt.Errorf("无效的CIDR格式: %s, error: %v", ipRange, err)
			}
		} else {
			ip := net.ParseIP(ipRange)
			if ip == nil {
				return fmt.Errorf("无效的IP地址: %s", ipRange)
			}
		}
	}

	// 至少需要一个规则
	if len(rule.ClientIDs) == 0 && len(rule.IPRanges) == 0 && rule.Percentage == 0 {
		return fmt.Errorf("灰度规则至少需要指定一个条件(客户端ID、IP段或百分比)")
	}

	return nil
}

// CalculateCanaryClients 计算灰度客户端数量（基于百分比）
// 用于预估灰度影响范围
func (e *CanaryRuleEngine) CalculateCanaryClients(totalClients int, percentage int) int {
	if percentage <= 0 || totalClients <= 0 {
		return 0
	}
	return (totalClients * percentage) / 100
}
