package main

import (
	"fmt"
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/itcmdb/cmdb-service/internal/handlers"
	grpcserver "github.com/itcmdb/cmdb-service/internal/grpc"
	"github.com/itcmdb/cmdb-service/internal/repository"
	"github.com/itcmdb/cmdb-service/internal/service"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/audit"
	"github.com/itcmdb/shared/pkg/cache"
	"github.com/itcmdb/shared/pkg/database"
	kafkapkg "github.com/itcmdb/shared/pkg/kafka"
	"github.com/itcmdb/shared/pkg/logger"
	pb "github.com/itcmdb/shared/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	// 初始化依赖
	jwtManager := auth.NewJWTManager(
		viper.GetString("jwt.secret"),
		viper.GetDuration("jwt.expiration"),
	)

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

	db := database.Get()
	ciRepo := repository.NewCIRepository(db)
	ciService := service.NewCIService(ciRepo)
	ciHandler := handlers.NewCIHandler(ciService)

	// 角色和标签服务
	roleRepo := repository.NewRoleRepository(db)
	tagRepo := repository.NewTagRepository(db)
	roleService := service.NewRoleService(roleRepo, ciRepo)
	tagService := service.NewTagService(tagRepo, ciRepo)
	roleHandler := handlers.NewRoleHandler(roleService, jwtManager)
	tagHandler := handlers.NewTagHandler(tagService, jwtManager)

	if viper.GetString("env") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 启动gRPC服务器
	go startGRPCServer(ciService)

	// 启动REST API服务器
	r := gin.Default()
	setupRoutes(r, jwtManager, ciHandler, roleHandler, tagHandler)

	addr := fmt.Sprintf(":%s", viper.GetString("server.port"))
	logger.Info("CMDB REST API service starting", zap.String("addr", addr))
	r.Run(addr)
}

func startGRPCServer(ciService service.CIService) {
	grpcPort := viper.GetString("grpc.port")
	if grpcPort == "" {
		grpcPort = "50002"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	cmdbServer := grpcserver.NewCMDBServer(ciService)
	pb.RegisterCMDBServiceServer(grpcServer, cmdbServer)

	// 注册反射服务，用于grpcurl等工具
	reflection.Register(grpcServer)

	logger.Info("CMDB gRPC service starting", zap.String("port", grpcPort))
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

	viper.AutomaticEnv()
	viper.SetEnvPrefix("CMDB")

	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", "5002")
	viper.SetDefault("grpc.port", "50002")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration", "24h")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "itcmdb")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("redis.host", "redis")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "itcmdb_redis_pass_2026")
	viper.SetDefault("redis.db", 0)

	viper.ReadInConfig()
	return nil
}

func setupRoutes(r *gin.Engine, jwtManager *auth.JWTManager, ciHandler *handlers.CIHandler, roleHandler *handlers.RoleHandler, tagHandler *handlers.TagHandler) {
	api := r.Group("/api/v1")
	api.Use(jwtManager.AuthMiddleware())
	{
		// CI管理
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
			ci.GET("/export", ciHandler.ExportCIInstances)
			ci.POST("/import", ciHandler.ImportCIInstances)
		}

		// 角色管理
		roles := api.Group("/roles")
		{
			// CI角色
			roles.GET("/ci", roleHandler.GetCIRoles)
			roles.POST("/ci", roleHandler.CreateCIRole)
			roles.PUT("/ci/:id", roleHandler.UpdateCIRole)
			roles.DELETE("/ci/:id", roleHandler.DeleteCIRole)

			// 负责人角色
			roles.GET("/owner", roleHandler.GetOwnerRoles)
			roles.POST("/owner", roleHandler.CreateOwnerRole)
			roles.PUT("/owner/:id", roleHandler.UpdateOwnerRole)
			roles.DELETE("/owner/:id", roleHandler.DeleteOwnerRole)
		}

		// CI实例角色关联
		api.GET("/ci/instances/:id/roles", roleHandler.GetCIInstanceRoles)
		api.POST("/ci/instances/:id/roles", roleHandler.AssignCIRole)
		api.DELETE("/ci/instances/:id/roles", roleHandler.RemoveCIRole)

		// 标签分类
		tags := api.Group("/tags")
		{
			tags.GET("/categories", tagHandler.GetTagCategories)
			tags.POST("/categories", tagHandler.CreateTagCategory)
			tags.PUT("/categories/:id", tagHandler.UpdateTagCategory)
			tags.DELETE("/categories/:id", tagHandler.DeleteTagCategory)

			tags.GET("", tagHandler.GetTags)
			tags.POST("", tagHandler.CreateTag)
			tags.GET("/stats", tagHandler.GetTagStats)
			tags.PUT("/:id", tagHandler.UpdateTag)
			tags.DELETE("/:id", tagHandler.DeleteTag)
		}

		// CI实例标签操作
		api.GET("/ci/instances/:id/tags", tagHandler.GetCITags)
		api.POST("/ci/instances/:id/tags", tagHandler.AssignTag)
		api.DELETE("/ci/instances/:id/tags/:tagId", tagHandler.RemoveTag)

		// 批量操作
		api.POST("/tags/batch/assign", tagHandler.BatchAssignTags)
		api.DELETE("/tags/batch/remove", tagHandler.BatchRemoveTags)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
