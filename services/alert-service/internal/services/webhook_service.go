package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// WebhookService Webhook服务
type WebhookService struct {
	db               *gorm.DB
	httpClient       *http.Client
	notifyService    *NotificationService
	routingService   *RoutingService
	dlqService       *DeadLetterService
	metricsService   *MetricsService
	rateLimiter      *RateLimiter
	circuitBreakers  map[int]*CircuitBreaker // webhookID -> CircuitBreaker
	retryConfig      *RetryConfig
	mu               sync.RWMutex
}

// NewWebhookService 创建Webhook服务
func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		notifyService:   NewNotificationService(),
		routingService:  NewRoutingService(db),
		dlqService:      NewDeadLetterService(db),
		metricsService:  NewMetricsService(db),
		rateLimiter:     NewRateLimiter(100, 200), // 100 req/s, 容量200
		circuitBreakers: make(map[int]*CircuitBreaker),
		retryConfig:     DefaultRetryConfig(),
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
	// 1. 提取告警信息 - labels和annotations是map[string]string类型
	labelsMap := getStringMapValue(alertData, "labels")
	annotationsMap := getStringMapValue(alertData, "annotations")
	status := getStringValue(alertData, "status", "firing")
	fingerprint := getStringValue(alertData, "fingerprint", "")

	// 如果没有fingerprint，生成一个
	if fingerprint == "" {
		// 使用labels生成指纹
		fingerprint = generateFingerprint(labelsMap)
	}

	// 生成alert_id（唯一标识）- 使用fingerprint作为唯一标识
	alertID := fmt.Sprintf("%s-%s", webhook.SourceType, fingerprint)

	// 匹配路由规则（用于记录和通知）
	receiverGroupIDs, err := s.routingService.MatchReceiverGroups(labelsMap, nil)
	if err != nil {
		// 路由匹配失败不应阻止告警创建，仅记录错误
		// TODO: 使用结构化日志记录错误
	}

	// 获取匹配的路由规则ID（用于记录）
	var routingRuleID *int
	if len(receiverGroupIDs) > 0 {
		// 查找匹配的路由规则
		rules, err := s.routingService.ruleRepo.FindEnabled()
		if err == nil {
			for _, rule := range rules {
				if rule.Matches(labelsMap) && rule.ReceiverGroupID != nil {
					for _, groupID := range receiverGroupIDs {
						if *rule.ReceiverGroupID == groupID {
							ruleID := rule.ID
							routingRuleID = &ruleID
							break
						}
					}
					if routingRuleID != nil {
						break
					}
				}
			}
		}
	}

	// 提取告警标题和描述
	title := labelsMap["alertname"]
	if title == "" {
		title = annotationsMap["summary"]
	}
	if title == "" {
		title = "告警"
	}
	description := annotationsMap["description"]

	// 提取严重程度
	severity := labelsMap["severity"]
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

	// 2. 检查是否已存在（去重）- 使用fingerprint查找
	var existingAlert models.AlertInstance
	err = s.db.Where("fingerprint = ? AND category = ?", fingerprint, webhook.SourceType).First(&existingAlert).Error

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
		s.recordAlertHistory(&existingAlert, "updated", fmt.Sprintf("Status changed to %s", status))

	} else if err == gorm.ErrRecordNotFound {
		// 3. 创建新告警记录
		newAlert := models.AlertInstance{
			AlertID:           alertID,
			Title:             title,
			Description:       description,
			Severity:          severity,
			Status:            status,
			Category:          webhook.SourceType,
			ObjectType:        labelsMap["instance"], // 使用instance字段作为空间名
			Fingerprint:       fingerprint,
			FirstTriggered:    now,
			LastTriggered:     now,
			Tags:              convertStringMapToJSONMap(labelsMap),
			TriggerConditions: convertStringMapToJSONMap(annotationsMap),
		}

		// 设置通知渠道（记录匹配到的接收人组）
		if len(receiverGroupIDs) > 0 {
			receiverGroupsMap := make(map[string]interface{})
			for i, groupID := range receiverGroupIDs {
				receiverGroupsMap[fmt.Sprintf("group_%d", i)] = groupID
			}
			newAlert.NotificationChannels = receiverGroupsMap
		}

		if err := s.db.Create(&newAlert).Error; err != nil {
			return fmt.Errorf("failed to create alert: %w", err)
		}

		// 记录历史
		s.recordAlertHistory(&newAlert, "triggered", "Alert received from webhook")

	} else {
		return fmt.Errorf("failed to query alert: %w", err)
	}

	// 5. 使用路由规则匹配接收人并发送通知
	go func() {
		// 构造告警数据用于路由和通知
		alertDataForRouting := map[string]interface{}{
			"alert_id":    alertID,
			"title":       title,
			"content":     description,
			"severity":    severity,
			"status":      status,
			"category":    webhook.SourceType,
			"instance":    labelsMap["instance"],
			"labels":     labelsMap,
			"annotations": annotationsMap,
			"timestamp":   now.Format("2006-01-02 15:04:05"),
			"fingerprint": fingerprint,
		}

		// 使用路由服务进行匹配和通知
		if err := s.routingService.RouteAndNotify(alertDataForRouting, webhook); err != nil {
			// 记录错误但不影响告警创建
			// TODO: 使用结构化日志记录错误
		}
	}()

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
	if val, ok := m[key]; ok && val != nil {
		// 尝试直接类型断言
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal
		}
		// 如果不是map[string]interface{}，可能是其他类型
		// TODO: 使用结构化日志记录类型不匹配
	}
	// 如果键不存在或为nil，返回空map而不是nil
	return make(map[string]interface{})
}

