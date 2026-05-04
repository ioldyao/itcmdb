package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/auth-service/internal/constants"
	"github.com/itcmdb/auth-service/internal/service"
	"github.com/itcmdb/shared/pkg/audit"
	"github.com/itcmdb/shared/pkg/logger"
	"github.com/itcmdb/shared/pkg/rbac"
	"github.com/itcmdb/shared/pkg/response"
	"go.uber.org/zap"
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
		audit.LogError(c, "create", "roles", nil, err.Error(), req)
		return
	}

	audit.LogSuccess(c, "create", "roles", &role.ID, map[string]interface{}{
		"role_name":     role.Name,
		"role_id":       role.ID,
		"description":   req.Description,
		"created_by":     "user",
	})

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
		audit.LogError(c, "update", "roles", nil, err.Error(), req)
		return
	}

	roleID := uint(role.ID)
	audit.LogSuccess(c, "update", "roles", &roleID, map[string]interface{}{
		"role_name":   role.Name,
		"description": req.Description,
	})

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

	// 删除前查出角色信息用于审计
	role, _ := h.roleService.GetRoleByID(uint(id))

	if err := h.roleService.DeleteRole(uint(id)); err != nil {
		c.JSON(500, response.Error("500", "failed to delete role"))
		audit.LogError(c, "delete", "roles", nil, err.Error(), nil)
		return
	}

	auditID := uint(id)
	details := map[string]interface{}{}
	if role != nil {
		details["role_name"] = role.Name
	}
	audit.LogSuccess(c, "delete", "roles", &auditID, details)

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

	// 验证资源和操作是否有效
	if !constants.IsValidResource(req.Resource) {
		c.JSON(400, response.Error("400", "invalid resource: "+req.Resource+". Please use one of the predefined resources."))
		return
	}

	if !constants.IsValidAction(req.Action) {
		c.JSON(400, response.Error("400", "invalid action: "+req.Action+". Please use one of the predefined actions."))
		return
	}

	permission, err := h.roleService.CreatePermission(req.Resource, req.Action)
	if err != nil {
		c.JSON(500, response.Error("500", "failed to create permission"))
		audit.LogError(c, "create", "permission", nil, err.Error(), req)
		return
	}

	permID := uint(permission.ID)
	audit.LogSuccess(c, "create", "permission", &permID, map[string]interface{}{
		"resource": req.Resource,
		"action":   req.Action,
	})

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
		audit.LogError(c, "delete", "permission", nil, err.Error(), nil)
		return
	}

	auditID := uint(id)
	audit.LogSuccess(c, "delete", "permission", &auditID, nil)

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
		audit.LogError(c, "assign_permission", "role", nil, err.Error(), req)
		return
	}

	// 清除该角色所有用户的权限缓存
	h.clearRoleUsersCache(c, req.RoleID)

	roleID := uint(req.RoleID)
	audit.LogSuccess(c, "assign_permission", "role", &roleID, map[string]interface{}{
		"permission_id": req.PermissionID,
	})

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
		audit.LogError(c, "remove_permission", "role", nil, err.Error(), req)
		return
	}

	// 清除该角色所有用户的权限缓存
	h.clearRoleUsersCache(c, req.RoleID)

	roleID := uint(req.RoleID)
	audit.LogSuccess(c, "remove_permission", "role", &roleID, map[string]interface{}{
		"permission_id": req.PermissionID,
	})

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
		audit.LogError(c, "assign_role", "user", nil, err.Error(), req)
		return
	}

	// 清除用户权限缓存
	if err := rbac.ClearUserPermissionsCache(c.Request.Context(), int64(req.UserID)); err != nil {
		logger.Warn("Failed to clear user permission cache", zap.Error(err), zap.Uint("user_id", req.UserID))
	}

	userID := uint(req.UserID)
	audit.LogSuccess(c, "assign_role", "user", &userID, map[string]interface{}{
		"role_id": req.RoleID,
	})

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
		audit.LogError(c, "remove_role", "user", nil, err.Error(), req)
		return
	}

	// 清除用户权限缓存
	if err := rbac.ClearUserPermissionsCache(c.Request.Context(), int64(req.UserID)); err != nil {
		logger.Warn("Failed to clear user permission cache", zap.Error(err), zap.Uint("user_id", req.UserID))
	}

	userID := uint(req.UserID)
	audit.LogSuccess(c, "remove_role", "user", &userID, map[string]interface{}{
		"role_id": req.RoleID,
	})

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

// GetValidResources 获取所有有效的资源类型
func (h *RoleHandler) GetValidResources(c *gin.Context) {
	c.JSON(200, response.Success(constants.ValidResources))
}

// GetValidActions 获取所有有效的操作类型
func (h *RoleHandler) GetValidActions(c *gin.Context) {
	c.JSON(200, response.Success(constants.ValidActions))
}

// clearRoleUsersCache 清除角色下所有用户的权限缓存
func (h *RoleHandler) clearRoleUsersCache(c *gin.Context, roleID uint) {
	users, err := h.roleService.GetRoleUsers(roleID)
	if err != nil {
		logger.Warn("Failed to get role users for cache clearing", zap.Error(err), zap.Uint("role_id", roleID))
		return
	}

	for _, user := range users {
		if err := rbac.ClearUserPermissionsCache(c.Request.Context(), int64(user.ID)); err != nil {
			logger.Warn("Failed to clear user permission cache", zap.Error(err), zap.Uint("user_id", user.ID))
		}
	}

	logger.Info("Cleared permission cache for role users", zap.Uint("role_id", roleID), zap.Int("user_count", len(users)))
}
