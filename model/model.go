package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// DateTime SQLite 兼容的时间类型（存储为 TEXT，格式: 2006-01-02 15:04:05）
type DateTime time.Time

// Scan 实现 sql.Scanner 接口
func (d *DateTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			t, err = time.Parse(time.RFC3339, v)
			if err != nil {
				return fmt.Errorf("无法解析时间: %s", v)
			}
		}
		*d = DateTime(t)
	case time.Time:
		*d = DateTime(v)
	default:
		return fmt.Errorf("无法将 %T 转换为 DateTime", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (d DateTime) Value() (driver.Value, error) {
	t := time.Time(d)
	if t.IsZero() {
		return nil, nil
	}
	return t.Format("2006-01-02 15:04:05"), nil
}

// Time 转换回 time.Time
func (d DateTime) Time() time.Time {
	return time.Time(d)
}

// MarshalJSON JSON 序列化
func (d DateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(t.Format("2006-01-02 15:04:05"))
}

// UnmarshalJSON JSON 反序列化
func (d *DateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*d = DateTime(time.Time{})
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return err
		}
	}
	*d = DateTime(t)
	return nil
}

// Format 格式化输出
func (d DateTime) Format(layout string) string {
	return time.Time(d).Format(layout)
}

// Now 返回当前时间
func Now() DateTime {
	return DateTime(time.Now())
}

// Branch 代码分支
type Branch struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	ParentBranchID *int64    `json:"parent_branch_id"` // NULL 表示初始分支
	BranchType     string    `json:"branch_type"`      // main / release / feature / hotfix / custom
	Description    string    `json:"description"`
	IsActive       bool      `json:"is_active"`
	PulledAt       *DateTime `json:"pulled_at"` // 实际拉取分支的时间（NULL = 未填写）
	CreatedAt      DateTime  `json:"created_at"`
	UpdatedAt      DateTime  `json:"updated_at"`
}

// Version 版本记录
type Version struct {
	ID            int64    `json:"id"`
	BranchID      int64    `json:"branch_id"`
	ProductName   string   `json:"product_name"`
	VersionNumber string   `json:"version_number"` // 版本号，用户自由输入，如 v1.2.3 / V2.0 / 2024Q1
	Description   string   `json:"description"`
	ReleaseNotes  string   `json:"release_notes"`
	BuildTime     DateTime `json:"build_time"`
	CommitHash    string   `json:"commit_hash"`
	ArtifactURL   string   `json:"artifact_url"`
	Status        string   `json:"status"` // draft / released / deprecated / revoked
	CreatedAt     DateTime `json:"created_at"`

	// 关联查询字段
	BranchName string `json:"branch_name,omitempty"`
}

// VersionQuery 版本查询条件
type VersionQuery struct {
	BranchID    *int64
	ProductName string
	Status      string
	Limit       int
	Offset      int
}
