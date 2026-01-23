package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/cmdb-service/internal/handlers"
	"github.com/itcmdb/cmdb-service/internal/repository"
	"github.com/itcmdb/cmdb-service/internal/service"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/database"
	"github.com/itcmdb/shared/pkg/logger"
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

	// 初始化依赖
	jwtManager := auth.NewJWTManager(
		viper.GetString("jwt.secret"),
		viper.GetDuration("jwt.expiration"),
	)

	db := database.GetDB()
	ciRepo := repository.NewCIRepository(db)
	ciService := service.NewCIService(ciRepo)
	ciHandler := handlers.NewCIHandler(ciService)

	if viper.GetString("env") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	setupRoutes(r, jwtManager, ciHandler)

	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))
	logger.Info("CMDB service starting", zap.String("addr", addr))
	r.Run(addr)
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./internal/config")

	// 绑定嵌套配置到环境变量
	viper.BindEnv("database.host", "CMDB_DATABASE_HOST")
	viper.BindEnv("database.port", "CMDB_DATABASE_PORT")
	viper.BindEnv("database.user", "CMDB_DATABASE_USER")
	viper.BindEnv("database.password", "CMDB_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "CMDB_DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "CMDB_DATABASE_SSLMODE")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("CMDB")

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
	return nil
}

func setupRoutes(r *gin.Engine, jwtManager *auth.JWTManager, ciHandler *handlers.CIHandler) {
	api := r.Group("/api/v1")
	api.Use(jwtManager.AuthMiddleware())
	{
		ci := api.Group("/ci")
		{
			ci.GET("/types", ciHandler.GetCITypes)
			ci.POST("/instances", ciHandler.CreateCIInstance)
			ci.GET("/instances", ciHandler.GetCIInstances)
			ci.GET("/instances/:id", ciHandler.GetCIInstance)
			ci.PUT("/instances/:id", ciHandler.UpdateCIInstance)
			ci.DELETE("/instances/:id", ciHandler.DeleteCIInstance)
			ci.GET("/instances/:id/history", ciHandler.GetCIHistory)
			ci.GET("/relations", ciHandler.GetCIRelations)
			ci.POST("/relations", ciHandler.CreateCIRelation)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
