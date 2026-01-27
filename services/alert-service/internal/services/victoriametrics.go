package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// VictoriaMetricsClient VictoriaMetrics客户端
type VictoriaMetricsClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Username   string
	Password   string
}

// NewVictoriaMetricsClient 创建VictoriaMetrics客户端
func NewVictoriaMetricsClient(baseURL, username, password string) *VictoriaMetricsClient {
	return &VictoriaMetricsClient{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Username: username,
		Password: password,
	}
}

// QueryResult 查询结果
type QueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}      `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// QueryRangeResult 范围查询结果
type QueryRangeResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// Query 瞬时查询
func (c *VictoriaMetricsClient) Query(query string, timestamp time.Time) (*QueryResult, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query", c.BaseURL)

	// 构建查询参数
	params := url.Values{}
	params.Set("query", query)
	if !timestamp.IsZero() {
		params.Set("time", strconv.FormatInt(timestamp.Unix(), 10))
	}

	// 构建请求URL
	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// 创建请求
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置基本认证
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %s, 响应: %s", resp.Status, string(body))
	}

	// 解析响应
	var result QueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// QueryRange 范围查询
func (c *VictoriaMetricsClient) QueryRange(query string, start, end time.Time, step time.Duration) (*QueryRangeResult, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query_range", c.BaseURL)

	// 构建查询参数
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(end.Unix(), 10))
	params.Set("step", strconv.FormatInt(int64(step.Seconds()), 10))

	// 构建请求URL
	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// 创建请求
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置基本认证
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %s, 响应: %s", resp.Status, string(body))
	}

	// 解析响应
	var result QueryRangeResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// GetMetricValue 获取指标值（辅助方法）
func (c *VictoriaMetricsClient) GetMetricValue(query string, timestamp time.Time) (float64, error) {
	result, err := c.Query(query, timestamp)
	if err != nil {
		return 0, err
	}

	if len(result.Data.Result) == 0 {
		return 0, fmt.Errorf("查询结果为空")
	}

	// 获取第一个结果
	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("无效的值类型")
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("解析值失败: %w", err)
	}

	return value, nil
}

// CheckThreshold 检查阈值（辅助方法）
func (c *VictoriaMetricsClient) CheckThreshold(query string, operator string, threshold float64, timestamp time.Time) (bool, float64, error) {
	value, err := c.GetMetricValue(query, timestamp)
	if err != nil {
		return false, 0, err
	}

	// 根据运算符比较
	var exceeded bool
	switch operator {
	case ">":
		exceeded = value > threshold
	case "<":
		exceeded = value < threshold
	case ">=":
		exceeded = value >= threshold
	case "<=":
		exceeded = value <= threshold
	case "==":
		exceeded = value == threshold
	case "!=":
		exceeded = value != threshold
	default:
		return false, value, fmt.Errorf("不支持的运算符: %s", operator)
	}

	return exceeded, value, nil
}

// HealthCheck 健康检查
func (c *VictoriaMetricsClient) HealthCheck() error {
	endpoint := fmt.Sprintf("%s/health", c.BaseURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查失败: %s", resp.Status)
	}

	return nil
}
