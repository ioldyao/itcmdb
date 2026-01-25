package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/itcmdb/cmdb-service/internal/cadvisor"
	"github.com/itcmdb/cmdb-service/internal/prometheus"
	"github.com/itcmdb/cmdb-service/internal/repository"
	"github.com/itcmdb/shared/pkg/cache"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

type MonitoringService interface {
	GetContainerStats(ctx context.Context, ciID uint) (*cadvisor.ContainerStats, error)
	HealthCheckCAdvisor(ctx context.Context, endpoint string) error
	HealthCheckVictoriaMetrics(ctx context.Context) error
	GetPrometheusClient() *prometheus.Client
}

type monitoringService struct {
	ciRepo           repository.CIRepository
	prometheusClient *prometheus.Client
}

func NewMonitoringService(ciRepo repository.CIRepository, vmEndpoint, vmUsername, vmPassword string) MonitoringService {
	var promClient *prometheus.Client
	if vmEndpoint != "" {
		promClient = prometheus.NewClient(vmEndpoint, vmUsername, vmPassword)
		logger.Info("VictoriaMetrics client initialized", zap.String("endpoint", vmEndpoint))
	}

	return &monitoringService{
		ciRepo:           ciRepo,
		prometheusClient: promClient,
	}
}

// GetContainerStats 获取容器监控数据（带Redis缓存）
func (s *monitoringService) GetContainerStats(ctx context.Context, ciID uint) (*cadvisor.ContainerStats, error) {
	// 获取CI实例信息
	instance, err := s.ciRepo.GetCIInstanceByID(ciID)
	if err != nil {
		return nil, fmt.Errorf("CI instance not found: %w", err)
	}

	// 从attributes中提取容器名称或ID
	containerName, ok := instance.Attributes["container_name"].(string)
	if !ok || containerName == "" {
		// 兼容旧的container_id字段
		containerName, ok = instance.Attributes["container_id"].(string)
		if !ok || containerName == "" {
			return nil, fmt.Errorf("container_name or container_id not configured for CI instance %d", ciID)
		}
	}

	// 尝试从Redis缓存获取
	cacheKey := fmt.Sprintf("monitoring:container:%d", ciID)
	redisClient := cache.Get()
	if redisClient != nil {
		if cachedData, err := redisClient.Get(ctx, cacheKey).Result(); err == nil && cachedData != "" {
			var stats cadvisor.ContainerStats
			if err := json.Unmarshal([]byte(cachedData), &stats); err == nil {
				logger.Debug("Retrieved container stats from cache",
					zap.Uint("ci_id", ciID),
					zap.String("container_name", containerName))
				return &stats, nil
			}
		}
	}

	var stats *cadvisor.ContainerStats

	// 优先使用VictoriaMetrics
	if s.prometheusClient != nil {
		stats, err = s.prometheusClient.GetContainerStats(ctx, containerName)
		if err != nil {
			logger.Error("Failed to fetch container stats from VictoriaMetrics",
				zap.Error(err),
				zap.Uint("ci_id", ciID),
				zap.String("container_name", containerName))
			return nil, fmt.Errorf("failed to fetch container stats from VictoriaMetrics: %w", err)
		}
	} else {
		// 回退到cAdvisor（兼容旧配置）
		cadvisorEndpoint, ok := instance.Attributes["cadvisor_endpoint"].(string)
		if !ok || cadvisorEndpoint == "" {
			return nil, fmt.Errorf("neither VictoriaMetrics nor cadvisor_endpoint configured")
		}

		client := cadvisor.NewClient(cadvisorEndpoint)
		stats, err = client.GetContainerStats(ctx, containerName)
		if err != nil {
			logger.Error("Failed to fetch container stats from cAdvisor",
				zap.Error(err),
				zap.Uint("ci_id", ciID),
				zap.String("container_name", containerName),
				zap.String("endpoint", cadvisorEndpoint))
			return nil, fmt.Errorf("failed to fetch container stats from cAdvisor: %w", err)
		}
	}

	// 缓存结果（30秒TTL）
	if redisClient != nil {
		if data, err := json.Marshal(stats); err == nil {
			if err := redisClient.Set(ctx, cacheKey, string(data), 30*time.Second).Err(); err != nil {
				logger.Warn("Failed to cache container stats", zap.Error(err), zap.Uint("ci_id", ciID))
			}
		}
	}

	return stats, nil
}

// HealthCheckCAdvisor 检查cAdvisor服务健康状态
func (s *monitoringService) HealthCheckCAdvisor(ctx context.Context, endpoint string) error {
	client := cadvisor.NewClient(endpoint)
	return client.HealthCheck(ctx)
}

// HealthCheckVictoriaMetrics 检查VictoriaMetrics服务健康状态
func (s *monitoringService) HealthCheckVictoriaMetrics(ctx context.Context) error {
	if s.prometheusClient == nil {
		return fmt.Errorf("VictoriaMetrics client not configured")
	}
	return s.prometheusClient.HealthCheck(ctx)
}

// GetPrometheusClient 获取Prometheus客户端（用于容器同步服务）
func (s *monitoringService) GetPrometheusClient() *prometheus.Client {
	return s.prometheusClient
}
