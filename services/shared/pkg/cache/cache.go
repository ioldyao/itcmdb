package cache

import (
	"context"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

var client *redis.Client

// Config Redis配置
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Init 初始化Redis连接
func Init(cfg Config) error {
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}

	logger.Info("Redis connected", zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)))
	return nil
}

// Get 获取Redis客户端
func Get() *redis.Client {
	return client
}

// Close 关闭连接
func Close() error {
	return client.Close()
}
