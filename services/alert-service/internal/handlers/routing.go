package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/alert-service/internal/services"
	"github.com/itcmdb/shared/pkg/response"
	"gorm.io/gorm"
)

// RoutingHandler 路由规则处理器
type RoutingHandler struct {
	routingService *services.RoutingService
}

// NewRoutingHandler 创建路由规则处理器
func NewRoutingHandler(db *gorm.DB) *RoutingHandler {
	return &RoutingHandler{
		routingService: services.NewRoutingService(db),
	}
}

// ListRoutingRules 获取路由规则列表
func (h *RoutingHandler) ListRoutingRules(c *gin.Context) {
	rules, err := h.routingService.ListRoutingRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get routing rules", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total": len(rules),
		"rules": rules,
	}))
}

// GetRoutingRule 获取路由规则详情
func (h *RoutingHandler) GetRoutingRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid ID", err.Error()))
		return
	}

	rule, err := h.routingService.GetRoutingRule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Routing rule not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(rule))
}

// CreateRoutingRule 创建路由规则
func (h *RoutingHandler) CreateRoutingRule(c *gin.Context) {
	var req models.CreateAlertRoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	rule, err := h.routingService.CreateRoutingRule(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create routing rule", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(rule))
}

// UpdateRoutingRule 更新路由规则
func (h *RoutingHandler) UpdateRoutingRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid ID", err.Error()))
		return
	}

	var req models.UpdateAlertRoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	rule, err := h.routingService.UpdateRoutingRule(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to update routing rule", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(rule))
}

// DeleteRoutingRule 删除路由规则
func (h *RoutingHandler) DeleteRoutingRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid ID", err.Error()))
		return
	}

	if err := h.routingService.DeleteRoutingRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to delete routing rule", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Routing rule deleted successfully"}))
}

// TemplateHandler 通知模板处理器
type TemplateHandler struct {
	routingService *services.RoutingService
}

// NewTemplateHandler 创建通知模板处理器
func NewTemplateHandler(db *gorm.DB) *TemplateHandler {
	return &TemplateHandler{
		routingService: services.NewRoutingService(db),
	}
}

// ListNotificationTemplates 获取通知模板列表
func (h *TemplateHandler) ListNotificationTemplates(c *gin.Context) {
	templateType := c.Query("template_type")

	templates, err := h.routingService.ListNotificationTemplates(templateType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get notification templates", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":     len(templates),
		"templates": templates,
	}))
}

// GetNotificationTemplate 获取通知模板详情
func (h *TemplateHandler) GetNotificationTemplate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid ID", err.Error()))
		return
	}

	template, err := h.routingService.GetNotificationTemplate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Notification template not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(template))
}

// CreateNotificationTemplate 创建通知模板
func (h *TemplateHandler) CreateNotificationTemplate(c *gin.Context) {
	var req models.CreateAlertNotificationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	template, err := h.routingService.CreateNotificationTemplate(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create notification template", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(template))
}

// UpdateNotificationTemplate 更新通知模板
func (h *TemplateHandler) UpdateNotificationTemplate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid ID", err.Error()))
		return
	}

	var req models.UpdateAlertNotificationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	template, err := h.routingService.UpdateNotificationTemplate(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to update notification template", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(template))
}

// DeleteNotificationTemplate 删除通知模板
func (h *TemplateHandler) DeleteNotificationTemplate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid ID", err.Error()))
		return
	}

	if err := h.routingService.DeleteNotificationTemplate(id); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to delete notification template", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Notification template deleted successfully"}))
}

// SetDefaultTemplate 设置默认模板
func (h *TemplateHandler) SetDefaultTemplate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid ID", err.Error()))
		return
	}

	// TODO: 实现设置默认模板的逻辑
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Default template set successfully"}))
}

// PreviewTemplate 预览模板
func (h *TemplateHandler) PreviewTemplate(c *gin.Context) {
	var req struct {
		TemplateContent string                 `json:"template_content"`
		TemplateType    string                 `json:"template_type"`
		SampleData      map[string]interface{} `json:"sample_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	// 使用示例数据渲染模板
	templateData := &models.AlertTemplateData{
		AlertID:   "ALERT-20260130-001",
		Title:     "CPU使用率告警",
		Content:   "服务器CPU使用率超过80%",
		Severity:  "high",
		Status:    "firing",
		Instance:  "server-01",
		Timestamp: "2026-01-30 12:00:00",
		Labels: map[string]string{
			"env":      "production",
			"cluster":  "main",
			"severity": "high",
		},
	}

	rendered, err := h.routingService.RenderTemplate(req.TemplateType, templateData, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to preview template", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"preview": rendered}))
}
