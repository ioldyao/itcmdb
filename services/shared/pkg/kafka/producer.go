package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

type Producer struct {
	producer sarama.SyncProducer
	topic   string
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

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Version = sarama.V2_8_0_0

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		topic:   topic,
	}, nil
}

func (p *Producer) SendAuditEvent(event AuditEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.Resource),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		logger.Error("Failed to send audit event",
			zap.String("action", event.Action),
			zap.String("resource", event.Resource),
			zap.Error(err),
		)
		return err
	}

	logger.Debug("Audit event sent",
		zap.String("action", event.Action),
		zap.String("resource", event.Resource),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset)),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
