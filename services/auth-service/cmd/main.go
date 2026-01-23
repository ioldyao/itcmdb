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
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
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

	// 初始化JWT管理器
	jwtManager := auth.NewJWTManager(
		viper.GetString("jwt.secret"),
		viper.GetDuration("jwt.expiration"),
	)

	// 设置Gin
	if viper.GetString("env") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 注册路由
	setupRoutes(r, jwtManager)

	// 启动服务
	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))
	logger.Info("Auth service starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/auth-service")
	viper.AddConfigPath("./internal/config")

	// 绑定嵌套配置到环境变量
	viper.BindEnv("database.host", "AUTH_DATABASE_HOST")
	viper.BindEnv("database.port", "AUTH_DATABASE_PORT")
	viper.BindEnv("database.user", "AUTH_DATABASE_USER")
	viper.BindEnv("database.password", "AUTH_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "AUTH_DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "AUTH_DATABASE_SSLMODE")
	viper.BindEnv("jwt.secret", "AUTH_JWT_SECRET")
	viper.BindEnv("jwt.expiration", "AUTH_JWT_EXPIRATION")
	viper.BindEnv("server.port", "AUTH_SERVER_PORT")
	viper.BindEnv("log.level", "AUTH_LOG_LEVEL")

	// 必须在 SetDefault 之前调用 AutomaticEnv
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTH")

	// 设置默认值
	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5001")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration", "24h")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "itcmdb")
	viper.SetDefault("database.sslmode", "disable")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn("Config file not found, using defaults", zap.Error(err))
		} else {
			return err
		}
	}

	return nil
}

func setupRoutes(r *gin.Engine, jwtManager *auth.JWTManager) {
	api := r.Group("/api/v1")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", loginHandler(jwtManager))
			auth.POST("/logout", logoutHandler())
			auth.POST("/refresh", refreshHandler(jwtManager))
		}

		// 用户相关（需要认证）
		users := api.Group("/users")
		users.Use(jwtManager.AuthMiddleware())
		{
			users.GET("/me", getMeHandler())
			users.PUT("/me", updateMeHandler())
			users.GET("/:id/permissions", getPermissionsHandler())
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

// Handler functions
func loginHandler(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, response.Error(400, "invalid request"))
			return
		}

		// TODO: 实现真实的用户认证逻辑
		// 这里简化处理，实际应该从数据库验证
		if req.Username == "admin" && req.Password == "admin123" {
			token, err := jwtManager.Generate(1, "admin", []string{"admin"})
			if err != nil {
				c.JSON(500, response.Error(500, "failed to generate token"))
				return
			}

			c.JSON(200, response.Success(gin.H{
				"token": token,
				"user": gin.H{
					"id":       1,
					"username": "admin",
					"email":    "admin@itcmdb.com",
					"fullName": "管理员",
				},
				"permissions": []string{"*:*"},
			}))
			return
		}

		c.JSON(401, response.Error(401, "invalid username or password"))
	}
}

func logoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现token黑名单逻辑
		c.JSON(200, response.Success(nil))
	}
}

func refreshHandler(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, response.Error(400, "invalid request"))
			return
		}

		newToken, err := jwtManager.Refresh(req.Token)
		if err != nil {
			c.JSON(401, response.Error(401, err.Error()))
			return
		}

		c.JSON(200, response.Success(gin.H{"token": newToken}))
	}
}

func getMeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := auth.GetUserID(c)
		username, _ := auth.GetUsername(c)

		c.JSON(200, response.Success(gin.H{
			"id":       userID,
			"username": username,
			"email":    username + "@itcmdb.com",
			"fullName": username,
		}))
	}
}

func updateMeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现更新用户信息逻辑
		c.JSON(200, response.Success(nil))
	}
}

func getPermissionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现获取用户权限逻辑
		c.JSON(200, response.Success([]string{"*:*"}))
	}
}
