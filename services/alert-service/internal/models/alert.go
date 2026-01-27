package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ============================================
// 数据模型定义
// ============================================

// AlertRule 告警规则
type AlertRule struct {
	ID                   int                    `json:"id" gorm:"primaryKey"`
	Name                 string                 `json:"name" gorm:"uniqueIndex;size:255;notNull"`
	Description          string                 `json:"description" gorm:"type:text"`
	MetricQuery          string                 `json:"metric_query" gorm:"type:text;notNull"`
	ThresholdOperator    string                 `json:"threshold_operator" gorm:"size:10;notNull;check:threshold_operator IN ('>','<','>=','<=','==','!=')"`
	ThresholdValue       float64                `json:"threshold_value" gorm:"notNull"`
	Duration             int                    `json:"duration" gorm:"default:300"` // 持续时间（秒）
	Severity             string                 `json:"severity" gorm:"size:20;notNull;check:severity IN ('critical','high','medium','low')"`
	Enabled              bool                   `json:"enabled" gorm:"default:true"`
	CITypeID             *int                   `json:"ci_type_id" gorm:"index"`
	NotificationChannels JSONMap                `json:"notification_channels" gorm:"type:jsonb"`
	SilencedUntil        *time.Time             `json:"silenced_until"`
	CreatedBy            *int                   `json:"created_by"`
	UpdatedBy            *int                   `json:"updated_by"`
	CreatedAt            time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt            *time.Time             `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (AlertRule) TableName() string {
	return "alert_rules"
}

// AlertInstance 告警实例
type AlertInstance struct {
	ID                   int                    `json:"id" gorm:"primaryKey"`
	AlertID              string                 `json:"alert_id" gorm:"uniqueIndex;size:64;notNull"`
	RuleID               *int                   `json:"rule_id" gorm:"index"`

	// 基本信息
	Title                string                 `json:"title" gorm:"size:255;notNull"`
	Description          string                 `json:"description" gorm:"type:text"`
	Severity             string                 `json:"severity" gorm:"size:20;notNull;check:severity IN ('critical','high','medium','low')"`
	Status               string                 `json:"status" gorm:"size:20;notNull;default:'firing';check:status IN ('firing','acknowledged','resolved','closed')"`

	// 分类信息
	Category             string                 `json:"category" gorm:"size:100;index"`
	Tags                 JSONMap                `json:"tags" gorm:"type:jsonb"`
	ObjectType           string                 `json:"object_type" gorm:"size:100"`

	// 目标信息
	TargetInfo           JSONMap                `json:"target_info" gorm:"type:jsonb"`
	AffectedCIID         *int                   `json:"affected_ci_id" gorm:"index"`

	// 触发条件
	TriggerConditions    JSONMap                `json:"trigger_conditions" gorm:"type:jsonb"`
	Metrics              JSONMap                `json:"metrics" gorm:"type:jsonb"`

	// 去重指纹
	Fingerprint          string                 `json:"fingerprint" gorm:"size:64;notNull;index"`

	// 时间信息
	FirstTriggered       time.Time              `json:"first_triggered" gorm:"notNull;index"`
	LastTriggered        time.Time              `json:"last_triggered" gorm:"notNull;index"`
	RecoveredAt          *time.Time             `json:"recovered_at"`
	ClosedAt             *time.Time             `json:"closed_at"`

	// 计数
	Count                int                    `json:"count" gorm:"default:1"`

	// 处理信息
	Handler              *int                   `json:"handler" gorm:"index"`
	HandlingStatus       string                 `json:"handling_status" gorm:"size:20"`
	HandlingNotes        string                 `json:"handling_notes" gorm:"type:text"`
	AcknowledgedAt       *time.Time             `json:"acknowledged_at"`

	// 通知信息
	NotificationSent     bool                   `json:"notification_sent" gorm:"default:false"`
	NotificationChannels JSONMap                `json:"notification_channels" gorm:"type:jsonb"`

	// 审计字段
	CreatedAt            time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AlertInstance) TableName() string {
	return "alert_instances"
}

// AlertHistory 告警历史
type AlertHistory struct {
	ID          int                    `json:"id" gorm:"primaryKey"`
	AlertID     int                    `json:"alert_id" gorm:"index:alert_id;notNull"`
	EventType   string                 `json:"event_type" gorm:"size:50;notNull;check:event_type IN ('triggered','updated','acknowledged','resolved','closed')"`
	OldStatus   string                 `json:"old_status" gorm:"size:20"`
	NewStatus   string                 `json:"new_status" gorm:"size:20"`
	OperatedBy  *int                   `json:"operated_by"`
	OperatedAt  time.Time              `json:"operated_at" gorm:"autoCreateTime;index"`
	Message     string                 `json:"message" gorm:"type:text"`
	Details     JSONMap                `json:"details" gorm:"type:jsonb"`
}

// TableName 指定表名
func (AlertHistory) TableName() string {
	return "alert_history"
}

// AlertSilence 告警静默规则
type AlertSilence struct {
	ID         int                    `json:"id" gorm:"primaryKey"`
	Name       string                 `json:"name" gorm:"size:255;notNull"`
	Comment    string                 `json:"comment" gorm:"type:text"`
	Matchers   JSONMap                `json:"matchers" gorm:"type:jsonb;notNull"`
	StartsAt   time.Time              `json:"starts_at" gorm:"notNull"`
	EndsAt     time.Time              `json:"ends_at" gorm:"notNull"`
	CreatedBy  *int                   `json:"created_by"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	Active     bool                   `json:"active" gorm:"default:true;index"`
}

