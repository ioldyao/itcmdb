package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// AlertEngine 告警引擎
type AlertEngine struct {
	vmClient *VictoriaMetricsClient
}

// NewAlertEngine 创建告警引擎
func NewAlertEngine(vmClient *VictoriaMetricsClient) *AlertEngine {
	return &AlertEngine{
		vmClient: vmClient,
	}
}

// EvaluateRule 评估告警规则
func (e *AlertEngine) EvaluateRule(ruleID int, metricQuery string, operator string, threshold float64, duration int) (*RuleEvaluationResult, error) {
	// 获取当前时间
	now := time.Now()

	// 检查是否超过阈值
	exceeded, currentValue, err := e.vmClient.CheckThreshold(metricQuery, operator, threshold, now)
	if err != nil {
		return nil, fmt.Errorf("检查阈值失败: %w", err)
	}

	result := &RuleEvaluationResult{
		RuleID:        ruleID,
		Exceeded:      exceeded,
		CurrentValue:  currentValue,
		Threshold:     threshold,
		EvaluatedAt:   now,
		Duration:      duration,
	}

	// 计算偏差
	if threshold != 0 {
		result.Deviation = ((currentValue - threshold) / threshold) * 100
	}

	return result, nil
}

// RuleEvaluationResult 规则评估结果
type RuleEvaluationResult struct {
	RuleID       int       `json:"rule_id"`
	Exceeded     bool      `json:"exceeded"`
	CurrentValue float64   `json:"current_value"`
	Threshold    float64   `json:"threshold"`
	Deviation    float64   `json:"deviation"`
	EvaluatedAt  time.Time `json:"evaluated_at"`
	Duration     int       `json:"duration"`
}

// GenerateFingerprint 生成告警指纹
func GenerateFingerprint(ruleID int, targetInfo map[string]interface{}) string {
	// 使用规则ID + 目标信息生成唯一指纹
	data := fmt.Sprintf("%d-%v", ruleID, targetInfo)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// GenerateAlertID 生成告警ID（UUID格式）
func GenerateAlertID() string {
	// 简单的UUID生成（生产环境建议使用google/uuid）
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%d-%d", timestamp, time.Now().Nanosecond())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CalculateSeverity 根据指标值计算严重程度（可选）
func CalculateSeverity(value, threshold float64, baseSeverity string) string {
	// 计算偏差百分比
	deviation := ((value - threshold) / threshold) * 100

	// 根据偏差调整严重程度
	switch baseSeverity {
	case "low":
		if deviation > 100 {
			return "critical"
		} else if deviation > 50 {
			return "high"
		} else if deviation > 20 {
			return "medium"
		}
	case "medium":
		if deviation > 100 {
			return "critical"
		} else if deviation > 50 {
			return "high"
		}
	case "high":
		if deviation > 50 {
			return "critical"
		}
	}

	return baseSeverity
}

// ShouldAggregate 判断是否应该聚合告警
func ShouldAggregate(existingFingerprint string, newFingerprint string, timeWindow time.Duration, lastTriggered time.Time) bool {
	// 相同指纹且在时间窗口内
	if existingFingerprint == newFingerprint {
		timeSinceLastTrigger := time.Since(lastTriggered)
		return timeSinceLastTrigger <= timeWindow
	}
	return false
}

// FormatDuration 格式化持续时间
func FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		return fmt.Sprintf("%d分钟", minutes)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		if minutes > 0 {
			return fmt.Sprintf("%d小时%d分钟", hours, minutes)
		}
		return fmt.Sprintf("%d小时", hours)
	}
}

// ParseTimeRange 解析时间范围
func ParseTimeRange(startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("解析开始时间失败: %w", err)
		}
	} else {
		// 默认开始时间为24小时前
		startTime = time.Now().Add(-24 * time.Hour)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("解析结束时间失败: %w", err)
		}
	} else {
		// 默认结束时间为当前时间
		endTime = time.Now()
	}

	return startTime, endTime, nil
}

// GetTimeBucket 获取时间桶（用于时间序列聚合）
func GetTimeBucket(t time.Time, interval string) time.Time {
	switch interval {
	case "1m":
		return t.Truncate(time.Minute)
	case "5m":
		return t.Truncate(5 * time.Minute)
	case "15m":
		return t.Truncate(15 * time.Minute)
	case "1h":
		return t.Truncate(time.Hour)
	case "1d":
		return t.Truncate(24 * time.Hour)
	default:
		return t.Truncate(time.Hour)
	}
}

// CalculatePercentage 计算百分比
func CalculatePercentage(value, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(value) / float64(total)) * 100
}

// FormatTimestamp 格式化时间戳
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// IsValidSeverity 验证严重程度
func IsValidSeverity(severity string) bool {
	switch severity {
	case "critical", "high", "medium", "low":
		return true
	default:
		return false
	}
}

// IsValidStatus 验证状态
func IsValidStatus(status string) bool {
	switch status {
	case "firing", "acknowledged", "resolved", "closed":
		return true
	default:
		return false
	}
}

// IsValidOperator 验证运算符
func IsValidOperator(operator string) bool {
	switch operator {
	case ">", "<", ">=", "<=", "==", "!=":
		return true
	default:
		return false
	}
}

// IsAlertSilenced 检查告警是否被静默规则匹配
// matchers 格式: {"severity": "critical", "alertname": "xxx", "env": "production"}
// alertLabels 格式: 同上
func IsAlertSilenced(matchers, alertLabels map[string]interface{}) bool {
	if len(matchers) == 0 {
		return false
	}
	for key, matcherVal := range matchers {
		alertVal, ok := alertLabels[key]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", matcherVal) != fmt.Sprintf("%v", alertVal) {
			return false
		}
	}
	return true
}
