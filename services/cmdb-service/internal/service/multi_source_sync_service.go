package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/itcmdb/cmdb-service/internal/models"
	"github.com/itcmdb/cmdb-service/internal/prometheus"
	"github.com/itcmdb/cmdb-service/internal/repository"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

// MultiSourceContainerSyncService 多数据源容器同步服务
type MultiSourceContainerSyncService struct {
	ciRepo       repository.CIRepository
	multiClient  *prometheus.MultiDataSourceClient
	syncInterval time.Duration
	stopChan     chan struct{}
}

// NewMultiSourceContainerSyncService 创建多数据源容器同步服务
func NewMultiSourceContainerSyncService(
	ciRepo repository.CIRepository,
	multiClient *prometheus.MultiDataSourceClient,
	syncInterval time.Duration,
) *MultiSourceContainerSyncService {
	return &MultiSourceContainerSyncService{
		ciRepo:       ciRepo,
		multiClient:  multiClient,
		syncInterval: syncInterval,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动同步服务
func (s *MultiSourceContainerSyncService) Start() {
	if s.multiClient == nil || len(s.multiClient.GetDatasourceInfo()) == 0 {
		logger.Warn("Multi-source client not configured or no datasources available, container sync disabled")
		return
	}

	logger.Info("Starting multi-source container sync service",
		zap.Duration("interval", s.syncInterval),
		zap.Int("datasources", len(s.multiClient.GetDatasourceInfo())))

	// 立即执行一次同步
	go func() {
		if err := s.syncContainers(); err != nil {
			logger.Error("Initial container sync failed", zap.Error(err))
		}
	}()

	// 启动定时同步
	go func() {
		ticker := time.NewTicker(s.syncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.syncContainers(); err != nil {
					logger.Error("Container sync failed", zap.Error(err))
				}
			case <-s.stopChan:
				logger.Info("Multi-source container sync service stopped")
				return
			}
		}
	}()

	// 启动健康检查
	go s.healthCheckLoop()
}

// Stop 停止同步服务
func (s *MultiSourceContainerSyncService) Stop() {
	close(s.stopChan)
}

// syncContainers 同步容器信息
func (s *MultiSourceContainerSyncService) syncContainers() error {
	ctx := context.Background()
	logger.Info("Starting multi-source container synchronization")

	// 1. 从所有数据源发现容器
	containers, err := s.multiClient.DiscoverContainersFromAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover containers: %w", err)
	}

	logger.Info("Discovered containers from all datasources",
		zap.Int("total", len(containers)))

	// 按数据源分组统计
	dsStats := make(map[string]int)
	for _, container := range containers {
		dsStats[container.DataSourceID]++
	}
	for dsID, count := range dsStats {
		logger.Info("Containers from datasource",
			zap.String("datasource_id", dsID),
			zap.Int("count", count))
	}

	// 2. 确保容器 CI 类型存在
	containerType, err := s.ensureContainerCIType()
	if err != nil {
		return fmt.Errorf("failed to ensure container CI type: %w", err)
	}

	// 3. 获取所有现有的容器 CI 实例
	existingContainers, err := s.getExistingContainerCIs(containerType.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing containers: %w", err)
	}

	// 4. 同步容器
	stats := s.performSync(ctx, containers, existingContainers, containerType.ID)

	logger.Info("Multi-source container synchronization completed",
		zap.Int("created", stats.Created),
		zap.Int("updated", stats.Updated),
		zap.Int("marked_offline", stats.MarkedOffline),
		zap.Int("marked_online", stats.MarkedOnline),
		zap.Int("rebuilt", stats.Rebuilt))

	return nil
}

// performSync 执行同步
func (s *MultiSourceContainerSyncService) performSync(
	ctx context.Context,
	discoveredContainers []prometheus.DataSourceContainerInfo,
	existingContainers map[string]*models.CIInstance,
	ciTypeID uint,
) SyncStats {
	stats := SyncStats{}

	// 创建容器名称到发现容器的映射
	discoveredMap := make(map[string]prometheus.DataSourceContainerInfo)
	for _, container := range discoveredContainers {
		discoveredMap[container.Name] = container
	}

	// 处理发现的容器
	for _, container := range discoveredContainers {
		existing, exists := existingContainers[container.Name]

		if !exists {
			// 创建新的 CI 实例
			if err := s.createContainerCI(ctx, container, ciTypeID); err != nil {
				logger.Error("Failed to create container CI",
					zap.String("container", container.Name),
					zap.String("datasource", container.DataSourceName),
					zap.Error(err))
			} else {
				stats.Created++
			}
		} else {
			// 检查数据源是否变更
			if existingDS, ok := existing.Attributes["datasource_id"].(string); ok && existingDS != container.DataSourceID {
				logger.Info("Container datasource changed",
					zap.String("container", container.Name),
					zap.String("old_datasource", existingDS),
					zap.String("new_datasource", container.DataSourceID))
			}

			// 更新现有的 CI 实例
			updated, rebuilt := s.updateContainerCI(ctx, existing, container)
			if updated {
				stats.Updated++
			}
			if rebuilt {
				stats.Rebuilt++
			}
			if !existing.Attributes["is_online"].(bool) && container.IsRunning {
				stats.MarkedOnline++
			}
		}
	}

	// 标记未发现的容器为离线
	for name, existing := range existingContainers {
		if _, found := discoveredMap[name]; !found {
			// 容器未被发现，检查是否需要标记为离线
			if isOnline, ok := existing.Attributes["is_online"].(bool); ok && isOnline {
				if err := s.markContainerOffline(ctx, existing); err != nil {
					logger.Error("Failed to mark container offline",
						zap.String("container", name),
						zap.Error(err))
				} else {
					stats.MarkedOffline++
				}
			}
		}
	}

	return stats
}

