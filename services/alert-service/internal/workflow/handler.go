package workflow

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler 工作流处理器
type Handler struct {
	service *Service
	baseURL string // 用于生成 inbound Webhook URL
}

// NewHandler 创建处理器
func NewHandler(db *gorm.DB, workflowEngineURL, baseURL string) *Handler {
	return &Handler{
		service: NewService(db, workflowEngineURL),
		baseURL: baseURL,
	}
}

// ==================== 工作流管理 ====================

// CreateWorkflow 创建工作流
func (h *Handler) CreateWorkflow(c *gin.Context) {
	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.service.CreateWorkflow(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

// GetWorkflow 获取工作流
func (h *Handler) GetWorkflow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	workflow, err := h.service.GetWorkflow(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// ListWorkflows 获取工作流列表
func (h *Handler) ListWorkflows(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	offset := (page - 1) * pageSize

	resp, err := h.service.ListWorkflows(offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateWorkflow 更新工作流
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	var req UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.service.UpdateWorkflow(id, &req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// DeleteWorkflow 删除工作流
func (h *Handler) DeleteWorkflow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	if err := h.service.DeleteWorkflow(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow deleted successfully"})
}

// ==================== Webhook 管理 ====================

// CreateWebhook 创建 Webhook
func (h *Handler) CreateWebhook(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.service.CreateWebhook(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 预加载 Workflow
	h.service.db.Preload("Workflow").First(webhook, webhook.ID)

	// 如果是 inbound，生成 Webhook URL
	if webhook.Direction == "inbound" {
		webhook.WebhookURL = h.service.GetWebhookURL(webhook, h.baseURL)
	}

	c.JSON(http.StatusCreated, webhook)
}

// GetWebhook 获取 Webhook
func (h *Handler) GetWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	webhook, err := h.service.GetWebhook(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 如果是 inbound，生成 Webhook URL
	if webhook.Direction == "inbound" {
		webhook.WebhookURL = h.service.GetWebhookURL(webhook, h.baseURL)
	}

	c.JSON(http.StatusOK, webhook)
}

// ListWebhooks 获取 Webhook 列表
func (h *Handler) ListWebhooks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	offset := (page - 1) * pageSize

	resp, err := h.service.ListWebhooks(offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 为 inbound webhook 生成 URL
	for i := range resp.Webhooks {
		if resp.Webhooks[i].Direction == "inbound" {
			resp.Webhooks[i].WebhookURL = h.service.GetWebhookURL(&resp.Webhooks[i], h.baseURL)
		}
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateWebhook 更新 Webhook
func (h *Handler) UpdateWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.service.UpdateWebhook(id, &req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// DeleteWebhook 删除 Webhook
func (h *Handler) DeleteWebhook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	if err := h.service.DeleteWebhook(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted successfully"})
}

// ==================== Webhook 接收 ====================

// ReceiveWebhookAlert 接收 Webhook 告警
func (h *Handler) ReceiveWebhookAlert(c *gin.Context) {
	token := c.Param("token")

	var alertData map[string]interface{}
	if err := c.ShouldBindJSON(&alertData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	execution, err := h.service.ReceiveWebhookAlert(token, alertData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Alert received and workflow started",
		"execution_id": execution.ID,
		"pipeline_id": execution.PipelineID,
	})
}

// ==================== 执行管理 ====================

// ExecuteWorkflow 执行工作流
func (h *Handler) ExecuteWorkflow(c *gin.Context) {
	var req ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 临时工作流执行（不从数据库加载）
	// TODO: 支持从数据库加载工作流并执行
	pipelineJSON, _ := json.Marshal(req.Pipeline)
	workflow := &Workflow{
		ID:       0,
		Pipeline: string(pipelineJSON),
		Enabled:  true,
	}

	execution, err := h.service.ExecuteWorkflow(workflow, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"execution_id": execution.ID,
		"pipeline_id": execution.PipelineID,
	})
}

// GetExecutionStatus 获取执行状态
func (h *Handler) GetExecutionStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	var execution WorkflowExecution
	if err := h.service.db.Preload("Workflow").First(&execution, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}

	status, err := h.service.GetExecutionStatus(&execution)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ListExecutions 获取执行记录列表
func (h *Handler) ListExecutions(c *gin.Context) {
	workflowID, _ := strconv.Atoi(c.Query("workflow_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	offset := (page - 1) * pageSize

	resp, err := h.service.ListExecutions(workflowID, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// PauseWorkflow 暂停工作流
func (h *Handler) PauseWorkflow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	var execution WorkflowExecution
	if err := h.service.db.First(&execution, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}

	// 调用 Python Engine
	if _, err := h.service.client.PauseWorkflow(execution.PipelineID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新状态
	execution.Status = "paused"
	h.service.db.Save(&execution)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Workflow paused"})
}

// ResumeWorkflow 恢复工作流
func (h *Handler) ResumeWorkflow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	var execution WorkflowExecution
	if err := h.service.db.First(&execution, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}

	// 调用 Python Engine
	if _, err := h.service.client.ResumeWorkflow(execution.PipelineID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新状态
	execution.Status = "running"
	h.service.db.Save(&execution)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Workflow resumed"})
}
