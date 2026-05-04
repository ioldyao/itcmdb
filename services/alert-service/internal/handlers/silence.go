package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/shared/pkg/response"
	"gorm.io/gorm"
)

// SilenceHandler 告警静默处理器
type SilenceHandler struct {
	db *gorm.DB
}

// NewSilenceHandler 创建静默处理器
func NewSilenceHandler(db *gorm.DB) *SilenceHandler {
	return &SilenceHandler{db: db}
}

// ListSilences 获取静默规则列表
func (h *SilenceHandler) ListSilences(c *gin.Context) {
	var silences []models.AlertSilence
	query := h.db.Order("created_at DESC")

	// 状态筛选
	if active := c.Query("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}

	if err := query.Find(&silences).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(silences))
}

// GetSilence 获取单个静默规则
func (h *SilenceHandler) GetSilence(c *gin.Context) {
	id := c.Param("id")
	silenceID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", err.Error()))
		return
	}

	var silence models.AlertSilence
	if err := h.db.Where("id = ?", silenceID).First(&silence).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("静默规则不存在", ""))
		return
	}

	c.JSON(http.StatusOK, response.Success(silence))
}

// CreateSilence 创建静默规则
func (h *SilenceHandler) CreateSilence(c *gin.Context) {
	var req struct {
		Name     string         `json:"name" binding:"required"`
		Comment  string         `json:"comment"`
		Matchers models.JSONMap `json:"matchers" binding:"required"`
		StartsAt string         `json:"starts_at" binding:"required"`
		EndsAt   string         `json:"ends_at" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("开始时间格式错误", err.Error()))
		return
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("结束时间格式错误", err.Error()))
		return
	}

	silence := models.AlertSilence{
		Name:      req.Name,
		Comment:   req.Comment,
		Matchers:  req.Matchers,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Active:    true,
		CreatedBy: nil, // TODO: 从上下文获取用户ID
	}

	if err := h.db.Create(&silence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("创建失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(silence))
}

// UpdateSilence 更新静默规则
func (h *SilenceHandler) UpdateSilence(c *gin.Context) {
	id := c.Param("id")
	silenceID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", err.Error()))
		return
	}

	var silence models.AlertSilence
	if err := h.db.Where("id = ?", silenceID).First(&silence).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("静默规则不存在", ""))
		return
	}

	var req struct {
		Name     *string         `json:"name"`
		Comment  *string         `json:"comment"`
		Matchers *models.JSONMap `json:"matchers"`
		StartsAt *string         `json:"starts_at"`
		EndsAt   *string         `json:"ends_at"`
		Active   *bool           `json:"active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	if req.Name != nil {
		silence.Name = *req.Name
	}
	if req.Comment != nil {
		silence.Comment = *req.Comment
	}
	if req.Matchers != nil {
		silence.Matchers = *req.Matchers
	}
	if req.StartsAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.StartsAt); err == nil {
			silence.StartsAt = t
		}
	}
	if req.EndsAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.EndsAt); err == nil {
			silence.EndsAt = t
		}
	}
	if req.Active != nil {
		silence.Active = *req.Active
	}

	if err := h.db.Save(&silence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("更新失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(silence))
}

// DeleteSilence 删除静默规则
func (h *SilenceHandler) DeleteSilence(c *gin.Context) {
	id := c.Param("id")
	silenceID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的ID", err.Error()))
		return
	}

	result := h.db.Delete(&models.AlertSilence{}, silenceID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("删除失败", result.Error.Error()))
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.Error("静默规则不存在", ""))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
