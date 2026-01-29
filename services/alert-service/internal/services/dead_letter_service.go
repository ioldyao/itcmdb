package services

import (
	"fmt"
	"time"

	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// DeadLetterService 死信队列服务
type DeadLetterService struct {
	db *gorm.DB
}

// NewDeadLetterService 创建死信队列服务
func NewDeadLetterService(db *gorm.DB) *DeadLetterService {
	return &DeadLetterService{db: db}
}

// AddToDeadLetter 添加到死信队列
func (s *DeadLetterService) AddToDeadLetter(webhookID int, webhookType string, alertData map[string]interface{}, err error) error {
	dlq := models.DeadLetterQueue{
		WebhookID:    webhookID,
		WebhookType:  webhookType,
		AlertData:    alertData,
		ErrorMessage: err.Error(),
		RetryCount:   0,
		Status:       "pending",
	}

	if dbErr := s.db.Create(&dlq).Error; dbErr != nil {
		return fmt.Errorf("failed to add to dead letter queue: %w", dbErr)
	}

	return nil
}

// RetryPendingItems 重试待处理的死信队列项
func (s *DeadLetterService) RetryPendingItems(maxRetries int) error {
	var items []models.DeadLetterQueue

	// 查找待重试的项（状态为pending且重试次数未超限）
	if err := s.db.Where("status = ? AND retry_count < ?", "pending", maxRetries).
		Order("created_at ASC").
		Limit(100).
		Find(&items).Error; err != nil {
		return fmt.Errorf("failed to query dead letter queue: %w", err)
	}

	for _, item := range items {
		// 更新状态为processing
		if err := s.db.Model(&item).Updates(map[string]interface{}{
			"status":        "processing",
			"last_retry_at": time.Now(),
		}).Error; err != nil {
			continue
		}

		// 这里应该调用实际的重试逻辑
		// 由于需要访问 WebhookService，这部分逻辑应该在外部实现
		// 这里只是标记为待处理状态
	}

	return nil
}

// MarkAsResolved 标记为已解决
func (s *DeadLetterService) MarkAsResolved(id int) error {
	return s.db.Model(&models.DeadLetterQueue{}).
		Where("id = ?", id).
		Update("status", "resolved").Error
}

// MarkAsFailed 标记为失败
func (s *DeadLetterService) MarkAsFailed(id int) error {
	return s.db.Model(&models.DeadLetterQueue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "failed",
			"retry_count": gorm.Expr("retry_count + 1"),
		}).Error
}

// GetPendingCount 获取待处理数量
func (s *DeadLetterService) GetPendingCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.DeadLetterQueue{}).
		Where("status = ?", "pending").
		Count(&count).Error
	return count, err
}

// CleanupOldItems 清理旧的已解决项（保留30天）
func (s *DeadLetterService) CleanupOldItems() error {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return s.db.Where("status = ? AND updated_at < ?", "resolved", thirtyDaysAgo).
		Delete(&models.DeadLetterQueue{}).Error
}
