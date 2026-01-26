package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

// ContainerInfo 容器信息
type ContainerInfo struct {
	Name            string            `json:"name"`
	ID              string            `json:"id"`
	Image           string            `json:"image"`
	LastSeen        time.Time         `json:"last_seen"`
	IsRunning       bool              `json:"is_running"`
	DataSourceID    string            `json:"datasource_id"`    // 数据源ID
	DataSourceName  string            `json:"datasource_name"`  // 数据源名称
	DataSourceLabels map[string]string `json:"datasource_labels"` // 数据源标签
}

// DiscoverContainers 从 VictoriaMetrics 发现所有容器
func (c *Client) DiscoverContainers(ctx context.Context) ([]ContainerInfo, error) {
	if c.multiSource {
		// 多数据源模式：从所有数据源发现容器
		return c.discoverContainersFromAll(ctx)
	}

	// 单数据源模式（原有逻辑）
	return c.discoverContainersFromEndpoint(ctx, c.baseURL, c.username, c.password, "", "", nil)
}

// discoverContainersFromAll 从所有数据源发现容器
func (c *Client) discoverContainersFromAll(ctx context.Context) ([]ContainerInfo, error) {
	allContainers := make([]ContainerInfo, 0)

	for _, ds := range c.dataSources {
		logger.Info("Discovering containers from datasource",
			zap.String("datasource", ds.Name),
			zap.String("id", ds.ID))

		containers, err := c.discoverContainersFromEndpoint(ctx, ds.Endpoint, ds.Username, ds.Password, ds.ID, ds.Name, ds.Labels)
		if err != nil {
			logger.Error("Failed to discover containers from datasource",
				zap.String("datasource", ds.Name),
				zap.String("id", ds.ID),
				zap.Error(err))
			continue
		}

		// 应用容器名前缀过滤
		for _, container := range containers {
			if c.matchContainerPrefix(container.Name, ds.ContainerPrefix) {
				allContainers = append(allContainers, container)
			}
		}

		logger.Info("Discovered containers from datasource",
			zap.String("datasource", ds.Name),
			zap.String("id", ds.ID),
			zap.Int("count", len(containers)))
	}

	logger.Info("Total containers discovered from all datasources",
		zap.Int("total", len(allContainers)))

	return allContainers, nil
}

// discoverContainersFromEndpoint 从指定端点发现容器
func (c *Client) discoverContainersFromEndpoint(ctx context.Context, endpoint, username, password, datasourceID, datasourceName string, datasourceLabels map[string]string) ([]ContainerInfo, error) {
	// 查询所有容器名称和 ID
	// 使用 container_last_seen 指标来获取容器列表
	query := `group by (name, id, image) (container_last_seen{name!=""})`

	results, err := c.queryInstantMultipleWithAuth(ctx, query, endpoint, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to query containers: %w", err)
	}

	containers := make([]ContainerInfo, 0, len(results))
	now := time.Now()

	for _, result := range results {
		name := result.Metric["name"]
		idPath := result.Metric["id"]
		image := result.Metric["image"]

		if name == "" {
			continue
		}

		// 从 id 路径中提取实际的容器 ID
		// 格式: /system.slice/docker-{container_id}.scope
		containerID := extractContainerID(idPath)

		// 检查容器是否在线（最近 5 分钟内有数据）
		isRunning := c.isContainerRunningWithAuth(ctx, name, endpoint, username, password)

		containers = append(containers, ContainerInfo{
			Name:             name,
			ID:               containerID,
			Image:            image,
			LastSeen:         now,
			IsRunning:        isRunning,
			DataSourceID:     datasourceID,
			DataSourceName:   datasourceName,
			DataSourceLabels: datasourceLabels,
		})
	}

	return containers, nil
}

// extractContainerID 从 cgroup 路径中提取容器 ID
// 输入格式: /system.slice/docker-{container_id}.scope
// 输出: {container_id}
func extractContainerID(idPath string) string {
	if idPath == "" {
		return ""
	}

	// 查找 "docker-" 前缀
	dockerPrefix := "docker-"
	startIdx := -1
	for i := 0; i < len(idPath)-len(dockerPrefix); i++ {
		if idPath[i:i+len(dockerPrefix)] == dockerPrefix {
			startIdx = i + len(dockerPrefix)
			break
		}
	}

	if startIdx == -1 {
		// 如果没有找到 docker- 前缀，返回原始值
		return idPath
	}

	// 查找 ".scope" 后缀
	endIdx := len(idPath)
	scopeSuffix := ".scope"
	for i := startIdx; i < len(idPath)-len(scopeSuffix); i++ {
		if idPath[i:i+len(scopeSuffix)] == scopeSuffix {
			endIdx = i
			break
		}
	}

	return idPath[startIdx:endIdx]
}

// isContainerRunning 检查容器是否在运行（最近 5 分钟内有指标数据）
func (c *Client) isContainerRunning(ctx context.Context, containerName string) bool {
	if c.multiSource {
		// 多数据源模式：尝试所有数据源
		for _, ds := range c.dataSources {
			if c.isContainerRunningWithAuth(ctx, containerName, ds.Endpoint, ds.Username, ds.Password) {
				return true
			}
		}
		return false
	}

	// 单数据源模式
	return c.isContainerRunningWithAuth(ctx, containerName, c.baseURL, c.username, c.password)
}

// isContainerRunningWithAuth 检查容器是否在运行（支持自定义认证）
func (c *Client) isContainerRunningWithAuth(ctx context.Context, containerName, endpoint, username, password string) bool {
	// 查询最近 5 分钟内是否有 CPU 使用数据
	query := fmt.Sprintf(`count_over_time(container_cpu_usage_seconds_total{name="%s"}[5m])`, containerName)

	value, err := c.queryInstantWithAuth(ctx, query, endpoint, username, password)
	if err != nil || value == nil {
		return false
	}

	return *value > 0
}

// queryInstantMultiple 执行即时查询并返回多个结果
func (c *Client) queryInstantMultiple(ctx context.Context, query string) ([]prometheusResult, error) {
	if c.multiSource {
		return nil, fmt.Errorf("queryInstantMultiple not supported in multi-source mode")
	}
	return c.queryInstantMultipleWithAuth(ctx, query, c.baseURL, c.username, c.password)
}

// queryInstantMultipleWithAuth 执行即时查询并返回多个结果（支持自定义认证）
func (c *Client) queryInstantMultipleWithAuth(ctx context.Context, query, endpoint, username, password string) ([]prometheusResult, error) {
	queryURL := fmt.Sprintf("%s/api/v1/query", endpoint)
	params := url.Values{}
	params.Add("query", query)
	fullURL := fmt.Sprintf("%s?%s", queryURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加基本认证
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	var promResp prometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if promResp.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed with status: %s", promResp.Status)
	}

	return promResp.Data.Result, nil
}

// matchContainerPrefix 检查容器名是否匹配前缀过滤
func (c *Client) matchContainerPrefix(containerName string, prefixes []string) bool {
	if len(prefixes) == 0 {
		// 没有配置前缀过滤，匹配所有容器
		return true
	}

	for _, prefix := range prefixes {
		if len(containerName) >= len(prefix) && containerName[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

func countRunning(containers []ContainerInfo) int {
	count := 0
	for _, c := range containers {
		if c.IsRunning {
			count++
		}
	}
	return count
}
