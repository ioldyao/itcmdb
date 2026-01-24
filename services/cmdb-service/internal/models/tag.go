package models

import (
	"time"
)

// TagCategory 标签分类
type TagCategory struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	DisplayName string    `gorm:"size:100;not null" json:"display_name"`
	Description string    `gorm:"type:text" json:"description"`
	Color       string    `gorm:"size:20" json:"color"`
	Icon        string    `gorm:"size:50" json:"icon"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Tags []Tag `gorm:"foreignKey:CategoryID" json:"tags,omitempty"`
}

// Tag 标签定义
type Tag struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CategoryID  *uint     `gorm:"index" json:"category_id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	DisplayName string    `gorm:"size:100;not null" json:"display_name"`
	Color       string    `gorm:"size:20" json:"color"`
	Description string    `gorm:"type:text" json:"description"`
	UsageCount  int       `gorm:"default:0" json:"usage_count"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Category *TagCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// CITag CI实例标签关联
type CITag struct {
	ID       uint      `gorm:"primarykey" json:"id"`
	CIID     uint      `gorm:"not null;index" json:"ci_id"`
	TagID    uint      `gorm:"not null;index" json:"tag_id"`
	TaggedBy *uint     `json:"tagged_by"`
	TaggedAt time.Time `gorm:"default:now()" json:"tagged_at"`

	// 关联
	CI *CIInstance `gorm:"foreignKey:CIID" json:"ci_instance,omitempty"`
	Tag *Tag        `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

// TagHistory 标签使用历史
type TagHistory struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CIID      *uint     `json:"ci_id"`
	TagID     *uint     `json:"tag_id"`
	Action    string    `gorm:"size:20;not null" json:"action"` // added 或 removed
	UserID    *uint     `json:"user_id"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`

	// 关联
	CI  *CIInstance `gorm:"foreignKey:CIID" json:"ci_instance,omitempty"`
	Tag *Tag        `gorm:"foreignKey:TagID" json:"tag,omitempty"`
	User *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
