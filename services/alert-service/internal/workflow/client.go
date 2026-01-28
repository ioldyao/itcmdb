package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client Workflow Engine 客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExecuteWorkflow 执行工作流
func (c *Client) ExecuteWorkflow(pipeline map[string]interface{}, data map[string]interface{}) (*ExecuteWorkflowResponse, error) {
	req := ExecuteWorkflowRequest{
		Pipeline: pipeline,
		Data:     data,
	}

	var resp ExecuteWorkflowResponse
	err := c.doRequest("POST", "/api/v1/workflow/execute", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ReceiveWebhookAlert 接收 Webhook 告警
func (c *Client) ReceiveWebhookAlert(token string, alertData map[string]interface{}) (*ExecuteWorkflowResponse, error) {
	var resp ExecuteWorkflowResponse
	err := c.doRequest("POST", fmt.Sprintf("/api/v1/webhooks/%s", token), alertData, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetWorkflowStatus 获取工作流状态
func (c *Client) GetWorkflowStatus(pipelineID string) (*WorkflowStatusResponse, error) {
	var resp WorkflowStatusResponse
	err := c.doRequest("GET", fmt.Sprintf("/api/v1/workflow/%s/status", pipelineID), nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// PauseWorkflow 暂停工作流
func (c *Client) PauseWorkflow(pipelineID string) (*SimpleResponse, error) {
	var resp SimpleResponse
	err := c.doRequest("POST", fmt.Sprintf("/api/v1/workflow/%s/pause", pipelineID), nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ResumeWorkflow 恢复工作流
func (c *Client) ResumeWorkflow(pipelineID string) (*SimpleResponse, error) {
	var resp SimpleResponse
	err := c.doRequest("POST", fmt.Sprintf("/api/v1/workflow/%s/resume", pipelineID), nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateWorkflow 创建工作流
func (c *Client) CreateWorkflow(req *CreateWorkflowRequest) (*CreateWorkflowResponse, error) {
	var resp CreateWorkflowResponse
	err := c.doRequest("POST", "/api/v1/workflow", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetWebhook 获取 Webhook 详情
func (c *Client) GetWebhook(webhookID string) (*WebhookDetailResponse, error) {
	var resp WebhookDetailResponse
	err := c.doRequest("GET", fmt.Sprintf("/api/v1/webhooks/%s", webhookID), nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// ==================== 响应类型 ====================

// ExecuteWorkflowResponse 执行工作流响应
type ExecuteWorkflowResponse struct {
	Success   bool   `json:"success"`
	PipelineID string `json:"pipeline_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// WorkflowStatusResponse 工作流状态响应
type WorkflowStatusResponse struct {
	Success bool                   `json:"success"`
	States  map[string]interface{} `json:"states,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// SimpleResponse 简单响应
type SimpleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// CreateWorkflowResponse 创建工作流响应
type CreateWorkflowResponse struct {
	Success     bool   `json:"success"`
	WorkflowID  string `json:"workflow_id,omitempty"`
	WebhookID   string `json:"webhook_id,omitempty"`
	WebhookToken string `json:"webhook_token,omitempty"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	Message     string `json:"message,omitempty"`
}

// WebhookDetailResponse Webhook 详情响应
type WebhookDetailResponse struct {
	Success  bool     `json:"success"`
	Webhook  *Webhook `json:"webhook,omitempty"`
	Workflow *Workflow `json:"workflow,omitempty"`
}