// createContainerCI 创建容器 CI 实例
func (s *MultiSourceContainerSyncService) createContainerCI(
	ctx context.Context,
	container prometheus.DataSourceContainerInfo,
	ciTypeID uint,
) error {
	// 获取容器的资源信息
	stats, err := s.multiClient.GetContainerStatsFromDatasource(ctx, container.Name, container.DataSourceID)

	// 合并数据源标签到容器属性
	attributes := s.buildContainerAttributes(container, stats, true)

	instance := &models.CIInstance{
		CITypeID:   ciTypeID,
		Name:       container.Name,
		Attributes: attributes,
		Status:     "active",
	}

	if err := s.ciRepo.CreateCIInstance(instance); err != nil {
		return err
	}

	logger.Info("Created container CI instance",
		zap.String("name", container.Name),
		zap.String("id", container.ID),
		zap.String("datasource", container.DataSourceName),
		zap.Bool("has_stats", stats != nil))

	return nil
}

// updateContainerCI 更新容器 CI 实例
func (s *MultiSourceContainerSyncService) updateContainerCI(
	ctx context.Context,
	existing *models.CIInstance,
	container prometheus.DataSourceContainerInfo,
) (updated bool, rebuilt bool) {
	needsUpdate := false

	// 检查容器 ID 是否变化（容器重建）
	oldID, _ := existing.Attributes["container_id"].(string)
	if oldID != container.ID && container.ID != "" {
		rebuilt = true
		needsUpdate = true

		// 更新 ID 历史
		history := []string{}
		if historyData, ok := existing.Attributes["container_id_history"]; ok {
			if historyJSON, err := json.Marshal(historyData); err == nil {
				json.Unmarshal(historyJSON, &history)
			}
		}
		history = append(history, container.ID)
		existing.Attributes["container_id_history"] = history
		existing.Attributes["container_id"] = container.ID
		existing.Attributes["rebuild_detected_at"] = time.Now().Format(time.RFC3339)

		logger.Info("Container rebuild detected",
			zap.String("name", container.Name),
			zap.String("datasource", container.DataSourceName),
			zap.String("old_id", oldID),
			zap.String("new_id", container.ID))
	}

	// 更新在线状态
	oldOnline, _ := existing.Attributes["is_online"].(bool)
	if oldOnline != container.IsRunning {
		needsUpdate = true
		existing.Attributes["is_online"] = container.IsRunning
		if container.IsRunning {
			existing.Attributes["last_online_at"] = time.Now().Format(time.RFC3339)
			logger.Info("Container came back online",
				zap.String("name", container.Name),
				zap.String("datasource", container.DataSourceName))
		} else {
			existing.Attributes["last_offline_at"] = time.Now().Format(time.RFC3339)
			logger.Info("Container went offline",
				zap.String("name", container.Name),
				zap.String("datasource", container.DataSourceName))
		}
	}

	// 更新最后发现时间
	existing.Attributes["last_seen"] = container.LastSeen.Format(time.RFC3339)
	needsUpdate = true

	// 更新镜像信息（如果变化）
	if oldImage, _ := existing.Attributes["container_image"].(string); oldImage != container.Image {
		existing.Attributes["container_image"] = container.Image
		needsUpdate = true
	}

	// 更新数据源信息
	existing.Attributes["datasource_id"] = container.DataSourceID
	existing.Attributes["datasource_name"] = container.DataSourceName
	needsUpdate = true

	// 更新数据源标签
	if container.Labels != nil {
		if existing.Attributes["datasource_labels"] == nil {
			existing.Attributes["datasource_labels"] = make(map[string]string)
		}
		for k, v := range container.Labels {
			if existing.Attributes["datasource_labels"].(map[string]string)[k] != v {
				existing.Attributes["datasource_labels"].(map[string]string)[k] = v
				needsUpdate = true
			}
		}
	}

	// 更新资源信息
	stats, err := s.multiClient.GetContainerStatsFromDatasource(ctx, container.Name, container.DataSourceID)
	if err == nil && stats != nil {
		existing.Attributes["cpu_usage_percent"] = stats.CPUUsagePercent
		existing.Attributes["memory_usage_mb"] = stats.MemoryUsageMB
		existing.Attributes["memory_limit_mb"] = stats.MemoryLimitMB
		existing.Attributes["network_rx_bytes"] = stats.NetworkRxBytes
		existing.Attributes["network_tx_bytes"] = stats.NetworkTxBytes
		existing.Attributes["disk_usage_mb"] = stats.DiskUsageMB
		existing.Attributes["uptime_seconds"] = stats.UptimeSeconds
		existing.Attributes["last_stats_update"] = time.Now().Format(time.RFC3339)
		needsUpdate = true
	}

	if needsUpdate {
		if err := s.ciRepo.UpdateCIInstance(existing); err != nil {
			logger.Error("Failed to update container CI",
				zap.String("name", container.Name),
				zap.String("datasource", container.DataSourceName),
				zap.Error(err))
			return false, rebuilt
		}
		updated = true
	}

	return updated, rebuilt
}

