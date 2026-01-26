package prometheus

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

// VictoriaMetricsDataSource VictoriaMetrics数据源配置
type VictoriaMetricsDataSource struct {
	Name            string            `json:"name"`             // 数据源名称
	ID              string            `json:"id"`               // 数据源唯一标识
	Endpoint        string            `json:"endpoint"`         // VictoriaMetrics endpoint
	Username        string            `json:"username"`         // 认证用户名
	Password        string            `json:"password"`         // 认证密码
	Enabled         bool              `json:"enabled"`          // 是否启用
	ContainerPrefix []string          `json:"container_prefix"` // 容器名前缀过滤（可选）
	Labels          map[string]string `json:"labels"`           // 自动添加到容器的标签
}

// MultiDataSourceClient 多数据源客户端
type MultiDataSourceClient struct {
	datasources []*DataSourceClient
}

// DataSourceClient 单个数据源客户端
type DataSourceClient struct {
	config *VictoriaMetricsDataSource
	client *Client
}

// NewMultiDataSourceClient 从配置创建多数据源客户端
func NewMultiDataSourceClient(datasources []VictoriaMetricsDataSource) *MultiDataSourceClient {
	if len(datasources) == 0 {
		logger.Warn("No VictoriaMetrics datasources configured")
		return &MultiDataSourceClient{
			datasources: []*DataSourceClient{},
		}
	}

	clients := make([]*DataSourceClient, 0, len(datasources))

	for i := range datasources {
		ds := &datasources[i]
		if !ds.Enabled {
			logger.Info("Skipping disabled datasource",
				zap.String("name", ds.Name),
				zap.String("id", ds.ID))
			continue
		}

		client := &Client{
			baseURL:  ds.Endpoint,
			username: ds.Username,
			password: ds.Password,
		}

		// 初始化 HTTP 客户端
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.httpClient = &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		}

		clients = append(clients, &DataSourceClient{
			config: ds,
			client: client,
		})

		logger.Info("Added VictoriaMetrics datasource",
			zap.String("name", ds.Name),
			zap.String("id", ds.ID),
			zap.String("endpoint", ds.Endpoint),
			zap.Int("labels", len(ds.Labels)))
	}

	return &MultiDataSourceClient{
		datasources: clients,
	}
}

// DiscoverContainersFromAll 从所有启用的数据源发现容器
func (m *MultiDataSourceClient) DiscoverContainersFromAll(ctx context.Context) ([]DataSourceContainerInfo, error) {
	allContainers := make([]DataSourceContainerInfo, 0)

	for _, ds := range m.datasources {
		logger.Info("Discovering containers from datasource",
			zap.String("datasource", ds.config.Name),
			zap.String("id", ds.config.ID))

		containers, err := ds.client.DiscoverContainers(ctx)
		if err != nil {
			logger.Error("Failed to discover containers from datasource",
				zap.String("datasource", ds.config.Name),
				zap.String("id", ds.config.ID),
				zap.Error(err))
			continue
		}

		// 为每个容器添加数据源信息
		for _, container := range containers {
			// 检查容器名前缀过滤
			if !ds.matchContainerPrefix(container.Name) {
				continue
			}

			dsContainer := DataSourceContainerInfo{
				ContainerInfo: container,
				DataSourceID:  ds.config.ID,
				DataSourceName: ds.config.Name,
				Labels:        ds.config.Labels,
			}

			allContainers = append(allContainers, dsContainer)
		}

		logger.Info("Discovered containers from datasource",
			zap.String("datasource", ds.config.Name),
			zap.String("id", ds.config.ID),
			zap.Int("count", len(containers)))
	}

	logger.Info("Total containers discovered from all datasources",
		zap.Int("total", len(allContainers)))

	return allContainers, nil
}

// GetContainerStatsFromDatasource 从指定数据源获取容器统计信息
func (m *MultiDataSourceClient) GetContainerStatsFromDatasource(
	ctx context.Context,
	containerName string,
	datasourceID string,
) (*DataSourceContainerStats, error) {
	ds := m.getDataSourceByID(datasourceID)
	if ds == nil {
		return nil, fmt.Errorf("datasource not found: %s", datasourceID)
	}

	stats, err := ds.client.GetContainerStats(ctx, containerName)
	if err != nil {
		return nil, err
	}

	return &DataSourceContainerStats{
		ContainerStats: stats,
		DataSourceID:   datasourceID,
		DataSourceName: ds.config.Name,
	}, nil
}

// GetContainerStatsFromAll 尝试从所有数据源获取容器统计信息
// 返回第一个成功的结果
func (m *MultiDataSourceClient) GetContainerStatsFromAll(
	ctx context.Context,
	containerName string,
) (*DataSourceContainerStats, error) {
	for _, ds := range m.datasources {
		stats, err := ds.client.GetContainerStats(ctx, containerName)
		if err != nil {
			logger.Debug("Failed to get stats from datasource, trying next",
				zap.String("datasource", ds.config.Name),
				zap.String("container", containerName),
				zap.Error(err))
			continue
		}

		return &DataSourceContainerStats{
			ContainerStats: stats,
			DataSourceID:   ds.config.ID,
			DataSourceName: ds.config.Name,
		}, nil
	}

	return nil, fmt.Errorf("failed to get stats from all datasources for container: %s", containerName)
}

// HealthCheckAll 检查所有数据源的健康状态
func (m *MultiDataSourceClient) HealthCheckAll(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for _, ds := range m.datasources {
		err := ds.client.HealthCheck(ctx)
		results[ds.config.ID] = err

		if err != nil {
			logger.Error("Datasource health check failed",
				zap.String("datasource", ds.config.Name),
				zap.String("id", ds.config.ID),
				zap.Error(err))
		} else {
			logger.Info("Datasource health check passed",
				zap.String("datasource", ds.config.Name),
				zap.String("id", ds.config.ID))
		}
	}

	return results
}

// GetDatasourceInfo 获取所有数据源信息
func (m *MultiDataSourceClient) GetDatasourceInfo() []VictoriaMetricsDataSource {
	info := make([]VictoriaMetricsDataSource, 0, len(m.datasources))

	for _, ds := range m.datasources {
		info = append(info, *ds.config)
	}

	return info
}

// getDataSourceByID 根据ID获取数据源
func (m *MultiDataSourceClient) getDataSourceByID(id string) *DataSourceClient {
	for _, ds := range m.datasources {
		if ds.config.ID == id {
			return ds
		}
	}
	return nil
}

// matchContainerPrefix 检查容器名是否匹配前缀过滤
func (d *DataSourceClient) matchContainerPrefix(containerName string) bool {
	if len(d.config.ContainerPrefix) == 0 {
		// 没有配置前缀过滤，匹配所有容器
		return true
	}

	for _, prefix := range d.config.ContainerPrefix {
		if len(containerName) >= len(prefix) && containerName[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// DataSourceContainerInfo 带数据源信息的容器信息
type DataSourceContainerInfo struct {
	ContainerInfo
	DataSourceID   string            `json:"datasource_id"`   // 数据源ID
	DataSourceName string            `json:"datasource_name"`  // 数据源名称
	Labels         map[string]string `json:"labels"`           // 数据源标签
}

// DataSourceContainerStats 带数据源信息的容器统计
type DataSourceContainerStats struct {
	*ContainerStats
	DataSourceID   string `json:"datasource_id"`   // 数据源ID
	DataSourceName string `json:"datasource_name"`  // 数据源名称
}
