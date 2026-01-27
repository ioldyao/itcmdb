package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/alert-service/internal/handlers"
	"github.com/itcmdb/alert-service/internal/services"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/database"
	"github.com/itcmdb/shared/pkg/logger"
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

	// 初始化VictoriaMetrics客户端
	vmClient := services.NewVictoriaMetricsClient(
		viper.GetString("victoriametrics.endpoint"),
		viper.GetString("victoriametrics.username"),
		viper.GetString("victoriametrics.password"),
	)

	// 初始化告警引擎
	alertEngine := services.NewAlertEngine(vmClient)

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
	setupRoutes(r, db, alertEngine, vmClient, jwtManager)

	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))
	logger.Info("Alert service starting", zap.String("addr", addr))
	r.Run(addr)
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

func setupRoutes(r *gin.Engine, db *gorm.DB, alertEngine *services.AlertEngine, vmClient *services.VictoriaMetricsClient, jwtManager *auth.JWTManager) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
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

		// 受保护的端点：需要JWT认证
		protected := api.Group("")
		protected.Use(jwtManager.AuthMiddleware())
		{
			// 告警管理
			alertHandler := handlers.NewAlertHandler(db, alertEngine, vmClient)
			protected.GET("/alerts", alertHandler.GetAlerts)
			protected.GET("/alerts/:id", alertHandler.GetAlertByID)
			protected.GET("/alerts/:id/history", alertHandler.GetAlertHistory)
			protected.POST("/alerts/:id/ack", alertHandler.AcknowledgeAlert)
			protected.POST("/alerts/:id/close", alertHandler.CloseAlert)
			protected.POST("/alerts/batch/ack", alertHandler.BatchAcknowledgeAlerts)
			protected.POST("/alerts/batch/close", alertHandler.BatchCloseAlerts)
			protected.GET("/alerts/statistics", alertHandler.GetAlertStatistics)
			protected.GET("/alerts/analytics", alertHandler.GetAlertAnalytics)

			// 规则管理
			ruleHandler := handlers.NewRuleHandler(db)
			protected.GET("/rules", ruleHandler.GetRules)
			protected.GET("/rules/:id", ruleHandler.GetRuleByID)
			protected.POST("/rules", ruleHandler.CreateRule)
			protected.PUT("/rules/:id", ruleHandler.UpdateRule)
			protected.DELETE("/rules/:id", ruleHandler.DeleteRule)
			protected.POST("/rules/:id/enable", ruleHandler.EnableRule)
			protected.POST("/rules/:id/disable", ruleHandler.DisableRule)
			protected.POST("/rules/test", ruleHandler.TestRule)
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
