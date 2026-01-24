package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itcmdb/cmdb-service/internal/service"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/response"
	"net/http"
	"strconv"
)

type TagHandler struct {
	tagService service.TagService
	jwtManager *auth.JWTManager
}

func NewTagHandler(tagService service.TagService, jwtManager *auth.JWTManager) *TagHandler {
	return &TagHandler{
		tagService: tagService,
		jwtManager: jwtManager,
	}
}

// ==================== 标签分类管理 ====================

// GetTagCategories 获取标签分类列表
func (h *TagHandler) GetTagCategories(c *gin.Context) {
	categories, err := h.tagService.GetTagCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("获取标签分类失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(categories))
}

// CreateTagCategory 创建标签分类
func (h *TagHandler) CreateTagCategory(c *gin.Context) {
	var req service.CreateTagCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	category, err := h.tagService.CreateTagCategory(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("创建标签分类失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(category))
}

// UpdateTagCategory 更新标签分类
func (h *TagHandler) UpdateTagCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var req service.UpdateTagCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	if err := h.tagService.UpdateTagCategory(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("更新标签分类失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// DeleteTagCategory 删除标签分类
func (h *TagHandler) DeleteTagCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	if err := h.tagService.DeleteTagCategory(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("删除标签分类失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// ==================== 标签管理 ====================

// GetTags 获取标签列表
func (h *TagHandler) GetTags(c *gin.Context) {
	var categoryID *uint

	if categoryStr := c.Query("category_id"); categoryStr != "" {
		id, err := strconv.ParseUint(categoryStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error("无效的分类ID", ""))
			return
		}
		cid := uint(id)
		categoryID = &cid
	}

	tags, err := h.tagService.GetTags(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("获取标签列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(tags))
}

// CreateTag 创建标签
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req service.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	tag, err := h.tagService.CreateTag(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("创建标签失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(tag))
}

// UpdateTag 更新标签
func (h *TagHandler) UpdateTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var req service.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	if err := h.tagService.UpdateTag(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("更新标签失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// DeleteTag 删除标签
func (h *TagHandler) DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	if err := h.tagService.DeleteTag(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("删除标签失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// GetTagStats 获取标签统计
func (h *TagHandler) GetTagStats(c *gin.Context) {
	stats, err := h.tagService.GetTagStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("获取标签统计失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(stats))
}

// ==================== CI实例标签操作 ====================

// GetCITags 获取CI实例的标签列表
func (h *TagHandler) GetCITags(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	tags, err := h.tagService.GetCITags(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("获取标签失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(tags))
}

// AssignTag 为CI实例添加标签
func (h *TagHandler) AssignTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	var req struct {
		TagID uint `json:"tag_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	userID, _ := auth.GetUserID(c)

	if err := h.tagService.AssignTag(uint(id), req.TagID, uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("添加标签失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// RemoveTag 移除CI实例的标签
func (h *TagHandler) RemoveTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", ""))
		return
	}

	tagStr := c.Param("tagId")
	tagID, err := strconv.ParseUint(tagStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的标签ID", ""))
		return
	}

	if err := h.tagService.RemoveTag(uint(id), uint(tagID)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("移除标签失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// ==================== 批量操作 ====================

// BatchAssignTags 批量为CI实例添加标签
func (h *TagHandler) BatchAssignTags(c *gin.Context) {
	var req struct {
		CIIDs []uint `json:"ci_ids" binding:"required"`
		TagID uint   `json:"tag_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	userID, _ := auth.GetUserID(c)

	if err := h.tagService.BatchAssignTags(req.CIIDs, req.TagID, uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("批量添加标签失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// BatchRemoveTags 批量移除CI实例的标签
func (h *TagHandler) BatchRemoveTags(c *gin.Context) {
	var req struct {
		CIIDs []uint `json:"ci_ids" binding:"required"`
		TagID uint   `json:"tag_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请求参数错误", err.Error()))
		return
	}

	if err := h.tagService.BatchRemoveTags(req.CIIDs, req.TagID); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("批量移除标签失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