// markContainerOffline 标记容器为离线
func (s *MultiSourceContainerSyncService) markContainerOffline(
	ctx context.Context,
	instance *models.CIInstance,
) error {
	instance.Attributes["is_online"] = false
	instance.Attributes["last_offline_at"] = time.Now().Format(time.RFC3339)

	if err := s.ciRepo.UpdateCIInstance(instance); err != nil {
		return err
	}

	containerName, _ := instance.Attributes["container_name"].(string)
	logger.Info("Marked container as offline", zap.String("name", containerName))

	return nil
}

// buildContainerAttributes 构建容器属性
func (s *MultiSourceContainerSyncService) buildContainerAttributes(
	container prometheus.DataSourceContainerInfo,
	stats *prometheus.DataSourceContainerStats,
	isNew bool,
) map[string]interface{} {
	attributes := map[string]interface{}{
		"container_name":       container.Name,
		"container_id":         container.ID,
		"container_image":      container.Image,
		"is_online":            container.IsRunning,
		"last_seen":            container.LastSeen.Format(time.RFC3339),
		"container_id_history": []string{container.ID},
		"sync_source":          "victoriametrics",
		"auto_discovered":      true,
		"datasource_id":         container.DataSourceID,
		"datasource_name":       container.DataSourceName,
	}

	// 添加数据源标签
	if container.Labels != nil {
		attributes["datasource_labels"] = container.Labels
	}

	// 如果成功获取资源信息，添加到 attributes
	if stats != nil {
		attributes["cpu_usage_percent"] = stats.CPUUsagePercent
		attributes["memory_usage_mb"] = stats.MemoryUsageMB
		attributes["memory_limit_mb"] = stats.MemoryLimitMB
		attributes["network_rx_bytes"] = stats.NetworkRxBytes
		attributes["network_tx_bytes"] = stats.NetworkTxBytes
		attributes["disk_usage_mb"] = stats.DiskUsageMB
		attributes["uptime_seconds"] = stats.UptimeSeconds
		attributes["last_stats_update"] = time.Now().Format(time.RFC3339)
	}

	return attributes
}

// ensureContainerCIType 确保容器 CI 类型存在
func (s *MultiSourceContainerSyncService) ensureContainerCIType() (*models.CIType, error) {
	// 查找容器类型
	containerType, err := s.ciRepo.GetCITypeByName("container")
	if err == nil {
		return containerType, nil
	}

	// 兼容：也尝试查找中文名称
	ciTypes, err := s.ciRepo.GetCITypes()
	if err != nil {
		return nil, err
	}

	for _, ciType := range ciTypes {
		if ciType.Name == "container" || ciType.Name == "容器" || ciType.Name == "Container" {
			return &ciType, nil
		}
	}

	return nil, fmt.Errorf("容器 CI 类型不存在")
}

// getExistingContainerCIs 获取所有现有的容器 CI 实例
func (s *MultiSourceContainerSyncService) getExistingContainerCIs(ciTypeID uint) (map[string]*models.CIInstance, error) {
	filters := make(map[string]interface{})
	instances, _, err := s.ciRepo.GetCIInstances(ciTypeID, filters, 1, 10000)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.CIInstance)
	for i := range instances {
		if containerName, ok := instances[i].Attributes["container_name"].(string); ok {
			result[containerName] = &instances[i]
		}
	}

	return result, nil
}

// healthCheckLoop 定期检查所有数据源的健康状态
func (s *MultiSourceContainerSyncService) healthCheckLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			results := s.multiClient.HealthCheckAll(ctx)

			for dsID, err := range results {
				if err != nil {
					logger.Error("Datasource unhealthy",
						zap.String("datasource_id", dsID),
						zap.Error(err))
				}
			}
		case <-s.stopChan:
			return
		}
	}
}