// TableName 指定表名
func (AlertSilence) TableName() string {
	return "alert_silences"
}

// AlertAggregation 告警聚合
type AlertAggregation struct {
	ID                int                    `json:"id" gorm:"primaryKey"`
	AggregationKey    string                 `json:"aggregation_key" gorm:"uniqueIndex;size:255;notNull"`
	BaseAlertID       *int                   `json:"base_alert_id" gorm:"index"`
	AlertCount        int                    `json:"alert_count" gorm:"default:1"`
	RelatedAlertIDs   JSONMap                `json:"related_alert_ids" gorm:"type:jsonb"`
	FirstTriggered    time.Time              `json:"first_triggered" gorm:"notNull"`
	LastTriggered     time.Time              `json:"last_triggered" gorm:"notNull;index"`
	UpdatedAt         time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AlertAggregation) TableName() string {
	return "alert_aggregations"
}

// ============================================
// 请求/响应模型
// ============================================

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	Name                 string                 `json:"name" binding:"required"`
	Description          string                 `json:"description"`
	MetricQuery          string                 `json:"metric_query" binding:"required"`
	ThresholdOperator    string                 `json:"threshold_operator" binding:"required,oneof=> < >= <= == !="`
	ThresholdValue       float64                `json:"threshold_value" binding:"required"`
	Duration             int                    `json:"duration"`
	Severity             string                 `json:"severity" binding:"required,oneof=critical high medium low"`
	Enabled              bool                   `json:"enabled"`
	CITypeID             *int                   `json:"ci_type_id"`
	NotificationChannels map[string]interface{} `json:"notification_channels"`
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	Description          *string                `json:"description"`
	MetricQuery          *string                `json:"metric_query"`
	ThresholdOperator    *string                `json:"threshold_operator" binding:"omitempty,oneof=> < >= <= == !="`
	ThresholdValue       *float64               `json:"threshold_value"`
	Duration             *int                   `json:"duration"`
	Severity             *string                `json:"severity" binding:"omitempty,oneof=critical high medium low"`
	Enabled              *bool                  `json:"enabled"`
	CITypeID             *int                   `json:"ci_type_id"`
	NotificationChannels map[string]interface{} `json:"notification_channels"`
	SilencedUntil        *string                `json:"silenced_until"` // ISO 8601格式
}

// AcknowledgeAlertRequest 确认告警请求
type AcknowledgeAlertRequest struct {
	Handler *int    `json:"handler" binding:"required"`
	Notes   string `json:"notes"`
}

// CloseAlertRequest 关闭告警请求
type CloseAlertRequest struct {
	Handler *int    `json:"handler" binding:"required"`
	Notes   string `json:"notes"`
}

// AlertListRequest 告警列表查询请求
type AlertListRequest struct {
	Page          int                    `form:"page,default=1"`
	PageSize      int                    `form:"page_size,default=20"`
	Status        []string               `form:"status"`
	Severity      []string               `form:"severity"`
	Category      string                 `form:"category"`
	SearchKeyword string                 `form:"search_keyword"`
	StartTime     string                 `form:"start_time"`
	EndTime       string                 `form:"end_time"`
	SortField     string                 `form:"sort_field,default=last_triggered"`
	SortOrder     string                 `form:"sort_order,default=desc"`
}

// AlertListResponse 告警列表响应
type AlertListResponse struct {
	Total  int                      `json:"total"`
	Alerts []AlertInstance          `json:"alerts"`
}

// AlertAnalyticsRequest 告警分析请求
type AlertAnalyticsRequest struct {
	StartTime string   `form:"start_time" binding:"required"`
	EndTime   string   `form:"end_time" binding:"required"`
	GroupBy   []string `form:"group_by"`
}

// AlertAnalyticsResponse 告警分析响应
type AlertAnalyticsResponse struct {
	Dimensions []AnalyticsDimension `json:"dimensions"`
	TimeSeries TimeSeriesData        `json:"time_series"`
}

// AnalyticsDimension 分析维度
type AnalyticsDimension struct {
	DimensionType string         `json:"dimension_type"`
	DimensionName string         `json:"dimension_name"`
	TotalCount    int            `json:"total_count"`
	Items         []DimensionItem `json:"items"`
}

// DimensionItem 维度项目
type DimensionItem struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	Dates  []string            `json:"dates"`
	Series []TimeSeriesSeries  `json:"series"`
}

// TimeSeriesSeries 时间序列
type TimeSeriesSeries struct {
	Name string   `json:"name"`
	Data []int    `json:"data"`
}

// ============================================
// JSONB 自定义类型
// ============================================

// JSONMap JSONB map类型
type JSONMap map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
