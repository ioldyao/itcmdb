package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONB 自定义类型用于存储动态属性
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// AuditLog 审计日志模型
type AuditLog struct {
	ID         uint       `gorm:"primarykey" json:"id"`
	UserID     *uint      `gorm:"index" json:"user_id,omitempty"`
	Action     string     `gorm:"size:100;not null;index" json:"action"`
	Resource   string     `gorm:"size:100;not null;index" json:"resource"`
	ResourceID *uint      `json:"resource_id,omitempty"`
	Details    JSONB      `gorm:"type:jsonb" json:"details"`
	IPAddress  string     `gorm:"size:45" json:"ip_address"`
	UserAgent  string     `gorm:"type:text" json:"user_agent"`
	Status     string     `gorm:"size:20;default:'success'" json:"status"` // success, failed
	ErrorMsg   string     `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}
