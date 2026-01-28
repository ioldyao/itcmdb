package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// InboundWebhook 接收外部告警的Webhook配置
type InboundWebhook struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	Name         string     `json:"name" gorm:"notNull" validate:"required"`
	WebhookURL   string     `json:"webhook_url" gorm:"uniqueIndex;notNull"`
	SourceType   string     `json:"source_type" gorm:"notNull;type:varchar(50)" validate:"required,oneof=alertmanager prometheus victoriametrics custom"`
	Enabled      bool       `json:"enabled" gorm:"default:true"`
	Description  string     `json:"description"`
	LastReceived *time.Time `json:"last_received"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// OutboundWebhook ITCMDB推送到外部系统的Webhook配置
type OutboundWebhook struct {
	ID          int        `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"notNull" validate:"required"`
	TargetType  string     `json:"target_type" gorm:"notNull;type:varchar(50)" validate:"required,oneof=alertmanager receiver"`
	// 对于 receiver 类型，使用 receiver_id；对于 alertmanager 类型，使用 endpoint_url
	ReceiverID  *int       `json:"receiver_id,omitempty"`
	EndpointURL string     `json:"endpoint_url,omitempty" gorm:"type:text"`
	Enabled     bool       `json:"enabled" gorm:"default:true"`
	Description string     `json:"description"`
	LastSent    *time.Time `json:"last_sent"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 关联的接收人（仅当 targetType=receiver 时有效）
	Receiver    *AlertReceiver `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
}

// InboundWebhookLog Webhook接收日志
type InboundWebhookLog struct {
	ID         int                `json:"id" gorm:"primaryKey"`
	WebhookID  int                `json:"webhook_id" gorm:"index"`
	SourceIP   string             `json:"source_ip"`
	UserAgent  string             `json:"user_agent"`
	StatusCode int                `json:"status_code"`
	RequestData JSONMap           `json:"request_data" gorm:"type:jsonb"`
	ResponseData string           `json:"response_data" gorm:"type:text"`
	ErrorMessage string           `json:"error_message"`
	ProcessedAt  time.Time        `json:"processed_at"`
	CreatedAt    time.Time        `json:"created_at"`
}

// OutboundWebhookLog Webhook推送日志
type OutboundWebhookLog struct {
	ID           int                `json:"id" gorm:"primaryKey"`
	WebhookID    int                `json:"webhook_id" gorm:"index"`
	AlertID      string             `json:"alert_id" gorm:"index"`
	TargetURL    string             `json:"target_url"`
	StatusCode   int                `json:"status_code"`
	RequestData  JSONMap            `json:"request_data" gorm:"type:jsonb"`
	ResponseData string             `json:"response_data" gorm:"type:text"`
	ErrorMessage string            `json:"error_message"`
	RetryCount   int                `json:"retry_count" gorm:"default:0"`
	SentAt       time.Time          `json:"sent_at"`
	CreatedAt    time.Time          `json:"created_at"`
}

// GenerateWebhookToken 生成唯一的Webhook token
func GenerateWebhookToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// 请求和响应模型

// CreateInboundWebhookRequest 创建接收Webhook请求
type CreateInboundWebhookRequest struct {
	Name        string `json:"name" validate:"required"`
	SourceType  string `json:"source_type" validate:"required,oneof=alertmanager prometheus victoriametrics custom"`
	Description string `json:"description"`
}

// UpdateInboundWebhookRequest 更新接收Webhook请求
type UpdateInboundWebhookRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1"`
	Enabled     *bool   `json:"enabled"`
	Description *string `json:"description"`
}

// CreateOutboundWebhookRequest 创建推送Webhook请求
type CreateOutboundWebhookRequest struct {
	Name        string `json:"name" validate:"required"`
	TargetType  string `json:"target_type" validate:"required,oneof=alertmanager receiver"`
	ReceiverID  *int    `json:"receiver_id,omitempty"`  // 当target_type=receiver时必填
	EndpointURL *string `json:"endpoint_url,omitempty"` // 当target_type=alertmanager时必填
	Description string `json:"description"`
}

// UpdateOutboundWebhookRequest 更新推送Webhook请求
type UpdateOutboundWebhookRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1"`
	ReceiverID  *int    `json:"receiver_id,omitempty"`
	EndpointURL *string `json:"endpoint_url,omitempty"`
	Enabled     *bool   `json:"enabled"`
	Description *string `json:"description"`
}

// InboundWebhookListResponse 接收Webhook列表响应
type InboundWebhookListResponse struct {
	Total    int               `json:"total"`
	Webhooks []InboundWebhook  `json:"webhooks"`
}

// OutboundWebhookListResponse 推送Webhook列表响应
type OutboundWebhookListResponse struct {
	Total    int                `json:"total"`
	Webhooks []OutboundWebhook  `json:"webhooks"`
}

// AlertmanagerWebhookPayload Alertmanager Webhook标准格式
type AlertmanagerWebhookPayload struct {
	Version           string                         `json:"version"`
	GroupKey          string                         `json:"groupKey"`
	TruncatedAlerts   int                            `json:"truncatedAlerts,omitempty"`
	Status            string                         `json:"status"`
	Receiver          string                         `json:"receiver"`
	GroupLabels       map[string]string              `json:"groupLabels"`
	CommonLabels      map[string]string              `json:"commonLabels"`
	CommonAnnotations map[string]string              `json:"commonAnnotations"`
	ExternalURL       string                         `json:"externalURL"`
	Alerts            []AlertmanagerAlert            `json:"alerts"`
}

type AlertmanagerAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// PrometheusWebhookPayload Prometheus Webhook格式
type PrometheusWebhookPayload struct {
	Version string                     `json:"version"`
	Alerts  []PrometheusAlert          `json:"alerts"`
}

type PrometheusAlert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

// VictoriaMetricsWebhookPayload VictoriaMetrics Webhook格式
type VictoriaMetricsWebhookPayload struct {
	ReceiverName string                     `json:"receiverName"`
	Status       string                     `json:"status"`
	Alerts       []VictoriaMetricsAlert      `json:"alerts"`
	GroupLabels  map[string]string          `json:"groupLabels"`
}

type VictoriaMetricsAlert struct {
	ID       string            `json:"id"`
	Status   string            `json:"status"`
	Labels   map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt time.Time         `json:"startsAt"`
	EndsAt   time.Time         `json:"endsAt"`
}

// CustomWebhookPayload 自定义Webhook格式
type CustomWebhookPayload struct {
	AlertID    string                 `json:"alert_id,omitempty"`
	Title      string                 `json:"title,omitempty"`
	Content    string                 `json:"content,omitempty"`
	Severity   string                 `json:"severity,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  *time.Time             `json:"timestamp,omitempty"`
}
