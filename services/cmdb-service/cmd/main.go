package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/cmdb-service/internal/handlers"
	grpcserver "github.com/itcmdb/cmdb-service/internal/grpc"
	"github.com/itcmdb/cmdb-service/internal/models"
	"github.com/itcmdb/cmdb-service/internal/prometheus"
	"github.com/itcmdb/cmdb-service/internal/repository"
	"github.com/itcmdb/cmdb-service/internal/service"
	"github.com/itcmdb/shared/pkg/audit"
	"github.com/itcmdb/shared/pkg/cache"
	"github.com/itcmdb/shared/pkg/database"
	grpcclient "github.com/itcmdb/shared/pkg/grpc"
	kafkapkg "github.com/itcmdb/shared/pkg/kafka"
	"github.com/itcmdb/shared/pkg/logger"
	"github.com/itcmdb/shared/pkg/middleware"
	pb "github.com/itcmdb/shared/proto/cmdb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
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
	db := database.Get()
	if err := models.AutoMigrate(db); err != nil {
		logger.Warn("Failed to migrate database", zap.Error(err))
		// 不终止服务启动，允许服务在迁移失败时继续运行
	} else {
		logger.Info("Database migration completed successfully")
	}

	// 初始化Auth服务gRPC客户端
	authServiceAddr := viper.GetString("auth.grpc.address")
	if authServiceAddr == "" {
		authServiceAddr = "auth-service:50001"
	}
	authClient, err := grpcclient.NewAuthClient(authServiceAddr)
	if err != nil {
		logger.Fatal("Failed to connect to Auth service", zap.Error(err))
	}
	defer authClient.Close()
	logger.Info("Connected to Auth service via gRPC", zap.String("address", authServiceAddr))

	// 初始化审计日志Kafka生产者
	kafkaBrokers := []string{"kafka:9092"}
	if err := audit.InitProducer(kafkaBrokers); err != nil {
		logger.Warn("Failed to initialize audit producer, audit logging disabled", zap.Error(err))
	}

	// 初始化Kafka事件生产者
	if err := kafkapkg.InitEventProducer(kafkaBrokers); err != nil {
		logger.Warn("Failed to initialize event producer, event publishing disabled", zap.Error(err))
	} else {
		logger.Info("Kafka event producer initialized")
	}

	// 初始化服务和处理器
	ciRepo := repository.NewCIRepository(db)
	ciService := service.NewCIService(ciRepo)
	ciHandler := handlers.NewCIHandler(ciService)

	// 角色和标签服务
	roleRepo := repository.NewRoleRepository(db)
	tagRepo := repository.NewTagRepository(db)
	roleService := service.NewRoleService(roleRepo, ciRepo)
	tagService := service.NewTagService(tagRepo, ciRepo)
	roleHandler := handlers.NewRoleHandler(roleService)
	tagHandler := handlers.NewTagHandler(tagService)

	// 配置服务
	configRepo := repository.NewConfigRepository(db)
	encryptionKey := viper.GetString("config.encryption_key")
	configService := service.NewConfigService(configRepo, encryptionKey)
	configHandler := handlers.NewConfigHandler(configService)

	// 监控服务（支持多数据源配置）
	// 优先从数据库读取配置，否则使用配置文件
	vmDataSources := loadVictoriaMetricsDataSources(configService)

	var monitoringHandler *handlers.MonitoringHandler

	if len(vmDataSources) > 0 {
		// 使用多数据源配置
		logger.Info("Using multiple VictoriaMetrics datasources", zap.Int("count", len(vmDataSources)))

		// 转换为DataSource指针数组
		dataSourcePtrs := make([]*prometheus.DataSource, len(vmDataSources))
		for i := range vmDataSources {
			dataSourcePtrs[i] = &vmDataSources[i]
		}

		// 创建多数据源客户端
		multiClient := prometheus.NewMultiSourceClient(dataSourcePtrs)

		monitoringService := service.NewMonitoringServiceWithMultiSource(ciRepo, multiClient)
		monitoringHandler = handlers.NewMonitoringHandler(monitoringService)

		// 启动多数据源容器自动同步服务
		syncInterval := viper.GetDuration("victoriametrics.sync_interval")
		if syncInterval == 0 {
			syncInterval = 5 * time.Minute // 默认5分钟同步一次
		}

		containerSyncService := service.NewContainerSyncService(ciRepo, multiClient, syncInterval)
		containerSyncService.Start()
		logger.Info("Multi-source container auto-sync service started", zap.Duration("interval", syncInterval))
	} else {
		// 使用单数据源配置（向后兼容）
		vmEndpoint := viper.GetString("victoriametrics.endpoint")
		vmUsername := viper.GetString("victoriametrics.username")
		vmPassword := viper.GetString("victoriametrics.password")

		// 尝试从数据库读取 VictoriaMetrics 配置
		if dbEndpoint, err := configService.GetConfigValue("monitoring", "victoriametrics_endpoint"); err == nil && dbEndpoint != "" {
			vmEndpoint = dbEndpoint
			logger.Info("Using VictoriaMetrics endpoint from database", zap.String("endpoint", vmEndpoint))
		}
		if dbUsername, err := configService.GetConfigValue("monitoring", "victoriametrics_username"); err == nil && dbUsername != "" {
			vmUsername = dbUsername
		}
		if dbPassword, err := configService.GetConfigValue("monitoring", "victoriametrics_password"); err == nil && dbPassword != "" {
			vmPassword = dbPassword
		}

		monitoringService := service.NewMonitoringService(ciRepo, vmEndpoint, vmUsername, vmPassword)
		monitoringHandler = handlers.NewMonitoringHandler(monitoringService)

		// 启动单数据源容器自动同步服务（如果配置了VictoriaMetrics）
		if vmEndpoint != "" {
			syncInterval := viper.GetDuration("victoriametrics.sync_interval")
			if syncInterval == 0 {
				syncInterval = 5 * time.Minute // 默认5分钟同步一次
			}

			promClient := monitoringService.GetPrometheusClient()
			if promClient != nil {
				containerSyncService := service.NewContainerSyncService(ciRepo, promClient, syncInterval)
				containerSyncService.Start()
				logger.Info("Container auto-sync service started", zap.Duration("interval", syncInterval))
			}
		}
	}

	if viper.GetString("env") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 启动gRPC服务器（需要authClient用于认证）
	go startGRPCServer(ciService, authClient)

	// 启动REST API服务器
	r := gin.Default()
	setupRoutes(r, authClient, ciHandler, roleHandler, tagHandler, monitoringHandler, configHandler)

	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))

	// 记录平台启动事件
	audit.LogPlatformEvent("platform_start", "cmdb-service", map[string]interface{}{
		"addr": addr,
	})

	logger.Info("CMDB REST API service starting", zap.String("addr", addr))

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
	audit.LogPlatformEvent("platform_stop", "cmdb-service", nil)

	logger.Info("Shutting down cmdb service...")
}

