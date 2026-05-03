package audit

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/shared/pkg/kafka"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

var producer *kafka.Producer

// 敏感字段列表 - 这些字段的值会被脱敏
var sensitiveFields = []string{
	"password",
	"pwd",
	"passwd",
	"token",
	"secret",
	"api_key",
	"apikey",
	"access_token",
	"refresh_token",
	"private_key",
	"credential",
	"auth",
	"authorization",
}

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

// filterSensitiveData 过滤敏感数据
func filterSensitiveData(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case map[string]interface{}:
		filtered := make(map[string]interface{})
		for key, value := range v {
			// 检查字段名是否为敏感字段
			if isSensitiveField(key) {
				filtered[key] = "***FILTERED***"
			} else {
				// 递归过滤嵌套的map
				filtered[key] = filterSensitiveData(value)
			}
		}
		return filtered
	case []interface{}:
		// 过滤数组中的每个元素
		filtered := make([]interface{}, len(v))
		for i, item := range v {
			filtered[i] = filterSensitiveData(item)
		}
		return filtered
	default:
		return v
	}
}

// isSensitiveField 检查字段名是否为敏感字段
func isSensitiveField(fieldName string) bool {
	lowerField := toLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if contains(lowerField, sensitive) {
			return true
		}
	}
	return false
}

// toLower 简单的字符串转小写
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Log 记录审计日志
func Log(c *gin.Context, action, resource string, resourceID *uint, opts LogOptions) {
	if producer == nil {
		logger.Warn("Audit producer is nil, skipping audit log")
		return
	}

	logger.Info("Recording audit log",
		zap.String("action", action),
		zap.String("resource", resource),
	)

	// 获取用户ID
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case uint64:
			uidUint := uint(v)
			userID = &uidUint
		case int64:
			uidUint := uint(v)
			userID = &uidUint
		case uint:
			userID = &v
		case int:
			uidUint := uint(v)
			userID = &uidUint
		}
	}

	// 获取IP和UserAgent
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 过滤敏感数据
	filteredDetails := filterSensitiveData(opts.Details)

	event := kafka.AuditEvent{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    filteredDetails,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     opts.Status,
		ErrorMsg:   opts.ErrorMsg,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	// 异步发送，不阻塞业务逻辑
	go func() {
		if err := producer.SendAuditEvent(event); err != nil {
			logger.Error("Failed to send audit event",
				zap.String("action", action),
				zap.String("resource", resource),
				zap.Error(err),
			)
		} else {
			logger.Info("Audit event sent successfully",
				zap.String("action", action),
				zap.String("resource", resource),
			)
		}
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
