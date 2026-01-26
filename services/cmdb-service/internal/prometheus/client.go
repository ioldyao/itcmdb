package prometheus

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/itcmdb/cmdb-service/internal/cadvisor"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

// DataSource 数据源配置
type DataSource struct {
	ID              string            `json:"id"`               // 数据源唯一标识
	Name            string            `json:"name"`             // 数据源名称
	Endpoint        string            `json:"endpoint"`         // VictoriaMetrics endpoint
	Username        string            `json:"username"`         // 认证用户名
	Password        string            `json:"password"`         // 认证密码
	Enabled         bool              `json:"enabled"`          // 是否启用
	ContainerPrefix []string          `json:"container_prefix"` // 容器名前缀过滤
	Labels          map[string]string `json:"labels"`           // 自动标签
}

// Client Prometheus/VictoriaMetrics 客户端
type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
	dataSources []*DataSource // 支持的数据源列表
	multiSource bool          // 是否为多数据源模式
}

// Prometheus API 响应结构
type prometheusResponse struct {
	Status string         `json:"status"`
	Data   prometheusData `json:"data"`
}

type prometheusData struct {
	ResultType string             `json:"resultType"`
	Result     []prometheusResult `json:"result"`
}

type prometheusResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

// NewClient 创建 Prometheus 客户端（单数据源模式，向后兼容）
func NewClient(baseURL, username, password string) *Client {
	// 创建跳过证书验证的 HTTP 客户端
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
		baseURL:     baseURL,
		username:    username,
		password:    password,
		dataSources: nil,
		multiSource: false,
	}
}

// NewMultiSourceClient 创建多数据源客户端
func NewMultiSourceClient(dataSources []*DataSource) *Client {
	// 创建跳过证书验证的 HTTP 客户端
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 过滤出启用的数据源
	enabledSources := make([]*DataSource, 0, len(dataSources))
	for _, ds := range dataSources {
		if ds.Enabled {
			enabledSources = append(enabledSources, ds)
			logger.Info("Added VictoriaMetrics datasource",
				zap.String("name", ds.Name),
				zap.String("id", ds.ID),
				zap.String("endpoint", ds.Endpoint))
		}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
		baseURL:     "", // 多数据源模式下不使用
		dataSources: enabledSources,
		multiSource: true,
	}
}

// GetContainerStats 获取容器统计信息（多数据源模式：尝试所有数据源）
func (c *Client) GetContainerStats(ctx context.Context, containerName string) (*cadvisor.ContainerStats, error) {
	if c.multiSource {
		// 多数据源模式：尝试从所有数据源获取
		return c.getStatsFromAllSources(ctx, containerName)
	}

	// 单数据源模式（原有逻辑）
	return c.getStatsFromEndpoint(ctx, containerName, c.baseURL, c.username, c.password)
}

// GetContainerStatsFromSource 从指定数据源获取容器统计信息
func (c *Client) GetContainerStatsFromSource(ctx context.Context, containerName string, datasourceID string) (*cadvisor.ContainerStats, error) {
	if !c.multiSource {
		return nil, fmt.Errorf("not in multi-source mode")
	}

	// 查找指定的数据源
	for _, ds := range c.dataSources {
		if ds.ID == datasourceID {
			return c.getStatsFromEndpoint(ctx, containerName, ds.Endpoint, ds.Username, ds.Password)
		}
	}

	return nil, fmt.Errorf("datasource not found: %s", datasourceID)
}

// getStatsFromAllSources 从所有数据源尝试获取统计信息
func (c *Client) getStatsFromAllSources(ctx context.Context, containerName string) (*cadvisor.ContainerStats, error) {
	for _, ds := range c.dataSources {
		stats, err := c.getStatsFromEndpoint(ctx, containerName, ds.Endpoint, ds.Username, ds.Password)
		if err != nil {
			logger.Debug("Failed to fetch stats from datasource, trying next",
				zap.String("datasource", ds.Name),
				zap.String("container", containerName),
				zap.Error(err))
			continue
		}
		return stats, nil
	}

	return nil, fmt.Errorf("failed to fetch stats from all datasources for container: %s", containerName)
}

