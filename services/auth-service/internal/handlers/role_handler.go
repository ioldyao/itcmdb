package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/auth-service/internal/service"
	"github.com/itcmdb/shared/pkg/response"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// GetRoles 获取所有角色
func (h *RoleHandler) GetRoles(c *gin.Context) {
	roles, err := h.roleService.GetRoles()
	if err != nil {
		c.JSON(500, response.Error("500", "failed to get roles"))
		return
	}

	c.JSON(200, response.Success(roles))
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("400", "invalid request"))
		return
	}

	role, err := h.roleService.CreateRole(req.Name, req.Description)
	if err != nil {
		c.JSON(400, response.Error("400", err.Error()))
		return
	}

	c.JSON(200, response.Success(role))
}

// UpdateRole 更新角色
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, response.Error("400", "invalid role id"))
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("400", "invalid request"))
		return
	}

	role, err := h.roleService.UpdateRole(uint(id), req.Name, req.Description)
	if err != nil {
		c.JSON(400, response.Error("400", err.Error()))
		return
	}

	c.JSON(200, response.Success(role))
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, response.Error("400", "invalid role id"))
		return
	}

	if err := h.roleService.DeleteRole(uint(id)); err != nil {
		c.JSON(500, response.Error("500", "failed to delete role"))
		return
	}

	c.JSON(200, response.Success(nil))
}

// GetPermissions 获取所有权限
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	permissions, err := h.roleService.GetPermissions()
	if err != nil {
		c.JSON(500, response.Error("500", "failed to get permissions"))
		return
	}

	c.JSON(200, response.Success(permissions))
}

// GetRolePermissions 获取角色的权限列表
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, response.Error("400", "invalid role id"))
		return
	}

	permissions, err := h.roleService.GetRolePermissions(uint(id))
	if err != nil {
		c.JSON(500, response.Error("500", "failed to get role permissions"))
		return
	}

	c.JSON(200, response.Success(permissions))
}

// CreatePermission 创建权限
func (h *RoleHandler) CreatePermission(c *gin.Context) {
	var req struct {
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("400", "invalid request"))
		return
	}

	permission, err := h.roleService.CreatePermission(req.Resource, req.Action)
	if err != nil {
		c.JSON(500, response.Error("500", "failed to create permission"))
		return
	}

	c.JSON(200, response.Success(permission))
}

// DeletePermission 删除权限
func (h *RoleHandler) DeletePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, response.Error("400", "invalid permission id"))
		return
	}

	if err := h.roleService.DeletePermission(uint(id)); err != nil {
		c.JSON(500, response.Error("500", "failed to delete permission"))
		return
	}

	c.JSON(200, response.Success(nil))
}

// AssignPermissionToRole 为角色分配权限
func (h *RoleHandler) AssignPermissionToRole(c *gin.Context) {
	var req struct {
		RoleID       uint `json:"role_id" binding:"required"`
		PermissionID uint `json:"permission_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("400", "invalid request"))
		return
	}

	if err := h.roleService.AssignPermissionToRole(req.RoleID, req.PermissionID); err != nil {
		c.JSON(500, response.Error("500", "failed to assign permission"))
		return
	}

	c.JSON(200, response.Success(nil))
}

// RemovePermissionFromRole 移除角色权限
func (h *RoleHandler) RemovePermissionFromRole(c *gin.Context) {
	var req struct {
		RoleID       uint `json:"role_id" binding:"required"`
		PermissionID uint `json:"permission_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("400", "invalid request"))
		return
	}

	if err := h.roleService.RemovePermissionFromRole(req.RoleID, req.PermissionID); err != nil {
		c.JSON(500, response.Error("500", "failed to remove permission"))
		return
	}

	c.JSON(200, response.Success(nil))
}

// AssignRoleToUser 为用户分配角色
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
		RoleID uint `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("400", "invalid request"))
		return
	}

	if err := h.roleService.AssignRoleToUser(req.UserID, req.RoleID); err != nil {
		c.JSON(500, response.Error("500", "failed to assign role"))
		return
	}

	c.JSON(200, response.Success(nil))
}

// RemoveRoleFromUser 移除用户角色
func (h *RoleHandler) RemoveRoleFromUser(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
		RoleID uint `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("400", "invalid request"))
		return
	}

	if err := h.roleService.RemoveRoleFromUser(req.UserID, req.RoleID); err != nil {
		c.JSON(500, response.Error("500", "failed to remove role"))
		return
	}

	c.JSON(200, response.Success(nil))
}

// GetUserRoles 获取用户角色列表
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, response.Error("400", "invalid user id"))
		return
	}

	roles, err := h.roleService.GetUserRoles(uint(id))
	if err != nil {
		c.JSON(500, response.Error("500", "failed to get user roles"))
		return
	}

	c.JSON(200, response.Success(roles))
}

// GetRoleUsers 获取角色用户列表
func (h *RoleHandler) GetRoleUsers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, response.Error("400", "invalid role id"))
		return
	}

	users, err := h.roleService.GetRoleUsers(uint(id))
	if err != nil {
		c.JSON(500, response.Error("500", "failed to get role users"))
		return
	}

	c.JSON(200, response.Success(users))
}
