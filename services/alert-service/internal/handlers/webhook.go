package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/alert-service/internal/services"
	"github.com/itcmdb/shared/pkg/response"
	"gorm.io/gorm"
)

// InboundWebhookHandler 接收Webhook处理器
type InboundWebhookHandler struct {
	db                *gorm.DB
	webhookService    *services.WebhookService
}

// NewInboundWebhookHandler 创建接收Webhook处理器
func NewInboundWebhookHandler(db *gorm.DB, webhookService *services.WebhookService) *InboundWebhookHandler {
	return &InboundWebhookHandler{
		db:             db,
		webhookService: webhookService,
	}
}

// ListInboundWebhooks 获取接收Webhook列表
func (h *InboundWebhookHandler) ListInboundWebhooks(c *gin.Context) {
	var webhooks []models.InboundWebhook
	var total int64

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sourceType := c.Query("source_type")
	enabled := c.Query("enabled")

	query := h.db.Model(&models.InboundWebhook{})

	// 筛选条件
	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	if enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&webhooks)

	c.JSON(http.StatusOK, models.InboundWebhookListResponse{
		Total:    int(total),
		Webhooks: webhooks,
	})
}

// GetInboundWebhook 获取接收Webhook详情
func (h *InboundWebhookHandler) GetInboundWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid webhook ID", err.Error()))
		return
	}

	var webhook models.InboundWebhook
	if err := h.db.First(&webhook, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("Webhook not found", ""))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("Database error", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// CreateInboundWebhook 创建接收Webhook
func (h *InboundWebhookHandler) CreateInboundWebhook(c *gin.Context) {
	var req models.CreateInboundWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	// 生成Webhook URL
	token, err := models.GenerateWebhookToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to generate webhook token", err.Error()))
		return
	}

	// 构建完整的Webhook URL
	// 优先使用X-Forwarded头（来自nginx反向代理）
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}

	webhookURL := scheme + "://" + host + "/api/v1/webhooks/inbound/" + token

	webhook := models.InboundWebhook{
		Name:       req.Name,
		SourceType: req.SourceType,
		WebhookURL: webhookURL,
		Enabled:    true,
		Description: req.Description,
	}

	if err := h.db.Create(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create webhook", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

// UpdateInboundWebhook 更新接收Webhook
func (h *InboundWebhookHandler) UpdateInboundWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid webhook ID", err.Error()))
		return
	}

	var webhook models.InboundWebhook
	if err := h.db.First(&webhook, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("Webhook not found", ""))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("Database error", err.Error()))
		}
		return
	}

	var req models.UpdateInboundWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	// 更新字段
	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.Enabled != nil {
		webhook.Enabled = *req.Enabled
	}
	if req.Description != nil {
		webhook.Description = *req.Description
	}

	if err := h.db.Save(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to update webhook", err.Error()))
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// DeleteInboundWebhook 删除接收Webhook
func (h *InboundWebhookHandler) DeleteInboundWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid webhook ID", err.Error()))
		return
	}

	if err := h.db.Delete(&models.InboundWebhook{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to delete webhook", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Webhook deleted successfully"}))
}

// OutboundWebhookHandler 推送Webhook处理器
type OutboundWebhookHandler struct {
	db             *gorm.DB
	webhookService *services.WebhookService
}

// NewOutboundWebhookHandler 创建推送Webhook处理器
func NewOutboundWebhookHandler(db *gorm.DB, webhookService *services.WebhookService) *OutboundWebhookHandler {
	return &OutboundWebhookHandler{
		db:             db,
		webhookService: webhookService,
	}
}

// ListOutboundWebhooks 获取推送Webhook列表
func (h *OutboundWebhookHandler) ListOutboundWebhooks(c *gin.Context) {
	var webhooks []models.OutboundWebhook
	var total int64

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	targetType := c.Query("target_type")
	enabled := c.Query("enabled")

	query := h.db.Model(&models.OutboundWebhook{})

	// 筛选条件
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}

	// 计算总数
	query.Count(&total)

	// 分页查询（预加载接收人）
	offset := (page - 1) * pageSize
	query.Preload("Receiver").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&webhooks)

	c.JSON(http.StatusOK, models.OutboundWebhookListResponse{
		Total:    int(total),
		Webhooks: webhooks,
	})
}

// GetOutboundWebhook 获取推送Webhook详情
func (h *OutboundWebhookHandler) GetOutboundWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid webhook ID", err.Error()))
		return
	}

	var webhook models.OutboundWebhook
	if err := h.db.Preload("Receiver").First(&webhook, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("Webhook not found", ""))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("Database error", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// CreateOutboundWebhook 创建推送Webhook
func (h *OutboundWebhookHandler) CreateOutboundWebhook(c *gin.Context) {
	var req models.CreateOutboundWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	webhook := models.OutboundWebhook{
		Name:        req.Name,
		TargetType:  req.TargetType,
		ReceiverID:  req.ReceiverID,
		EndpointURL: "",
		Description: req.Description,
		Enabled:     true,
	}

	if req.EndpointURL != nil {
		webhook.EndpointURL = *req.EndpointURL
	}

	if err := h.db.Create(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create webhook", err.Error()))
		return
	}

	// 重新加载包含接收人的数据
	if err := h.db.Preload("Receiver").First(&webhook, webhook.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to reload webhook", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

// UpdateOutboundWebhook 更新推送Webhook
func (h *OutboundWebhookHandler) UpdateOutboundWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid webhook ID", err.Error()))
		return
	}

	var webhook models.OutboundWebhook
	if err := h.db.First(&webhook, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("Webhook not found", ""))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("Database error", err.Error()))
		}
		return
	}

	var req models.UpdateOutboundWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request", err.Error()))
		return
	}

	// 更新字段
	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.ReceiverID != nil {
		webhook.ReceiverID = req.ReceiverID
	}
	if req.EndpointURL != nil {
		webhook.EndpointURL = *req.EndpointURL
	}
	if req.Enabled != nil {
		webhook.Enabled = *req.Enabled
	}
	if req.Description != nil {
		webhook.Description = *req.Description
	}

	if err := h.db.Save(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to update webhook", err.Error()))
		return
	}

	// 重新加载包含接收人的数据
	if err := h.db.Preload("Receiver").First(&webhook, webhook.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to reload webhook", err.Error()))
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// DeleteOutboundWebhook 删除推送Webhook
func (h *OutboundWebhookHandler) DeleteOutboundWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid webhook ID", err.Error()))
		return
	}

	if err := h.db.Delete(&models.OutboundWebhook{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to delete webhook", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Webhook deleted successfully"}))
}

