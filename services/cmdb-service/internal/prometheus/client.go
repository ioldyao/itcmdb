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

// Client Prometheus/VictoriaMetrics 客户端
type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
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

// NewClient 创建 Prometheus 客户端
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
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

// GetContainerStats 获取容器统计信息
func (c *Client) GetContainerStats(ctx context.Context, containerName string) (*cadvisor.ContainerStats, error) {
	logger.Info("Fetching container stats from VictoriaMetrics",
		zap.String("container_name", containerName),
		zap.String("endpoint", c.baseURL))

	// 查询所有需要的指标
	stats := &cadvisor.ContainerStats{
		ContainerID: containerName,
		Timestamp:   time.Now().Unix(),
	}

	// 1. 查询 CPU 使用率 (使用 rate 计算)
	cpuQuery := fmt.Sprintf(`rate(container_cpu_usage_seconds_total{name="%s",cpu="total"}[1m])`, containerName)
	cpuValue, err := c.queryInstant(ctx, cpuQuery)
	if err != nil {
		logger.Warn("Failed to query CPU usage", zap.Error(err))
	} else if cpuValue != nil {
		stats.CPUUsagePercent = *cpuValue * 100.0 // 转换为百分比
	}

	// 2. 查询内存使用量
	memQuery := fmt.Sprintf(`container_memory_working_set_bytes{name="%s"}`, containerName)
	memValue, err := c.queryInstant(ctx, memQuery)
	if err != nil {
		logger.Warn("Failed to query memory usage", zap.Error(err))
	} else if memValue != nil {
		stats.MemoryUsageMB = *memValue / 1024 / 1024
	}

	// 3. 查询内存限制
	memLimitQuery := fmt.Sprintf(`container_spec_memory_limit_bytes{name="%s"}`, containerName)
	memLimitValue, err := c.queryInstant(ctx, memLimitQuery)
	if err != nil {
		logger.Warn("Failed to query memory limit", zap.Error(err))
	} else if memLimitValue != nil {
		stats.MemoryLimitMB = *memLimitValue / 1024 / 1024
	}

	// 4. 查询网络接收字节数 (所有接口总和)
	netRxQuery := fmt.Sprintf(`sum(container_network_receive_bytes_total{name="%s"})`, containerName)
	netRxValue, err := c.queryInstant(ctx, netRxQuery)
	if err != nil {
		logger.Warn("Failed to query network RX", zap.Error(err))
	} else if netRxValue != nil {
		stats.NetworkRxBytes = uint64(*netRxValue)
	}

	// 5. 查询网络发送字节数 (所有接口总和)
	netTxQuery := fmt.Sprintf(`sum(container_network_transmit_bytes_total{name="%s"})`, containerName)
	netTxValue, err := c.queryInstant(ctx, netTxQuery)
	if err != nil {
		logger.Warn("Failed to query network TX", zap.Error(err))
	} else if netTxValue != nil {
		stats.NetworkTxBytes = uint64(*netTxValue)
	}

	// 6. 查询磁盘使用量
	diskQuery := fmt.Sprintf(`sum(container_fs_usage_bytes{name="%s"})`, containerName)
	diskValue, err := c.queryInstant(ctx, diskQuery)
	if err != nil {
		logger.Warn("Failed to query disk usage", zap.Error(err))
	} else if diskValue != nil {
		stats.DiskUsageMB = *diskValue / 1024 / 1024
	}

	// 7. 查询容器启动时间并计算运行时长
	startTimeQuery := fmt.Sprintf(`container_start_time_seconds{name="%s"}`, containerName)
	startTimeValue, err := c.queryInstant(ctx, startTimeQuery)
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

// queryInstant 执行即时查询
func (c *Client) queryInstant(ctx context.Context, query string) (*float64, error) {
	// 构建查询 URL
	queryURL := fmt.Sprintf("%s/api/v1/query", c.baseURL)
	params := url.Values{}
	params.Add("query", query)

	fullURL := fmt.Sprintf("%s?%s", queryURL, params.Encode())

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加基本认证
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
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
	url := fmt.Sprintf("%s/api/v1/query", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url+"?query=up", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 添加基本认证
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
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
