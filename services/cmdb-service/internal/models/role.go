package models

import (
	"time"
	"gorm.io/gorm"
)

// CIRole CI技术角色
type CIRole struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	DisplayName string    `gorm:"size:100;not null" json:"display_name"`
	Description string    `gorm:"type:text" json:"description"`
	Color       string    `gorm:"size:20" json:"color"`
	Icon        string    `gorm:"size:50" json:"icon"`
	Priority    int       `gorm:"default:0" json:"priority"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OwnerRole 负责人角色
type OwnerRole struct {
	ID              uint             `gorm:"primarykey" json:"id"`
	Name            string           `gorm:"size:50;uniqueIndex;not null" json:"name"`
	DisplayName     string           `gorm:"size:100;not null" json:"display_name"`
	Description     string           `gorm:"type:text" json:"description"`
	Level           int              `gorm:"default:0" json:"level"`
	Responsibilities JSONB            `gorm:"type:jsonb;default:'{}'" json:"responsibilities"`
	IsActive         bool             `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// RolePermission 角色权限
type RolePermission struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RoleName    string    `gorm:"size:50;uniqueIndex;not null" json:"role_name"`
	Permissions JSONB     `gorm:"type:jsonb;not null" json:"permissions"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for CMDB RolePermission
func (RolePermission) TableName() string {
	return "cmdb_role_permissions"
}

// CIInstanceRole CI实例角色关联
type CIInstanceRole struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CIID       uint      `gorm:"not null;index" json:"ci_id"`
	RoleType   string    `gorm:"size:20;not null" json:"role_type"` // ci_role 或 owner_role
	RoleID     uint      `gorm:"not null" json:"role_id"`
	UserID     *uint     `json:"user_id"`
	AssignedAt time.Time `gorm:"default:now()" json:"assigned_at"`
	AssignedBy *uint     `json:"assigned_by"`

	// 关联
	CIInstance *CIInstance `gorm:"foreignKey:CIID" json:"ci_instance,omitempty"`
	User       *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// User 用户模型（简化版，用于关联）
type User struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Username string `gorm:"size:50;not null" json:"username"`
	FullName string `gorm:"size:100" json:"full_name"`
	Email    string `gorm:"size:100" json:"email"`
}

// BeforeCreate GORM hook
func (c *CIInstanceRole) BeforeCreate(tx *gorm.DB) error {
	// 验证 role_type
	if c.RoleType != "ci_role" && c.RoleType != "owner_role" {
		return gorm.ErrInvalidData
	}
	return nil
}
