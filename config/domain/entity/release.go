package entity

import (
	"encoding/json"
	"time"

	baseGorm "config-client/share/repository/gorm"
)

// ReleaseStatus 发布状态
type ReleaseStatus string

const (
	// ReleaseStatusTesting 测试中
	ReleaseStatusTesting ReleaseStatus = "testing"
	// ReleaseStatusPublished 已发布
	ReleaseStatusPublished ReleaseStatus = "published"
	// ReleaseStatusRollback 已回滚
	ReleaseStatusRollback ReleaseStatus = "rollback"
)

// ReleaseType 发布类型
type ReleaseType string

const (
	// ReleaseTypeFull 全量发布
	ReleaseTypeFull ReleaseType = "full"
	// ReleaseTypeIncremental 增量发布
	ReleaseTypeIncremental ReleaseType = "incremental"
	// ReleaseTypeCanary 灰度发布
	ReleaseTypeCanary ReleaseType = "canary"
)

// Release 发布版本领域实体（聚合根）
// 用于支持配置的版本发布和灰度发布
type Release struct {
	baseGorm.BaseEntity               // 组合通用审计字段
	NamespaceID         int           `json:"namespace_id"`          // 命名空间ID
	Environment         string        `json:"environment"`           // 发布环境
	Version             int           `json:"version"`               // 版本号
	VersionName         string        `json:"version_name"`          // 版本名称，例如：v1.0.0
	ConfigSnapshot      string        `json:"config_snapshot"`       // 配置快照（JSON格式）
	ConfigCount         int           `json:"config_count"`          // 包含的配置项数量
	Status              ReleaseStatus `json:"status"`                // 状态
	ReleaseType         ReleaseType   `json:"release_type"`          // 发布类型
	CanaryRule          string        `json:"canary_rule"`           // 灰度规则（JSON格式）
	CanaryPercentage    int           `json:"canary_percentage"`     // 灰度比例（0-100）
	ReleasedBy          string        `json:"released_by"`           // 发布人
	ReleasedAt          *time.Time    `json:"released_at"`           // 发布时间
	RollbackFromVersion int           `json:"rollback_from_version"` // 从哪个版本回滚
	RollbackBy          string        `json:"rollback_by"`           // 回滚人
	RollbackAt          *time.Time    `json:"rollback_at"`           // 回滚时间
	RollbackReason      string        `json:"rollback_reason"`       // 回滚原因
}

// ConfigSnapshotItem 配置快照项
type ConfigSnapshotItem struct {
	ConfigID             int    `json:"config_id"`
	Key                  string `json:"key"`
	Value                string `json:"value"`
	ValueType            string `json:"value_type"`
	GroupName            string `json:"group_name"`
	ContentHash          string `json:"content_hash"`
	ContentHashAlgorithm string `json:"content_hash_algorithm"`
	Description          string `json:"description"`
	Version              int    `json:"version"`
}

// CanaryRule 灰度规则
type CanaryRule struct {
	ClientIDs  []string `json:"client_ids"` // 客户端ID白名单
	IPRanges   []string `json:"ip_ranges"`  // IP段白名单
	Percentage int      `json:"percentage"` // 灰度百分比（0-100）
}

// ==================== 领域行为方法 ====================

// Publish 发布版本
func (r *Release) Publish(publishedBy string) {
	r.Status = ReleaseStatusPublished
	now := time.Now()
	r.ReleasedAt = &now
	r.ReleasedBy = publishedBy
}

// Rollback 回滚版本
func (r *Release) Rollback(rollbackBy string, reason string) {
	r.Status = ReleaseStatusRollback
	now := time.Now()
	r.RollbackAt = &now
	r.RollbackBy = rollbackBy
	r.RollbackReason = reason
}

// IsPublished 是否已发布
func (r *Release) IsPublished() bool {
	return r.Status == ReleaseStatusPublished
}

// IsTesting 是否测试中
func (r *Release) IsTesting() bool {
	return r.Status == ReleaseStatusTesting
}

// IsRollback 是否已回滚
func (r *Release) IsRollback() bool {
	return r.Status == ReleaseStatusRollback
}

// IsCanaryRelease 是否灰度发布
func (r *Release) IsCanaryRelease() bool {
	return r.ReleaseType == ReleaseTypeCanary
}

// GetConfigSnapshot 获取配置快照
func (r *Release) GetConfigSnapshot() ([]ConfigSnapshotItem, error) {
	var snapshot []ConfigSnapshotItem
	if r.ConfigSnapshot == "" {
		return snapshot, nil
	}
	err := json.Unmarshal([]byte(r.ConfigSnapshot), &snapshot)
	return snapshot, err
}

// SetConfigSnapshot 设置配置快照
func (r *Release) SetConfigSnapshot(snapshot []ConfigSnapshotItem) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	r.ConfigSnapshot = string(data)
	r.ConfigCount = len(snapshot)
	return nil
}

// GetCanaryRule 获取灰度规则
func (r *Release) GetCanaryRule() (*CanaryRule, error) {
	if r.CanaryRule == "" {
		return &CanaryRule{}, nil
	}
	var rule CanaryRule
	err := json.Unmarshal([]byte(r.CanaryRule), &rule)
	return &rule, err
}

// SetCanaryRule 设置灰度规则
func (r *Release) SetCanaryRule(rule *CanaryRule) error {
	if rule == nil {
		r.CanaryRule = ""
		return nil
	}
	data, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	r.CanaryRule = string(data)
	r.CanaryPercentage = rule.Percentage
	return nil
}

// CanPublish 判断是否可以发布
func (r *Release) CanPublish() bool {
	return r.Status == ReleaseStatusTesting
}

// CanRollback 判断是否可以回滚
func (r *Release) CanRollback() bool {
	return r.Status == ReleaseStatusPublished
}
