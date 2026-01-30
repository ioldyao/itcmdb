package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/alert-service/internal/repositories"
	"gorm.io/gorm"
)

// RuleEvaluator 规则评估器
type RuleEvaluator struct {
	db             *gorm.DB
	alertEngine    *AlertEngine
	routingService *RoutingService
	ruleRepo       *repositories.AlertRuleRepository
	interval       time.Duration
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
	running        bool
}

// NewRuleEvaluator 创建规则评估器
func NewRuleEvaluator(db *gorm.DB, alertEngine *AlertEngine, routingService *RoutingService, interval time.Duration) *RuleEvaluator {
	if interval == 0 {
		interval = 1 * time.Minute // 默认1分钟评估一次
	}

	return &RuleEvaluator{
		db:             db,
		alertEngine:    alertEngine,
		routingService: routingService,
		ruleRepo:       repositories.NewAlertRuleRepository(db),
		interval:       interval,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动规则评估器
func (e *RuleEvaluator) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("rule evaluator is already running")
	}

	e.running = true
	e.wg.Add(1)

	go e.evaluationLoop()

	fmt.Printf("[INFO] Rule evaluator started with interval: %v\n", e.interval)
	return nil
}

// Stop 停止规则评估器
func (e *RuleEvaluator) Stop() error {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return fmt.Errorf("rule evaluator is not running")
	}
	e.mu.Unlock()

	close(e.stopChan)
	e.wg.Wait()

	e.mu.Lock()
	e.running = false
	e.mu.Unlock()

	fmt.Printf("[INFO] Rule evaluator stopped\n")
	return nil
}

// IsRunning 检查评估器是否正在运行
func (e *RuleEvaluator) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// evaluationLoop 评估循环
func (e *RuleEvaluator) evaluationLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	// 立即执行一次评估
	e.evaluateAllRules()

	for {
		select {
		case <-ticker.C:
			e.evaluateAllRules()
		case <-e.stopChan:
			return
		}
	}
}

// evaluateAllRules 评估所有启用的规则
func (e *RuleEvaluator) evaluateAllRules() {
	// 获取所有启用的规则
	rules, err := e.ruleRepo.FindEnabled()
	if err != nil {
		fmt.Printf("[ERROR] Failed to get enabled rules: %v\n", err)
		return
	}

	if len(rules) == 0 {
		return
	}

	fmt.Printf("[INFO] Evaluating %d alert rules...\n", len(rules))

	// 并发评估所有规则
	var wg sync.WaitGroup
	for _, rule := range rules {
		wg.Add(1)
		go func(r models.AlertRule) {
			defer wg.Done()
			e.evaluateRule(&r)
		}(rule)
	}

	wg.Wait()
}

// evaluateRule 评估单个规则
func (e *RuleEvaluator) evaluateRule(rule *models.AlertRule) {
	// 检查规则是否被静默
	if rule.SilencedUntil != nil && time.Now().Before(*rule.SilencedUntil) {
		return
	}

	// 使用告警引擎评估规则
	result, err := e.alertEngine.EvaluateRule(
		rule.ID,
		rule.MetricQuery,
		rule.ThresholdOperator,
		rule.ThresholdValue,
		rule.Duration,
	)
	if err != nil {
		fmt.Printf("[ERROR] Failed to evaluate rule %s: %v\n", rule.Name, err)
		return
	}

	// 如果超过阈值，创建或更新告警实例
	if result.Exceeded {
		e.handleExceededThreshold(rule, result)
	} else {
		// 如果未超过阈值，检查是否有需要恢复的告警
		e.handleRecoveredAlert(rule)
	}
}

