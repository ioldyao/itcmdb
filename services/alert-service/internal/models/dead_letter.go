package models

import "time"

// DeadLetterQueue 死信队列记录
type DeadLetterQueue struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	WebhookID    int       `json:"webhook_id" gorm:"index"`
	WebhookType  string    `json:"webhook_type" gorm:"type:varchar(20)"` // inbound/outbound
	AlertData    JSONMap   `json:"alert_data" gorm:"type:jsonb"`
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	RetryCount   int       `json:"retry_count" gorm:"default:0"`
	LastRetryAt  *time.Time `json:"last_retry_at"`
	Status       string    `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending/processing/failed/resolved
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WebhookMetrics Webhook指标
type WebhookMetrics struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	WebhookID       int       `json:"webhook_id" gorm:"uniqueIndex"`
	WebhookType     string    `json:"webhook_type" gorm:"type:varchar(20)"` // inbound/outbound
	TotalRequests   int64     `json:"total_requests" gorm:"default:0"`
	SuccessRequests int64     `json:"success_requests" gorm:"default:0"`
	FailedRequests  int64     `json:"failed_requests" gorm:"default:0"`
	AvgResponseTime float64   `json:"avg_response_time" gorm:"default:0"` // 毫秒
	LastRequestAt   *time.Time `json:"last_request_at"`
	CircuitState    string    `json:"circuit_state" gorm:"type:varchar(20);default:'closed'"` // closed/open/half_open
	UpdatedAt       time.Time `json:"updated_at"`
}
