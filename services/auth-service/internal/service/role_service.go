package service

import (
	"errors"
	"github.com/itcmdb/auth-service/internal/models"
	"github.com/itcmdb/auth-service/internal/repository"
)

type RoleService interface {
	// Role operations
	GetRoles() ([]models.Role, error)
	GetRoleByID(id uint) (*models.Role, error)
	CreateRole(name, description string) (*models.Role, error)
	UpdateRole(id uint, name, description string) (*models.Role, error)
	DeleteRole(id uint) error

	// Permission operations
	GetPermissions() ([]models.Permission, error)
	GetRolePermissions(roleID uint) ([]models.Permission, error)
	CreatePermission(resource, action string) (*models.Permission, error)
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

type roleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) GetRoles() ([]models.Role, error) {
	return s.roleRepo.GetRoles()
}

func (s *roleService) GetRoleByID(id uint) (*models.Role, error) {
	return s.roleRepo.GetRoleByID(id)
}

func (s *roleService) CreateRole(name, description string) (*models.Role, error) {
	// Check if role already exists
	_, err := s.roleRepo.GetRoleByName(name)
	if err == nil {
		return nil, errors.New("role already exists")
	}

	role := &models.Role{
		Name:        name,
		Description: description,
	}

	if err := s.roleRepo.CreateRole(role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) UpdateRole(id uint, name, description string) (*models.Role, error) {
	role, err := s.roleRepo.GetRoleByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	role.Name = name
	role.Description = description

	if err := s.roleRepo.UpdateRole(role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) DeleteRole(id uint) error {
	return s.roleRepo.DeleteRole(id)
}

func (s *roleService) GetPermissions() ([]models.Permission, error) {
	return s.roleRepo.GetPermissions()
}

func (s *roleService) GetRolePermissions(roleID uint) ([]models.Permission, error) {
	return s.roleRepo.GetPermissionsByRoleID(roleID)
}

func (s *roleService) CreatePermission(resource, action string) (*models.Permission, error) {
	permission := &models.Permission{
		Resource: resource,
		Action:   action,
	}

	if err := s.roleRepo.CreatePermission(permission); err != nil {
		return nil, err
	}

	return permission, nil
}

func (s *roleService) DeletePermission(id uint) error {
	return s.roleRepo.DeletePermission(id)
}

func (s *roleService) AssignPermissionToRole(roleID, permissionID uint) error {
	return s.roleRepo.AssignPermissionToRole(roleID, permissionID)
}

func (s *roleService) RemovePermissionFromRole(roleID, permissionID uint) error {
	return s.roleRepo.RemovePermissionFromRole(roleID, permissionID)
}

func (s *roleService) AssignRoleToUser(userID, roleID uint) error {
	return s.roleRepo.AssignRoleToUser(userID, roleID)
}

func (s *roleService) RemoveRoleFromUser(userID, roleID uint) error {
	return s.roleRepo.RemoveRoleFromUser(userID, roleID)
}

func (s *roleService) GetUserRoles(userID uint) ([]models.Role, error) {
	return s.roleRepo.GetUserRoles(userID)
}

func (s *roleService) GetRoleUsers(roleID uint) ([]models.User, error) {
	return s.roleRepo.GetRoleUsers(roleID)
}
