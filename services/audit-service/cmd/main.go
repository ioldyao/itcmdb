package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/audit-service/internal/consumer"
	"github.com/itcmdb/audit-service/internal/handlers"
	"github.com/itcmdb/audit-service/internal/repository"
	"github.com/itcmdb/shared/pkg/database"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err := logger.Init(viper.GetString("server.log_level")); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	// 初始化数据库
	if err := database.Init(database.Config{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
		SSLMode:  viper.GetString("database.sslmode"),
	}); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 初始化repository
	db := database.Get()
	auditRepo := repository.NewAuditRepository(db)
	auditHandler := handlers.NewAuditHandler(auditRepo)

	// 创建Kafka消费者
	brokers := viper.GetStringSlice("kafka.brokers")
	groupID := viper.GetString("kafka.group_id")
	consumerConfig := getKafkaConfig()
	kafkaConsumer, err := sarama.NewConsumerGroup(brokers, groupID, consumerConfig)
	if err != nil {
		logger.Fatal("Failed to create Kafka consumer", zap.Error(err))
	}

	// 创建审计消费者
	auditConsumer := consumer.NewAuditConsumer(kafkaConsumer, auditRepo)

	// 启动消费者
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := auditConsumer.Start(ctx); err != nil {
			logger.Error("Consumer stopped", zap.Error(err))
		}
	}()

	// 设置Gin
	if viper.GetString("server.env") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 注册路由
	setupRoutes(r, auditHandler)

	// 启动HTTP服务器
	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))
	go func() {
		if err := r.Run(addr); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	logger.Info("Audit service started", zap.String("addr", addr))

	// 等待中断信号
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	logger.Info("Shutting down audit service...")
	cancel()
	time.Sleep(2 * time.Second)
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/audit-service")
	viper.AddConfigPath("./internal/config")

	// 绑定环境变量
	viper.BindEnv("server.port", "AUDIT_SERVER_PORT")
	viper.BindEnv("server.log_level", "AUDIT_LOG_LEVEL")
	viper.BindEnv("database.host", "AUDIT_DATABASE_HOST")
	viper.BindEnv("database.port", "AUDIT_DATABASE_PORT")
	viper.BindEnv("database.user", "AUDIT_DATABASE_USER")
	viper.BindEnv("database.password", "AUDIT_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "AUDIT_DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "AUDIT_DATABASE_SSLMODE")
	viper.BindEnv("kafka.brokers", "AUDIT_KAFKA_BROKERS")
	viper.BindEnv("kafka.topic", "AUDIT_KAFKA_TOPIC")
	viper.BindEnv("kafka.group_id", "AUDIT_KAFKA_GROUP_ID")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUDIT")

	// 设置默认值
	viper.SetDefault("server.port", "5007")
	viper.SetDefault("server.log_level", "info")
	viper.SetDefault("database.host", "postgres")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "itcmdb")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.SetDefault("kafka.topic", "audit_logs")
	viper.SetDefault("kafka.group_id", "audit_consumer")
	viper.SetDefault("consumer.batch_size", 100)
	viper.SetDefault("consumer.batch_timeout", "1s")
	viper.SetDefault("consumer.workers", 4)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn("Config file not found, using defaults", zap.Error(err))
		} else {
			return err
		}
	}

	return nil
}

func getKafkaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0 // Kafka 4.1.0
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Net.KeepAlive = 30 * time.Second

	return config
}

func setupRoutes(r *gin.Engine, auditHandler *handlers.AuditHandler) {
	api := r.Group("/api/v1")
	{
		audit := api.Group("/audit")
		{
			audit.GET("", auditHandler.GetAuditLogs)
			audit.GET("/stats", auditHandler.GetAuditStats)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
