package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/alert-service/internal/handlers"
	"github.com/itcmdb/alert-service/internal/services"
	"github.com/itcmdb/shared/pkg/audit"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/database"
	"github.com/itcmdb/shared/pkg/logger"
	"github.com/itcmdb/shared/pkg/rbac"
	"github.com/itcmdb/shared/pkg/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := logger.Init(viper.GetString("log.level")); err != nil {
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

	db := database.Get()

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	// 初始化审计日志Kafka生产者
	kafkaBrokers := viper.GetStringSlice("kafka.brokers")
	if err := audit.InitProducer(kafkaBrokers); err != nil {
		logger.Warn("Failed to init audit producer, audit logging disabled", zap.Error(err))
	} else {
		defer audit.CloseProducer()
	}

	// 初始化VictoriaMetrics客户端
	vmClient := services.NewVictoriaMetricsClient(
		viper.GetString("victoriametrics.endpoint"),
		viper.GetString("victoriametrics.username"),
		viper.GetString("victoriametrics.password"),
	)

	// 初始化告警引擎
	alertEngine := services.NewAlertEngine(vmClient)

	// 初始化Webhook服务
	webhookService := services.NewWebhookService(db)

	// 初始化JWT管理器
	jwtManager := auth.NewJWTManager(
		viper.GetString("jwt.secret"),
		viper.GetDuration("jwt.expiration"),
	)

	// 设置Gin模式
	if viper.GetString("env") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	setupRoutes(r, db, alertEngine, vmClient, webhookService, jwtManager)

	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))

	// 记录平台启动事件
	audit.LogPlatformEvent("platform_start", "alert-service", map[string]interface{}{
		"addr": addr,
	})

	logger.Info("Alert service starting", zap.String("addr", addr))

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
	audit.LogPlatformEvent("platform_stop", "alert-service", nil)

	logger.Info("Shutting down alert service...")
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 绑定嵌套配置到环境变量
	viper.BindEnv("database.host", "ALERT_DATABASE_HOST")
	viper.BindEnv("database.port", "ALERT_DATABASE_PORT")
	viper.BindEnv("database.user", "ALERT_DATABASE_USER")
	viper.BindEnv("database.password", "ALERT_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "ALERT_DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "ALERT_DATABASE_SSLMODE")
	viper.BindEnv("jwt.secret", "ALERT_JWT_SECRET")
	viper.BindEnv("jwt.expiration", "ALERT_JWT_EXPIRATION")
	viper.BindEnv("server.port", "ALERT_SERVER_PORT")
	viper.BindEnv("log.level", "ALERT_LOG_LEVEL")
	viper.BindEnv("victoriametrics.endpoint", "ALERT_VICTORIAMETRICS_ENDPOINT")
	viper.BindEnv("victoriametrics.username", "ALERT_VICTORIAMETRICS_USERNAME")
	viper.BindEnv("victoriametrics.password", "ALERT_VICTORIAMETRICS_PASSWORD")

	// 必须在 SetDefault 之前调用 AutomaticEnv
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ALERT")

	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5004")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5433)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "itcmdb")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration", "24h")
	viper.SetDefault("victoriametrics.endpoint", "http://localhost:8428")
	viper.SetDefault("victoriametrics.username", "")
	viper.SetDefault("victoriametrics.password", "")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.ReadInConfig()
	return nil
}

func autoMigrate(db *gorm.DB) error {
	// 自动迁移表结构
	// 注意：生产环境建议使用SQL迁移脚本而不是自动迁移
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 检查表是否存在
	var tableExists int
	sqlDB.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'alert_rules'").Scan(&tableExists)

	if tableExists == 0 {
		logger.Info("Alert tables not found, please run migration script first")
		logger.Warn("Please run: psql -h localhost -U postgres -d itcmdb -f services/alert-service/migrations/001_init_alerts.sql")
	}

	return nil
}

