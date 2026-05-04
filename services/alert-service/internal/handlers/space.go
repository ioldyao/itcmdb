package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/shared/pkg/response"
	"gorm.io/gorm"
)

// SpaceHandler 空间处理器
type SpaceHandler struct {
	db *gorm.DB
}

func NewSpaceHandler(db *gorm.DB) *SpaceHandler {
	return &SpaceHandler{db: db}
}

// ============================================
// 空间管理
// ============================================

// ListSpaces 获取空间列表
func (h *SpaceHandler) ListSpaces(c *gin.Context) {
	var spaces []models.AlertSpace
	if err := h.db.Order("id ASC").Find(&spaces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	// 附加每个空间关联的角色
	type SpaceWithRoles struct {
		models.AlertSpace
		Roles []models.Role `json:"roles"`
	}
	var result []SpaceWithRoles
	for _, s := range spaces {
		swr := SpaceWithRoles{AlertSpace: s, Roles: []models.Role{}}
		var roleIDs []int
		h.db.Model(&models.AlertSpaceRole{}).Where("space_id = ?", s.ID).Pluck("role_id", &roleIDs)
		if len(roleIDs) > 0 {
			h.db.Where("id IN ?", roleIDs).Find(&swr.Roles)
		}
		result = append(result, swr)
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// CreateSpace 创建空间
func (h *SpaceHandler) CreateSpace(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		RoleIDs     []int  `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	space := models.AlertSpace{Name: req.Name, Description: req.Description}
	if err := h.db.Create(&space).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("创建失败", err.Error()))
		return
	}

	// 关联角色
	for _, roleID := range req.RoleIDs {
		h.db.Create(&models.AlertSpaceRole{SpaceID: space.ID, RoleID: roleID})
	}

	c.JSON(http.StatusOK, response.Success(space))
}

// UpdateSpace 更新空间
func (h *SpaceHandler) UpdateSpace(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var space models.AlertSpace
	if err := h.db.Where("id = ?", id).First(&space).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("空间不存在", ""))
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		RoleIDs     *[]int  `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	if req.Name != nil {
		space.Name = *req.Name
	}
	if req.Description != nil {
		space.Description = *req.Description
	}
	h.db.Save(&space)

	// 更新角色关联
	if req.RoleIDs != nil {
		h.db.Where("space_id = ?", id).Delete(&models.AlertSpaceRole{})
		for _, roleID := range *req.RoleIDs {
			h.db.Create(&models.AlertSpaceRole{SpaceID: id, RoleID: roleID})
		}
	}

	c.JSON(http.StatusOK, response.Success(space))
}

// DeleteSpace 删除空间
func (h *SpaceHandler) DeleteSpace(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	// 清理关联
	h.db.Where("space_id = ?", id).Delete(&models.AlertSpaceRole{})
	h.db.Where("space_id = ?", id).Delete(&models.AlertSpaceRoute{})

	result := h.db.Delete(&models.AlertSpace{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("删除失败", result.Error.Error()))
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.Error("空间不存在", ""))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// ============================================
// 路由规则管理
// ============================================

// ListSpaceRoutes 获取路由规则列表
func (h *SpaceHandler) ListSpaceRoutes(c *gin.Context) {
	var routes []models.AlertSpaceRoute
	if err := h.db.Order("priority ASC, id ASC").Find(&routes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	// 附加空间名
	type RouteWithSpace struct {
		models.AlertSpaceRoute
		SpaceName string `json:"space_name"`
	}
	var result []RouteWithSpace
	for _, r := range routes {
		rws := RouteWithSpace{AlertSpaceRoute: r, SpaceName: ""}
		var space models.AlertSpace
		if h.db.Where("id = ?", r.SpaceID).First(&space).Error == nil {
			rws.SpaceName = space.Name
		}
		result = append(result, rws)
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// CreateSpaceRoute 创建路由规则
func (h *SpaceHandler) CreateSpaceRoute(c *gin.Context) {
	var req struct {
		FieldName  string `json:"field_name" binding:"required"`
		FieldValue string `json:"field_value" binding:"required"`
		SpaceID    int    `json:"space_id" binding:"required"`
		Priority   int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	// 验证空间存在
	var space models.AlertSpace
	if err := h.db.Where("id = ?", req.SpaceID).First(&space).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.Error("空间不存在", ""))
		return
	}

	route := models.AlertSpaceRoute{
		FieldName:  req.FieldName,
		FieldValue: req.FieldValue,
		SpaceID:    req.SpaceID,
		Priority:   req.Priority,
		Enabled:    true,
	}
	if err := h.db.Create(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("创建失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(route))
}

// UpdateSpaceRoute 更新路由规则
func (h *SpaceHandler) UpdateSpaceRoute(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var route models.AlertSpaceRoute
	if err := h.db.Where("id = ?", id).First(&route).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("路由规则不存在", ""))
		return
	}

	var req struct {
		FieldName  *string `json:"field_name"`
		FieldValue *string `json:"field_value"`
		SpaceID    *int    `json:"space_id"`
		Priority   *int    `json:"priority"`
		Enabled    *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	if req.FieldName != nil {
		route.FieldName = *req.FieldName
	}
	if req.FieldValue != nil {
		route.FieldValue = *req.FieldValue
	}
	if req.SpaceID != nil {
		route.SpaceID = *req.SpaceID
	}
	if req.Priority != nil {
		route.Priority = *req.Priority
	}
	if req.Enabled != nil {
		route.Enabled = *req.Enabled
	}

	h.db.Save(&route)
	c.JSON(http.StatusOK, response.Success(route))
}

// DeleteSpaceRoute 删除路由规则
func (h *SpaceHandler) DeleteSpaceRoute(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	result := h.db.Delete(&models.AlertSpaceRoute{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("删除失败", result.Error.Error()))
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.Error("路由规则不存在", ""))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// MatchSpaceByLabels 根据告警标签匹配空间，返回命中的空间ID列表
func (h *SpaceHandler) MatchSpaceByLabels(labels map[string]interface{}) []int {
	var routes []models.AlertSpaceRoute
	h.db.Where("enabled = ?", true).Order("priority ASC").Find(&routes)

	spaceIDs := make(map[int]bool)
	for _, route := range routes {
		if val, ok := labels[route.FieldName]; ok {
			if val == route.FieldValue {
				spaceIDs[route.SpaceID] = true
			}
		}
	}

	var ids []int
	for id := range spaceIDs {
		ids = append(ids, id)
	}
	return ids
}

// ListRoles 获取角色列表（从共享DB直接读取）
func (h *SpaceHandler) ListRoles(c *gin.Context) {
	var roles []models.Role
	if err := h.db.Order("id ASC").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(roles))
}

// GetSpaceUserIDs 获取空间下所有用户ID（通过角色关联）
func (h *SpaceHandler) GetSpaceUserIDs(spaceIDs []int) []int {
	if len(spaceIDs) == 0 {
		return nil
	}

	var roleIDs []int
	h.db.Model(&models.AlertSpaceRole{}).Where("space_id IN ?", spaceIDs).Pluck("role_id", &roleIDs)
	if len(roleIDs) == 0 {
		return nil
	}

	var userIDs []int
	h.db.Model(&models.UserRole{}).Where("role_id IN ?", roleIDs).Pluck("user_id", &userIDs)
	return userIDs
}
