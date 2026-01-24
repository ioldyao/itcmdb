package service

import (
	"errors"
	"github.com/itcmdb/cmdb-service/internal/models"
	"github.com/itcmdb/cmdb-service/internal/repository"
)

type RoleService interface {
	// CI角色管理
	GetCIRoles() ([]models.CIRole, error)
	CreateCIRole(req *CreateCIRoleRequest) (*models.CIRole, error)
	UpdateCIRole(id uint, req *UpdateCIRoleRequest) error
	DeleteCIRole(id uint) error

	// 负责人角色管理
	GetOwnerRoles() ([]models.OwnerRole, error)
	CreateOwnerRole(req *CreateOwnerRoleRequest) (*models.OwnerRole, error)
	UpdateOwnerRole(id uint, req *UpdateOwnerRoleRequest) error
	DeleteOwnerRole(id uint) error

	// CI实例角色关联
	GetCIInstanceRoles(ciID uint) ([]models.CIInstanceRole, error)
	AssignCIRole(ciID uint, req *AssignRoleRequest, userID uint) error
	RemoveCIRole(ciID uint, req *RemoveRoleRequest) error
}

type roleService struct {
	roleRepo repository.RoleRepository
	ciRepo   repository.CIRepository
}

func NewRoleService(roleRepo repository.RoleRepository, ciRepo repository.CIRepository) RoleService {
	return &roleService{
		roleRepo: roleRepo,
		ciRepo:   ciRepo,
	}
}

type CreateCIRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	Priority    int    `json:"priority"`
}

type UpdateCIRoleRequest struct {
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	Priority    *int    `json:"priority"`
	IsActive    *bool   `json:"is_active"`
}

type CreateOwnerRoleRequest struct {
	Name            string                 `json:"name" binding:"required"`
	DisplayName     string                 `json:"display_name" binding:"required"`
	Description     string                 `json:"description"`
	Level           int                    `json:"level"`
	Responsibilities map[string]interface{} `json:"responsibilities"`
}

type UpdateOwnerRoleRequest struct {
	DisplayName     *string                `json:"display_name"`
	Description     *string                `json:"description"`
	Level           *int                   `json:"level"`
	Responsibilities map[string]interface{} `json:"responsibilities"`
	IsActive        *bool                  `json:"is_active"`
}

type AssignRoleRequest struct {
	RoleType string `json:"role_type" binding:"required,oneof=ci_role owner_role"`
	RoleID   uint   `json:"role_id" binding:"required"`
	UserID   *uint  `json:"user_id"`
}

type RemoveRoleRequest struct {
	RoleType string `json:"role_type" binding:"required,oneof=ci_role owner_role"`
	RoleID   uint   `json:"role_id" binding:"required"`
	UserID   *uint  `json:"user_id"`
}

// ==================== CI角色管理 ====================

func (s *roleService) GetCIRoles() ([]models.CIRole, error) {
	return s.roleRepo.GetCIRoles()
}

func (s *roleService) CreateCIRole(req *CreateCIRoleRequest) (*models.CIRole, error) {
	role := &models.CIRole{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		Priority:    req.Priority,
		IsActive:    true,
	}

	if err := s.roleRepo.CreateCIRole(role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) UpdateCIRole(id uint, req *UpdateCIRoleRequest) error {
	role, err := s.roleRepo.GetCIRoleByID(id)
	if err != nil {
		return err
	}

	if req.DisplayName != nil {
		role.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Color != nil {
		role.Color = *req.Color
	}
	if req.Icon != nil {
		role.Icon = *req.Icon
	}
	if req.Priority != nil {
		role.Priority = *req.Priority
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	return s.roleRepo.UpdateCIRole(role)
}

func (s *roleService) DeleteCIRole(id uint) error {
	// 检查是否有关联的CI实例
	roles, err := s.roleRepo.GetCIInstanceRoles(id)
	if err == nil && len(roles) > 0 {
		return errors.New("角色正在使用中，无法删除")
	}

	return s.roleRepo.DeleteCIRole(id)
}

// ==================== 负责人角色管理 ====================

func (s *roleService) GetOwnerRoles() ([]models.OwnerRole, error) {
	return s.roleRepo.GetOwnerRoles()
}

func (s *roleService) CreateOwnerRole(req *CreateOwnerRoleRequest) (*models.OwnerRole, error) {
	role := &models.OwnerRole{
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		Level:           req.Level,
		Responsibilities: models.JSONB(req.Responsibilities),
		IsActive:         true,
	}

	if err := s.roleRepo.CreateOwnerRole(role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) UpdateOwnerRole(id uint, req *UpdateOwnerRoleRequest) error {
	role, err := s.roleRepo.GetOwnerRoleByID(id)
	if err != nil {
		return err
	}

	if req.DisplayName != nil {
		role.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Level != nil {
		role.Level = *req.Level
	}
	if req.Responsibilities != nil {
		role.Responsibilities = models.JSONB(req.Responsibilities)
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	return s.roleRepo.UpdateOwnerRole(role)
}

func (s *roleService) DeleteOwnerRole(id uint) error {
	return s.roleRepo.DeleteOwnerRole(id)
}

// ==================== CI实例角色关联 ====================

func (s *roleService) GetCIInstanceRoles(ciID uint) ([]models.CIInstanceRole, error) {
	// 验证CI存在
	_, err := s.ciRepo.GetCIInstanceByID(ciID)
	if err != nil {
		return nil, errors.New("CI实例不存在")
	}

	return s.roleRepo.GetCIInstanceRoles(ciID)
}

func (s *roleService) AssignCIRole(ciID uint, req *AssignRoleRequest, userID uint) error {
	// 验证CI存在
	_, err := s.ciRepo.GetCIInstanceByID(ciID)
	if err != nil {
		return errors.New("CI实例不存在")
	}

	// 验证角色存在
	if req.RoleType == "ci_role" {
		_, err := s.roleRepo.GetCIRoleByID(req.RoleID)
		if err != nil {
			return errors.New("CI角色不存在")
		}
	} else {
		_, err := s.roleRepo.GetOwnerRoleByID(req.RoleID)
		if err != nil {
			return errors.New("负责人角色不存在")
		}
	}

	// 如果没有指定用户，使用当前用户
	assignUserID := userID
	if req.UserID != nil {
		assignUserID = *req.UserID
	}

	return s.roleRepo.AssignCIRole(ciID, req.RoleType, req.RoleID, assignUserID, userID)
}

func (s *roleService) RemoveCIRole(ciID uint, req *RemoveRoleRequest) error {
	// 如果没有指定用户，删除该角色类型的所有关联
	if req.UserID == nil {
		return s.roleRepo.RemoveCIRole(ciID, req.RoleType, req.RoleID, 0)
	}

	return s.roleRepo.RemoveCIRole(ciID, req.RoleType, req.RoleID, *req.UserID)
}