// handleExceededThreshold 处理超过阈值的情况
func (e *RuleEvaluator) handleExceededThreshold(rule *models.AlertRule, result *RuleEvaluationResult) {
	// 生成告警指纹
	fingerprint := GenerateFingerprint(rule.ID, map[string]interface{}{
		"rule_id": rule.ID,
		"query":   rule.MetricQuery,
	})

	// 查找是否已存在相同指纹的告警
	var existingAlert models.AlertInstance
	err := e.db.Where("fingerprint = ? AND status IN ('firing', 'acknowledged')", fingerprint).
		First(&existingAlert).Error

	now := time.Now()

	if err == gorm.ErrRecordNotFound {
		// 创建新告警
		alert := models.AlertInstance{
			AlertID:     GenerateAlertID(),
			RuleID:      &rule.ID,
			Title:       rule.Name,
			Description: rule.Description,
			Severity:    rule.Severity,
			Status:      "firing",
			Category:    "internal_rule",
			Fingerprint: fingerprint,
			FirstTriggered: now,
			LastTriggered:  now,
			Count:          1,
			Metrics: models.JSONMap{
				"current_value": result.CurrentValue,
				"threshold":     result.Threshold,
				"deviation":     result.Deviation,
			},
			TriggerConditions: models.JSONMap{
				"query":    rule.MetricQuery,
				"operator": rule.ThresholdOperator,
				"value":    rule.ThresholdValue,
				"duration": rule.Duration,
			},
		}

		if err := e.db.Create(&alert).Error; err != nil {
			fmt.Printf("[ERROR] Failed to create alert for rule %s: %v\n", rule.Name, err)
			return
		}

		fmt.Printf("[INFO] Created new alert for rule %s (current: %.2f, threshold: %.2f)\n",
			rule.Name, result.CurrentValue, result.Threshold)

		// 发送通知
		e.sendNotification(&alert, rule)

	} else if err == nil {
		// 更新现有告警
		updates := map[string]interface{}{
			"last_triggered": now,
			"count":          existingAlert.Count + 1,
			"metrics": models.JSONMap{
				"current_value": result.CurrentValue,
				"threshold":     result.Threshold,
				"deviation":     result.Deviation,
			},
		}

		if err := e.db.Model(&existingAlert).Updates(updates).Error; err != nil {
			fmt.Printf("[ERROR] Failed to update alert for rule %s: %v\n", rule.Name, err)
			return
		}

		fmt.Printf("[INFO] Updated existing alert for rule %s (count: %d, current: %.2f)\n",
			rule.Name, existingAlert.Count+1, result.CurrentValue)
	} else {
		fmt.Printf("[ERROR] Failed to query alert for rule %s: %v\n", rule.Name, err)
	}
}

// handleRecoveredAlert 处理恢复的告警
func (e *RuleEvaluator) handleRecoveredAlert(rule *models.AlertRule) {
	// 生成告警指纹
	fingerprint := GenerateFingerprint(rule.ID, map[string]interface{}{
		"rule_id": rule.ID,
		"query":   rule.MetricQuery,
	})

	// 查找处于firing状态的告警
	var alert models.AlertInstance
	err := e.db.Where("fingerprint = ? AND status = 'firing'", fingerprint).
		First(&alert).Error

	if err == gorm.ErrRecordNotFound {
		// 没有需要恢复的告警
		return
	} else if err != nil {
		fmt.Printf("[ERROR] Failed to query alert for rule %s: %v\n", rule.Name, err)
		return
	}

	// 标记告警为已恢复
	now := time.Now()
	updates := map[string]interface{}{
		"status":       "resolved",
		"recovered_at": &now,
	}

	if err := e.db.Model(&alert).Updates(updates).Error; err != nil {
		fmt.Printf("[ERROR] Failed to update recovered alert for rule %s: %v\n", rule.Name, err)
		return
	}

	fmt.Printf("[INFO] Alert recovered for rule %s\n", rule.Name)

	// 发送恢复通知
	e.sendRecoveryNotification(&alert, rule)
}

// sendNotification 发送告警通知
func (e *RuleEvaluator) sendNotification(alert *models.AlertInstance, rule *models.AlertRule) {
	// 准备告警数据
	alertData := map[string]interface{}{
		"alert_id":    alert.AlertID,
		"title":       alert.Title,
		"content":     alert.Description,
		"severity":    alert.Severity,
		"status":      alert.Status,
		"instance":    fmt.Sprintf("rule_%d", rule.ID),
		"labels":      map[string]string{
			"alertname": rule.Name,
			"severity":  rule.Severity,
			"rule_id":   fmt.Sprintf("%d", rule.ID),
		},
		"annotations": map[string]string{
			"description": rule.Description,
			"query":       rule.MetricQuery,
		},
		"timestamp": alert.LastTriggered.Format(time.RFC3339),
	}

	// 使用路由服务发送通知
	go func() {
		if err := e.routingService.RouteAndNotify(alertData, nil, alert.ID, nil); err != nil {
			fmt.Printf("[ERROR] Failed to send notification for alert %s: %v\n", alert.AlertID, err)
		}
	}()
}

// sendRecoveryNotification 发送恢复通知
func (e *RuleEvaluator) sendRecoveryNotification(alert *models.AlertInstance, rule *models.AlertRule) {
	// 准备恢复通知数据
	alertData := map[string]interface{}{
		"alert_id":    alert.AlertID,
		"title":       fmt.Sprintf("[已恢复] %s", alert.Title),
		"content":     fmt.Sprintf("告警已恢复: %s", alert.Description),
		"severity":    "info",
		"status":      "resolved",
		"instance":    fmt.Sprintf("rule_%d", rule.ID),
		"labels":      map[string]string{
			"alertname": rule.Name,
			"severity":  "info",
			"rule_id":   fmt.Sprintf("%d", rule.ID),
		},
		"annotations": map[string]string{
			"description": "告警已恢复",
			"query":       rule.MetricQuery,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// 使用路由服务发送通知
	go func() {
		if err := e.routingService.RouteAndNotify(alertData, nil, alert.ID, nil); err != nil {
			fmt.Printf("[ERROR] Failed to send recovery notification for alert %s: %v\n", alert.AlertID, err)
		}
	}()
}
