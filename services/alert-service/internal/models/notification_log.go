package models

import (
	"time"
)

// NotificationLog 通知日志 - 统一的通知记录表
type NotificationLog struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	AlertInstanceID  int       `json:"alert_instance_id" gorm:"notNull;index"`
	ReceiverID       int       `json:"receiver_id" gorm:"notNull;index"`
	ReceiverGroupID  int       `json:"receiver_group_id" gorm:"notNull;index"`
	RoutingRuleID    *int      `json:"routing_rule_id" gorm:"index"`

	// 通知详情
	Status           string    `json:"status" gorm:"size:20;notNull;default:'pending';index;check:status IN ('pending','sent','failed','retrying','max_retries_exceeded')"`
	NotificationType string    `json:"notification_type" gorm:"size:50;notNull"` // email, webhook, slack, dingtalk, feishu, wechat, sms

	// 内容
	Subject          string    `json:"subject" gorm:"type:text"`
	Body             string    `json:"body" gorm:"type:text"`
	RenderedTemplate string    `json:"rendered_template" gorm:"type:text"`

	// 投递跟踪
	SentAt           *time.Time `json:"sent_at"`
	DeliveredAt      *time.Time `json:"delivered_at"`
	FailedAt         *time.Time `json:"failed_at"`

	// 错误处理
	ErrorMessage     string    `json:"error_message" gorm:"type:text"`
	RetryCount       int       `json:"retry_count" gorm:"default:0"`
	MaxRetries       int       `json:"max_retries" gorm:"default:3"`
	NextRetryAt      *time.Time `json:"next_retry_at" gorm:"index"`

	// 元数据
	RequestPayload   JSONMap   `json:"request_payload" gorm:"type:jsonb"`
	ResponsePayload  JSONMap   `json:"response_payload" gorm:"type:jsonb"`
	DeliveryMetadata JSONMap   `json:"delivery_metadata" gorm:"type:jsonb;default:'{}'"`

	// 时间戳
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联
	AlertInstance    *AlertInstance    `json:"alert_instance,omitempty" gorm:"foreignKey:AlertInstanceID"`
	Receiver         *AlertReceiver    `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
	ReceiverGroup    *AlertReceiverGroup `json:"receiver_group,omitempty" gorm:"foreignKey:ReceiverGroupID"`
	RoutingRule      *AlertRoutingRule `json:"routing_rule,omitempty" gorm:"foreignKey:RoutingRuleID"`
}

// TableName 指定表名
func (NotificationLog) TableName() string {
	return "notification_logs"
}

// IsRetryable 检查是否可以重试
func (nl *NotificationLog) IsRetryable() bool {
	return nl.Status == "failed" && nl.RetryCount < nl.MaxRetries
}

// CanRetryNow 检查是否可以立即重试
func (nl *NotificationLog) CanRetryNow() bool {
	if !nl.IsRetryable() {
		return false
	}
	if nl.NextRetryAt == nil {
		return true
	}
	return time.Now().After(*nl.NextRetryAt)
}

// MarkAsSent 标记为已发送
func (nl *NotificationLog) MarkAsSent() {
	now := time.Now()
	nl.Status = "sent"
	nl.SentAt = &now
	nl.DeliveredAt = &now
}

// MarkAsFailed 标记为失败
func (nl *NotificationLog) MarkAsFailed(errorMsg string) {
	now := time.Now()
	nl.Status = "failed"
	nl.FailedAt = &now
	nl.ErrorMessage = errorMsg
}

// MarkAsRetrying 标记为重试中
func (nl *NotificationLog) MarkAsRetrying(nextRetryAt time.Time) {
	nl.Status = "retrying"
	nl.RetryCount++
	nl.NextRetryAt = &nextRetryAt
}

// MarkAsMaxRetriesExceeded 标记为超过最大重试次数
func (nl *NotificationLog) MarkAsMaxRetriesExceeded() {
	nl.Status = "max_retries_exceeded"
}

// ============================================
// 请求/响应模型
// ============================================

// NotificationLogListRequest 通知日志列表查询请求
type NotificationLogListRequest struct {
	Page            int      `form:"page,default=1"`
	PageSize        int      `form:"page_size,default=20"`
	AlertInstanceID *int     `form:"alert_instance_id"`
	ReceiverID      *int     `form:"receiver_id"`
	ReceiverGroupID *int     `form:"receiver_group_id"`
	Status          []string `form:"status"`
	StartTime       string   `form:"start_time"`
	EndTime         string   `form:"end_time"`
	SortField       string   `form:"sort_field,default=created_at"`
	SortOrder       string   `form:"sort_order,default=desc"`
}

// NotificationLogListResponse 通知日志列表响应
type NotificationLogListResponse struct {
	Total int               `json:"total"`
	Logs  []NotificationLog `json:"logs"`
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalSent              int     `json:"total_sent"`
	TotalFailed            int     `json:"total_failed"`
	TotalRetrying          int     `json:"total_retrying"`
	TotalMaxRetriesExceeded int    `json:"total_max_retries_exceeded"`
	SuccessRate            float64 `json:"success_rate"`
	AverageRetryCount      float64 `json:"average_retry_count"`
}
