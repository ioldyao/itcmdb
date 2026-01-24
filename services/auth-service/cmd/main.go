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
	"github.com/itcmdb/auth-service/internal/handlers"
	"github.com/itcmdb/auth-service/internal/repository"
	"github.com/itcmdb/auth-service/internal/service"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/audit"
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

	// 自动迁移数据库表
	// TODO: AutoMigrate has issues with existing tables, skipping for now
	// if err := models.AutoMigrate(); err != nil {
	// 	logger.Fatal("Failed to migrate database", zap.Error(err))
	// }
	logger.Info("Skipping database migration")

	// 初始化用户服务
	userService := service.NewUserService()

	// 初始化角色服务
	db := database.Get()
	roleRepo := repository.NewRoleRepository(db)
	roleService := service.NewRoleService(roleRepo)
	roleHandler := handlers.NewRoleHandler(roleService)

	// 初始化审计日志Kafka生产者
	kafkaBrokers := []string{"kafka:9092"}
	if err := audit.InitProducer(kafkaBrokers); err != nil {
		logger.Warn("Failed to initialize audit producer, audit logging disabled", zap.Error(err))
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
	setupRoutes(r, jwtManager, userService, roleHandler)

	// 启动服务
	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))
	logger.Info("Auth service starting", zap.String("addr", addr))

	// 优雅关闭
	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm

		logger.Info("Shutting down auth service...")
		audit.CloseProducer()
		time.Sleep(1 * time.Second)
	}()

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

func setupRoutes(r *gin.Engine, jwtManager *auth.JWTManager, userService service.UserService, roleHandler *handlers.RoleHandler) {
	api := r.Group("/api/v1")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/register", registerHandler(userService, jwtManager))
			auth.POST("/login", loginHandler(userService, jwtManager))
			auth.POST("/logout", logoutHandler())
			auth.POST("/refresh", refreshHandler(jwtManager))
		}

		// 用户相关（需要认证）
		users := api.Group("/users")
		users.Use(jwtManager.AuthMiddleware())
		{
			users.GET("/me", getMeHandler(userService))
			users.PUT("/me", updateMeHandler(userService))
			users.GET("/me/permissions", getMyPermissionsHandler(userService))
			users.GET("/:id/permissions", getPermissionsHandler(userService))
		}

		// 角色管理（需要认证）
		roles := api.Group("/roles")
		roles.Use(jwtManager.AuthMiddleware())
		{
			roles.GET("", roleHandler.GetRoles)
			roles.POST("", roleHandler.CreateRole)
			roles.PUT("/:id", roleHandler.UpdateRole)
			roles.DELETE("/:id", roleHandler.DeleteRole)
			roles.GET("/:id/permissions", roleHandler.GetRolePermissions)
			roles.GET("/:id/users", roleHandler.GetRoleUsers)
		}

		// 权限管理（需要认证）
		permissions := api.Group("/permissions")
		permissions.Use(jwtManager.AuthMiddleware())
		{
			permissions.GET("", roleHandler.GetPermissions)
			permissions.POST("", roleHandler.CreatePermission)
			permissions.DELETE("/:id", roleHandler.DeletePermission)
		}

		// 角色权限关联（需要认证）
		rolePermissions := api.Group("/role-permissions")
		rolePermissions.Use(jwtManager.AuthMiddleware())
		{
			rolePermissions.POST("", roleHandler.AssignPermissionToRole)
			rolePermissions.DELETE("", roleHandler.RemovePermissionFromRole)
		}

		// 用户角色关联（需要认证）
		userRoles := api.Group("/user-roles")
		userRoles.Use(jwtManager.AuthMiddleware())
		{
			userRoles.POST("", roleHandler.AssignRoleToUser)
			userRoles.DELETE("", roleHandler.RemoveRoleFromUser)
			userRoles.GET("/user/:id", roleHandler.GetUserRoles)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

// Handler functions

// registerHandler 用户注册
func registerHandler(userService service.UserService, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
			FullName string `json:"full_name"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, response.Error("Bad Request", "invalid request: "+err.Error()))
			return
		}

		// 创建用户
		user, err := userService.Register(req.Username, req.Email, req.Password, req.FullName)
		if err != nil {
			c.JSON(400, response.Error("Bad Request", err.Error()))
			return
		}

		// 获取用户权限
		permissions, err := userService.GetUserPermissions(user.ID)
		if err != nil {
			logger.Warn("Failed to get user permissions", zap.Error(err))
			permissions = []string{}
		}

		// 生成token
		token, err := jwtManager.Generate(int64(user.ID), user.Username, permissions)
		if err != nil {
			c.JSON(500, response.Error("Internal Error", "failed to generate token"))
			return
		}

		c.JSON(200, response.Success(gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"fullName": user.FullName,
			},
			"permissions": permissions,
		}))
	}
}

// loginHandler 用户登录
func loginHandler(userService service.UserService, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, response.Error("Bad Request", "invalid request"))
			return
		}

		// 验证用户
		user, err := userService.ValidateUser(req.Username, req.Password)
		if err != nil {
			c.JSON(401, response.Error("Unauthorized", "invalid username or password"))
			return
		}

		// 获取用户权限
		permissions, err := userService.GetUserPermissions(user.ID)
		if err != nil {
			logger.Warn("Failed to get user permissions", zap.Error(err))
			permissions = []string{}
		}

		// 生成token
		token, err := jwtManager.Generate(int64(user.ID), user.Username, permissions)
		if err != nil {
			c.JSON(500, response.Error("Internal Error", "failed to generate token"))
			return
		}

		c.JSON(200, response.Success(gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"fullName": user.FullName,
			},
			"permissions": permissions,
		}))
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
			c.JSON(400, response.Error("Bad Request", "invalid request"))
			return
		}

		newToken, err := jwtManager.Refresh(req.Token)
		if err != nil {
			c.JSON(401, response.Error("Unauthorized", err.Error()))
			return
		}

		c.JSON(200, response.Success(gin.H{"token": newToken}))
	}
}

func getMeHandler(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := auth.GetUserID(c)

		user, err := userService.GetUserByID(uint(userID))
		if err != nil {
			c.JSON(404, response.Error("Not Found", "user not found"))
			return
		}

		c.JSON(200, response.Success(gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"fullName": user.FullName,
			"status":   user.Status,
		}))
	}
}

func updateMeHandler(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := auth.GetUserID(c)

		var req struct {
			FullName string `json:"full_name"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, response.Error("Bad Request", "invalid request"))
			return
		}

		updates := make(map[string]interface{})
		if req.FullName != "" {
			updates["full_name"] = req.FullName
		}
		if req.Password != "" {
			updates["password"] = req.Password
		}

		if err := userService.UpdateUser(uint(userID), updates); err != nil {
			c.JSON(400, response.Error("Bad Request", err.Error()))
			return
		}

		c.JSON(200, response.Success(nil))
	}
}

func getMyPermissionsHandler(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := auth.GetUserID(c)

		permissions, err := userService.GetUserPermissions(uint(userID))
		if err != nil {
			c.JSON(500, response.Error("Internal Error", "failed to get permissions"))
			return
		}

		c.JSON(200, response.Success(permissions))
	}
}

func getPermissionsHandler(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var id uint
		if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
			c.JSON(400, response.Error("Bad Request", "invalid user id"))
			return
		}

		permissions, err := userService.GetUserPermissions(id)
		if err != nil {
			c.JSON(500, response.Error("Internal Error", "failed to get permissions"))
			return
		}

		c.JSON(200, response.Success(permissions))
	}
}
