package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/itcmdb/audit-service/internal/models"
	"github.com/itcmdb/audit-service/internal/repository"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

type AuditConsumer struct {
	consumer sarama.ConsumerGroupHandler
	repo     repository.AuditRepository
}

type AuditEvent struct {
	UserID     *uint              `json:"user_id"`
	Action     string             `json:"action"`
	Resource   string             `json:"resource"`
	ResourceID *uint              `json:"resource_id,omitempty"`
	Details     interface{}        `json:"details"`
	IPAddress   string             `json:"ip_address"`
	UserAgent   string             `json:"user_agent"`
	Status      string             `json:"status"`
	ErrorMsg    string             `json:"error_msg,omitempty"`
	Timestamp   string             `json:"timestamp"`
}

func NewAuditConsumer(consumer sarama.ConsumerGroupHandler, repo repository.AuditRepository) *AuditConsumer {
	return &AuditConsumer{
		consumer: consumer,
		repo:     repo,
	}
}

func (c *AuditConsumer) Start(ctx context.Context) error {
	topic := "audit_logs"

	consumer := sarama.ConsumerGroupHandler(c)

	// 订阅topic
	if err := c.consumer.Consume(topic, consumer); err != nil {
		return fmt.Errorf("failed to consume topic %s: %w", topic, err)
	}

	logger.Info("Audit consumer started", zap.String("topic", topic))

	// 等待context取消
	<-ctx.Done()

	logger.Info("Audit consumer stopped")
	return nil
}

// Setup 实现sarama.ConsumerGroupHandler接口
func (c *AuditConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 实现sarama.ConsumerGroupHandler接口
func (c *AuditConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 实现sarama.ConsumerGroupHandler接口
func (c *AuditConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 批量处理消息
	batchSize := 100
	logs := make([]models.AuditLog, 0, batchSize)

	for _, message := range claim.Messages {
		var event AuditEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			logger.Error("Failed to unmarshal audit event", zap.Error(err), zap.String("value", string(message.Value)))
			continue
		}

		// 转换为模型
		log := models.AuditLog{
			UserID:     event.UserID,
			Action:     event.Action,
			Resource:   event.Resource,
			ResourceID: event.ResourceID,
			IPAddress:  event.IPAddress,
			UserAgent:  event.UserAgent,
			Status:     event.Status,
			ErrorMsg:   event.ErrorMsg,
		}

		// 序列化details
		if event.Details != nil {
			if detailsBytes, err := json.Marshal(event.Details); err == nil {
				if err := json.Unmarshal(detailsBytes, &log.Details); err != nil {
					logger.Error("Failed to unmarshal details", zap.Error(err))
				}
			}
		}

		logs = append(logs, log)

		// 批量写入
		if len(logs) >= batchSize {
			if err := c.repo.CreateBatch(logs); err != nil {
				logger.Error("Failed to create audit logs batch", zap.Error(err))
			} else {
				logger.Debug("Created audit logs batch", zap.Int("count", len(logs)))
			}
			logs = make([]models.AuditLog, 0, batchSize)
		}
	}

	// 写入剩余日志
	if len(logs) > 0 {
		if err := c.repo.CreateBatch(logs); err != nil {
			logger.Error("Failed to create audit logs final batch", zap.Error(err))
		} else {
			logger.Debug("Created audit logs final batch", zap.Int("count", len(logs)))
		}
	}

	return nil
}
