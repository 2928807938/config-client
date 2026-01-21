package vo

import "time"

// ReleaseVO 发布版本值对象
type ReleaseVO struct {
	ID                  int                    `json:"id"`
	NamespaceID         int                    `json:"namespace_id"`
	Environment         string                 `json:"environment"`
	Version             int                    `json:"version"`
	VersionName         string                 `json:"version_name"`
	ConfigCount         int                    `json:"config_count"`
	Status              string                 `json:"status"`
	ReleaseType         string                 `json:"release_type"`
	CanaryPercentage    int                    `json:"canary_percentage"`
	CanaryRule          *CanaryRuleVO          `json:"canary_rule,omitempty"`
	ConfigSnapshot      []ConfigSnapshotItemVO `json:"config_snapshot,omitempty"`
	ReleasedBy          string                 `json:"released_by"`
	ReleasedAt          *time.Time             `json:"released_at"`
	RollbackFromVersion int                    `json:"rollback_from_version"`
	RollbackBy          string                 `json:"rollback_by"`
	RollbackAt          *time.Time             `json:"rollback_at"`
	RollbackReason      string                 `json:"rollback_reason"`
	CreatedBy           string                 `json:"created_by"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// CanaryRuleVO 灰度规则值对象
type CanaryRuleVO struct {
	ClientIDs  []string `json:"client_ids"`
	IPRanges   []string `json:"ip_ranges"`
	Percentage int      `json:"percentage"`
}

// ConfigSnapshotItemVO 配置快照项值对象
type ConfigSnapshotItemVO struct {
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

// ReleaseListVO 发布版本列表值对象
type ReleaseListVO struct {
	Items []*ReleaseVO `json:"items"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}

// ReleaseCompareVO 版本对比值对象
type ReleaseCompareVO struct {
	FromReleaseID int                    `json:"from_release_id"`
	ToReleaseID   int                    `json:"to_release_id"`
	FromVersion   int                    `json:"from_version"`
	ToVersion     int                    `json:"to_version"`
	Added         []ConfigSnapshotItemVO `json:"added"`
	Deleted       []ConfigSnapshotItemVO `json:"deleted"`
	Modified      []ConfigDiffVO         `json:"modified"`
}

// ConfigDiffVO 配置差异值对象
type ConfigDiffVO struct {
	Key      string `json:"key"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}
