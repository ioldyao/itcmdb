package workflow

import "time"

// Workflow 工作流
type Workflow struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Direction   string    `json:"direction" gorm:"not null;type:varchar(20)"` // inbound | outbound
	Type        string    `json:"type" gorm:"not null;type:varchar(50)"`      // alertmanager | prometheus | victoriametrics | workflow
	Pipeline    string    `json:"pipeline" gorm:"type:text"`                   // JSON Pipeline
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Webhook Webhook 配置
type Webhook struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	WorkflowID  int       `json:"workflow_id" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Direction   string    `json:"direction" gorm:"not null;type:varchar(20)"` // inbound | outbound

	// inbound 字段
	WebhookToken string `json:"webhook_token,omitempty" gorm:"type:varchar(100);uniqueIndex"`

	// outbound 字段
	WebhookURL string `json:"webhook_url,omitempty" gorm:"type:text"`

	Type        string    `json:"type" gorm:"not null;type:varchar(50)"` // alertmanager | prometheus | victoriametrics | workflow
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Workflow *Workflow `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
}

// WorkflowExecution 工作流执行记录
type WorkflowExecution struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	WorkflowID int       `json:"workflow_id" gorm:"not null"`
	WebhookID  *int      `json:"webhook_id,omitempty"`
	PipelineID string    `json:"pipeline_id" gorm:"not null;index"`
	Status     string    `json:"status" gorm:"not null;type:varchar(20)"` // running | completed | failed | paused
	InputData  string    `json:"input_data" gorm:"type:text"`             // JSON
	OutputData string    `json:"output_data" gorm:"type:text"`            // JSON
	ErrorMsg   string    `json:"error_msg" gorm:"type:text"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`

	// 关联
	Workflow *Workflow `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
	Webhook  *Webhook  `json:"webhook,omitempty" gorm:"foreignKey:WebhookID"`
}

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Direction   string          `json:"direction" validate:"required,oneof=inbound outbound"`
	Type        string          `json:"type" validate:"required"`
	Pipeline    string          `json:"pipeline" validate:"required"`
	Enabled     bool            `json:"enabled"`
}

// UpdateWorkflowRequest 更新工作流请求
type UpdateWorkflowRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1"`
	Description *string `json:"description"`
	Direction   *string `json:"direction" validate:"omitempty,oneof=inbound outbound"`
	Type        *string `json:"type" validate:"omitempty"`
	Pipeline    *string `json:"pipeline" validate:"omitempty"`
	Enabled     *bool   `json:"enabled"`
}

// CreateWebhookRequest 创建 Webhook 请求
type CreateWebhookRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Direction   string `json:"direction" validate:"required,oneof=inbound outbound"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	Type        string `json:"type" validate:"required"`
	WorkflowID  int    `json:"workflow_id" validate:"required"`
	Enabled     bool   `json:"enabled"`
}

// UpdateWebhookRequest 更新 Webhook 请求
type UpdateWebhookRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1"`
	Description *string `json:"description"`
	WebhookURL  *string `json:"webhook_url"`
	Enabled     *bool   `json:"enabled"`
}

// ExecuteWorkflowRequest 执行工作流请求
type ExecuteWorkflowRequest struct {
	Pipeline map[string]interface{} `json:"pipeline" validate:"required"`
	Data     map[string]interface{} `json:"data"`
}

// ReceiveWebhookRequest 接收 Webhook 请求
type ReceiveWebhookRequest map[string]interface{}

// WorkflowListResponse 工作流列表响应
type WorkflowListResponse struct {
	Total     int        `json:"total"`
	Workflows []Workflow `json:"workflows"`
}

// WebhookListResponse Webhook 列表响应
type WebhookListResponse struct {
	Total    int      `json:"total"`
	Webhooks []Webhook `json:"webhooks"`
}

// ExecutionListResponse 执行记录列表响应
type ExecutionListResponse struct {
	Total       int                `json:"total"`
	Executions  []WorkflowExecution `json:"executions"`
}
