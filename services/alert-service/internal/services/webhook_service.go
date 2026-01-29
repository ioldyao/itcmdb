package services

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// WebhookService Webhook服务
type WebhookService struct {
	db            *gorm.DB
	httpClient    *http.Client
	notifyService *NotificationService
}

// NewWebhookService 创建Webhook服务
func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		notifyService: NewNotificationService(),
	}
}

// ParseAlertmanagerPayload 解析Alertmanager payload
func (s *WebhookService) ParseAlertmanagerPayload(r *http.Request) ([]map[string]interface{}, error) {
	var payload models.AlertmanagerWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode alertmanager payload: %w", err)
	}

	alerts := make([]map[string]interface{}, 0, len(payload.Alerts))
	for _, alert := range payload.Alerts {
		alertMap := map[string]interface{}{
			"fingerprint":       alert.Fingerprint,
			"status":            alert.Status,
			"labels":            alert.Labels,
			"annotations":       alert.Annotations,
			"starts_at":         alert.StartsAt,
			"ends_at":           alert.EndsAt,
			"generator_url":     alert.GeneratorURL,
			"group_key":         payload.GroupKey,
			"receiver":          payload.Receiver,
			"group_labels":      payload.GroupLabels,
			"common_labels":     payload.CommonLabels,
			"common_annotations": payload.CommonAnnotations,
			"external_url":      payload.ExternalURL,
			"source_type":       "alertmanager",
		}
		alerts = append(alerts, alertMap)
	}

	return alerts, nil
}

// ParsePrometheusPayload 解析Prometheus payload
func (s *WebhookService) ParsePrometheusPayload(r *http.Request) ([]map[string]interface{}, error) {
	var payload models.PrometheusWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode prometheus payload: %w", err)
	}

	alerts := make([]map[string]interface{}, 0, len(payload.Alerts))
	for _, alert := range payload.Alerts {
		alertMap := map[string]interface{}{
			"status":       "firing",
			"labels":       alert.Labels,
			"annotations":  alert.Annotations,
			"starts_at":    alert.StartsAt,
			"ends_at":      alert.EndsAt,
			"source_type":  "prometheus",
		}
		alerts = append(alerts, alertMap)
	}

	return alerts, nil
}

// ParseVictoriaMetricsPayload 解析VictoriaMetrics payload
func (s *WebhookService) ParseVictoriaMetricsPayload(r *http.Request) ([]map[string]interface{}, error) {
	var payload models.VictoriaMetricsWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode victoriametrics payload: %w", err)
	}

	alerts := make([]map[string]interface{}, 0, len(payload.Alerts))
	for _, alert := range payload.Alerts {
		alertMap := map[string]interface{}{
			"id":           alert.ID,
			"status":       alert.Status,
			"labels":       alert.Labels,
			"annotations":  alert.Annotations,
			"starts_at":    alert.StartsAt,
			"ends_at":      alert.EndsAt,
			"group_labels": payload.GroupLabels,
			"receiver":     payload.ReceiverName,
			"source_type":  "victoriametrics",
		}
		alerts = append(alerts, alertMap)
	}

	return alerts, nil
}

// ParseCustomPayload 解析自定义payload
func (s *WebhookService) ParseCustomPayload(r *http.Request) ([]map[string]interface{}, error) {
	var payload models.CustomWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode custom payload: %w", err)
	}

	alertMap := map[string]interface{}{
		"alert_id":    payload.AlertID,
		"title":       payload.Title,
		"content":     payload.Content,
		"severity":    payload.Severity,
		"status":      payload.Status,
		"metadata":    payload.Metadata,
		"timestamp":   payload.Timestamp,
		"source_type": "custom",
	}

	return []map[string]interface{}{alertMap}, nil
}

