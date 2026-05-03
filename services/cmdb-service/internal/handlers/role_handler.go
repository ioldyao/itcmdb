package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itcmdb/cmdb-service/internal/service"
	"github.com/itcmdb/shared/pkg/response"
	"net/http"
	"strconv"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// ==================== CI角色管理 ====================

// GetCIRoles 获取CI角色列表
func (h *RoleHandler) GetCIRoles(c *gin.Context) {
	roles, err := h.roleService.GetCIRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("获取角色列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(roles))
}

// CreateCIRole 创建CI角色
func (h *RoleHandler) CreateCIRole(c *gin.Context) {
	var req service.CreateCIRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	role, err := h.roleService.CreateCIRole(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("创建角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(role))
}

// UpdateCIRole 更新CI角色
func (h *RoleHandler) UpdateCIRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var req service.UpdateCIRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	if err := h.roleService.UpdateCIRole(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("更新角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// DeleteCIRole 删除CI角色
func (h *RoleHandler) DeleteCIRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	if err := h.roleService.DeleteCIRole(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("删除角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// ==================== 负责人角色管理 ====================

// GetOwnerRoles 获取负责人角色列表
func (h *RoleHandler) GetOwnerRoles(c *gin.Context) {
	roles, err := h.roleService.GetOwnerRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("获取角色列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(roles))
}

// CreateOwnerRole 创建负责人角色
func (h *RoleHandler) CreateOwnerRole(c *gin.Context) {
	var req service.CreateOwnerRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	role, err := h.roleService.CreateOwnerRole(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("创建角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(role))
}

// UpdateOwnerRole 更新负责人角色
func (h *RoleHandler) UpdateOwnerRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var req service.UpdateOwnerRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	if err := h.roleService.UpdateOwnerRole(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("更新角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// DeleteOwnerRole 删除负责人角色
func (h *RoleHandler) DeleteOwnerRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	if err := h.roleService.DeleteOwnerRole(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("删除角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// ==================== CI实例角色关联 ====================

// GetCIInstanceRoles 获取CI实例的角色列表
func (h *RoleHandler) GetCIInstanceRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	roles, err := h.roleService.GetCIInstanceRoles(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("获取角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(roles))
}

// AssignCIRole 为CI实例分配角色
func (h *RoleHandler) AssignCIRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var req service.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	uid, ok := response.GetUserID(c)
	if !ok {
		return
	}

	if err := h.roleService.AssignCIRole(uint(id), &req, uint(uid)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("分配角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// RemoveCIRole 移除CI实例的角色
func (h *RoleHandler) RemoveCIRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var req service.RemoveRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	if err := h.roleService.RemoveCIRole(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("移除角色失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
