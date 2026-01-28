package workflow

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Service 工作流服务
type Service struct {
	db     *gorm.DB
	client *Client
}

// NewService 创建服务
func NewService(db *gorm.DB, workflowEngineURL string) *Service {
	return &Service{
		db:     db,
		client: NewClient(workflowEngineURL),
	}
}

// ==================== 工作流管理 ====================

// CreateWorkflow 创建工作流
func (s *Service) CreateWorkflow(req *CreateWorkflowRequest) (*Workflow, error) {
	workflow := &Workflow{
		Name:        req.Name,
		Description: req.Description,
		Direction:   req.Direction,
		Type:        req.Type,
		Pipeline:    req.Pipeline,
		Enabled:     req.Enabled,
	}

	if err := s.db.Create(workflow).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow, nil
}

// GetWorkflow 获取工作流
func (s *Service) GetWorkflow(id int) (*Workflow, error) {
	var workflow Workflow
	if err := s.db.First(&workflow, id).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

// ListWorkflows 获取工作流列表
func (s *Service) ListWorkflows(offset, limit int) (*WorkflowListResponse, error) {
	var workflows []Workflow
	var total int64

	if err := s.db.Model(&Workflow{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := s.db.Offset(offset).Limit(limit).Find(&workflows).Error; err != nil {
		return nil, err
	}

	return &WorkflowListResponse{
		Total:     int(total),
		Workflows: workflows,
	}, nil
}

// UpdateWorkflow 更新工作流
func (s *Service) UpdateWorkflow(id int, req *UpdateWorkflowRequest) (*Workflow, error) {
	var workflow Workflow
	if err := s.db.First(&workflow, id).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Direction != nil {
		updates["direction"] = *req.Direction
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Pipeline != nil {
		updates["pipeline"] = *req.Pipeline
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := s.db.Model(&workflow).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	return &workflow, nil
}

// DeleteWorkflow 删除工作流
func (s *Service) DeleteWorkflow(id int) error {
	return s.db.Delete(&Workflow{}, id).Error
}

// ==================== Webhook 管理 ====================

// CreateWebhook 创建 Webhook
func (s *Service) CreateWebhook(req *CreateWebhookRequest) (*Webhook, error) {
	webhook := &Webhook{
		WorkflowID:  req.WorkflowID,
		Name:        req.Name,
		Description: req.Description,
		Direction:   req.Direction,
		WebhookURL:  req.WebhookURL,
		Type:        req.Type,
		Enabled:     req.Enabled,
	}

	// 如果是 inbound，生成 token
	if req.Direction == "inbound" {
		token, err := generateWebhookToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}
		webhook.WebhookToken = token
	}

	if err := s.db.Create(webhook).Error; err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return webhook, nil
}

// GetWebhook 获取 Webhook
func (s *Service) GetWebhook(id int) (*Webhook, error) {
	var webhook Webhook
	if err := s.db.Preload("Workflow").First(&webhook, id).Error; err != nil {
		return nil, err
	}
	return &webhook, nil
}

// GetWebhookByToken 根据 token 获取 Webhook
func (s *Service) GetWebhookByToken(token string) (*Webhook, error) {
	var webhook Webhook
	if err := s.db.Preload("Workflow").Where("webhook_token = ? AND enabled = ?", token, true).First(&webhook).Error; err != nil {
		return nil, err
	}
	return &webhook, nil
}

// ListWebhooks 获取 Webhook 列表
func (s *Service) ListWebhooks(offset, limit int) (*WebhookListResponse, error) {
	var webhooks []Webhook
	var total int64

	if err := s.db.Model(&Webhook{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("Workflow").Offset(offset).Limit(limit).Find(&webhooks).Error; err != nil {
		return nil, err
	}

	return &WebhookListResponse{
		Total:    int(total),
		Webhooks: webhooks,
	}, nil
}

// UpdateWebhook 更新 Webhook
func (s *Service) UpdateWebhook(id int, req *UpdateWebhookRequest) (*Webhook, error) {
	var webhook Webhook
	if err := s.db.First(&webhook, id).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.WebhookURL != nil {
		updates["webhook_url"] = *req.WebhookURL
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := s.db.Model(&webhook).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	return &webhook, nil
}

// DeleteWebhook 删除 Webhook
func (s *Service) DeleteWebhook(id int) error {
	return s.db.Delete(&Webhook{}, id).Error
}

// GetWebhookURL 获取 Webhook URL
func (s *Service) GetWebhookURL(webhook *Webhook, baseURL string) string {
	if webhook.Direction == "inbound" {
		return fmt.Sprintf("%s/api/v1/webhooks/%s", baseURL, webhook.WebhookToken)
	}
	return webhook.WebhookURL
}

// ==================== 工作流执行 ====================

// ExecuteWorkflow 执行工作流
func (s *Service) ExecuteWorkflow(workflow *Workflow, data map[string]interface{}) (*WorkflowExecution, error) {
	// 解析 Pipeline
	var pipeline map[string]interface{}
	if err := json.Unmarshal([]byte(workflow.Pipeline), &pipeline); err != nil {
		return nil, fmt.Errorf("failed to parse pipeline: %w", err)
	}

	// 调用 Python Engine
	resp, err := s.client.ExecuteWorkflow(pipeline, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute workflow: %w", err)
	}

	// 保存执行记录
	execution := &WorkflowExecution{
		WorkflowID: workflow.ID,
		PipelineID: resp.PipelineID,
		Status:     "running",
		StartedAt:  time.Now(),
	}

	inputJSON, _ := json.Marshal(data)
	execution.InputData = string(inputJSON)

	if err := s.db.Create(execution).Error; err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	return execution, nil
}

// ReceiveWebhookAlert 接收 Webhook 告警
func (s *Service) ReceiveWebhookAlert(token string, alertData map[string]interface{}) (*WorkflowExecution, error) {
	// 根据 token 获取 Webhook
	webhook, err := s.GetWebhookByToken(token)
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}

	if !webhook.Enabled {
		return nil, fmt.Errorf("webhook is disabled")
	}

	// 获取工作流
	workflow := webhook.Workflow
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found")
	}

	if !workflow.Enabled {
		return nil, fmt.Errorf("workflow is disabled")
	}

	// 执行工作流
	execution, err := s.ExecuteWorkflow(workflow, map[string]interface{}{
		"alert": alertData,
	})
	if err != nil {
		return nil, err
	}

	// 关联 Webhook
	execution.WebhookID = &webhook.ID
	s.db.Save(execution)

	return execution, nil
}

// GetExecutionStatus 获取执行状态
func (s *Service) GetExecutionStatus(execution *WorkflowExecution) (*WorkflowStatusResponse, error) {
	// 调用 Python Engine 获取状态
	resp, err := s.client.GetWorkflowStatus(execution.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// 更新本地状态
	// TODO: 根据 resp.States 更新 execution.Status

	return resp, nil
}

// ListExecutions 获取执行记录列表
func (s *Service) ListExecutions(workflowID int, offset, limit int) (*ExecutionListResponse, error) {
	var executions []WorkflowExecution
	var total int64

	query := s.db.Model(&WorkflowExecution{})
	if workflowID > 0 {
		query = query.Where("workflow_id = ?", workflowID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Preload("Workflow").Preload("Webhook").Offset(offset).Limit(limit).Order("started_at DESC").Find(&executions).Error; err != nil {
		return nil, err
	}

	return &ExecutionListResponse{
		Total:      int(total),
		Executions: executions,
	}, nil
}

// ==================== 辅助函数 ====================

// generateWebhookToken 生成 Webhook Token
func generateWebhookToken() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "wh_" + hex.EncodeToString(b), nil
}
