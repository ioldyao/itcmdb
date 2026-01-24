package repository

import (
	"github.com/itcmdb/auth-service/internal/models"
	"github.com/itcmdb/shared/pkg/database"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByUsername(username string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	GetUserPermissions(userID uint) ([]string, error)
}

type userRepository struct{}

// NewUserRepository 创建用户仓库实例
func NewUserRepository() UserRepository {
	return &userRepository{}
}

// Create 创建用户
func (r *userRepository) Create(user *models.User) error {
	db := database.Get()
	if err := db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// FindByUsername 根据用户名查找用户
func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	db := database.Get()
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据ID查找用户
func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	db := database.Get()
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers 获取所有用户
func (r *userRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	db := database.Get()
	err := db.Find(&users).Error
	return users, err
}

// Update 更新用户
func (r *userRepository) Update(user *models.User) error {
	db := database.Get()
	return db.Save(user).Error
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	db := database.Get()
	return db.Delete(&models.User{}, id).Error
}

// GetUserPermissions 获取用户所有权限
func (r *userRepository) GetUserPermissions(userID uint) ([]string, error) {
	db := database.Get()

	// 通过角色获取的权限
	var rolePermissions []string
	err := db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Pluck("CONCAT(resource, ':', action)", &rolePermissions).
		Error

	if err != nil {
		return nil, err
	}

	// 直接分配给用户的权限
	var userPermissions []string
	err = db.Table("permissions").
		Joins("JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ?", userID).
		Pluck("CONCAT(resource, ':', action)", &userPermissions).
		Error

	if err != nil {
		return nil, err
	}

	// 合并权限并去重
	permissionMap := make(map[string]bool)
	for _, p := range rolePermissions {
		permissionMap[p] = true
	}
	for _, p := range userPermissions {
		permissionMap[p] = true
	}

	// 转换为切片
	permissions := make([]string, 0, len(permissionMap))
	for p := range permissionMap {
		permissions = append(permissions, p)
	}

	return permissions, nil
}
