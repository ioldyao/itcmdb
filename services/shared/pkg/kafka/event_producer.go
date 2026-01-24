package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

var eventProducer sarama.SyncProducer

// InitEventProducer 初始化事件生产者
func InitEventProducer(brokers []string) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return err
	}

	eventProducer = producer
	logger.Info("Event producer initialized", zap.Strings("brokers", brokers))
	return nil
}

// PublishEvent 发布事件到Kafka
func PublishEvent(topic string, event *Event) error {
	if eventProducer == nil {
		logger.Warn("Event producer not initialized, skipping event publish")
		return nil
	}

	data, err := event.ToJSON()
	if err != nil {
		logger.Error("Failed to marshal event", zap.Error(err))
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := eventProducer.SendMessage(msg)
	if err != nil {
		logger.Error("Failed to publish event",
			zap.String("topic", topic),
			zap.String("event_type", event.EventType),
			zap.Error(err))
		return err
	}

	logger.Info("Event published",
		zap.String("topic", topic),
		zap.String("event_type", event.EventType),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// PublishCIEvent 发布CI事件
func PublishCIEvent(eventType string, data interface{}) error {
	dataMap, err := structToMap(data)
	if err != nil {
		return err
	}

	event := NewEvent(eventType, dataMap)
	return PublishEvent("ci_events", event)
}

// PublishTicketEvent 发布工单事件
func PublishTicketEvent(eventType string, data interface{}) error {
	dataMap, err := structToMap(data)
	if err != nil {
		return err
	}

	event := NewEvent(eventType, dataMap)
	return PublishEvent("ticket_events", event)
}

// PublishAlertEvent 发布告警事件
func PublishAlertEvent(eventType string, data interface{}) error {
	dataMap, err := structToMap(data)
	if err != nil {
		return err
	}

	event := NewEvent(eventType, dataMap)
	return PublishEvent("alert_events", event)
}

// PublishNotificationEvent 发布通知事件
func PublishNotificationEvent(data interface{}) error {
	dataMap, err := structToMap(data)
	if err != nil {
		return err
	}

	event := NewEvent(EventNotificationSend, dataMap)
	return PublishEvent("notification_events", event)
}

// CloseEventProducer 关闭事件生产者
func CloseEventProducer() error {
	if eventProducer != nil {
		return eventProducer.Close()
	}
	return nil
}

// structToMap 将结构体转换为map
func structToMap(data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
