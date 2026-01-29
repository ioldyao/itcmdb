package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ============================================
// 告警路由规则模型
// ============================================

// AlertRoutingRule 告警路由规则
type AlertRoutingRule struct {
	ID                int                    `json:"id" gorm:"primaryKey"`
	Name              string                 `json:"name" gorm:"size:255;notNull"`
	Description       string                 `json:"description" gorm:"type:text"`
	Matchers          JSONMap                `json:"matchers" gorm:"type:jsonb;notNull;default:'{}'"`
	MatchType         string                 `json:"match_type" gorm:"size:20;notNull;default:'match';check:match_type IN ('match','match_re')"`
	ReceiverGroupID   *int                   `json:"receiver_group_id" gorm:"index"`
	ReceiverGroup     *AlertReceiverGroup    `json:"receiver_group,omitempty" gorm:"foreignKey:ReceiverGroupID"`
	Continue          bool                   `json:"continue" gorm:"default:false"`
	Priority          int                    `json:"priority" gorm:"notNull;default:0"`
	Enabled           bool                   `json:"enabled" gorm:"default:true;index"`
	CreatedBy         *int                   `json:"created_by"`
	UpdatedBy         *int                   `json:"updated_by"`
	CreatedAt         time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AlertRoutingRule) TableName() string {
	return "alert_routing_rules"
}

// Matches 检查给定的labels是否匹配此路由规则
func (r *AlertRoutingRule) Matches(labels map[string]string) bool {
	if !r.Enabled {
		return false
	}

	// 遍历matchers中的所有条件
	for key, expectedValue := range r.Matchers {
		actualValue, exists := labels[key]

		// 如果labels中没有这个key，不匹配
		if !exists {
			return false
		}

		// 根据match_type进行匹配
		if r.MatchType == "match" {
			// 完全匹配
			if actualValue != expectedValue {
				return false
			}
		} else if r.MatchType == "match_re" {
			// TODO: 实现正则匹配 (需要编译正则表达式)
			// 简化实现：先按完全匹配处理
			if actualValue != expectedValue {
				return false
			}
		}
	}

	return true
}

// ============================================
// 告警通知模板模型
// ============================================

// AlertNotificationTemplate 告警通知模板
type AlertNotificationTemplate struct {
	ID              int        `json:"id" gorm:"primaryKey"`
	Name            string     `json:"name" gorm:"size:255;uniqueIndex;notNull"`
	Description     string     `json:"description" gorm:"type:text"`
	TemplateType    string     `json:"template_type" gorm:"size:50;notNull;index;check:template_type IN ('dingtalk','feishu','wechat','email')"`
	TemplateContent string     `json:"template_content" gorm:"type:text;notNull"`
	IsDefault       bool       `json:"is_default" gorm:"default:false;index"`
	CreatedBy       *int       `json:"created_by"`
	UpdatedBy       *int       `json:"updated_by"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AlertNotificationTemplate) TableName() string {
	return "alert_notification_templates"
}

// ============================================
// 模板数据结构
// ============================================

// AlertTemplateData 告警模板数据
type AlertTemplateData struct {
	AlertID     string            `json:"alert_id"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	Instance    string            `json:"instance"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Timestamp   string            `json:"timestamp"`
}

// ============================================
// 请求/响应模型
// ============================================

// CreateAlertRoutingRuleRequest 创建路由规则请求
type CreateAlertRoutingRuleRequest struct {
	Name            string                 `json:"name" binding:"required"`
	Description     string                 `json:"description"`
	Matchers        map[string]interface{} `json:"matchers" binding:"required"`
	MatchType       string                 `json:"match_type" binding:"required,oneof=match match_re"`
	ReceiverGroupID *int                   `json:"receiver_group_id"`
	Continue        bool                   `json:"continue"`
	Priority        int                    `json:"priority"`
	Enabled         bool                   `json:"enabled"`
}

// UpdateAlertRoutingRuleRequest 更新路由规则请求
type UpdateAlertRoutingRuleRequest struct {
	Name            *string                `json:"name"`
	Description     *string                `json:"description"`
	Matchers        map[string]interface{} `json:"matchers"`
	MatchType       *string                `json:"match_type" binding:"omitempty,oneof=match match_re"`
	ReceiverGroupID *int                   `json:"receiver_group_id"`
	Continue        *bool                  `json:"continue"`
	Priority        *int                   `json:"priority"`
	Enabled         *bool                  `json:"enabled"`
}

// CreateAlertNotificationTemplateRequest 创建通知模板请求
type CreateAlertNotificationTemplateRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	TemplateType    string `json:"template_type" binding:"required,oneof=dingtalk feishu wechat email"`
	TemplateContent string `json:"template_content" binding:"required"`
	IsDefault       bool   `json:"is_default"`
}

// UpdateAlertNotificationTemplateRequest 更新通知模板请求
type UpdateAlertNotificationTemplateRequest struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	TemplateContent *string `json:"template_content"`
	IsDefault       *bool   `json:"is_default"`
}

// ============================================
// JSONB 自定义类型 (已在alert.go中定义，这里不需要重复定义)
// ============================================

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
