package kafka

import (
	"encoding/json"
	"time"
)

// Event 事件基础结构
type Event struct {
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// CI事件类型
const (
	EventCIChanged              = "ci.changed"
	EventCIDeleted              = "ci.deleted"
	EventCIRelationshipChanged  = "ci.relationship.changed"
)

// 工单事件类型
const (
	EventTicketCreated       = "ticket.created"
	EventTicketStatusChanged = "ticket.status.changed"
	EventTicketAssigned      = "ticket.assigned"
	EventTicketSLABreached   = "ticket.sla.breached"
)

// 告警事件类型
const (
	EventAlertTriggered     = "alert.triggered"
	EventAlertAcknowledged  = "alert.acknowledged"
	EventAlertClosed        = "alert.closed"
	EventAlertEscalated     = "alert.escalated"
)

// 通知事件类型
const (
	EventNotificationSend = "notification.send"
)

// NewEvent 创建新事件
func NewEvent(eventType string, data map[string]interface{}) *Event {
	return &Event{
		EventType: eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// ToJSON 转换为JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON 从JSON解析
func FromJSON(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// CI事件数据结构
type CIChangedEvent struct {
	CIID       uint                   `json:"ci_id"`
	CITypeID   uint                   `json:"ci_type_id"`
	CIName     string                 `json:"ci_name"`
	Action     string                 `json:"action"` // create, update, delete
	ChangedBy  uint                   `json:"changed_by"`
	OldValue   map[string]interface{} `json:"old_value,omitempty"`
	NewValue   map[string]interface{} `json:"new_value,omitempty"`
}

// 工单事件数据结构
type TicketEvent struct {
	TicketID    uint   `json:"ticket_id"`
	Title       string `json:"title"`
	Status      string `json:"status,omitempty"`
	Priority    string `json:"priority,omitempty"`
	AssigneeID  uint   `json:"assignee_id,omitempty"`
	RequesterID uint   `json:"requester_id"`
	Action      string `json:"action"`
}

// 告警事件数据结构
type AlertEvent struct {
	AlertID       uint   `json:"alert_id"`
	RuleID        uint   `json:"rule_id"`
	Title         string `json:"title"`
	Severity      string `json:"severity"`
	Status        string `json:"status"`
	AffectedCIID  uint   `json:"affected_ci_id,omitempty"`
	AcknowledgedBy uint  `json:"acknowledged_by,omitempty"`
	Action        string `json:"action"`
}

// 通知事件数据结构
type NotificationEvent struct {
	NotificationID uint     `json:"notification_id"`
	Type           string   `json:"type"` // email, wechat, sms
	Recipients     []string `json:"recipients"`
	Subject        string   `json:"subject"`
	Content        string   `json:"content"`
	RelatedType    string   `json:"related_type,omitempty"` // ticket, alert, ci
	RelatedID      uint     `json:"related_id,omitempty"`
}
