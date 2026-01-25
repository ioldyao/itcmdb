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
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	Image     string    `json:"image"`
	LastSeen  time.Time `json:"last_seen"`
	IsRunning bool      `json:"is_running"`
}

// DiscoverContainers 从 VictoriaMetrics 发现所有容器
func (c *Client) DiscoverContainers(ctx context.Context) ([]ContainerInfo, error) {
	logger.Info("Discovering containers from VictoriaMetrics")

	// 查询所有容器名称和 ID
	// 使用 container_last_seen 指标来获取容器列表
	query := `group by (name, id, image) (container_last_seen{name!=""})`

	results, err := c.queryInstantMultiple(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query containers: %w", err)
	}

	containers := make([]ContainerInfo, 0, len(results))
	now := time.Now()

	for _, result := range results {
		name := result.Metric["name"]
		id := result.Metric["id"]
		image := result.Metric["image"]

		if name == "" {
			continue
		}

		// 检查容器是否在线（最近 5 分钟内有数据）
		isRunning := c.isContainerRunning(ctx, name)

		containers = append(containers, ContainerInfo{
			Name:      name,
			ID:        id,
			Image:     image,
			LastSeen:  now,
			IsRunning: isRunning,
		})
	}

	logger.Info("Discovered containers",
		zap.Int("total", len(containers)),
		zap.Int("running", countRunning(containers)))

	return containers, nil
}

// isContainerRunning 检查容器是否在运行（最近 5 分钟内有指标数据）
func (c *Client) isContainerRunning(ctx context.Context, containerName string) bool {
	// 查询最近 5 分钟内是否有 CPU 使用数据
	query := fmt.Sprintf(`count_over_time(container_cpu_usage_seconds_total{name="%s"}[5m])`, containerName)

	value, err := c.queryInstant(ctx, query)
	if err != nil || value == nil {
		return false
	}

	return *value > 0
}

// queryInstantMultiple 执行即时查询并返回多个结果
func (c *Client) queryInstantMultiple(ctx context.Context, query string) ([]prometheusResult, error) {
	queryURL := fmt.Sprintf("%s/api/v1/query", c.baseURL)
	params := url.Values{}
	params.Add("query", query)
	fullURL := fmt.Sprintf("%s?%s", queryURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加基本认证
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
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

func countRunning(containers []ContainerInfo) int {
	count := 0
	for _, c := range containers {
		if c.IsRunning {
			count++
		}
	}
	return count
}