func startGRPCServer(ciService service.CIService, authClient *grpcclient.AuthClient) {
	grpcPort := viper.GetString("grpc.port")
	if grpcPort == "" {
		grpcPort = "50002"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	// 检查是否启用mTLS
	enableMTLS := viper.GetBool("grpc.mtls.enabled")
	serverCert := viper.GetString("grpc.mtls.server_cert")
	serverKey := viper.GetString("grpc.mtls.server_key")
	caCert := viper.GetString("grpc.mtls.ca_cert")

	var creds credentials.TransportCredentials

	if enableMTLS {
		// 加载服务端证书
		cert, err := tls.LoadX509KeyPair(serverCert, serverKey)
		if err != nil {
			logger.Fatal("Failed to load server certificates",
				zap.Error(err),
				zap.String("cert", serverCert),
				zap.String("key", serverKey))
		}

		// 加载CA证书用于验证客户端证书
		caCertPool := x509.NewCertPool()
		caCertBytes, err := os.ReadFile(caCert)
		if err != nil {
			logger.Fatal("Failed to read CA certificate",
				zap.Error(err),
				zap.String("ca_cert", caCert))
		}
		if !caCertPool.AppendCertsFromPEM(caCertBytes) {
			logger.Fatal("Failed to append CA certificate to pool")
		}

		// 创建TLS配置（mTLS）
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert, // 强制要求客户端证书
			ClientCAs:    caCertPool,
			MinVersion:   tls.VersionTLS12,
		}

		creds = credentials.NewTLS(tlsConfig)

		logger.Info("mTLS enabled",
			zap.String("server_cert", serverCert),
			zap.String("ca_cert", caCert))
	} else {
		logger.Warn("mTLS disabled, using insecure credentials (not recommended for production)")
		// 开发环境可以使用不安全连接
		creds = insecure.NewCredentials()
	}

	// 创建gRPC服务器，添加拦截器
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			middleware.LoggingInterceptor(),
		),
	)

	cmdbServer := grpcserver.NewCMDBServer(ciService)
	pb.RegisterCMDBServiceServer(grpcServer, cmdbServer)
	pb.RegisterHardwareServiceServer(grpcServer, cmdbServer)

	// 注册反射服务，用于grpcurl等工具
	reflection.Register(grpcServer)

	if enableMTLS {
		logger.Info("CMDB gRPC service starting with mTLS",
			zap.String("port", grpcPort),
			zap.String("auth", "mtls"))
	} else {
		logger.Info("CMDB gRPC service starting without mTLS",
			zap.String("port", grpcPort),
			zap.String("auth", "none"))
	}

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", zap.Error(err))
	}
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
	viper.BindEnv("jwt.secret", "CMDB_JWT_SECRET")
	viper.BindEnv("jwt.expiration", "CMDB_JWT_EXPIRATION")
	viper.BindEnv("redis.host", "CMDB_REDIS_HOST")
	viper.BindEnv("redis.port", "CMDB_REDIS_PORT")
	viper.BindEnv("redis.password", "CMDB_REDIS_PASSWORD")
	viper.BindEnv("redis.db", "CMDB_REDIS_DB")
	viper.BindEnv("grpc.port", "CMDB_GRPC_PORT")
	viper.BindEnv("grpc.mtls.enabled", "CMDB_GRPC_MTLS_ENABLED")
	viper.BindEnv("grpc.mtls.server_cert", "CMDB_GRPC_MTLS_SERVER_CERT")
	viper.BindEnv("grpc.mtls.server_key", "CMDB_GRPC_MTLS_SERVER_KEY")
	viper.BindEnv("grpc.mtls.ca_cert", "CMDB_GRPC_MTLS_CA_CERT")
	viper.BindEnv("auth.grpc.address", "CMDB_AUTH_GRPC_ADDRESS")
	viper.BindEnv("victoriametrics.endpoint", "CMDB_VICTORIAMETRICS_ENDPOINT")
	viper.BindEnv("victoriametrics.username", "CMDB_VICTORIAMETRICS_USERNAME")
	viper.BindEnv("victoriametrics.password", "CMDB_VICTORIAMETRICS_PASSWORD")
	viper.BindEnv("victoriametrics.sync_interval", "CMDB_VICTORIAMETRICS_SYNC_INTERVAL")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("CMDB")

	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5002")
	viper.SetDefault("grpc.port", "50002")
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
	viper.SetDefault("grpc.mtls.enabled", false)
	viper.SetDefault("grpc.mtls.server_cert", "./certificates/server_cert.pem")
	viper.SetDefault("grpc.mtls.server_key", "./certificates/server_key.pem")
	viper.SetDefault("grpc.mtls.ca_cert", "./certificates/ca_cert.pem")
	viper.SetDefault("auth.grpc.address", "auth-service:50001")
	viper.SetDefault("victoriametrics.endpoint", "")
	viper.SetDefault("victoriametrics.username", "")
	viper.SetDefault("victoriametrics.password", "")
	viper.SetDefault("victoriametrics.sync_interval", "5m")

	viper.ReadInConfig()
	return nil
}

