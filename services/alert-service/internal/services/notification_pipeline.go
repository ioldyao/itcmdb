package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/itcmdb/services/alert-service/internal/models"
	"gorm.io/gorm"
)

// NotificationPipeline 通知管道 - 统一的通知工作流
type NotificationPipeline struct {
	db              *gorm.DB
	routingEngine   *RoutingEngine
	notificationSvc *NotificationService
}

// NewNotificationPipeline 创建通知管道实例
func NewNotificationPipeline(db *gorm.DB, routingEngine *RoutingEngine, notificationSvc *NotificationService) *NotificationPipeline {
	return &NotificationPipeline{
		db:              db,
		routingEngine:   routingEngine,
		notificationSvc: notificationSvc,
	}
}

// ProcessAlert 处理告警并发送通知
func (np *NotificationPipeline) ProcessAlert(alert *models.AlertInstance) error {
	// 将 alert.Tags 转换为 labels map
	labels := make(map[string]string)
	if alert.Tags != nil {
		for key, value := range alert.Tags {
			if strValue, ok := value.(string); ok {
				labels[key] = strValue
			}
		}
	}

	// 添加基本标签
	labels["alert_id"] = alert.AlertID
	labels["severity"] = alert.Severity
	labels["status"] = alert.Status
	if alert.Category != "" {
		labels["category"] = alert.Category
	}

	// 使用路由引擎进行路由
	routingResult, err := np.routingEngine.RouteAlert(labels)
	if err != nil {
		return fmt.Errorf("failed to route alert: %w", err)
	}

	// 如果没有匹配的接收组，记录警告但不报错
	if len(routingResult.ReceiverGroupIDs) == 0 {
		fmt.Printf("Warning: No receiver groups matched for alert %s\n", alert.AlertID)
		return nil
	}

	// 为每个接收组发送通知
	for i, groupID := range routingResult.ReceiverGroupIDs {
		var routingRuleID *int
		if i < len(routingResult.MatchedRules) {
			routingRuleID = &routingResult.MatchedRules[i].ID
		}

		err := np.sendToReceiverGroup(alert, groupID, routingRuleID)
		if err != nil {
			// 记录错误但继续处理其他接收组
			fmt.Printf("Error sending to receiver group %d: %v\n", groupID, err)
		}
	}

	return nil
}

// sendToReceiverGroup 向接收组发送通知
func (np *NotificationPipeline) sendToReceiverGroup(alert *models.AlertInstance, groupID int, routingRuleID *int) error {
	// 获取接收组及其接收人
	var group models.AlertReceiverGroup
	err := np.db.Preload("Receivers").First(&group, groupID).Error
	if err != nil {
		return fmt.Errorf("failed to load receiver group: %w", err)
	}

	if !group.Enabled {
		fmt.Printf("Receiver group %d is disabled, skipping\n", groupID)
		return nil
	}

	// 为每个接收人创建通知日志并发送
	for _, receiver := range group.Receivers {
		if !receiver.Enabled {
			continue
		}

		// 创建通知日志
		notifLog := &models.NotificationLog{
			AlertInstanceID:  alert.ID,
			ReceiverID:       receiver.ID,
			ReceiverGroupID:  groupID,
			RoutingRuleID:    routingRuleID,
			Status:           "pending",
			NotificationType: receiver.Type,
			Subject:          alert.Title,
			Body:             alert.Description,
			MaxRetries:       3,
		}

		// 保存通知日志
		if err := np.db.Create(notifLog).Error; err != nil {
			fmt.Printf("Failed to create notification log: %v\n", err)
			continue
		}

		// 异步发送通知
		go np.sendNotification(alert, &receiver, notifLog)
	}

	return nil
}

// sendNotification 发送单个通知
func (np *NotificationPipeline) sendNotification(alert *models.AlertInstance, receiver *models.AlertReceiver, notifLog *models.NotificationLog) {
	// 构建通知内容
	message, err := np.notificationSvc.BuildAlertMessage(
		receiver.Type,
		alert.AlertID,
		alert.Title,
		alert.Description,
		alert.Severity,
		alert.Status,
		map[string]interface{}{
			"category":    alert.Category,
			"object_type": alert.ObjectType,
		},
	)

	if err != nil {
		np.handleNotificationFailure(notifLog, fmt.Sprintf("Failed to build message: %v", err))
		return
	}

	// 序列化请求payload
	requestPayload, _ := json.Marshal(message)
	notifLog.RequestPayload = models.JSONMap{"message": message}

	// 发送通知
	err = np.notificationSvc.SendNotification(receiver.Type, receiver.WebhookURL, receiver.Secret, message)

	if err != nil {
		// 发送失败，安排重试
		np.handleNotificationFailure(notifLog, err.Error())
	} else {
		// 发送成功
		np.handleNotificationSuccess(notifLog)
	}
}

