package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemConfig 系统配置
type SystemConfig struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Category    string         `gorm:"size:50;not null;index" json:"category"` // monitoring, database, cache, etc.
	Key         string         `gorm:"size:100;not null;uniqueIndex:idx_category_key" json:"key"`
	Value       string         `gorm:"type:text" json:"value"`
	Description string         `gorm:"size:255" json:"description"`
	IsEncrypted bool           `gorm:"default:false" json:"is_encrypted"` // 是否加密存储（如密码）
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	UpdatedBy   *uint          `gorm:"index" json:"updated_by,omitempty"` // 改为指针类型，允许NULL
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides
func (SystemConfig) TableName() string {
	return "system_configs"
}
