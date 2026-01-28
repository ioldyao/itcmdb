package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AlertReceiverGroup 告警接收组
type AlertReceiverGroup struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"unique;not null" validate:"required"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Receivers   []AlertReceiver `json:"receivers,omitempty" gorm:"many2many:alert_receiver_group_members"`
}

// AlertReceiver 告警接收人
type AlertReceiver struct {
	ID          int                `json:"id" gorm:"primaryKey"`
	Name        string             `json:"name" gorm:"not null" validate:"required"`
	Type        string             `json:"type" gorm:"not null;type:varchar(50)" validate:"required,oneof=wechat dingtalk feishu email sms"`
	WebhookURL  string             `json:"webhook_url" gorm:"type:text"`
	AtMobiles   StringArray        `json:"at_mobiles" gorm:"type:text[]"`
	AtUserIDs   StringArray        `json:"at_user_ids" gorm:"type:text[]"`
	Secret      string             `json:"secret"`
	Config      JSONMap            `json:"config" gorm:"type:jsonb"`
	Enabled     bool               `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// StringArray 字符串数组类型
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	return json.Unmarshal(value.([]byte), a)
}

// Value 实现 driver.Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// AlertNotificationRequest 告警通知请求
type AlertNotificationRequest struct {
	AlertID    string                 `json:"alert_id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Severity   string                 `json:"severity"`
	Status     string                 `json:"status"`
	Metadata   map[string]interface{} `json:"metadata"`
	ReceiverIDs []int                 `json:"receiver_ids"`
}

// DingTalkMessage 钉钉消息格式
type DingTalkMessage struct {
	MsgType string      `json:"msgtype"`
	Text    *TextContent `json:"text,omitempty"`
	Markdown *MarkdownContent `json:"markdown,omitempty"`
	ActionCard *ActionCardContent `json:"action_card,omitempty"`
	At      *At         `json:"at,omitempty"`
}

type TextContent struct {
	Content string `json:"content"`
}

type MarkdownContent struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type ActionCardContent struct {
	Title          string              `json:"title"`
	Markdown       string              `json:"markdown"`
	SingleTitle    string              `json:"single_title,omitempty"`
	SingleURL      string              `json:"single_url,omitempty"`
	BtnOrientation string              `json:"btn_orientation,omitempty"`
	BtnJsonList    []ActionButton      `json:"btn_json_list,omitempty"`
}

type ActionButton struct {
	Title     string `json:"title"`
	ActionURL string `json:"action_url"`
}

type At struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIDs []string `json:"atUserIds,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

// FeishuMessage 飞书消息格式
type FeishuMessage struct {
	MsgType string      `json:"msg_type"`
	Content interface{} `json:"content"`
	Card    string      `json:"card,omitempty"`
}

type FeishuTextContent struct {
	Text string `json:"text"`
}

type FeishuPostContent struct {
	Title   string                    `json:"title"`
	Content []FeishuPostContentElement `json:"content"`
}

type FeishuPostContentElement struct {
	Tag  string `json:"tag"`
	Text string `json:"text,omitempty"`
}

// WechatMessage 企业微信消息格式
type WechatMessage struct {
	MsgType       string             `json:"msgtype"`
	Text          *WechatTextContent `json:"text,omitempty"`
	Markdown      *WechatMarkdownContent `json:"markdown,omitempty"`
	News          *WechatNewsContent `json:"news,omitempty"`
	TemplateCard  interface{}        `json:"template_card,omitempty"`
}

type WechatTextContent struct {
	Content               string   `json:"content"`
	MentionedList         []string `json:"mentioned_list,omitempty"`
	MentionedMobileList   []string `json:"mentioned_mobile_list,omitempty"`
}

type WechatMarkdownContent struct {
	Content string `json:"content"`
}

type WechatNewsContent struct {
	Articles []WechatArticle `json:"articles"`
}

type WechatArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl"`
}

// 请求和响应模型

// CreateReceiverGroupRequest 创建接收组请求
type CreateReceiverGroupRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	ReceiverIDs []int    `json:"receiver_ids"`
}

// UpdateReceiverGroupRequest 更新接收组请求
type UpdateReceiverGroupRequest struct {
	Name        *string  `json:"name" validate:"omitempty,min=1"`
	Description *string  `json:"description"`
	Enabled     *bool    `json:"enabled"`
	ReceiverIDs []int    `json:"receiver_ids"`
}

// CreateReceiverRequest 创建接收人请求
type CreateReceiverRequest struct {
	Name       string                 `json:"name" validate:"required"`
	Type       string                 `json:"type" validate:"required,oneof=wechat dingtalk feishu email sms"`
	WebhookURL string                 `json:"webhook_url"`
	AtMobiles  []string               `json:"at_mobiles"`
	AtUserIDs  []string               `json:"at_user_ids"`
	Secret     string                 `json:"secret"`
	Config     map[string]interface{} `json:"config"`
}

// UpdateReceiverRequest 更新接收人请求
type UpdateReceiverRequest struct {
	Name       *string                 `json:"name" validate:"omitempty,min=1"`
	WebhookURL *string                 `json:"webhook_url"`
	AtMobiles  *[]string               `json:"at_mobiles"`
	AtUserIDs  *[]string               `json:"at_user_ids"`
	Secret     *string                 `json:"secret"`
	Config     map[string]interface{}  `json:"config"`
	Enabled    *bool                   `json:"enabled"`
}

// ReceiverGroupListResponse 接收组列表响应
type ReceiverGroupListResponse struct {
	Total  int                   `json:"total"`
	Groups []AlertReceiverGroup  `json:"groups"`
}

// ReceiverListResponse 接收人列表响应
type ReceiverListResponse struct {
	Total      int             `json:"total"`
	Receivers []AlertReceiver `json:"receivers"`
}
