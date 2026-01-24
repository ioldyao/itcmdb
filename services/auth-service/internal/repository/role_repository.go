package repository

import (
	"github.com/itcmdb/auth-service/internal/models"
	"gorm.io/gorm"
)

type RoleRepository interface {
	// Role operations
	GetRoles() ([]models.Role, error)
	GetRoleByID(id uint) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	CreateRole(role *models.Role) error
	UpdateRole(role *models.Role) error
	DeleteRole(id uint) error

	// Permission operations
	GetPermissions() ([]models.Permission, error)
	GetPermissionsByRoleID(roleID uint) ([]models.Permission, error)
	CreatePermission(permission *models.Permission) error
	DeletePermission(id uint) error

	// Role-Permission association
	AssignPermissionToRole(roleID, permissionID uint) error
	RemovePermissionFromRole(roleID, permissionID uint) error

	// User-Role association
	AssignRoleToUser(userID, roleID uint) error
	RemoveRoleFromUser(userID, roleID uint) error
	GetUserRoles(userID uint) ([]models.Role, error)
	GetRoleUsers(roleID uint) ([]models.User, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetRoles() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetRoleByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) CreateRole(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) UpdateRole(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) DeleteRole(id uint) error {
	return r.db.Delete(&models.Role{}, id).Error
}

func (r *roleRepository) GetPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) GetPermissionsByRoleID(roleID uint) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) CreatePermission(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

func (r *roleRepository) DeletePermission(id uint) error {
	return r.db.Delete(&models.Permission{}, id).Error
}

func (r *roleRepository) AssignPermissionToRole(roleID, permissionID uint) error {
	rolePermission := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	return r.db.Create(rolePermission).Error
}

func (r *roleRepository) RemovePermissionFromRole(roleID, permissionID uint) error {
	return r.db.
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{}).Error
}

func (r *roleRepository) AssignRoleToUser(userID, roleID uint) error {
	userRole := &models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.Create(userRole).Error
}

func (r *roleRepository) RemoveRoleFromUser(userID, roleID uint) error {
	return r.db.
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error
}

func (r *roleRepository) GetUserRoles(userID uint) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetRoleUsers(roleID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id = ?", roleID).
		Find(&users).Error
	return users, err
}
