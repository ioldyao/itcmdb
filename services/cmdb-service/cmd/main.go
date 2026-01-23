package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/database"
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
	logger.Info("CMDB service starting", zap.String("addr", addr))
	r.Run(addr)
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./internal/config")

	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5002")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration", "24h")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "itcmdb")
	viper.SetDefault("database.sslmode", "disable")

	viper.ReadInConfig()
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CMDB")
	return nil
}

func setupRoutes(r *gin.Engine, jwtManager *auth.JWTManager) {
	api := r.Group("/api/v1")
	api.Use(jwtManager.AuthMiddleware())
	{
		ci := api.Group("/ci")
		{
			ci.GET("/types", getCITypesHandler())
			ci.POST("/instances", createCIInstanceHandler())
			ci.GET("/instances", getCIInstancesHandler())
			ci.GET("/instances/:id", getCIInstanceHandler())
			ci.PUT("/instances/:id", updateCIInstanceHandler())
			ci.DELETE("/instances/:id", deleteCIInstanceHandler())
			ci.GET("/relations", getCIRelationsHandler())
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

// Handler functions
func getCITypesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现获取CI类型逻辑
		types := []interface{}{}
		c.JSON(200, response.Success(types))
	}
}

func createCIInstanceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}

func getCIInstancesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success([]interface{}{}))
	}
}

func getCIInstanceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, response.Success(gin.H{"id": id}))
	}
}

func updateCIInstanceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}

func deleteCIInstanceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success(nil))
	}
}

func getCIRelationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, response.Success([]interface{}{}))
	}
}