// getStringMapValue 从map中获取map[string]string值
func getStringMapValue(m map[string]interface{}, key string) map[string]string {
	result := make(map[string]string)
	if val, ok := m[key]; ok && val != nil {
		// 尝试类型断言
		if mapVal, ok := val.(map[string]string); ok {
			return mapVal
		}
		// 如果是map[string]interface{}，转换
		if mapVal, ok := val.(map[string]interface{}); ok {
			for k, v := range mapVal {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
			return result
		}
		// TODO: 使用结构化日志记录类型不匹配
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

// convertStringMapToJSONMap 将map[string]string转换为JSONMap
func convertStringMapToJSONMap(m map[string]string) models.JSONMap {
	if m == nil {
		return nil
	}
	result := make(models.JSONMap)
	for k, v := range m {
		result[k] = v
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

// sendHTTP 发送HTTP请求（带企业级特性）
func (s *WebhookService) sendHTTP(webhook *models.OutboundWebhook, url string, payload interface{}, alertData map[string]interface{}) error {
	// 1. 速率限制检查
	if !s.rateLimiter.Allow() {
		err := fmt.Errorf("rate limit exceeded")
		s.metricsService.RecordRequest(webhook.ID, "outbound", false, 0)
		return err
	}

	// 2. 获取或创建断路器
	cb := s.getOrCreateCircuitBreaker(webhook.ID)

	// 3. 使用断路器和重试机制发送请求
	startTime := time.Now()
	var lastErr error

	err := cb.Call(func() error {
		return RetryWithExponentialBackoff(s.retryConfig, func() error {
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
				lastErr = err
				return fmt.Errorf("failed to send webhook: %w", err)
			}
			defer resp.Body.Close()

			respBody, _ := io.ReadAll(resp.Body)

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				lastErr = fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
				return lastErr
			}

			// 成功响应
			s.logOutboundSuccess(webhook.ID, alertData, resp.StatusCode, string(respBody))
			return nil
		})
	})

	// 4. 记录指标
	responseTime := float64(time.Since(startTime).Milliseconds())
	success := err == nil
	s.metricsService.RecordRequest(webhook.ID, "outbound", success, responseTime)

	// 5. 更新断路器状态到数据库
	state := "closed"
	switch cb.GetState() {
	case StateOpen:
		state = "open"
	case StateHalfOpen:
		state = "half_open"
	}
	s.metricsService.UpdateCircuitState(webhook.ID, "outbound", state)

	// 6. 如果失败，添加到死信队列
	if err != nil {
		s.dlqService.AddToDeadLetter(webhook.ID, "outbound", alertData, err)
		s.logOutboundError(webhook.ID, alertData, err)
		return err
	}

	// 7. 更新最后发送时间
	now := time.Now()
	webhook.LastSent = &now
	if dbErr := s.db.Save(webhook).Error; dbErr != nil {
		// 更新失败不应影响主流程
		// TODO: 使用结构化日志记录错误
	}

	return nil
}

// getOrCreateCircuitBreaker 获取或创建断路器
func (s *WebhookService) getOrCreateCircuitBreaker(webhookID int) *CircuitBreaker {
	s.mu.RLock()
	cb, exists := s.circuitBreakers[webhookID]
	s.mu.RUnlock()

	if exists {
		return cb
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查
	if cb, exists := s.circuitBreakers[webhookID]; exists {
		return cb
	}

	// 创建新的断路器：5次失败后打开，60秒后尝试恢复
	cb = NewCircuitBreaker(5, 60*time.Second)
	s.circuitBreakers[webhookID] = cb
	return cb
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
	// 使用labels生成指纹（排序后拼接）
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
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

