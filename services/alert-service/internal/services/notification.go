package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NotificationService 通知服务
type NotificationService struct {
	client *http.Client
}

// NewNotificationService 创建通知服务
func NewNotificationService() *NotificationService {
	return &NotificationService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(receiverType, webhookURL, secret string, content interface{}) error {
	switch receiverType {
	case "dingtalk":
		return s.sendDingTalk(webhookURL, secret, content)
	case "feishu":
		return s.sendFeishu(webhookURL, content)
	case "wechat":
		return s.sendWechat(webhookURL, secret, content)
	default:
		return fmt.Errorf("unsupported receiver type: %s", receiverType)
	}
}

// sendDingTalk 发送钉钉通知
func (s *NotificationService) sendDingTalk(webhookURL, secret string, content interface{}) error {
	// 如果有签名密钥，计算签名
	url := webhookURL
	if secret != "" {
		timestamp := time.Now().UnixMilli()
		sign, err := s.generateDingTalkSign(timestamp, secret)
		if err != nil {
			return fmt.Errorf("failed to generate dingtalk sign: %w", err)
		}
		url = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
	}

	// 构建请求体
	body, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal dingtalk message: %w", err)
	}

	// 发送请求
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send dingtalk notification: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read dingtalk response: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse dingtalk response: %w", err)
	}

	// 检查错误码
	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("dingtalk api error: %v", result["errmsg"])
	}

	return nil
}

// generateDingTalkSign 生成钉钉签名
func (s *NotificationService) generateDingTalkSign(timestamp int64, secret string) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

// sendFeishu 发送飞书通知
func (s *NotificationService) sendFeishu(webhookURL string, content interface{}) error {
	// 构建请求体
	body, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal feishu message: %w", err)
	}

	// 发送请求
	resp, err := s.client.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send feishu notification: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read feishu response: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse feishu response: %w", err)
	}

	// 检查错误码
	if code, ok := result["code"].(float64); ok && code != 0 {
		return fmt.Errorf("feishu api error: %v", result["msg"])
	}

	return nil
}

// sendWechat 发送企业微信通知
func (s *NotificationService) sendWechat(webhookURL, secret string, content interface{}) error {
	// 构建请求体
	body, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal wechat message: %w", err)
	}

	// 发送请求
	resp, err := s.client.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send wechat notification: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read wechat response: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse wechat response: %w", err)
	}

	// 检查错误码
	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("wechat api error: %v", result["errmsg"])
	}

	return nil
}

// BuildDingTalkMarkdownMessage 构建钉钉Markdown消息
func (s *NotificationService) BuildDingTalkMarkdownMessage(title, text string, atMobiles, atUserIDs []string, isAtAll bool) map[string]interface{} {
	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  text,
		},
		"at": map[string]interface{}{
			"atMobiles": atMobiles,
			"atUserIds": atUserIDs,
			"isAtAll":   isAtAll,
		},
	}
}

// BuildFeishuTextMessage 构建飞书文本消息
func (s *NotificationService) BuildFeishuTextMessage(text string) map[string]interface{} {
	return map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": text,
		},
	}
}

// BuildFeishuPostMessage 构建飞书富文本消息
func (s *NotificationService) BuildFeishuPostMessage(title string, content []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   title,
					"content": content,
				},
			},
		},
	}
}

// BuildWechatMarkdownMessage 构建企业微信Markdown消息
func (s *NotificationService) BuildWechatMarkdownMessage(content string) map[string]interface{} {
	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}
}

// BuildWechatTextMessage 构建企业微信文本消息
func (s *NotificationService) BuildWechatTextMessage(content string, mentionedList, mentionedMobileList []string) map[string]interface{} {
	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content":               content,
			"mentioned_list":        mentionedList,
			"mentioned_mobile_list": mentionedMobileList,
		},
	}
}

