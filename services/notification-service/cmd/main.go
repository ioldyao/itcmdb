package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/logger"
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
	logger.Info("Notification service starting", zap.String("addr", addr))
	r.Run(addr)
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5005")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration", "24h")
	viper.ReadInConfig()
	viper.AutomaticEnv()
	viper.SetEnvPrefix("NOTIFICATION")
	return nil
}

func setupRoutes(r *gin.Engine, jwtManager *auth.JWTManager) {
	api := r.Group("/api/v1")
	api.Use(jwtManager.AuthMiddleware())
	{
		api.GET("/notifications", getNotificationsHandler())
		api.POST("/notifications/send", sendNotificationHandler())
		api.GET("/templates", getTemplatesHandler())
		api.POST("/templates", createTemplateHandler())
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func getNotificationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success([]interface{}{}))
	}
}

func sendNotificationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}

func getTemplatesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success([]interface{}{}))
	}
}

func createTemplateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}