func setupRoutes(r *gin.Engine, authClient *grpcclient.AuthClient, ciHandler *handlers.CIHandler, roleHandler *handlers.RoleHandler, tagHandler *handlers.TagHandler, monitoringHandler *handlers.MonitoringHandler, configHandler *handlers.ConfigHandler) {
	api := r.Group("/api/v1")
	api.Use(middleware.GRPCAuthMiddleware(authClient))
	{
		// CI管理
		ci := api.Group("/ci")
		{
			// 查看CI类型（需要 ci:view 权限）
			ci.GET("/types", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), ciHandler.GetCITypes)

			// 创建CI实例（需要 ci:create 权限）
			ci.POST("/instances", middleware.GRPCPermissionMiddleware(authClient, "ci", "create"), ciHandler.CreateCIInstance)

			// 查看CI实例列表（需要 ci:view 权限）
			ci.GET("/instances", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), ciHandler.GetCIInstances)

			// 查看单个CI实例（需要 ci:view 权限）
			ci.GET("/instances/:id", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), ciHandler.GetCIInstance)

			// 更新CI实例（需要 ci:update 权限）
			ci.PUT("/instances/:id", middleware.GRPCPermissionMiddleware(authClient, "ci", "update"), ciHandler.UpdateCIInstance)

			// 删除CI实例（需要 ci:delete 权限）
			ci.DELETE("/instances/:id", middleware.GRPCPermissionMiddleware(authClient, "ci", "delete"), ciHandler.DeleteCIInstance)

			// 查看CI历史（需要 ci:view 权限）
			ci.GET("/instances/:id/history", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), ciHandler.GetCIHistory)

			// 查看CI关系（需要 ci:view 权限）
			ci.GET("/relations", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), ciHandler.GetCIRelations)

			// 创建CI关系（需要 ci:create 权限）
			ci.POST("/relations", middleware.GRPCPermissionMiddleware(authClient, "ci", "create"), ciHandler.CreateCIRelation)

			// 导出CI实例（需要 ci:view 权限）
			ci.GET("/export", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), ciHandler.ExportCIInstances)

			// 导入CI实例（需要 ci:create 权限）
			ci.POST("/import", middleware.GRPCPermissionMiddleware(authClient, "ci", "create"), ciHandler.ImportCIInstances)
		}

		// 角色管理（CI角色和负责人角色）
		roles := api.Group("/roles")
		{
			// CI角色（需要 ci:view 权限查看，ci:create/update/delete 权限修改）
			roles.GET("/ci", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), roleHandler.GetCIRoles)
			roles.POST("/ci", middleware.GRPCPermissionMiddleware(authClient, "ci", "create"), roleHandler.CreateCIRole)
			roles.PUT("/ci/:id", middleware.GRPCPermissionMiddleware(authClient, "ci", "update"), roleHandler.UpdateCIRole)
			roles.DELETE("/ci/:id", middleware.GRPCPermissionMiddleware(authClient, "ci", "delete"), roleHandler.DeleteCIRole)

			// 负责人角色（需要 ci:view 权限查看，ci:create/update/delete 权限修改）
			roles.GET("/owner", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), roleHandler.GetOwnerRoles)
			roles.POST("/owner", middleware.GRPCPermissionMiddleware(authClient, "ci", "create"), roleHandler.CreateOwnerRole)
			roles.PUT("/owner/:id", middleware.GRPCPermissionMiddleware(authClient, "ci", "update"), roleHandler.UpdateOwnerRole)
			roles.DELETE("/owner/:id", middleware.GRPCPermissionMiddleware(authClient, "ci", "delete"), roleHandler.DeleteOwnerRole)
		}

		// CI实例角色关联（需要 ci:view 权限查看，ci:update 权限修改）
		api.GET("/ci/instances/:id/roles", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), roleHandler.GetCIInstanceRoles)
		api.POST("/ci/instances/:id/roles", middleware.GRPCPermissionMiddleware(authClient, "ci", "update"), roleHandler.AssignCIRole)
		api.DELETE("/ci/instances/:id/roles", middleware.GRPCPermissionMiddleware(authClient, "ci", "update"), roleHandler.RemoveCIRole)

		// 标签分类（需要 tag:view 权限查看，tag:create/update/delete 权限修改）
		tags := api.Group("/tags")
		{
			tags.GET("/categories", middleware.GRPCPermissionMiddleware(authClient, "tag", "view"), tagHandler.GetTagCategories)
			tags.POST("/categories", middleware.GRPCPermissionMiddleware(authClient, "tag", "create"), tagHandler.CreateTagCategory)
			tags.PUT("/categories/:id", middleware.GRPCPermissionMiddleware(authClient, "tag", "update"), tagHandler.UpdateTagCategory)
			tags.DELETE("/categories/:id", middleware.GRPCPermissionMiddleware(authClient, "tag", "delete"), tagHandler.DeleteTagCategory)

			tags.GET("", middleware.GRPCPermissionMiddleware(authClient, "tag", "view"), tagHandler.GetTags)
			tags.POST("", middleware.GRPCPermissionMiddleware(authClient, "tag", "create"), tagHandler.CreateTag)
			tags.GET("/stats", middleware.GRPCPermissionMiddleware(authClient, "tag", "view"), tagHandler.GetTagStats)
			tags.PUT("/:id", middleware.GRPCPermissionMiddleware(authClient, "tag", "update"), tagHandler.UpdateTag)
			tags.DELETE("/:id", middleware.GRPCPermissionMiddleware(authClient, "tag", "delete"), tagHandler.DeleteTag)
		}

		// CI实例标签操作（需要 tag:view 权限查看，tag:update 权限修改）
		api.GET("/ci/instances/:id/tags", middleware.GRPCPermissionMiddleware(authClient, "tag", "view"), tagHandler.GetCITags)
		api.POST("/ci/instances/:id/tags", middleware.GRPCPermissionMiddleware(authClient, "tag", "update"), tagHandler.AssignTag)
		api.DELETE("/ci/instances/:id/tags/:tagId", middleware.GRPCPermissionMiddleware(authClient, "tag", "update"), tagHandler.RemoveTag)

		// 批量操作（需要 tag:update 权限）
		api.POST("/tags/batch/assign", middleware.GRPCPermissionMiddleware(authClient, "tag", "update"), tagHandler.BatchAssignTags)
		api.DELETE("/tags/batch/remove", middleware.GRPCPermissionMiddleware(authClient, "tag", "update"), tagHandler.BatchRemoveTags)

		// 监控管理（需要 ci:view 权限）
		monitoring := api.Group("/monitoring")
		{
			monitoring.GET("/containers/:id/stats", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), monitoringHandler.GetContainerStats)
			monitoring.GET("/cadvisor/health", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), monitoringHandler.HealthCheckCAdvisor)
			monitoring.GET("/victoriametrics/health", middleware.GRPCPermissionMiddleware(authClient, "ci", "view"), monitoringHandler.HealthCheckVictoriaMetrics)
		}

		// 系统配置管理（仅管理员可访问）
		configs := api.Group("/configs")
		configs.Use(middleware.GRPCAdminOnlyMiddleware(authClient))
		{
			configs.GET("", configHandler.GetAllConfigs)
			configs.GET("/category/:category", configHandler.GetConfigsByCategory)
			configs.POST("", configHandler.CreateConfig)
			configs.PUT("/:id", configHandler.UpdateConfig)
			configs.DELETE("/:id", configHandler.DeleteConfig)
			configs.POST("/batch", configHandler.BatchUpdateConfigs)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

// loadVictoriaMetricsDataSources 从数据库加载VictoriaMetrics多数据源配置
func loadVictoriaMetricsDataSources(configService service.ConfigService) []prometheus.DataSource {
	// 尝试从数据库读取多数据源配置
	datasourcesJSON, err := configService.GetConfigValue("monitoring", "victoriametrics_datasources")
	if err != nil || datasourcesJSON == "" {
		logger.Debug("No VictoriaMetrics datasources configured in database")
		return []prometheus.DataSource{}
	}

	// 解析JSON配置
	var datasources []prometheus.DataSource
	if err := json.Unmarshal([]byte(datasourcesJSON), &datasources); err != nil {
		logger.Error("Failed to parse VictoriaMetrics datasources config", zap.Error(err))
		return []prometheus.DataSource{}
	}

	// 只返回启用的数据源
	enabledDatasources := make([]prometheus.DataSource, 0, len(datasources))
	for i := range datasources {
		if datasources[i].Enabled {
			enabledDatasources = append(enabledDatasources, datasources[i])
		}
	}

	logger.Info("Loaded VictoriaMetrics datasources from database",
		zap.Int("total", len(datasources)),
		zap.Int("enabled", len(enabledDatasources)))

	return enabledDatasources
}