// handleNotificationSuccess 处理通知发送成功
func (np *NotificationPipeline) handleNotificationSuccess(notifLog *models.NotificationLog) {
	notifLog.MarkAsSent()

	// 更新数据库
	err := np.db.Model(notifLog).Updates(map[string]interface{}{
		"status":       notifLog.Status,
		"sent_at":      notifLog.SentAt,
		"delivered_at": notifLog.DeliveredAt,
	}).Error

	if err != nil {
		fmt.Printf("Failed to update notification log %d: %v\n", notifLog.ID, err)
	}
}

// handleNotificationFailure 处理通知发送失败
func (np *NotificationPipeline) handleNotificationFailure(notifLog *models.NotificationLog, errorMsg string) {
	notifLog.MarkAsFailed(errorMsg)

	// 检查是否可以重试
	if notifLog.IsRetryable() {
		// 计算下次重试时间（指数退避）
		nextRetryAt := np.calculateNextRetry(notifLog.RetryCount)
		notifLog.MarkAsRetrying(nextRetryAt)

		// 更新数据库
		err := np.db.Model(notifLog).Updates(map[string]interface{}{
			"status":        notifLog.Status,
			"failed_at":     notifLog.FailedAt,
			"error_message": notifLog.ErrorMessage,
			"retry_count":   notifLog.RetryCount,
			"next_retry_at": notifLog.NextRetryAt,
		}).Error

		if err != nil {
			fmt.Printf("Failed to update notification log %d: %v\n", notifLog.ID, err)
		}

		fmt.Printf("Scheduled retry %d/%d for notification log %d at %s\n",
			notifLog.RetryCount, notifLog.MaxRetries, notifLog.ID, nextRetryAt.Format(time.RFC3339))
	} else {
		// 超过最大重试次数
		notifLog.MarkAsMaxRetriesExceeded()

		// 更新数据库
		err := np.db.Model(notifLog).Updates(map[string]interface{}{
			"status":        notifLog.Status,
			"failed_at":     notifLog.FailedAt,
			"error_message": notifLog.ErrorMessage,
		}).Error

		if err != nil {
			fmt.Printf("Failed to update notification log %d: %v\n", notifLog.ID, err)
		}

		fmt.Printf("Max retries exceeded for notification log %d\n", notifLog.ID)
	}
}

// calculateNextRetry 计算下次重试时间（指数退避）
func (np *NotificationPipeline) calculateNextRetry(retryCount int) time.Time {
	// 指数退避: 1分钟, 5分钟, 15分钟
	delays := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
	}

	delayIndex := retryCount
	if delayIndex >= len(delays) {
		delayIndex = len(delays) - 1
	}

	return time.Now().Add(delays[delayIndex])
}

// ProcessRetries 处理待重试的通知
func (np *NotificationPipeline) ProcessRetries() error {
	// 查询需要重试的通知
	var notifLogs []models.NotificationLog
	err := np.db.Where("status = ? AND next_retry_at <= ?", "retrying", time.Now()).
		Preload("AlertInstance").
		Preload("Receiver").
		Find(&notifLogs).Error

	if err != nil {
		return fmt.Errorf("failed to query retryable notifications: %w", err)
	}

	fmt.Printf("Found %d notifications to retry\n", len(notifLogs))

	// 重试每个通知
	for _, notifLog := range notifLogs {
		if notifLog.AlertInstance == nil || notifLog.Receiver == nil {
			fmt.Printf("Skipping notification log %d: missing alert or receiver\n", notifLog.ID)
			continue
		}

		// 异步重试
		go np.sendNotification(notifLog.AlertInstance, notifLog.Receiver, &notifLog)
	}

	return nil
}

// GetNotificationStats 获取通知统计信息
func (np *NotificationPipeline) GetNotificationStats(startTime, endTime time.Time) (*models.NotificationStats, error) {
	stats := &models.NotificationStats{}

	// 统计各状态的通知数量
	var counts []struct {
		Status string
		Count  int
	}

	err := np.db.Model(&models.NotificationLog{}).
		Select("status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Group("status").
		Find(&counts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	// 填充统计数据
	for _, c := range counts {
		switch c.Status {
		case "sent":
			stats.TotalSent = c.Count
		case "failed":
			stats.TotalFailed = c.Count
		case "retrying":
			stats.TotalRetrying = c.Count
		case "max_retries_exceeded":
			stats.TotalMaxRetriesExceeded = c.Count
		}
	}

	// 计算成功率
	total := stats.TotalSent + stats.TotalFailed + stats.TotalMaxRetriesExceeded
	if total > 0 {
		stats.SuccessRate = float64(stats.TotalSent) / float64(total) * 100
	}

	// 计算平均重试次数
	var avgRetry struct {
		AvgRetryCount float64
	}
	err = np.db.Model(&models.NotificationLog{}).
		Select("AVG(retry_count) as avg_retry_count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Where("retry_count > 0").
		Scan(&avgRetry).Error

	if err == nil {
		stats.AverageRetryCount = avgRetry.AvgRetryCount
	}

	return stats, nil
}
