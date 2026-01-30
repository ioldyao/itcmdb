package services

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"sync"
	"time"

	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/alert-service/internal/repositories"
	"gorm.io/gorm"
)

// routingCache 路由规则缓存
type routingCache struct {
	rules      []models.AlertRoutingRule
	lastUpdate time.Time
	ttl        time.Duration
	mu         sync.RWMutex
}

// RoutingService 路由服务
type RoutingService struct {
	db                *gorm.DB
	notifyService     *NotificationService
	ruleRepo          *repositories.AlertRoutingRuleRepository
	templateRepo      *repositories.AlertNotificationTemplateRepository
	receiverGroupRepo *repositories.AlertReceiverGroupRepository
	cache             *routingCache
}

// NewRoutingService 创建路由服务
func NewRoutingService(db *gorm.DB) *RoutingService {
	return &RoutingService{
		db:                db,
		notifyService:     NewNotificationService(),
		ruleRepo:          repositories.NewAlertRoutingRuleRepository(db),
		templateRepo:      repositories.NewAlertNotificationTemplateRepository(db),
		receiverGroupRepo: repositories.NewAlertReceiverGroupRepository(db),
		cache: &routingCache{
			rules:      []models.AlertRoutingRule{},
			lastUpdate: time.Time{},
			ttl:        5 * time.Minute, // 5分钟缓存TTL
		},
	}
}

// getRoutingRules 获取路由规则（带缓存）
func (s *RoutingService) getRoutingRules() ([]models.AlertRoutingRule, error) {
	s.cache.mu.RLock()
	// 检查缓存是否有效
	if time.Since(s.cache.lastUpdate) < s.cache.ttl && len(s.cache.rules) > 0 {
		rules := s.cache.rules
		s.cache.mu.RUnlock()
		return rules, nil
	}
	s.cache.mu.RUnlock()

	// 缓存失效，重新加载
	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	// 双重检查，防止并发重复加载
	if time.Since(s.cache.lastUpdate) < s.cache.ttl && len(s.cache.rules) > 0 {
		return s.cache.rules, nil
	}

	// 从数据库加载
	rules, err := s.ruleRepo.FindEnabled()
	if err != nil {
		return nil, fmt.Errorf("failed to get routing rules: %w", err)
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	// 更新缓存
	s.cache.rules = rules
	s.cache.lastUpdate = time.Now()

	return rules, nil
}

// InvalidateCache 使缓存失效
func (s *RoutingService) InvalidateCache() {
	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()
	s.cache.lastUpdate = time.Time{}
}

// MatchReceiverGroups 根据告警labels匹配接收人组
func (s *RoutingService) MatchReceiverGroups(labels map[string]string, defaultReceiverGroupID *int) ([]int, error) {
	// 获取所有启用的路由规则（使用缓存）
	rules, err := s.getRoutingRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get routing rules: %w", err)
	}

	// 匹配的接收人组ID集合
	matchedGroupIDs := make(map[int]bool)

	// 遍历路由规则进行匹配
	for _, rule := range rules {
		if rule.Matches(labels) {
			if rule.ReceiverGroupID != nil {
				matchedGroupIDs[*rule.ReceiverGroupID] = true

				// 如果不继续匹配，直接返回结果
				if !rule.Continue {
					break
				}
			}
		}
	}

	// 如果没有匹配到任何规则，使用默认接收人组
	if len(matchedGroupIDs) == 0 && defaultReceiverGroupID != nil {
		return []int{*defaultReceiverGroupID}, nil
	}

	// 转换为切片
	result := make([]int, 0, len(matchedGroupIDs))
	for id := range matchedGroupIDs {
		result = append(result, id)
	}

	return result, nil
}

