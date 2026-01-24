package audit

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/shared/pkg/kafka"
)

var producer *kafka.Producer

// InitProducer 初始化Kafka生产者
func InitProducer(brokers []string) error {
	var err error
	producer, err = kafka.NewProducer(brokers, "audit_logs")
	return err
}

// CloseProducer 关闭生产者
func CloseProducer() error {
	if producer != nil {
		return producer.Close()
	}
	return nil
}

// LogOptions 审计日志选项
type LogOptions struct {
	Details   interface{} `json:"details"`
	Status    string      `json:"status"` // success, failed
	ErrorMsg  string      `json:"error_msg,omitempty"`
}

// Log 记录审计日志
func Log(c *gin.Context, action, resource string, resourceID *uint, opts LogOptions) {
	if producer == nil {
		return
	}

	// 获取用户ID
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		if uidInt, ok := uid.(uint); ok {
			userID = &uidInt
		}
	}

	// 获取IP和UserAgent
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	event := kafka.AuditEvent{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    opts.Details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     opts.Status,
		ErrorMsg:   opts.ErrorMsg,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	// 异步发送，不阻塞业务逻辑
	go func() {
		_ = producer.SendAuditEvent(event)
	}()
}

// LogSuccess 记录成功操作
func LogSuccess(c *gin.Context, action, resource string, resourceID *uint, details interface{}) {
	Log(c, action, resource, resourceID, LogOptions{
		Details: details,
		Status:  "success",
	})
}

// LogError 记录失败操作
func LogError(c *gin.Context, action, resource string, resourceID *uint, errorMsg string, details interface{}) {
	Log(c, action, resource, resourceID, LogOptions{
		Details:  details,
		Status:  "failed",
		ErrorMsg: errorMsg,
	})
}
