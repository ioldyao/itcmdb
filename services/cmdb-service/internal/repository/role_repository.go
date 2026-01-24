package repository

import (
	"github.com/itcmdb/cmdb-service/internal/models"
	"gorm.io/gorm"
)

type RoleRepository interface {
	// CI角色管理
	GetCIRoles() ([]models.CIRole, error)
	GetCIRoleByID(id uint) (*models.CIRole, error)
	CreateCIRole(role *models.CIRole) error
	UpdateCIRole(role *models.CIRole) error
	DeleteCIRole(id uint) error

	// 负责人角色管理
	GetOwnerRoles() ([]models.OwnerRole, error)
	GetOwnerRoleByID(id uint) (*models.OwnerRole, error)
	CreateOwnerRole(role *models.OwnerRole) error
	UpdateOwnerRole(role *models.OwnerRole) error
	DeleteOwnerRole(id uint) error

	// 角色权限管理
	GetRolePermissions() ([]models.RolePermission, error)
	GetRolePermissionByName(name string) (*models.RolePermission, error)
	CreateRolePermission(perm *models.RolePermission) error
	UpdateRolePermission(perm *models.RolePermission) error
	DeleteRolePermission(name string) error

	// CI实例角色关联
	GetCIInstanceRoles(ciID uint) ([]models.CIInstanceRole, error)
	AssignCIRole(ciID uint, roleType string, roleID uint, userID uint, assignedBy uint) error
	RemoveCIRole(ciID uint, roleType string, roleID uint, userID uint) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// ==================== CI角色管理 ====================

func (r *roleRepository) GetCIRoles() ([]models.CIRole, error) {
	var roles []models.CIRole
	err := r.db.Where("is_active = ?", true).Order("priority ASC, id ASC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetCIRoleByID(id uint) (*models.CIRole, error) {
	var role models.CIRole
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) CreateCIRole(role *models.CIRole) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) UpdateCIRole(role *models.CIRole) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) DeleteCIRole(id uint) error {
	return r.db.Delete(&models.CIRole{}, id).Error
}

// ==================== 负责人角色管理 ====================

func (r *roleRepository) GetOwnerRoles() ([]models.OwnerRole, error) {
	var roles []models.OwnerRole
	err := r.db.Where("is_active = ?", true).Order("level ASC, id ASC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetOwnerRoleByID(id uint) (*models.OwnerRole, error) {
	var role models.OwnerRole
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) CreateOwnerRole(role *models.OwnerRole) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) UpdateOwnerRole(role *models.OwnerRole) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) DeleteOwnerRole(id uint) error {
	return r.db.Delete(&models.OwnerRole{}, id).Error
}

// ==================== 角色权限管理 ====================

func (r *roleRepository) GetRolePermissions() ([]models.RolePermission, error) {
	var perms []models.RolePermission
	err := r.db.Find(&perms).Error
	return perms, err
}

func (r *roleRepository) GetRolePermissionByName(name string) (*models.RolePermission, error) {
	var perm models.RolePermission
	err := r.db.Where("role_name = ?", name).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *roleRepository) CreateRolePermission(perm *models.RolePermission) error {
	return r.db.Create(perm).Error
}

func (r *roleRepository) UpdateRolePermission(perm *models.RolePermission) error {
	return r.db.Save(perm).Error
}

func (r *roleRepository) DeleteRolePermission(name string) error {
	return r.db.Where("role_name = ?", name).Delete(&models.RolePermission{}).Error
}

// ==================== CI实例角色关联 ====================

func (r *roleRepository) GetCIInstanceRoles(ciID uint) ([]models.CIInstanceRole, error) {
	var roles []models.CIInstanceRole
	err := r.db.Preload("User").Where("ci_id = ?", ciID).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) AssignCIRole(ciID uint, roleType string, roleID uint, userID uint, assignedBy uint) error {
	ciRole := &models.CIInstanceRole{
		CIID:       ciID,
		RoleType:   roleType,
		RoleID:     roleID,
		UserID:     &userID,
		AssignedBy: &assignedBy,
	}
	return r.db.Create(ciRole).Error
}

func (r *roleRepository) RemoveCIRole(ciID uint, roleType string, roleID uint, userID uint) error {
	return r.db.Where("ci_id = ? AND role_type = ? AND role_id = ? AND user_id = ?",
		ciID, roleType, roleID, userID).
		Delete(&models.CIInstanceRole{}).Error
}