// getStatsFromEndpoint 从指定端点获取容器统计信息
func (c *Client) getStatsFromEndpoint(ctx context.Context, containerName string, endpoint, username, password string) (*cadvisor.ContainerStats, error) {
	logger.Info("Fetching container stats from VictoriaMetrics",
		zap.String("container_name", containerName),
		zap.String("endpoint", endpoint))

	// 查询所有需要的指标
	stats := &cadvisor.ContainerStats{
		ContainerID: containerName,
		Timestamp:   time.Now().Unix(),
	}

	// 1. 查询 CPU 使用率 (使用 rate 计算)
	cpuQuery := fmt.Sprintf(`rate(container_cpu_usage_seconds_total{name="%s",cpu="total"}[1m])`, containerName)
	cpuValue, err := c.queryInstantWithAuth(ctx, cpuQuery, endpoint, username, password)
	if err != nil {
		logger.Warn("Failed to query CPU usage", zap.Error(err))
	} else if cpuValue != nil {
		stats.CPUUsagePercent = *cpuValue * 100.0 // 转换为百分比
	}

	// 2. 查询内存使用量
	memQuery := fmt.Sprintf(`container_memory_working_set_bytes{name="%s"}`, containerName)
	memValue, err := c.queryInstantWithAuth(ctx, memQuery, endpoint, username, password)
	if err != nil {
		logger.Warn("Failed to query memory usage", zap.Error(err))
	} else if memValue != nil {
		stats.MemoryUsageMB = *memValue / 1024 / 1024
	}

	// 3. 查询内存限制
	memLimitQuery := fmt.Sprintf(`container_spec_memory_limit_bytes{name="%s"}`, containerName)
	memLimitValue, err := c.queryInstantWithAuth(ctx, memLimitQuery, endpoint, username, password)
	if err != nil {
		logger.Warn("Failed to query memory limit", zap.Error(err))
	} else if memLimitValue != nil {
		stats.MemoryLimitMB = *memLimitValue / 1024 / 1024
	}

	// 4. 查询网络接收字节数 (所有接口总和)
	netRxQuery := fmt.Sprintf(`sum(container_network_receive_bytes_total{name="%s"})`, containerName)
	netRxValue, err := c.queryInstantWithAuth(ctx, netRxQuery, endpoint, username, password)
	if err != nil {
		logger.Warn("Failed to query network RX", zap.Error(err))
	} else if netRxValue != nil {
		stats.NetworkRxBytes = uint64(*netRxValue)
	}

	// 5. 查询网络发送字节数 (所有接口总和)
	netTxQuery := fmt.Sprintf(`sum(container_network_transmit_bytes_total{name="%s"})`, containerName)
	netTxValue, err := c.queryInstantWithAuth(ctx, netTxQuery, endpoint, username, password)
	if err != nil {
		logger.Warn("Failed to query network TX", zap.Error(err))
	} else if netTxValue != nil {
		stats.NetworkTxBytes = uint64(*netTxValue)
	}

	// 6. 查询磁盘使用量
	diskQuery := fmt.Sprintf(`sum(container_fs_usage_bytes{name="%s"})`, containerName)
	diskValue, err := c.queryInstantWithAuth(ctx, diskQuery, endpoint, username, password)
	if err != nil {
		logger.Warn("Failed to query disk usage", zap.Error(err))
	} else if diskValue != nil {
		stats.DiskUsageMB = *diskValue / 1024 / 1024
	}

	// 7. 查询容器启动时间并计算运行时长
	startTimeQuery := fmt.Sprintf(`container_start_time_seconds{name="%s"}`, containerName)
	startTimeValue, err := c.queryInstantWithAuth(ctx, startTimeQuery, endpoint, username, password)
	if err != nil {
		logger.Warn("Failed to query start time", zap.Error(err))
	} else if startTimeValue != nil {
		startTime := time.Unix(int64(*startTimeValue), 0)
		stats.UptimeSeconds = int64(time.Since(startTime).Seconds())
	}

	logger.Info("Successfully fetched container stats from VictoriaMetrics",
		zap.String("container_name", containerName),
		zap.Float64("cpu_percent", stats.CPUUsagePercent),
		zap.Float64("memory_mb", stats.MemoryUsageMB))

	return stats, nil
}