// TestOutboundWebhook 测试推送Webhook
func (h *OutboundWebhookHandler) TestOutboundWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid webhook ID", err.Error()))
		return
	}

	var webhook models.OutboundWebhook
	if err := h.db.First(&webhook, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("Webhook not found", ""))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("Database error", err.Error()))
		}
		return
	}

	// 发送测试消息
	testAlert := map[string]interface{}{
		"alert_id":   "test-" + strconv.FormatInt(time.Now().Unix(), 10),
		"title":      "Webhook测试通知",
		"content":    "这是一条测试告警消息，如果您看到此消息，说明推送配置正确！",
		"severity":   "info",
		"status":     "firing",
		"timestamp":  time.Now(),
	}

	err = h.webhookService.SendOutboundWebhook(&webhook, testAlert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Test failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "Test message sent successfully",
	}))
}

// HandleInboundWebhook 处理接收到的Webhook
func HandleInboundWebhook(db *gorm.DB, webhookService *services.WebhookService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")

		// 根据token查找webhook配置 - 使用精确匹配而不是LIKE查询
		var webhook models.InboundWebhook
		// 构建完整URL进行精确匹配
		scheme := c.GetHeader("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}
		}
		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}
		expectedURL := scheme + "://" + host + "/api/v1/webhooks/inbound/" + token

		if err := db.Where("webhook_url = ?", expectedURL).First(&webhook).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, response.Error("Webhook not found", ""))
			} else {
				c.JSON(http.StatusInternalServerError, response.Error("Database error", err.Error()))
			}
			return
		}

		// 检查是否启用
		if !webhook.Enabled {
			c.JSON(http.StatusForbidden, response.Error("Webhook is disabled", ""))
			return
		}

		// 记录接收日志
		now := time.Now()
		webhook.LastReceived = &now
		if err := db.Save(&webhook).Error; err != nil {
			// 记录错误但不阻止处理
			logInboundError(db, webhook.ID, c, err)
		}

		// 根据类型解析payload
		var alerts []map[string]interface{}
		var err error

		switch webhook.SourceType {
		case "alertmanager":
			alerts, err = webhookService.ParseAlertmanagerPayload(c.Request)
		case "prometheus":
			alerts, err = webhookService.ParsePrometheusPayload(c.Request)
		case "victoriametrics":
			alerts, err = webhookService.ParseVictoriaMetricsPayload(c.Request)
		case "custom":
			alerts, err = webhookService.ParseCustomPayload(c.Request)
		default:
			c.JSON(http.StatusBadRequest, response.Error("Unknown source type", ""))
			return
		}

		if err != nil {
			// 记录错误日志
			logInboundError(db, webhook.ID, c, err)
			c.JSON(http.StatusBadRequest, response.Error("Failed to parse payload", err.Error()))
			return
		}

		// 处理告警
		for _, alert := range alerts {
			if err := webhookService.ProcessInboundAlert(&webhook, alert); err != nil {
				// 记录错误但继续处理其他告警
				logInboundError(db, webhook.ID, c, err)
			}
		}

		// 记录成功日志
		logInboundSuccess(db, webhook.ID, c, len(alerts))

		c.JSON(http.StatusOK, response.Success(gin.H{
			"message":      "Alerts received successfully",
			"alert_count":  len(alerts),
		}))
	}
}

// logInboundSuccess 记录接收成功日志
func logInboundSuccess(db *gorm.DB, webhookID int, c *gin.Context, alertCount int) {
	log := models.InboundWebhookLog{
		WebhookID:    webhookID,
		SourceIP:     c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		StatusCode:   200,
		ProcessedAt:  time.Now(),
	}
	if err := db.Create(&log).Error; err != nil {
		// 日志记录失败不应影响主流程
		// TODO: 使用结构化日志记录错误
	}
}

// logInboundError 记录接收错误日志
func logInboundError(db *gorm.DB, webhookID int, c *gin.Context, err error) {
	log := models.InboundWebhookLog{
		WebhookID:     webhookID,
		SourceIP:      c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		StatusCode:    500,
		ErrorMessage:  err.Error(),
		ProcessedAt:   time.Now(),
	}
	if dbErr := db.Create(&log).Error; dbErr != nil {
		// 日志记录失败不应影响主流程
		// TODO: 使用结构化日志记录错误
	}
}
