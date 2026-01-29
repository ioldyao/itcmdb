package services

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"

	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/alert-service/internal/repositories"
	"gorm.io/gorm"
)

// RoutingService 路由服务
type RoutingService struct {
	db              *gorm.DB
	notifyService   *NotificationService
	ruleRepo        *repositories.AlertRoutingRuleRepository
	templateRepo    *repositories.AlertNotificationTemplateRepository
	receiverGroupRepo *repositories.AlertReceiverGroupRepository
}

// NewRoutingService 创建路由服务
func NewRoutingService(db *gorm.DB) *RoutingService {
	return &RoutingService{
		db:              db,
		notifyService:   NewNotificationService(),
		ruleRepo:        repositories.NewAlertRoutingRuleRepository(db),
		templateRepo:    repositories.NewAlertNotificationTemplateRepository(db),
		receiverGroupRepo: repositories.NewAlertReceiverGroupRepository(db),
	}
}

// MatchReceiverGroups 根据告警labels匹配接收人组
func (s *RoutingService) MatchReceiverGroups(labels map[string]string, defaultReceiverGroupID *int) ([]int, error) {
	// 获取所有启用的路由规则，按优先级排序
	rules, err := s.ruleRepo.FindEnabled()
	if err != nil {
		return nil, fmt.Errorf("failed to get routing rules: %w", err)
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

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

// RouteAndNotify 路由并通知
func (s *RoutingService) RouteAndNotify(alertData map[string]interface{}, inboundWebhook *models.InboundWebhook) error {
	// 提取labels
	labelsMap := getStringMapValue(alertData, "labels")

	// 匹配接收人组
	receiverGroupIDs, err := s.MatchReceiverGroups(labelsMap, inboundWebhook.DefaultReceiverGroupID)
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
		if group.Members != nil {
			for _, member := range group.Members {
				if member.Receiver == nil || !member.Receiver.Enabled {
					continue
				}

				// 渲染模板
				message, err := s.RenderTemplate(member.Receiver.Type, templateData, nil)
				if err != nil {
					fmt.Printf("[ERROR] Failed to render template: %v\n", err)
					continue
				}

				// 异步发送通知
				go func(receiver *models.AlertReceiver) {
					if err := s.notifyService.SendNotification(
						receiver.Type,
						receiver.WebhookURL,
						receiver.Secret,
						message,
					); err != nil {
						fmt.Printf("[ERROR] Failed to send notification to %s: %v\n", receiver.Name, err)
					} else {
						fmt.Printf("[INFO] Successfully sent notification to %s\n", receiver.Name)
					}
				}(member.Receiver)
			}
		}
	}

	return nil
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

	return rule, nil
}

// DeleteRoutingRule 删除路由规则
func (s *RoutingService) DeleteRoutingRule(id int) error {
	return s.ruleRepo.Delete(id)
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