// queryInstant 执行即时查询（向后兼容单数据源模式）
func (c *Client) queryInstant(ctx context.Context, query string) (*float64, error) {
	if c.multiSource {
		return nil, fmt.Errorf("queryInstant not supported in multi-source mode, use queryInstantWithAuth")
	}
	return c.queryInstantWithAuth(ctx, query, c.baseURL, c.username, c.password)
}

// queryInstantWithAuth 执行即时查询（支持自定义认证）
func (c *Client) queryInstantWithAuth(ctx context.Context, query, endpoint, username, password string) (*float64, error) {
	// 构建查询 URL
	queryURL := fmt.Sprintf("%s/api/v1/query", endpoint)
	params := url.Values{}
	params.Add("query", query)

	fullURL := fmt.Sprintf("%s?%s", queryURL, params.Encode())

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加基本认证
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var promResp prometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查状态
	if promResp.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed with status: %s", promResp.Status)
	}

	// 提取值
	if len(promResp.Data.Result) == 0 {
		logger.Debug("No results for query", zap.String("query", query))
		return nil, nil
	}

	result := promResp.Data.Result[0]
	if len(result.Value) < 2 {
		return nil, fmt.Errorf("invalid result format")
	}

	// 值在数组的第二个位置，是字符串格式
	valueStr, ok := result.Value[1].(string)
	if !ok {
		return nil, fmt.Errorf("value is not a string")
	}

	var value float64
	if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
		return nil, fmt.Errorf("failed to parse value: %w", err)
	}

	return &value, nil
}

// HealthCheck 检查 Prometheus/VictoriaMetrics 服务是否可用
func (c *Client) HealthCheck(ctx context.Context) error {
	if c.multiSource {
		// 多数据源模式：检查所有数据源
		return c.healthCheckAll(ctx)
	}

	// 单数据源模式（原有逻辑）
	return c.healthCheckEndpoint(ctx, c.baseURL, c.username, c.password)
}

// HealthCheckAll 检查所有数据源的健康状态（多数据源模式）
func (c *Client) HealthCheckAll(ctx context.Context) map[string]error {
	if !c.multiSource {
		return nil
	}

	results := make(map[string]error)
	for _, ds := range c.dataSources {
		err := c.healthCheckEndpoint(ctx, ds.Endpoint, ds.Username, ds.Password)
		results[ds.ID] = err

		if err != nil {
			logger.Error("Datasource health check failed",
				zap.String("datasource", ds.Name),
				zap.String("id", ds.ID),
				zap.Error(err))
		} else {
			logger.Info("Datasource health check passed",
				zap.String("datasource", ds.Name),
				zap.String("id", ds.ID))
		}
	}

	return results
}

// healthCheckAll 检查所有数据源，只要有一个健康就返回成功
func (c *Client) healthCheckAll(ctx context.Context) error {
	results := c.HealthCheckAll(ctx)
	healthyCount := 0
	for _, err := range results {
		if err == nil {
			healthyCount++
		}
	}

	if healthyCount == 0 {
		return fmt.Errorf("no healthy datasource")
	}

	logger.Info("VictoriaMetrics health check completed",
		zap.Int("healthy", healthyCount),
		zap.Int("total", len(results)))

	return nil
}

// healthCheckEndpoint 检查单个端点的健康状态
func (c *Client) healthCheckEndpoint(ctx context.Context, endpoint, username, password string) error {
	url := fmt.Sprintf("%s/api/v1/query", endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url+"?query=up", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 添加基本认证
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to VictoriaMetrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("VictoriaMetrics health check failed with status %d", resp.StatusCode)
	}

	return nil
}
