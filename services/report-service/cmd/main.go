package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	logger.Info("Report service starting", zap.String("addr", addr))
	r.Run(addr)
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 绑定嵌套配置到环境变量
	viper.BindEnv("database.host", "REPORT_DATABASE_HOST")
	viper.BindEnv("database.port", "REPORT_DATABASE_PORT")
	viper.BindEnv("database.user", "REPORT_DATABASE_USER")
	viper.BindEnv("database.password", "REPORT_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "REPORT_DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "REPORT_DATABASE_SSLMODE")
	viper.BindEnv("jwt.secret", "REPORT_JWT_SECRET")
	viper.BindEnv("jwt.expiration", "REPORT_JWT_EXPIRATION")
	viper.BindEnv("server.port", "REPORT_SERVER_PORT")
	viper.BindEnv("log.level", "REPORT_LOG_LEVEL")

	// 必须在 SetDefault 之前调用 AutomaticEnv
	viper.AutomaticEnv()
	viper.SetEnvPrefix("REPORT")

	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5006")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5433)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "itcmdb")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration", "24h")
	viper.ReadInConfig()
	return nil
}

func setupRoutes(r *gin.Engine, jwtManager *auth.JWTManager) {
	api := r.Group("/api/v1")
	api.Use(jwtManager.AuthMiddleware())
	{
		api.GET("/reports/cmdb/assets", rbac.RequirePermission("report", "view"), getCMDBAssetReportHandler())
		api.GET("/reports/tickets/stats", rbac.RequirePermission("report", "view"), getTicketStatsHandler())
		api.GET("/reports/alerts/trends", rbac.RequirePermission("report", "view"), getAlertTrendsHandler())
		api.POST("/reports/export", rbac.RequirePermission("report", "view"), exportReportHandler())
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func getCMDBAssetReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}

func getTicketStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}

func getAlertTrendsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}

func exportReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}
