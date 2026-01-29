package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// WebhookSignature Webhook签名验证
type WebhookSignature struct {
	secret string
}

// NewWebhookSignature 创建签名验证器
func NewWebhookSignature(secret string) *WebhookSignature {
	return &WebhookSignature{secret: secret}
}

// GenerateSignature 生成签名
func (ws *WebhookSignature) GenerateSignature(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(ws.secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature 验证签名
func (ws *WebhookSignature) VerifySignature(payload []byte, signature string) error {
	expectedSignature := ws.GenerateSignature(payload)

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// GetSignatureHeader 获取签名头名称
func (ws *WebhookSignature) GetSignatureHeader() string {
	return "X-Webhook-Signature"
}