// RenderTemplate 渲染通知模板
func (s *RoutingService) RenderTemplate(templateType string, data *models.AlertTemplateData, customTemplateID *int) (interface{}, error) {
	var tplContent string
	var err error

	// 如果指定了自定义模板ID，使用自定义模板
	if customTemplateID != nil {
		tpl, err := s.templateRepo.FindByID(*customTemplateID)
		if err != nil {
			return nil, fmt.Errorf("failed to get template: %w", err)
		}
		tplContent = tpl.TemplateContent
	} else {
		// 否则使用默认模板
		tpl, err := s.templateRepo.FindDefaultByType(templateType)
		if err != nil {
			return nil, fmt.Errorf("failed to get default template: %w", err)
		}
		tplContent = tpl.TemplateContent
	}

	// 渲染模板
	rendered, err := s.executeTemplate(tplContent, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// 根据模板类型构建不同的消息格式
	switch templateType {
	case "dingtalk":
		return s.notifyService.BuildDingTalkMarkdownMessage(data.Title, rendered, []string{}, []string{}, false), nil
	case "feishu":
		// 飞书富文本格式
		content := []map[string]interface{}{
			{
				"tag":  "text",
				"text": rendered,
			},
		}
		return s.notifyService.BuildFeishuPostMessage(data.Title, content), nil
	case "wechat":
		return s.notifyService.BuildWechatMarkdownMessage(rendered), nil
	case "email":
		// 邮件HTML格式
		return rendered, nil
	default:
		return nil, fmt.Errorf("unsupported template type: %s", templateType)
	}
}

// executeTemplate 执行Go模板渲染
func (s *RoutingService) executeTemplate(tplContent string, data *models.AlertTemplateData) (string, error) {
	tmpl, err := template.New("alert").Parse(tplContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RouteAndNotify 路由并通知（带通知日志和重试机制）
func (s *RoutingService) RouteAndNotify(alertData map[string]interface{}, inboundWebhook *models.InboundWebhook, alertInstanceID int, routingRuleID *int) error {
	// 提取labels
	labelsMap := getStringMapValue(alertData, "labels")

	// 匹配接收人组
	receiverGroupIDs, err := s.MatchReceiverGroups(labelsMap, nil)
	if err != nil {
		return fmt.Errorf("failed to match receiver groups: %w", err)
	}

	// 如果没有匹配到任何接收人组，记录日志但不报错
	if len(receiverGroupIDs) == 0 {
		fmt.Printf("[INFO] No receiver group matched for alert: %+v\n", labelsMap)
		return nil
	}

	// 准备模板数据
	templateData := &models.AlertTemplateData{
		AlertID:     getStringValue(alertData, "alert_id", ""),
		Title:       getStringValue(alertData, "title", ""),
		Content:     getStringValue(alertData, "content", ""),
		Severity:    getStringValue(alertData, "severity", ""),
		Status:      getStringValue(alertData, "status", ""),
		Instance:    getStringValue(alertData, "instance", ""),
		Labels:      labelsMap,
		Annotations: getStringMapValue(alertData, "annotations"),
		Timestamp:   getStringValue(alertData, "timestamp", ""),
	}

	// 获取接收人组中的所有接收人
	for _, groupID := range receiverGroupIDs {
		group, err := s.receiverGroupRepo.FindByID(groupID)
		if err != nil {
			fmt.Printf("[ERROR] Failed to get receiver group %d: %v\n", groupID, err)
			continue
		}

		// 获取组中的所有接收人
		if group.Receivers != nil {
			for _, receiver := range group.Receivers {
				if !receiver.Enabled {
					continue
				}

				// 选择模板：路由规则模板 > 接收人默认模板 > 系统默认模板
				var templateID *int
				if routingRuleID != nil {
					rule, err := s.ruleRepo.FindByID(*routingRuleID)
					if err == nil && rule.TemplateID != nil {
						templateID = rule.TemplateID
					}
				}
				if templateID == nil && receiver.DefaultTemplateID != nil {
					templateID = receiver.DefaultTemplateID
				}

				// 渲染模板
				message, err := s.RenderTemplate(receiver.Type, templateData, templateID)
				if err != nil {
					fmt.Printf("[ERROR] Failed to render template: %v\n", err)
					continue
				}

				// 创建通知日志
				notifLog := &models.NotificationLog{
					AlertInstanceID:  alertInstanceID,
					ReceiverID:       receiver.ID,
					ReceiverGroupID:  groupID,
					RoutingRuleID:    routingRuleID,
					Status:           "pending",
					NotificationType: receiver.Type,
					Subject:          templateData.Title,
					Body:             fmt.Sprintf("%v", message),
					MaxRetries:       3,
					RetryCount:       0,
				}

				// 保存通知日志
				if err := s.db.Create(notifLog).Error; err != nil {
					fmt.Printf("[ERROR] Failed to create notification log: %v\n", err)
					continue
				}

				// 异步发送通知
				go s.sendNotificationWithRetry(receiver, message, notifLog)
			}
		}
	}

	return nil
}

// sendNotificationWithRetry 发送通知并处理重试
func (s *RoutingService) sendNotificationWithRetry(receiver models.AlertReceiver, message interface{}, notifLog *models.NotificationLog) {
	err := s.notifyService.SendNotification(
		receiver.Type,
		receiver.WebhookURL,
		receiver.Secret,
		message,
	)

	if err != nil {
		fmt.Printf("[ERROR] Failed to send notification to %s: %v\n", receiver.Name, err)

		// 检查是否可以重试
		if notifLog.IsRetryable() {
			// 计算下次重试时间（指数退避：1分钟、5分钟、15分钟）
			retryDelays := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute}
			nextRetryDelay := retryDelays[0]
			if notifLog.RetryCount < len(retryDelays) {
				nextRetryDelay = retryDelays[notifLog.RetryCount]
			}
			nextRetryAt := time.Now().Add(nextRetryDelay)

			notifLog.MarkAsRetrying(nextRetryAt)
			if updateErr := s.db.Save(notifLog).Error; updateErr != nil {
				fmt.Printf("[ERROR] Failed to update notification log: %v\n", updateErr)
			}

			// 安排重试
			time.AfterFunc(nextRetryDelay, func() {
				s.sendNotificationWithRetry(receiver, message, notifLog)
			})
		} else {
			// 超过最大重试次数
			notifLog.MarkAsMaxRetriesExceeded()
			notifLog.MarkAsFailed(err.Error())
			if updateErr := s.db.Save(notifLog).Error; updateErr != nil {
				fmt.Printf("[ERROR] Failed to update notification log: %v\n", updateErr)
			}
		}
	} else {
		fmt.Printf("[INFO] Successfully sent notification to %s\n", receiver.Name)
		notifLog.MarkAsSent()
		if updateErr := s.db.Save(notifLog).Error; updateErr != nil {
			fmt.Printf("[ERROR] Failed to update notification log: %v\n", updateErr)
		}
	}
}

// GetRoutingRule 获取路由规则
func (s *RoutingService) GetRoutingRule(id int) (*models.AlertRoutingRule, error) {
	return s.ruleRepo.FindByID(id)
}

// ListRoutingRules 列出路由规则
func (s *RoutingService) ListRoutingRules() ([]models.AlertRoutingRule, error) {
	return s.ruleRepo.FindAll()
}

// CreateRoutingRule 创建路由规则
func (s *RoutingService) CreateRoutingRule(req *models.CreateAlertRoutingRuleRequest) (*models.AlertRoutingRule, error) {
	rule := &models.AlertRoutingRule{
		Name:            req.Name,
		Description:     req.Description,
		Matchers:        req.Matchers,
		MatchType:       req.MatchType,
		ReceiverGroupID: req.ReceiverGroupID,
		Continue:        req.Continue,
		Priority:        req.Priority,
		Enabled:         req.Enabled,
	}

	if err := s.db.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("failed to create routing rule: %w", err)
	}

	// 使缓存失效
	s.InvalidateCache()

	return rule, nil
}

// UpdateRoutingRule 更新路由规则
func (s *RoutingService) UpdateRoutingRule(id int, req *models.UpdateAlertRoutingRuleRequest) (*models.AlertRoutingRule, error) {
	rule, err := s.ruleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Matchers != nil {
		updates["matchers"] = req.Matchers
	}
	if req.MatchType != nil {
		updates["match_type"] = *req.MatchType
	}
	if req.ReceiverGroupID != nil {
		updates["receiver_group_id"] = *req.ReceiverGroupID
	}
	if req.Continue != nil {
		updates["continue"] = *req.Continue
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := s.db.Model(rule).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update routing rule: %w", err)
	}

	// 使缓存失效
	s.InvalidateCache()

	return rule, nil
}

// DeleteRoutingRule 删除路由规则
func (s *RoutingService) DeleteRoutingRule(id int) error {
	if err := s.ruleRepo.Delete(id); err != nil {
		return err
	}

	// 使缓存失效
	s.InvalidateCache()

	return nil
}

// GetNotificationTemplate 获取通知模板
func (s *RoutingService) GetNotificationTemplate(id int) (*models.AlertNotificationTemplate, error) {
	return s.templateRepo.FindByID(id)
}

// ListNotificationTemplates 列出通知模板
func (s *RoutingService) ListNotificationTemplates(templateType string) ([]models.AlertNotificationTemplate, error) {
	if templateType == "" {
		return s.templateRepo.FindAll()
	}
	return s.templateRepo.FindByType(templateType)
}

// CreateNotificationTemplate 创建通知模板
func (s *RoutingService) CreateNotificationTemplate(req *models.CreateAlertNotificationTemplateRequest) (*models.AlertNotificationTemplate, error) {
	tpl := &models.AlertNotificationTemplate{
		Name:            req.Name,
		Description:     req.Description,
		TemplateType:    req.TemplateType,
		TemplateContent: req.TemplateContent,
		IsDefault:       req.IsDefault,
	}

	if err := s.db.Create(tpl).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification template: %w", err)
	}

	return tpl, nil
}

// UpdateNotificationTemplate 更新通知模板
func (s *RoutingService) UpdateNotificationTemplate(id int, req *models.UpdateAlertNotificationTemplateRequest) (*models.AlertNotificationTemplate, error) {
	tpl, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.TemplateContent != nil {
		updates["template_content"] = *req.TemplateContent
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	if err := s.db.Model(tpl).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update notification template: %w", err)
	}

	return tpl, nil
}

// DeleteNotificationTemplate 删除通知模板
func (s *RoutingService) DeleteNotificationTemplate(id int) error {
	return s.templateRepo.Delete(id)
}
