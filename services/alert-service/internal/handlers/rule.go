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

// RuleHandler 规则处理器
type RuleHandler struct {
	db *gorm.DB
}

// NewRuleHandler 创建规则处理器
func NewRuleHandler(db *gorm.DB) *RuleHandler {
	return &RuleHandler{db: db}
}

// GetRules 获取告警规则列表
func (h *RuleHandler) GetRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	severity := c.Query("severity")
	enabled := c.Query("enabled")

	// 构建查询
	query := h.db.Model(&models.AlertRule{}).Where("deleted_at IS NULL")

	// 严重程度过滤
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// 启用状态过滤
	if enabled != "" {
		if enabled == "true" {
			query = query.Where("enabled = ?", true)
		} else if enabled == "false" {
			query = query.Where("enabled = ?", false)
		}
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var rules []models.AlertRule
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total": total,
		"rules": rules,
	}))
}

// GetRuleByID 获取单个规则
func (h *RuleHandler) GetRuleByID(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的规则ID", err.Error()))
		return
	}

	var rule models.AlertRule
	if err := h.db.Where("id = ? AND deleted_at IS NULL", ruleID).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("规则不存在", ""))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, response.Success(rule))
}

// CreateRule 创建告警规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var req models.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	// 从上下文获取用户ID（需要JWT中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error("未授权", ""))
		return
	}

	// 创建规则
	rule := models.AlertRule{
		Name:                 req.Name,
		Description:          req.Description,
		MetricQuery:          req.MetricQuery,
		ThresholdOperator:    req.ThresholdOperator,
		ThresholdValue:       req.ThresholdValue,
		Duration:             req.Duration,
		Severity:             req.Severity,
		Enabled:              req.Enabled,
		CITypeID:             req.CITypeID,
		NotificationChannels: req.NotificationChannels,
		CreatedBy:            userID.(*int),
		UpdatedBy:            userID.(*int),
	}

	if req.Duration == 0 {
		rule.Duration = 300 // 默认5分钟
	}

	if err := h.db.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("创建失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(rule))
}

// UpdateRule 更新告警规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的规则ID", err.Error()))
		return
	}

	var req models.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error("未授权", ""))
		return
	}

	// 查询规则
	var rule models.AlertRule
	if err := h.db.Where("id = ? AND deleted_at IS NULL", ruleID).First(&rule).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("规则不存在", err.Error()))
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.MetricQuery != nil {
		updates["metric_query"] = *req.MetricQuery
	}
	if req.ThresholdOperator != nil {
		updates["threshold_operator"] = *req.ThresholdOperator
	}
	if req.ThresholdValue != nil {
		updates["threshold_value"] = *req.ThresholdValue
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.CITypeID != nil {
		updates["ci_type_id"] = *req.CITypeID
	}
	if req.NotificationChannels != nil {
		updates["notification_channels"] = req.NotificationChannels
	}
	if req.SilencedUntil != nil {
		if *req.SilencedUntil == "" {
			updates["silenced_until"] = nil
		} else {
			silencedUntil, err := time.Parse(time.RFC3339, *req.SilencedUntil)
			if err != nil {
				c.JSON(http.StatusBadRequest, response.Error("静默时间格式错误", err.Error()))
				return
			}
			updates["silenced_until"] = silencedUntil
		}
	}
	updates["updated_by"] = userID.(*int)

	// 更新规则
	if err := h.db.Model(&rule).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("更新失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(rule))
}

// DeleteRule 删除告警规则（软删除）
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的规则ID", err.Error()))
		return
	}

	// 软删除
	if err := h.db.Model(&models.AlertRule{}).Where("id = ?", ruleID).Update("deleted_at", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("删除失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// EnableRule 启用规则
func (h *RuleHandler) EnableRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的规则ID", err.Error()))
		return
	}

	if err := h.db.Model(&models.AlertRule{}).Where("id = ?", ruleID).Update("enabled", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("启用失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// DisableRule 禁用规则
func (h *RuleHandler) DisableRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的规则ID", err.Error()))
		return
	}

	if err := h.db.Model(&models.AlertRule{}).Where("id = ?", ruleID).Update("enabled", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("禁用失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// TestRule 测试规则
func (h *RuleHandler) TestRule(c *gin.Context) {
	var req struct {
		MetricQuery       string  `json:"metric_query" binding:"required"`
		ThresholdOperator string  `json:"threshold_operator" binding:"required"`
		ThresholdValue    float64 `json:"threshold_value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	// TODO: 实现规则测试逻辑
	// 1. 查询VictoriaMetrics获取当前值
	// 2. 比较阈值
	// 3. 返回测试结果

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "规则测试功能待实现",
	}))
}