// ProcessInboundAlert 处理接收到的告警
func (s *WebhookService) ProcessInboundAlert(webhook *models.InboundWebhook, alertData map[string]interface{}) error {
	// 1. 提取告警信息
	labels := getInterfaceMapValue(alertData, "labels")
	annotations := getInterfaceMapValue(alertData, "annotations")
	status := getStringValue(alertData, "status", "firing")
	fingerprint := getStringValue(alertData, "fingerprint", "")

	// 如果没有fingerprint，生成一个
	if fingerprint == "" {
		// 使用labels生成指纹
		fingerprint = generateFingerprint(convertMapToString(labels))
	}

	// 生成alert_id（唯一标识）
	alertID := fmt.Sprintf("%s-%s", webhook.SourceType, fingerprint)

	// 提取告警标题和描述
	title := getMapStringValue(labels, "alertname", "")
	if title == "" {
		title = getMapStringValue(annotations, "summary", "告警")
	}
	description := getMapStringValue(annotations, "description", "")

	// 提取严重程度
	severity := getMapStringValue(labels, "severity", "warning")
	if severity == "" {
		severity = "warning"
	}
	// 标准化严重程度
	switch severity {
	case "critical", "crit", "fatal":
		severity = "critical"
	case "high", "major", "error":
		severity = "high"
	case "warning", "warn", "minor":
		severity = "medium"
	case "info", "low":
		severity = "low"
	default:
		severity = "medium"
	}

	// 2. 检查是否已存在（去重）
	var existingAlert models.AlertInstance
	err := s.db.Where("alert_id = ?", alertID).First(&existingAlert).Error

	now := time.Now()
	if err == nil {
		// 告警已存在，更新状态和时间
		updates := map[string]interface{}{
			"status":         status,
			"last_triggered": now,
		}

		// 如果状态从firing变为resolved，记录恢复时间
		if existingAlert.Status == "firing" && status == "resolved" {
			updates["recovered_at"] = &now
		}

		if err := s.db.Model(&existingAlert).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update alert: %w", err)
		}

		// 记录历史
		s.recordAlertHistory(&existingAlert, "status_update", fmt.Sprintf("Status changed to %s", status))

	} else if err == gorm.ErrRecordNotFound {
		// 3. 创建新告警记录
		newAlert := models.AlertInstance{
			AlertID:           alertID,
			Title:             title,
			Description:       description,
			Severity:          severity,
			Status:            status,
			Category:          webhook.SourceType,
			ObjectType:        webhook.Name,
			Fingerprint:       fingerprint,
			FirstTriggered:    now,
			LastTriggered:     now,
			Tags:              convertInterfaceMapToJSONMap(labels),
			TriggerConditions: convertInterfaceMapToJSONMap(annotations),
		}

		if err := s.db.Create(&newAlert).Error; err != nil {
			return fmt.Errorf("failed to create alert: %w", err)
		}

		// 记录历史
		s.recordAlertHistory(&newAlert, "triggered", "Alert received from webhook")

	} else {
		return fmt.Errorf("failed to query alert: %w", err)
	}

	// 5. 根据配置决定是否发送通知（这里暂时不做）
	// TODO: 根据webhook配置决定是否发送通知

	return nil
}

// convertToJSONMap 将map[string]string转换为JSONMap
func convertToJSONMap(m map[string]string) models.JSONMap {
	if m == nil {
		return nil
	}
	result := make(models.JSONMap)
	for k, v := range m {
		result[k] = v
	}
	return result
}

