package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

var writers = make(map[string]*kafka.Writer)
var readers = make(map[string]*kafka.Reader)

// Config Kafka配置
type Config struct {
	Brokers []string
	GroupID string
}

// Init 初始化Kafka
func Init(cfg Config) {
	logger.Info("Kafka initialized", zap.Strings("brokers", cfg.Brokers))
}

// CreateWriter 创建Writer
func CreateWriter(topic string, brokers []string) *kafka.Writer {
	if w, ok := writers[topic]; ok {
		return w
	}

	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:   topic,
		Balancer: &kafka.LeastBytes{},
	}

	writers[topic] = w
	return w
}

// CreateReader 创建Reader
func CreateReader(topic, groupID string, brokers []string) *kafka.Reader {
	key := topic + ":" + groupID
	if r, ok := readers[key]; ok {
		return r
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	readers[key] = r
	return r
}

// PublishMessage 发送消息
func PublishMessage(ctx context.Context, topic string, brokers []string, key, value []byte) error {
	w := CreateWriter(topic, brokers)
	return w.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

// Close 关闭所有连接
func Close() {
	for _, w := range writers {
		w.Close()
	}
	for _, r := range readers {
		r.Close()
	}
}
