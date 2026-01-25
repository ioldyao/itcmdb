package cadvisor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

// Client cAdvisor 客户端
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// ContainerStats 容器统计信息
type ContainerStats struct {
	ContainerID       string  `json:"container_id"`
	CPUUsagePercent   float64 `json:"cpu_usage_percent"`
	MemoryUsageMB     float64 `json:"memory_usage_mb"`
	MemoryLimitMB     float64 `json:"memory_limit_mb"`
	NetworkRxBytes    uint64  `json:"network_rx_bytes"`
	NetworkTxBytes    uint64  `json:"network_tx_bytes"`
	DiskUsageMB       float64 `json:"disk_usage_mb"`
	UptimeSeconds     int64   `json:"uptime_seconds"`
	Timestamp         int64   `json:"timestamp"`
}

// cAdvisor API 响应结构
type cAdvisorResponse struct {
	Containers map[string]cAdvisorContainer `json:"containers"`
}

type cAdvisorContainer struct {
	Spec  cAdvisorSpec  `json:"spec"`
	Stats []cAdvisorStat `json:"stats"`
}

type cAdvisorSpec struct {
	CreationTime time.Time       `json:"creation_time"`
	Memory       cAdvisorMemory  `json:"memory"`
	CPU          cAdvisorCPU     `json:"cpu"`
}

type cAdvisorMemory struct {
	Limit uint64 `json:"limit"`
}

type cAdvisorCPU struct {
	Limit uint64 `json:"limit"`
}

type cAdvisorStat struct {
	Timestamp time.Time          `json:"timestamp"`
	CPU       cAdvisorCPUStat    `json:"cpu"`
	Memory    cAdvisorMemoryStat `json:"memory"`
	Network   cAdvisorNetworkStat `json:"network"`
	Filesystem []cAdvisorFSStat   `json:"filesystem"`
}

type cAdvisorCPUStat struct {
	Usage cAdvisorCPUUsage `json:"usage"`
}

type cAdvisorCPUUsage struct {
	Total  uint64 `json:"total"`
	PerCPU []uint64 `json:"per_cpu_usage"`
}

type cAdvisorMemoryStat struct {
	Usage      uint64 `json:"usage"`
	WorkingSet uint64 `json:"working_set"`
}

type cAdvisorNetworkStat struct {
	Interfaces []cAdvisorNetworkInterface `json:"interfaces"`
}

type cAdvisorNetworkInterface struct {
	Name     string `json:"name"`
	RxBytes  uint64 `json:"rx_bytes"`
	TxBytes  uint64 `json:"tx_bytes"`
}

type cAdvisorFSStat struct {
	Device string `json:"device"`
	Usage  uint64 `json:"usage"`
}

// NewClient 创建 cAdvisor 客户端
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// GetContainerStats 获取容器统计信息
func (c *Client) GetContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	// 构建 API URL
	url := fmt.Sprintf("%s/api/v1.3/docker/%s", c.baseURL, containerID)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Error("Failed to create cAdvisor request", zap.Error(err), zap.String("url", url))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to fetch from cAdvisor", zap.Error(err), zap.String("url", url))
		return nil, fmt.Errorf("failed to fetch from cAdvisor: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("cAdvisor returned non-OK status",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("cAdvisor returned status %d", resp.StatusCode)
	}

	// 解析响应
	var apiResp map[string]cAdvisorContainer
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logger.Error("Failed to decode cAdvisor response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取容器数据
	containerKey := fmt.Sprintf("/docker/%s", containerID)
	container, ok := apiResp[containerKey]
	if !ok {
		logger.Warn("Container not found in cAdvisor response", zap.String("container_id", containerID))
		return nil, fmt.Errorf("container not found: %s", containerID)
	}

	// 检查是否有统计数据
	if len(container.Stats) == 0 {
		logger.Warn("No stats available for container", zap.String("container_id", containerID))
		return nil, fmt.Errorf("no stats available for container: %s", containerID)
	}

	// 获取最新的统计数据
	latestStat := container.Stats[len(container.Stats)-1]

	// 计算 CPU 使用率（需要两个数据点）
	cpuUsagePercent := 0.0
	if len(container.Stats) >= 2 {
		prevStat := container.Stats[len(container.Stats)-2]
		cpuDelta := float64(latestStat.CPU.Usage.Total - prevStat.CPU.Usage.Total)
		timeDelta := latestStat.Timestamp.Sub(prevStat.Timestamp).Seconds()
		if timeDelta > 0 {
			cpuUsagePercent = (cpuDelta / (timeDelta * 1e9)) * 100.0
		}
	}

	// 计算网络统计
	var rxBytes, txBytes uint64
	for _, iface := range latestStat.Network.Interfaces {
		rxBytes += iface.RxBytes
		txBytes += iface.TxBytes
	}

	// 计算磁盘使用
	var diskUsage uint64
	for _, fs := range latestStat.Filesystem {
		diskUsage += fs.Usage
	}

	// 计算运行时间
	uptime := time.Since(container.Spec.CreationTime).Seconds()

	// 构建返回结果
	stats := &ContainerStats{
		ContainerID:       containerID,
		CPUUsagePercent:   cpuUsagePercent,
		MemoryUsageMB:     float64(latestStat.Memory.WorkingSet) / 1024 / 1024,
		MemoryLimitMB:     float64(container.Spec.Memory.Limit) / 1024 / 1024,
		NetworkRxBytes:    rxBytes,
		NetworkTxBytes:    txBytes,
		DiskUsageMB:       float64(diskUsage) / 1024 / 1024,
		UptimeSeconds:     int64(uptime),
		Timestamp:         latestStat.Timestamp.Unix(),
	}

	logger.Info("Successfully fetched container stats",
		zap.String("container_id", containerID),
		zap.Float64("cpu_percent", cpuUsagePercent),
		zap.Float64("memory_mb", stats.MemoryUsageMB))

	return stats, nil
}

// HealthCheck 检查 cAdvisor 服务是否可用
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1.3/machine", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to cAdvisor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cAdvisor health check failed with status %d", resp.StatusCode)
	}

	return nil
}
