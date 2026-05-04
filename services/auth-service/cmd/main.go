package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/auth-service/internal/handlers"
	grpcserver "github.com/itcmdb/auth-service/internal/grpc"
	"github.com/itcmdb/auth-service/internal/middleware"
	"github.com/itcmdb/auth-service/internal/models"
	"github.com/itcmdb/auth-service/internal/repository"
	"github.com/itcmdb/auth-service/internal/service"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/audit"
	"github.com/itcmdb/shared/pkg/cache"
	"github.com/itcmdb/shared/pkg/database"
	"github.com/itcmdb/shared/pkg/logger"
	"github.com/itcmdb/shared/pkg/response"
	pb "github.com/itcmdb/shared/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// 初始化Redis缓存
	if err := cache.Init(cache.Config{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	}); err != nil {
		logger.Warn("Failed to connect to Redis, running without cache", zap.Error(err))
	} else {
		logger.Info("Redis cache connected")
	}

	// 自动迁移数据库表
	if err := models.AutoMigrate(); err != nil {
		logger.Warn("Failed to migrate database", zap.Error(err))
		// 不终止服务启动，允许服务在迁移失败时继续运行
	} else {
		logger.Info("Database migration completed successfully")
	}

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

	// 初始化Auth服务
	authService := service.NewAuthService(jwtManager)

	// 启动gRPC服务器
	go startGRPCServer(authService, userService)

	// 设置Gin
	if viper.GetString("env") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 注册路由
	setupRoutes(r, jwtManager, userService, roleHandler)

	// 启动REST API服务
	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))

	// 记录平台启动事件
	audit.LogPlatformEvent("platform_start", "auth-service", map[string]interface{}{
		"addr": addr,
	})

	logger.Info("Auth REST API service starting", zap.String("addr", addr))

	// 优雅关闭
	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm

		// 记录平台停止事件
		audit.LogPlatformEvent("platform_stop", "auth-service", nil)

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
	viper.BindEnv("redis.host", "AUTH_REDIS_HOST")
	viper.BindEnv("redis.port", "AUTH_REDIS_PORT")
	viper.BindEnv("redis.password", "AUTH_REDIS_PASSWORD")
	viper.BindEnv("redis.db", "AUTH_REDIS_DB")
	viper.BindEnv("grpc.port", "AUTH_GRPC_PORT")

	// 必须在 SetDefault 之前调用 AutomaticEnv
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTH")

	// 设置默认值
	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5001")
	viper.SetDefault("grpc.port", "50001")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration", "24h")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5433)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "itcmdb")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("redis.host", "redis")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "itcmdb_redis_pass_2026")
	viper.SetDefault("redis.db", 0)

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
			// 注册接口仅管理员可用
			auth.POST("/register", jwtManager.AuthMiddleware(), middleware.RequirePermission(userService, "user", "create"), registerHandler(userService, jwtManager))
			auth.POST("/login", loginHandler(userService, jwtManager))
			auth.POST("/logout", jwtManager.AuthMiddleware(), logoutHandler(jwtManager))
			auth.POST("/refresh", refreshHandler(jwtManager))
		}

		// 用户相关
		users := api.Group("/users")
		users.Use(jwtManager.AuthMiddleware())
		{
			// 查看自己的信息（所有认证用户）
			users.GET("/me", getMeHandler(userService))
			users.PUT("/me", updateMeHandler(userService))
			users.GET("/me/permissions", getMyPermissionsHandler(userService))

			// 查看用户列表（需要 user:view 权限）
			users.GET("", middleware.RequirePermission(userService, "user", "view"), getAllUsersHandler(userService))

			// 查看其他用户权限（需要 user:view 权限）
			users.GET("/:id/permissions", middleware.RequirePermission(userService, "user", "view"), getPermissionsHandler(userService))

			// 更新用户（需要 user:update 权限）
			users.PUT("/:id", middleware.RequirePermission(userService, "user", "update"), updateUserHandler(userService))

			// 删除用户（需要 user:delete 权限）
			users.DELETE("/:id", middleware.RequirePermission(userService, "user", "delete"), deleteUserHandler(userService))
		}

		// 角色管理
		roles := api.Group("/roles")
		roles.Use(jwtManager.AuthMiddleware())
		{
			// 查看角色列表（需要 role:view 权限）
			roles.GET("", middleware.RequirePermission(userService, "role", "view"), roleHandler.GetRoles)

			// 创建角色（需要 role:create 权限）
			roles.POST("", middleware.RequirePermission(userService, "role", "create"), roleHandler.CreateRole)

			// 更新角色（需要 role:update 权限）
			roles.PUT("/:id", middleware.RequirePermission(userService, "role", "update"), roleHandler.UpdateRole)

			// 删除角色（需要 role:delete 权限）
			roles.DELETE("/:id", middleware.RequirePermission(userService, "role", "delete"), roleHandler.DeleteRole)

			// 查看角色权限（需要 role:view 权限）
			roles.GET("/:id/permissions", middleware.RequirePermission(userService, "role", "view"), roleHandler.GetRolePermissions)

			// 查看角色用户（需要 role:view 权限）
			roles.GET("/:id/users", middleware.RequirePermission(userService, "role", "view"), roleHandler.GetRoleUsers)
		}

		// 权限管理
		permissions := api.Group("/permissions")
		permissions.Use(jwtManager.AuthMiddleware())
		{
			// 查看权限列表（需要 permission:view 权限）
			permissions.GET("", middleware.RequirePermission(userService, "permission", "view"), roleHandler.GetPermissions)

			// 获取有效的资源类型（所有认证用户可访问）
			permissions.GET("/resources", roleHandler.GetValidResources)

			// 获取有效的操作类型（所有认证用户可访问）
			permissions.GET("/actions", roleHandler.GetValidActions)

			// 创建权限（需要 permission:create 权限）
			permissions.POST("", middleware.RequirePermission(userService, "permission", "create"), roleHandler.CreatePermission)

			// 删除权限（需要 permission:delete 权限）
			permissions.DELETE("/:id", middleware.RequirePermission(userService, "permission", "delete"), roleHandler.DeletePermission)
		}

		// 角色权限关联
		rolePermissions := api.Group("/role-permissions")
		rolePermissions.Use(jwtManager.AuthMiddleware())
		{
			// 分配权限给角色（需要 role:manage 权限）
			rolePermissions.POST("", middleware.RequirePermission(userService, "role", "manage"), roleHandler.AssignPermissionToRole)

			// 移除角色权限（需要 role:manage 权限）
			rolePermissions.DELETE("", middleware.RequirePermission(userService, "role", "manage"), roleHandler.RemovePermissionFromRole)
		}

		// 用户角色关联
		userRoles := api.Group("/user-roles")
		userRoles.Use(jwtManager.AuthMiddleware())
		{
			// 分配角色给用户（需要 user:manage 权限）
			userRoles.POST("", middleware.RequirePermission(userService, "user", "manage"), roleHandler.AssignRoleToUser)

			// 移除用户角色（需要 user:manage 权限）
			userRoles.DELETE("", middleware.RequirePermission(userService, "user", "manage"), roleHandler.RemoveRoleFromUser)

			// 查看用户角色（需要 user:view 权限）
			userRoles.GET("/user/:id", middleware.RequirePermission(userService, "user", "view"), roleHandler.GetUserRoles)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		health := gin.H{"status": "ok", "service": "auth-service"}

		// 检查数据库连接
		db := database.Get()
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			health["status"] = "degraded"
			health["database"] = "unavailable"
		} else {
			health["database"] = "ok"
		}

		// 检查Redis连接
		redisClient := cache.Get()
		if redisClient != nil {
			if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
				health["status"] = "degraded"
				health["redis"] = "unavailable"
			} else {
				health["redis"] = "ok"
			}
		} else {
			health["redis"] = "not_configured"
		}

		status := 200
		if health["status"] == "degraded" {
			status = 503
		}
		c.JSON(status, health)
	})
}

