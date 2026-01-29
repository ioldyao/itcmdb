package services

import (
	"fmt"
	"time"

	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// MetricsService Webhook指标服务
type MetricsService struct {
	db *gorm.DB
}

// NewMetricsService 创建指标服务
func NewMetricsService(db *gorm.DB) *MetricsService {
	return &MetricsService{db: db}
}

// RecordRequest 记录请求
func (s *MetricsService) RecordRequest(webhookID int, webhookType string, success bool, responseTime float64) error {
	var metrics models.WebhookMetrics

	// 查找或创建指标记录
	err := s.db.Where("webhook_id = ? AND webhook_type = ?", webhookID, webhookType).
		FirstOrCreate(&metrics, models.WebhookMetrics{
			WebhookID:   webhookID,
			WebhookType: webhookType,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	// 更新指标
	now := time.Now()
	updates := map[string]interface{}{
		"total_requests":  gorm.Expr("total_requests + 1"),
		"last_request_at": now,
	}

	if success {
		updates["success_requests"] = gorm.Expr("success_requests + 1")
	} else {
		updates["failed_requests"] = gorm.Expr("failed_requests + 1")
	}

	// 更新平均响应时间（简单移动平均）
	if metrics.TotalRequests > 0 {
		newAvg := (metrics.AvgResponseTime*float64(metrics.TotalRequests) + responseTime) / float64(metrics.TotalRequests+1)
		updates["avg_response_time"] = newAvg
	} else {
		updates["avg_response_time"] = responseTime
	}

	return s.db.Model(&metrics).Updates(updates).Error
}

// UpdateCircuitState 更新断路器状态
func (s *MetricsService) UpdateCircuitState(webhookID int, webhookType string, state string) error {
	return s.db.Model(&models.WebhookMetrics{}).
		Where("webhook_id = ? AND webhook_type = ?", webhookID, webhookType).
		Update("circuit_state", state).Error
}

// GetMetrics 获取指标
func (s *MetricsService) GetMetrics(webhookID int, webhookType string) (*models.WebhookMetrics, error) {
	var metrics models.WebhookMetrics
	err := s.db.Where("webhook_id = ? AND webhook_type = ?", webhookID, webhookType).
		First(&metrics).Error
	if err != nil {
		return nil, err
	}
	return &metrics, nil
}

// GetSuccessRate 获取成功率
func (s *MetricsService) GetSuccessRate(webhookID int, webhookType string) (float64, error) {
	metrics, err := s.GetMetrics(webhookID, webhookType)
	if err != nil {
		return 0, err
	}

	if metrics.TotalRequests == 0 {
		return 0, nil
	}

	return float64(metrics.SuccessRequests) / float64(metrics.TotalRequests) * 100, nil
}
