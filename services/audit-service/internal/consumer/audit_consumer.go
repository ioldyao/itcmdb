package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/itcmdb/audit-service/internal/models"
	"github.com/itcmdb/audit-service/internal/repository"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

type AuditConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	repo          repository.AuditRepository
}

type AuditEvent struct {
	UserID     *uint       `json:"user_id"`
	Action     string      `json:"action"`
	Resource   string      `json:"resource"`
	ResourceID *uint       `json:"resource_id,omitempty"`
	Details    interface{} `json:"details"`
	IPAddress  string      `json:"ip_address"`
	UserAgent  string      `json:"user_agent"`
	Status     string      `json:"status"`
	ErrorMsg   string      `json:"error_msg,omitempty"`
	Timestamp  string      `json:"timestamp"`
}

func NewAuditConsumer(consumerGroup sarama.ConsumerGroup, repo repository.AuditRepository) *AuditConsumer {
	return &AuditConsumer{
		consumerGroup: consumerGroup,
		topic:         "audit_logs",
		repo:          repo,
	}
}

func (c *AuditConsumer) Start(ctx context.Context) error {
	// 创建一个handler来处理消息
	handler := &consumerHandler{repo: c.repo}

	// 启动消费循环
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("Stopping audit consumer...")
				return
			default:
				if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
					if err == sarama.ErrClosedConsumerGroup {
						return
					}
					logger.Error("Error from consumer", zap.Error(err))
				}
			}
		}
	}()

	logger.Info("Audit consumer started", zap.String("topic", c.topic))
	return nil
}

// consumerHandler 实现sarama.ConsumerGroupHandler接口
type consumerHandler struct {
	repo repository.AuditRepository
}

// Setup 实现sarama.ConsumerGroupHandler接口
func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 实现sarama.ConsumerGroupHandler接口
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 实现sarama.ConsumerGroupHandler接口
func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger.Info("Consumer claim started",
		zap.String("topic", claim.Topic()),
		zap.Int32("partition", claim.Partition()),
	)

	// 批量处理消息 - 减小批量大小以更快写入
	batchSize := 10
	logs := make([]models.AuditLog, 0, batchSize)
	messageCount := 0

	// 添加超时定时器，确保即使消息不足也能及时写入
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				// Channel关闭，写入剩余日志
				if len(logs) > 0 {
					if err := h.repo.CreateBatch(logs); err != nil {
						logger.Error("Failed to create audit logs final batch", zap.Error(err))
					} else {
						logger.Info("Created audit logs final batch", zap.Int("count", len(logs)))
					}
				}
				logger.Info("Consumer claim finished", zap.Int("total_messages", messageCount))
				return nil
			}

			// 重置定时器
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(1 * time.Second)

			messageCount++
			logger.Info("Received audit message",
				zap.Int("count", messageCount),
				zap.Int32("partition", message.Partition),
				zap.Int64("offset", message.Offset),
			)

			var event AuditEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				logger.Error("Failed to unmarshal audit event", zap.Error(err), zap.String("value", string(message.Value)))
				session.MarkMessage(message, "")
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

			// 标记消息为已处理
			session.MarkMessage(message, "")

			// 批量写入
			if len(logs) >= batchSize {
				if err := h.repo.CreateBatch(logs); err != nil {
					logger.Error("Failed to create audit logs batch", zap.Error(err))
				} else {
					logger.Info("Created audit logs batch", zap.Int("count", len(logs)))
				}
				logs = make([]models.AuditLog, 0, batchSize)
			}

		case <-timer.C:
			// 超时，写入累积的日志
			if len(logs) > 0 {
				if err := h.repo.CreateBatch(logs); err != nil {
					logger.Error("Failed to create audit logs on timeout", zap.Error(err))
				} else {
					logger.Info("Created audit logs on timeout", zap.Int("count", len(logs)))
				}
				logs = make([]models.AuditLog, 0, batchSize)
			}
			// 重置定时器
			timer.Reset(1 * time.Second)
		}
	}
}

// Close 关闭消费者
func (c *AuditConsumer) Close() error {
	if c.consumerGroup != nil {
		return c.consumerGroup.Close()
	}
	return nil
}
