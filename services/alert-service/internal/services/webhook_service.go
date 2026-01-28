package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	// TODO: 实现告警处理逻辑
	// 1. 提取告警信息
	// 2. 生成告警指纹
	// 3. 检查是否已存在（去重）
	// 4. 创建或更新告警记录
	// 5. 根据配置发送通知

	return nil
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