func setupRoutes(r *gin.Engine, db *gorm.DB, alertEngine *services.AlertEngine, vmClient *services.VictoriaMetricsClient, webhookService *services.WebhookService, jwtManager *auth.JWTManager) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		health := gin.H{"status": "ok", "service": "alert-service"}

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

	// VM健康检查
	r.GET("/health/vm", func(c *gin.Context) {
		if err := vmClient.HealthCheck(); err != nil {
			c.JSON(503, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok", "victoriametrics": "connected"})
	})

	api := r.Group("/api/v1")
	{
		// 公开端点：外部告警接入
		api.POST("/alerts/ingest", ingestAlertHandler(db, alertEngine))

		// 公开端点：接收Webhook（使用token）
		api.POST("/webhooks/inbound/:token", handlers.HandleInboundWebhook(db, webhookService))

		// 受保护的端点：需要JWT认证
		protected := api.Group("")
		protected.Use(jwtManager.AuthMiddleware())
		{
			// 告警管理
			alertHandler := handlers.NewAlertHandler(db, alertEngine, vmClient)
			protected.GET("/alerts", rbac.RequirePermission("alert", "view"), alertHandler.GetAlerts)
			protected.GET("/alerts/:id", rbac.RequirePermission("alert", "view"), alertHandler.GetAlertByID)
			protected.GET("/alerts/:id/history", rbac.RequirePermission("alert", "view"), alertHandler.GetAlertHistory)
			protected.POST("/alerts/:id/ack", rbac.RequirePermission("alert", "manage"), alertHandler.AcknowledgeAlert)
			protected.POST("/alerts/:id/close", rbac.RequirePermission("alert", "manage"), alertHandler.CloseAlert)
			protected.POST("/alerts/batch/ack", rbac.RequirePermission("alert", "manage"), alertHandler.BatchAcknowledgeAlerts)
			protected.POST("/alerts/batch/close", rbac.RequirePermission("alert", "manage"), alertHandler.BatchCloseAlerts)
			protected.GET("/alerts/statistics", rbac.RequirePermission("alert", "view"), alertHandler.GetAlertStatistics)
			protected.GET("/alerts/analytics", rbac.RequirePermission("alert", "view"), alertHandler.GetAlertAnalytics)

			// 静默管理
			silenceHandler := handlers.NewSilenceHandler(db)
			protected.GET("/silences", rbac.RequirePermission("alert", "view"), silenceHandler.ListSilences)
			protected.GET("/silences/:id", rbac.RequirePermission("alert", "view"), silenceHandler.GetSilence)
			protected.POST("/silences", rbac.RequirePermission("alert", "manage"), silenceHandler.CreateSilence)
			protected.PUT("/silences/:id", rbac.RequirePermission("alert", "manage"), silenceHandler.UpdateSilence)
			protected.DELETE("/silences/:id", rbac.RequirePermission("alert", "manage"), silenceHandler.DeleteSilence)

			// 空间管理（管理员）
			spaceHandler := handlers.NewSpaceHandler(db)
			protected.GET("/roles", rbac.RequirePermission("alert", "view"), spaceHandler.ListRoles)
			protected.GET("/spaces", rbac.RequirePermission("alert", "view"), spaceHandler.ListSpaces)
			protected.POST("/spaces", rbac.RequirePermission("alert", "manage"), spaceHandler.CreateSpace)
			protected.PUT("/spaces/:id", rbac.RequirePermission("alert", "manage"), spaceHandler.UpdateSpace)
			protected.DELETE("/spaces/:id", rbac.RequirePermission("alert", "manage"), spaceHandler.DeleteSpace)

			// 空间路由规则（管理员）
			protected.GET("/space-routes", rbac.RequirePermission("alert", "view"), spaceHandler.ListSpaceRoutes)
			protected.POST("/space-routes", rbac.RequirePermission("alert", "manage"), spaceHandler.CreateSpaceRoute)
			protected.PUT("/space-routes/:id", rbac.RequirePermission("alert", "manage"), spaceHandler.UpdateSpaceRoute)
			protected.DELETE("/space-routes/:id", rbac.RequirePermission("alert", "manage"), spaceHandler.DeleteSpaceRoute)

			// 规则管理
			ruleHandler := handlers.NewRuleHandler(db)
			protected.GET("/rules", rbac.RequirePermission("alert_rule", "view"), ruleHandler.GetRules)
			protected.GET("/rules/:id", rbac.RequirePermission("alert_rule", "view"), ruleHandler.GetRuleByID)
			protected.POST("/rules", rbac.RequirePermission("alert_rule", "create"), ruleHandler.CreateRule)
			protected.PUT("/rules/:id", rbac.RequirePermission("alert_rule", "update"), ruleHandler.UpdateRule)
			protected.DELETE("/rules/:id", rbac.RequirePermission("alert_rule", "delete"), ruleHandler.DeleteRule)
			protected.POST("/rules/:id/enable", rbac.RequirePermission("alert_rule", "manage"), ruleHandler.EnableRule)
			protected.POST("/rules/:id/disable", rbac.RequirePermission("alert_rule", "manage"), ruleHandler.DisableRule)
			protected.POST("/rules/test", rbac.RequirePermission("alert_rule", "test"), ruleHandler.TestRule)

			// 接收人管理
			receiverHandler := handlers.NewReceiverHandler(db)
			protected.GET("/receivers", rbac.RequirePermission("alert_receiver", "view"), receiverHandler.ListReceivers)
			protected.GET("/receivers/:id", rbac.RequirePermission("alert_receiver", "view"), receiverHandler.GetReceiver)
			protected.POST("/receivers", rbac.RequirePermission("alert_receiver", "create"), receiverHandler.CreateReceiver)
			protected.PUT("/receivers/:id", rbac.RequirePermission("alert_receiver", "update"), receiverHandler.UpdateReceiver)
			protected.DELETE("/receivers/:id", rbac.RequirePermission("alert_receiver", "delete"), receiverHandler.DeleteReceiver)
			protected.POST("/receivers/:id/test", rbac.RequirePermission("alert_receiver", "test"), receiverHandler.TestReceiver)

			// 接收组管理
			groupHandler := handlers.NewReceiverGroupHandler(db)
			protected.GET("/receiver-groups", rbac.RequirePermission("alert_receiver", "view"), groupHandler.ListReceiverGroups)
			protected.GET("/receiver-groups/:id", rbac.RequirePermission("alert_receiver", "view"), groupHandler.GetReceiverGroup)
			protected.POST("/receiver-groups", rbac.RequirePermission("alert_receiver", "create"), groupHandler.CreateReceiverGroup)
			protected.PUT("/receiver-groups/:id", rbac.RequirePermission("alert_receiver", "update"), groupHandler.UpdateReceiverGroup)
			protected.DELETE("/receiver-groups/:id", rbac.RequirePermission("alert_receiver", "delete"), groupHandler.DeleteReceiverGroup)

			// Webhook集成管理
			inboundHandler := handlers.NewInboundWebhookHandler(db, webhookService)
			protected.GET("/webhooks/inbound", rbac.RequirePermission("webhook", "view"), inboundHandler.ListInboundWebhooks)
			protected.GET("/webhooks/inbound/:id", rbac.RequirePermission("webhook", "view"), inboundHandler.GetInboundWebhook)
			protected.POST("/webhooks/inbound", rbac.RequirePermission("webhook", "create"), inboundHandler.CreateInboundWebhook)
			protected.PUT("/webhooks/inbound/:id", rbac.RequirePermission("webhook", "update"), inboundHandler.UpdateInboundWebhook)
			protected.DELETE("/webhooks/inbound/:id", rbac.RequirePermission("webhook", "delete"), inboundHandler.DeleteInboundWebhook)

			outboundHandler := handlers.NewOutboundWebhookHandler(db, webhookService)
			protected.GET("/webhooks/outbound", rbac.RequirePermission("webhook", "view"), outboundHandler.ListOutboundWebhooks)
			protected.GET("/webhooks/outbound/:id", rbac.RequirePermission("webhook", "view"), outboundHandler.GetOutboundWebhook)
			protected.POST("/webhooks/outbound", rbac.RequirePermission("webhook", "create"), outboundHandler.CreateOutboundWebhook)
			protected.PUT("/webhooks/outbound/:id", rbac.RequirePermission("webhook", "update"), outboundHandler.UpdateOutboundWebhook)
			protected.DELETE("/webhooks/outbound/:id", rbac.RequirePermission("webhook", "delete"), outboundHandler.DeleteOutboundWebhook)
			protected.POST("/webhooks/outbound/:id/test", rbac.RequirePermission("webhook", "test"), outboundHandler.TestOutboundWebhook)

			// 路由规则管理
			routingHandler := handlers.NewRoutingHandler(db)
			protected.GET("/alert-routing-rules", rbac.RequirePermission("routing", "view"), routingHandler.ListRoutingRules)
			protected.GET("/alert-routing-rules/:id", rbac.RequirePermission("routing", "view"), routingHandler.GetRoutingRule)
			protected.POST("/alert-routing-rules", rbac.RequirePermission("routing", "create"), routingHandler.CreateRoutingRule)
			protected.PUT("/alert-routing-rules/:id", rbac.RequirePermission("routing", "update"), routingHandler.UpdateRoutingRule)
			protected.DELETE("/alert-routing-rules/:id", rbac.RequirePermission("routing", "delete"), routingHandler.DeleteRoutingRule)

			// 通知模板管理
			templateHandler := handlers.NewTemplateHandler(db)
			protected.GET("/alert-notification-templates", rbac.RequirePermission("template", "view"), templateHandler.ListNotificationTemplates)
			protected.GET("/alert-notification-templates/:id", rbac.RequirePermission("template", "view"), templateHandler.GetNotificationTemplate)
			protected.POST("/alert-notification-templates", rbac.RequirePermission("template", "create"), templateHandler.CreateNotificationTemplate)
			protected.PUT("/alert-notification-templates/:id", rbac.RequirePermission("template", "update"), templateHandler.UpdateNotificationTemplate)
			protected.DELETE("/alert-notification-templates/:id", rbac.RequirePermission("template", "delete"), templateHandler.DeleteNotificationTemplate)
			protected.POST("/alert-notification-templates/:id/set-default", rbac.RequirePermission("template", "manage"), templateHandler.SetDefaultTemplate)
			protected.POST("/alert-notification-templates/preview", rbac.RequirePermission("template", "view"), templateHandler.PreviewTemplate)
		}
	}
}

// ingestAlertHandler 外部告警接入处理
func ingestAlertHandler(db *gorm.DB, alertEngine *services.AlertEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, response.Error("Invalid payload", err.Error()))
			return
		}

		// TODO: 实现告警接入逻辑
		// 1. 验证payload格式
		// 2. 提取告警信息
		// 3. 生成指纹并去重
		// 4. 存储到数据库
		// 5. 发送通知

		logger.Info("Received alert ingestion", zap.Any("payload", payload))

		c.JSON(200, response.Success(gin.H{
			"message":  "Alert ingested successfully",
			"alert_id": services.GenerateAlertID(),
		}))
	}
}

// 后台任务：定期评估告警规则
func startAlertEvaluator(db *gorm.DB, alertEngine *services.AlertEngine) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		logger.Info("Evaluating alert rules...")

		// TODO: 实现规则评估逻辑
		// 1. 查询所有启用的规则
		// 2. 评估每个规则
		// 3. 创建或更新告警实例
		// 4. 发送通知
	}
}