// getInterfaceMapValue 从map中获取map[string]interface{}值
func getInterfaceMapValue(m map[string]interface{}, key string) map[string]interface{} {
	result := make(map[string]interface{})
	if val, ok := m[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return result
}

// getMapStringValue 从map[string]interface{}中获取string值
func getMapStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// convertMapToString 将map[string]interface{}转换为map[string]string
func convertMapToString(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}

// convertInterfaceMapToJSONMap 将map[string]interface{}转换为JSONMap
func convertInterfaceMapToJSONMap(m map[string]interface{}) models.JSONMap {
	if m == nil {
		return nil
	}
	result := make(models.JSONMap)
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// SendOutboundWebhook 发送推送Webhook
func (s *WebhookService) SendOutboundWebhook(webhook *models.OutboundWebhook, alertData map[string]interface{}) error {
	// 根据目标类型处理
	switch webhook.TargetType {
	case "receiver":
		// 使用配置的告警接收人
		if webhook.ReceiverID == nil {
			return fmt.Errorf("receiver_id is required for receiver type")
		}
		var receiver models.AlertReceiver
		if err := s.db.First(&receiver, *webhook.ReceiverID).Error; err != nil {
			return fmt.Errorf("failed to get receiver: %w", err)
		}
		return s.sendToReceiver(&receiver, alertData)

	case "alertmanager":
		// 推送到 Alertmanager
		if webhook.EndpointURL == "" {
			return fmt.Errorf("endpoint_url is required for alertmanager type")
		}
		payload, err := s.buildAlertmanagerPayload(alertData)
		if err != nil {
			return err
		}
		return s.sendHTTP(webhook, webhook.EndpointURL, payload, alertData)

	default:
		return fmt.Errorf("unknown target type: %s", webhook.TargetType)
	}
}

// sendToReceiver 发送到接收人
func (s *WebhookService) sendToReceiver(receiver *models.AlertReceiver, alertData map[string]interface{}) error {
	title := getStringValue(alertData, "title", "告警通知")
	content := getStringValue(alertData, "content", "")
	severity := getStringValue(alertData, "severity", "info")
	status := getStringValue(alertData, "status", "firing")
	alertID := getStringValue(alertData, "alert_id", "")

	// 使用通知服务发送
	return s.notifyService.SendAlertNotification(
		receiver.Type,
		receiver.WebhookURL,
		receiver.Secret,
		alertID,
		title,
		content,
		severity,
		status,
		nil,
	)
}

// sendHTTP 发送HTTP请求
func (s *WebhookService) sendHTTP(webhook *models.OutboundWebhook, url string, payload interface{}, alertData map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logOutboundError(webhook.ID, alertData, err)
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// 更新最后发送时间
	now := time.Now()
	webhook.LastSent = &now
	s.db.Save(webhook)

	s.logOutboundSuccess(webhook.ID, alertData, resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// buildAlertmanagerPayload 构建Alertmanager格式的payload
func (s *WebhookService) buildAlertmanagerPayload(alertData map[string]interface{}) (interface{}, error) {
	alerts := []models.AlertmanagerAlert{
		{
			Status:      getStringValue(alertData, "status", "firing"),
			Labels:      getMapValue(alertData, "labels"),
			Annotations: getMapValue(alertData, "annotations"),
			StartsAt:    getTimeValue(alertData, "starts_at"),
			EndsAt:      getTimeValue(alertData, "ends_at"),
			GeneratorURL: getStringValue(alertData, "generator_url", ""),
			Fingerprint: getStringValue(alertData, "fingerprint", ""),
		},
	}

	payload := models.AlertmanagerWebhookPayload{
		Version:           "4",
		GroupKey:          getStringValue(alertData, "group_key", ""),
		Status:            getStringValue(alertData, "status", "firing"),
		Receiver:          getStringValue(alertData, "receiver", ""),
		GroupLabels:       getMapValue(alertData, "group_labels"),
		CommonLabels:      getMapValue(alertData, "common_labels"),
		CommonAnnotations: getMapValue(alertData, "common_annotations"),
		Alerts:            alerts,
	}

	return payload, nil
}

// logOutboundSuccess 记录推送成功日志
func (s *WebhookService) logOutboundSuccess(webhookID int, alertData map[string]interface{}, statusCode int, responseBody string) {
	log := models.OutboundWebhookLog{
		WebhookID:    webhookID,
		AlertID:      getStringValue(alertData, "alert_id", ""),
		StatusCode:   statusCode,
		RequestData:  alertData,
		ResponseData: responseBody,
		SentAt:       time.Now(),
	}
	s.db.Create(&log)
}

// logOutboundError 记录推送错误日志
func (s *WebhookService) logOutboundError(webhookID int, alertData map[string]interface{}, err error) {
	log := models.OutboundWebhookLog{
		WebhookID:     webhookID,
		AlertID:       getStringValue(alertData, "alert_id", ""),
		StatusCode:    500,
		RequestData:   alertData,
		ErrorMessage:  err.Error(),
		SentAt:        time.Now(),
		RetryCount:    0,
	}
	s.db.Create(&log)
}

// BroadcastAlert 广播告警到所有启用的推送目标
func (s *WebhookService) BroadcastAlert(alertData map[string]interface{}) error {
	var webhooks []models.OutboundWebhook
	if err := s.db.Where("enabled = ?", true).Find(&webhooks).Error; err != nil {
		return fmt.Errorf("failed to query webhooks: %w", err)
	}

	// 并发发送
	errChan := make(chan error, len(webhooks))
	for _, webhook := range webhooks {
		go func(wh models.OutboundWebhook) {
			errChan <- s.SendOutboundWebhook(&wh, alertData)
		}(webhook)
	}

	// 收集错误
	var errors []error
	for i := 0; i < len(webhooks); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send some webhooks: %d errors", len(errors))
	}

	return nil
}

// 辅助函数
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getMapValue(m map[string]interface{}, key string) map[string]string {
	result := make(map[string]string)
	if val, ok := m[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			for k, v := range m {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
		}
	}
	return result
}

func getTimeValue(m map[string]interface{}, key string) time.Time {
	if val, ok := m[key]; ok {
		if t, ok := val.(time.Time); ok {
			return t
		}
	}
	return time.Now()
}

// generateFingerprint 生成告警指纹
func generateFingerprint(labels map[string]string) string {
	// 使用labels生成指纹（简单实现：排序后拼接）
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha1.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte("="))
		h.Write([]byte(labels[k]))
		h.Write([]byte(","))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// recordAlertHistory 记录告警历史
func (s *WebhookService) recordAlertHistory(alert *models.AlertInstance, eventType string, message string) {
	history := models.AlertHistory{
		AlertID:      alert.ID,
		EventType:    eventType,
		Message:      message,
		OldStatus:    alert.Status,
		NewStatus:    alert.Status,
		OperatedBy:   nil,
		OperatedAt:   time.Now(),
	}
	s.db.Create(&history)
}