// BuildAlertMessage 构建告警消息（通用）
func (s *NotificationService) BuildAlertMessage(receiverType, alertID, title, content, severity, status string, metadata map[string]interface{}) (interface{}, error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	switch receiverType {
	case "dingtalk":
		markdown := fmt.Sprintf("#### %s\n", title)
		markdown += fmt.Sprintf("> **告警ID**: %s\n", alertID)
		markdown += fmt.Sprintf("> **严重级别**: %s\n", severity)
		markdown += fmt.Sprintf("> **状态**: %s\n", status)
		markdown += fmt.Sprintf("> **时间**: %s\n", timestamp)
		markdown += fmt.Sprintf("> **详情**:\n%s\n", content)
		return s.BuildDingTalkMarkdownMessage(title, markdown, []string{}, []string{}, false), nil

	case "feishu":
		content := []map[string]interface{}{
			{
				"tag":  "text",
				"text": fmt.Sprintf("%s\n", title),
			},
			{
				"tag":  "text",
				"text": fmt.Sprintf("告警ID: %s\n", alertID),
			},
			{
				"tag":  "text",
				"text": fmt.Sprintf("严重级别: %s\n", severity),
			},
			{
				"tag":  "text",
				"text": fmt.Sprintf("状态: %s\n", status),
			},
			{
				"tag":  "text",
				"text": fmt.Sprintf("时间: %s\n", timestamp),
			},
			{
				"tag":  "text",
				"text": fmt.Sprintf("详情:\n%s\n", content),
			},
		}
		return s.BuildFeishuPostMessage(title, content), nil

	case "wechat":
		markdown := fmt.Sprintf("### %s\n", title)
		markdown += fmt.Sprintf("**告警ID**: %s\n", alertID)
		markdown += fmt.Sprintf("**严重级别**: %s\n", severity)
		markdown += fmt.Sprintf("**状态**: %s\n", status)
		markdown += fmt.Sprintf("**时间**: %s\n", timestamp)
		markdown += fmt.Sprintf("**详情**:\n%s\n", content)
		return s.BuildWechatMarkdownMessage(markdown), nil

	default:
		return nil, fmt.Errorf("unsupported receiver type: %s", receiverType)
	}
}

// SendAlertNotification 发送告警通知
func (s *NotificationService) SendAlertNotification(receiverType, webhookURL, secret, alertID, title, content, severity, status string, metadata map[string]interface{}) error {
	message, err := s.BuildAlertMessage(receiverType, alertID, title, content, severity, status, metadata)
	if err != nil {
		return err
	}
	return s.SendNotification(receiverType, webhookURL, secret, message)
}

// BatchSendAlertNotification 批量发送告警通知
func (s *NotificationService) BatchSendAlertNotification(notifications []NotificationTask) []error {
	errors := make([]error, len(notifications))

	for i, task := range notifications {
		err := s.SendAlertNotification(
			task.ReceiverType,
			task.WebhookURL,
			task.Secret,
			task.AlertID,
			task.Title,
			task.Content,
			task.Severity,
			task.Status,
			task.Metadata,
		)
		errors[i] = err
	}

	return errors
}

// NotificationTask 通知任务
type NotificationTask struct {
	ReceiverType string
	WebhookURL   string
	Secret       string
	AlertID      string
	Title        string
	Content      string
	Severity     string
	Status       string
	Metadata     map[string]interface{}
}

// SendEmail 发送邮件通知
func (s *NotificationService) SendEmail(to, subject, body, severity, status string, metadata map[string]interface{}) error {
	// TODO: 实现邮件发送逻辑
	// 可以使用 SMTP 或调用邮件API
	// 这里提供一个简单实现示例

	// 示例：构建HTML邮件
	_ = s.buildEmailHTML(subject, body, severity, status, metadata)

	// TODO: 实际发送邮件
	// 1. 连接SMTP服务器
	// 2. 构建邮件
	// 3. 发送
	// 4. 处理错误

	return fmt.Errorf("email sending not implemented yet")
}

// buildEmailHTML 构建邮件HTML内容
func (s *NotificationService) buildEmailHTML(subject, body, severity, status string, metadata map[string]interface{}) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f4f4f4; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .severity { display: inline-block; padding: 5px 10px; border-radius: 3px; color: white; }
        .severity-critical { background-color: #d32f2f; }
        .severity-high { background-color: #f57c00; }
        .severity-warning { background-color: #fbc02d; }
        .severity-info { background-color: #1976d2; }
        .footer { background-color: #f4f4f4; padding: 10px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>` + subject + `</h2>
        </div>
        <div class="content">
            <p><strong>告警级别:</strong> <span class="severity severity-` + severity + `">` + severity + `</span></p>
            <p><strong>告警状态:</strong> ` + status + `</p>
            <hr>
            <p>` + body + `</p>
        </div>
        <div class="footer">
            <p>这是一封来自ITCMDB系统的自动告警邮件，请勿直接回复。</p>
        </div>
    </div>
</body>
</html>`
	return html
}
