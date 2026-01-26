package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/itcmdb/cmdb-service/internal/cadvisor"
	"github.com/itcmdb/cmdb-service/internal/models"
	"github.com/itcmdb/cmdb-service/internal/prometheus"
	"github.com/itcmdb/cmdb-service/internal/repository"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

// ContainerSyncService 容器同步服务
type ContainerSyncService struct {
	ciRepo           repository.CIRepository
	prometheusClient *prometheus.Client
	syncInterval     time.Duration
	stopChan         chan struct{}
}

// NewContainerSyncService 创建容器同步服务
func NewContainerSyncService(
	ciRepo repository.CIRepository,
	prometheusClient *prometheus.Client,
	syncInterval time.Duration,
) *ContainerSyncService {
	return &ContainerSyncService{
		ciRepo:           ciRepo,
		prometheusClient: prometheusClient,
		syncInterval:     syncInterval,
		stopChan:         make(chan struct{}),
	}
}

// Start 启动同步服务
func (s *ContainerSyncService) Start() {
	if s.prometheusClient == nil {
		logger.Warn("VictoriaMetrics client not configured, container sync disabled")
		return
	}

	logger.Info("Starting container sync service", zap.Duration("interval", s.syncInterval))

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
				logger.Info("Container sync service stopped")
				return
			}
		}
	}()
}

// Stop 停止同步服务
func (s *ContainerSyncService) Stop() {
	close(s.stopChan)
}

// syncContainers 同步容器信息
func (s *ContainerSyncService) syncContainers() error {
	ctx := context.Background()
	logger.Info("Starting container synchronization")

	// 1. 从 VictoriaMetrics 发现容器
	containers, err := s.prometheusClient.DiscoverContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover containers: %w", err)
	}

	logger.Info("Discovered containers from VictoriaMetrics",
		zap.Int("total", len(containers)))

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

	logger.Info("Container synchronization completed",
		zap.Int("created", stats.Created),
		zap.Int("updated", stats.Updated),
		zap.Int("marked_offline", stats.MarkedOffline),
		zap.Int("marked_online", stats.MarkedOnline),
		zap.Int("rebuilt", stats.Rebuilt))

	return nil
}

// SyncStats 同步统计
type SyncStats struct {
	Created       int
	Updated       int
	MarkedOffline int
	MarkedOnline  int
	Rebuilt       int
}

// performSync 执行同步
func (s *ContainerSyncService) performSync(
	ctx context.Context,
	discoveredContainers []prometheus.ContainerInfo,
	existingContainers map[string]*models.CIInstance,
	ciTypeID uint,
) SyncStats {
	stats := SyncStats{}

	// 创建容器名称到发现容器的映射
	discoveredMap := make(map[string]prometheus.ContainerInfo)
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
					zap.Error(err))
			} else {
				stats.Created++
			}
		} else {
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

// ensureContainerCIType 确保容器 CI 类型存在
func (s *ContainerSyncService) ensureContainerCIType() (*models.CIType, error) {
	// 查找容器类型（数据库初始化时创建的是 "container"）
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

	// 如果还是找不到，返回错误
	return nil, fmt.Errorf("容器 CI 类型不存在。数据库初始化时应该已创建 'container' 类型，请检查数据库")
}

// getExistingContainerCIs 获取所有现有的容器 CI 实例
func (s *ContainerSyncService) getExistingContainerCIs(ciTypeID uint) (map[string]*models.CIInstance, error) {
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

// createContainerCI 创建容器 CI 实例
func (s *ContainerSyncService) createContainerCI(
	ctx context.Context,
	container prometheus.ContainerInfo,
	ciTypeID uint,
) error {
	// 获取容器的资源信息
	var stats *cadvisor.ContainerStats
	var err error

	// 如果容器有数据源信息，尝试从指定数据源获取
	if container.DataSourceID != "" {
		stats, err = s.prometheusClient.GetContainerStatsFromSource(ctx, container.Name, container.DataSourceID)
		if err != nil {
			logger.Debug("Failed to get stats from datasource, will try all",
				zap.String("container", container.Name),
				zap.String("datasource", container.DataSourceName),
				zap.Error(err))
			// 回退到尝试所有数据源
			stats, err = s.prometheusClient.GetContainerStats(ctx, container.Name)
		}
	} else {
		stats, err = s.prometheusClient.GetContainerStats(ctx, container.Name)
	}

	attributes := map[string]interface{}{
		"container_name":       container.Name,
		"container_id":         container.ID,
		"container_image":      container.Image,
		"is_online":            container.IsRunning,
		"last_seen":            container.LastSeen.Format(time.RFC3339),
		"container_id_history": []string{container.ID},
		"sync_source":          "victoriametrics",
		"auto_discovered":      true,
	}

	// 添加数据源信息
	if container.DataSourceID != "" {
		attributes["datasource_id"] = container.DataSourceID
		attributes["datasource_name"] = container.DataSourceName
		if container.DataSourceLabels != nil {
			attributes["datasource_labels"] = container.DataSourceLabels
		}
	}

	// 如果成功获取资源信息，添加到 attributes
	if err == nil && stats != nil {
		attributes["cpu_usage_percent"] = stats.CPUUsagePercent
		attributes["memory_usage_mb"] = stats.MemoryUsageMB
		attributes["memory_limit_mb"] = stats.MemoryLimitMB
		attributes["network_rx_bytes"] = stats.NetworkRxBytes
		attributes["network_tx_bytes"] = stats.NetworkTxBytes
		attributes["disk_usage_mb"] = stats.DiskUsageMB
		attributes["uptime_seconds"] = stats.UptimeSeconds
		attributes["last_stats_update"] = time.Now().Format(time.RFC3339)
	}

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
func (s *ContainerSyncService) updateContainerCI(
	ctx context.Context,
	existing *models.CIInstance,
	container prometheus.ContainerInfo,
) (updated bool, rebuilt bool) {
	needsUpdate := false

	// 检查数据源是否变更
	oldDS, _ := existing.Attributes["datasource_id"].(string)
	if oldDS != container.DataSourceID && container.DataSourceID != "" {
		logger.Info("Container datasource changed",
			zap.String("container", container.Name),
			zap.String("old_datasource", oldDS),
			zap.String("new_datasource", container.DataSourceID))
		needsUpdate = true
		existing.Attributes["datasource_id"] = container.DataSourceID
		existing.Attributes["datasource_name"] = container.DataSourceName
		if container.DataSourceLabels != nil {
			existing.Attributes["datasource_labels"] = container.DataSourceLabels
		}
	}

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

	// 更新资源信息
	var stats *cadvisor.ContainerStats
	var err error

	// 如果容器有数据源信息，尝试从指定数据源获取
	if container.DataSourceID != "" {
		stats, err = s.prometheusClient.GetContainerStatsFromSource(ctx, container.Name, container.DataSourceID)
		if err != nil {
			// 回退到尝试所有数据源
			stats, err = s.prometheusClient.GetContainerStats(ctx, container.Name)
		}
	} else {
		stats, err = s.prometheusClient.GetContainerStats(ctx, container.Name)
	}

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
func (s *ContainerSyncService) markContainerOffline(
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