// Handler functions

// registerHandler 用户注册（仅管理员可用）
func registerHandler(userService service.UserService, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
			FullName string `json:"full_name"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, response.Error("Bad Request", "invalid request: "+err.Error()))
			return
		}

		user, err := userService.Register(req.Username, req.Email, req.Password, req.FullName)
		if err != nil {
			c.JSON(400, response.Error("Bad Request", err.Error()))
			return
		}

		audit.LogSuccess(c, "create", "user", &user.ID, map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
		})

		c.JSON(200, response.Success(gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"fullName": user.FullName,
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

		// 检查登录锁定
		lockKey := "login_lock:" + req.Username
		redisClient := cache.Get()
		if redisClient != nil {
			ctx := c.Request.Context()
			if locked, _ := redisClient.Get(ctx, lockKey).Bool(); locked {
				audit.LogError(c, "login", "user", nil, "account locked due to too many failed attempts", gin.H{
					"username": req.Username,
				})
				c.JSON(423, response.Error("Locked", "account locked due to too many failed login attempts, please try again later"))
				return
			}
		}

		// 验证用户
		user, err := userService.ValidateUser(req.Username, req.Password)
		if err != nil {
			// 记录登录失败次数
			failKey := "login_fail:" + req.Username
			if redisClient != nil {
				ctx := c.Request.Context()
				count, _ := redisClient.Incr(ctx, failKey).Result()
				redisClient.Expire(ctx, failKey, 15*time.Minute)

				// 5次失败后锁定15分钟
				if count >= 5 {
					redisClient.Set(ctx, lockKey, true, 15*time.Minute)
					logger.Warn("Account locked due to too many failed login attempts",
						zap.String("username", req.Username),
						zap.Int64("attempts", count))
				}
			}

			// 记录登录失败的审计日志
			audit.LogError(c, "login", "user", nil, "invalid username or password", gin.H{
				"username": req.Username,
			})
			c.JSON(401, response.Error("Unauthorized", "invalid username or password"))
			return
		}

		// 登录成功，清除失败计数
		if redisClient != nil {
			ctx := c.Request.Context()
			redisClient.Del(ctx, "login_fail:"+req.Username)
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

		// 记录登录成功的审计日志
		userIDUint := user.ID
		audit.LogSuccess(c, "login", "user", &userIDUint, gin.H{
			"username": user.Username,
			"email":    user.Email,
		})

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

func logoutHandler(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息（如果已认证）
		userID, exists := auth.GetUserID(c)
		username, _ := auth.GetUsername(c)

		// 提取token并加入黑名单
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// 解析token获取claims（用于计算TTL）
			claims, err := jwtManager.Verify(tokenString)
			if err == nil {
				// 将token加入黑名单
				if err := jwtManager.AddToBlacklist(c.Request.Context(), tokenString, claims); err != nil {
					logger.Warn("Failed to add token to blacklist", zap.Error(err))
					// 继续执行，不影响用户体验
				}
			}
		}

		// 记录登出审计日志
		if exists {
			userIDUint := uint(userID)
			audit.LogSuccess(c, "logout", "user", &userIDUint, gin.H{
				"username": username,
			})
		}

		c.JSON(200, response.Success(nil))
	}
}

func refreshHandler(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, response.Error("Unauthorized", "missing or invalid authorization header"))
			return
		}
		oldToken := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证旧token
		claims, err := jwtManager.Verify(oldToken)
		if err != nil {
			c.JSON(401, response.Error("Unauthorized", err.Error()))
			return
		}

		// 检查是否在刷新窗口内（剩余时间 < 1小时）
		if time.Until(claims.ExpiresAt.Time) > time.Hour {
			c.JSON(400, response.Error("Bad Request", "token not ready for refresh"))
			return
		}

		// 将旧token加入黑名单
		if err := jwtManager.AddToBlacklist(c.Request.Context(), oldToken, claims); err != nil {
			logger.Warn("Failed to blacklist old token during refresh", zap.Error(err))
		}

		// 获取最新权限
		var permissions []string
		userID := uint(claims.UserID)
		db := database.Get()
		if db != nil {
			var rolePerms []string
			db.Table("permissions").
				Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
				Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
				Where("user_roles.user_id = ?", userID).
				Pluck("CONCAT(resource, ':', action)", &rolePerms)
			var userPerms []string
			db.Table("permissions").
				Joins("JOIN user_permissions ON permissions.id = user_permissions.permission_id").
				Where("user_permissions.user_id = ?", userID).
				Pluck("CONCAT(resource, ':', action)", &userPerms)
			permMap := make(map[string]bool)
			for _, p := range rolePerms {
				permMap[p] = true
			}
			for _, p := range userPerms {
				permMap[p] = true
			}
			for p := range permMap {
				permissions = append(permissions, p)
			}
		}

		// 生成新token
		newToken, err := jwtManager.Generate(claims.UserID, claims.Username, permissions)
		if err != nil {
			c.JSON(500, response.Error("Internal Error", "failed to generate token"))
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

// getAllUsersHandler 获取所有用户列表
func getAllUsersHandler(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := userService.GetAllUsers()
		if err != nil {
			c.JSON(500, response.Error("Internal Error", "failed to get users"))
			return
		}

		// 过滤敏感信息
		var userList []gin.H
		for _, user := range users {
			userList = append(userList, gin.H{
				"id":        user.ID,
				"username":  user.Username,
				"email":     user.Email,
				"full_name": user.FullName,
				"status":    user.Status,
			})
		}

		c.JSON(200, response.Success(userList))
	}
}

// updateUserHandler 更新用户信息
func updateUserHandler(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(400, response.Error("Bad Request", "invalid user id"))
			return
		}

		var req struct {
			FullName string `json:"full_name"`
			Status   string `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, response.Error("Bad Request", "invalid request"))
			return
		}

		updates := make(map[string]interface{})
		if req.FullName != "" {
			updates["full_name"] = req.FullName
		}
		if req.Status != "" {
			if req.Status != "active" && req.Status != "inactive" {
				c.JSON(400, response.Error("Bad Request", "invalid status"))
				return
			}
			updates["status"] = req.Status
		}

		if err := userService.UpdateUser(uint(id), updates); err != nil {
			c.JSON(400, response.Error("Bad Request", err.Error()))
			audit.LogError(c, "update", "user", nil, err.Error(), updates)
			return
		}

		userID := uint(id)
		audit.LogSuccess(c, "update", "user", &userID, map[string]interface{}{
			"updates": updates,
		})

		c.JSON(200, response.Success(nil))
	}
}

// deleteUserHandler 删除用户
func deleteUserHandler(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(400, response.Error("Bad Request", "invalid user id"))
			return
		}

		if err := userService.DeleteUser(uint(id)); err != nil {
			c.JSON(500, response.Error("Internal Error", "failed to delete user"))
			audit.LogError(c, "delete", "user", nil, err.Error(), nil)
			return
		}

		userID := uint(id)
		audit.LogSuccess(c, "delete", "user", &userID, nil)

		c.JSON(200, response.Success(nil))
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

func startGRPCServer(authService service.AuthService, userService service.UserService) {
	grpcPort := viper.GetString("grpc.port")
	if grpcPort == "" {
		grpcPort = "50001"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	authServer := grpcserver.NewAuthServer(authService, userService)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	// 注册反射服务，用于grpcurl等工具
	reflection.Register(grpcServer)

	logger.Info("Auth gRPC service starting", zap.String("port", grpcPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}
