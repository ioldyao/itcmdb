package models

import (
	"time"

	"github.com/itcmdb/shared/pkg/database"
)

// User 用户模型
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email        string     `gorm:"uniqueIndex;not null;size:100" json:"email"`
	PasswordHash string     `gorm:"column:password_hash;not null;size:255" json:"-"`
	FullName     string     `gorm:"size:100" json:"full_name"`
	Status       string     `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"column:user_id" json:"user_id"`
	RoleID uint `gorm:"column:role_id" json:"role_id"`
}

// Role 角色模型
type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null;size:50" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Resource  string    `gorm:"not null;size:50" json:"resource"`
	Action    string    `gorm:"not null;size:50" json:"action"`
}

// RolePermission 角色权限关联
type RolePermission struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	RoleID       uint `gorm:"column:role_id" json:"role_id"`
	PermissionID uint `gorm:"column:permission_id" json:"permission_id"`
}

// UserPermission 用户权限关联（直接权限）
type UserPermission struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	UserID       uint `gorm:"column:user_id" json:"user_id"`
	PermissionID uint `gorm:"column:permission_id" json:"permission_id"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

// AutoMigrate 自动迁移表结构
func AutoMigrate() error {
	db := database.Get()
	return db.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
		&UserRole{},
		&RolePermission{},
		&UserPermission{},
	)
}
