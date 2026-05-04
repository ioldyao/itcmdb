package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/shared/pkg/audit"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/database"
	"github.com/itcmdb/shared/pkg/logger"
	"github.com/itcmdb/shared/pkg/rbac"
	"github.com/itcmdb/shared/pkg/response"
	"go.uber.org/zap"
)

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := logger.Init(viper.GetString("log.level")); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

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

	// 初始化审计日志Kafka生产者
	kafkaBrokers := viper.GetStringSlice("kafka.brokers")
	if err := audit.InitProducer(kafkaBrokers); err != nil {
		logger.Warn("Failed to init audit producer, audit logging disabled", zap.Error(err))
	} else {
		defer audit.CloseProducer()
	}

	jwtManager := auth.NewJWTManager(
		viper.GetString("jwt.secret"),
		viper.GetDuration("jwt.expiration"),
	)

	if viper.GetString("env") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	setupRoutes(r, jwtManager)

	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))

	// 记录平台启动事件
	audit.LogPlatformEvent("platform_start", "ticket-service", map[string]interface{}{
		"addr": addr,
	})

	logger.Info("Ticket service starting", zap.String("addr", addr))

	// 启动HTTP服务器
	go func() {
		if err := r.Run(addr); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// 等待中断信号
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	// 记录平台停止事件
	audit.LogPlatformEvent("platform_stop", "ticket-service", nil)

	logger.Info("Shutting down ticket service...")
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 绑定嵌套配置到环境变量
	viper.BindEnv("database.host", "TICKET_DATABASE_HOST")
	viper.BindEnv("database.port", "TICKET_DATABASE_PORT")
	viper.BindEnv("database.user", "TICKET_DATABASE_USER")
	viper.BindEnv("database.password", "TICKET_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "TICKET_DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "TICKET_DATABASE_SSLMODE")
	viper.BindEnv("jwt.secret", "TICKET_JWT_SECRET")
	viper.BindEnv("jwt.expiration", "TICKET_JWT_EXPIRATION")
	viper.BindEnv("server.port", "TICKET_SERVER_PORT")
	viper.BindEnv("log.level", "TICKET_LOG_LEVEL")

	// 必须在 SetDefault 之前调用 AutomaticEnv
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TICKET")

	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5003")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5433)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "itcmdb")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration", "24h")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.ReadInConfig()
	return nil
}

func setupRoutes(r *gin.Engine, jwtManager *auth.JWTManager) {
	api := r.Group("/api/v1")
	api.Use(jwtManager.AuthMiddleware())
	{
		api.GET("/tickets", rbac.RequirePermission("ticket", "view"), getTicketsHandler())
		api.POST("/tickets", rbac.RequirePermission("ticket", "create"), createTicketHandler())
		api.GET("/tickets/:id", rbac.RequirePermission("ticket", "view"), getTicketHandler())
		api.PUT("/tickets/:id/status", rbac.RequirePermission("ticket", "update"), updateTicketStatusHandler())
		api.POST("/tickets/:id/comments", rbac.RequirePermission("ticket", "update"), addCommentHandler())
		api.GET("/workflows", rbac.RequirePermission("ticket", "view"), getWorkflowsHandler())
		api.GET("/tickets/:id/sla", rbac.RequirePermission("ticket", "view"), getTicketSLAHandler())
	}

	r.GET("/health", func(c *gin.Context) {
		health := gin.H{"status": "ok", "service": "ticket-service"}

		db := database.Get()
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			health["status"] = "degraded"
			health["database"] = "unavailable"
		} else {
			health["database"] = "ok"
		}

		status := 200
		if health["status"] == "degraded" {
			status = 503
		}
		c.JSON(status, health)
	})
}

func getTicketsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success([]interface{}{}))
	}
}

func createTicketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		audit.LogSuccess(c, "create", "ticket", nil, nil)
		c.JSON(200, response.Success(nil))
	}
}

func getTicketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}

func updateTicketStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		audit.LogSuccess(c, "update_status", "ticket", nil, map[string]interface{}{
			"ticket_id": id,
		})
		c.JSON(200, response.Success(nil))
	}
}

func addCommentHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		audit.LogSuccess(c, "add_comment", "ticket", nil, map[string]interface{}{
			"ticket_id": id,
		})
		c.JSON(200, response.Success(nil))
	}
}

func getWorkflowsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success([]interface{}{}))
	}
}

func getTicketSLAHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}
