package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/webhook-demo/internal/models"
	"github.com/webhook-demo/internal/services"
)

// WebhookHandler 处理GitHub webhook请求
type WebhookHandler struct {
	eventProcessor *services.EventProcessor
	webhookSecret  string
}

// NewWebhookHandler 创建新的webhook处理器
func NewWebhookHandler(eventProcessor *services.EventProcessor, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		eventProcessor: eventProcessor,
		webhookSecret:  webhookSecret,
	}
}

// HandleWebhook 处理GitHub webhook请求
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	// 获取请求头信息
	eventType := c.GetHeader("X-GitHub-Event")
	deliveryID := c.GetHeader("X-GitHub-Delivery")
	signature := c.GetHeader("X-Hub-Signature-256")

	log.Printf("收到GitHub事件: Type=%s, DeliveryID=%s", eventType, deliveryID)

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
		return
	}

	// 验证签名
	if !h.verifySignature(signature, body) {
		log.Printf("签名验证失败: DeliveryID=%s", deliveryID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "签名验证失败"})
		return
	}

	log.Printf("签名验证成功: DeliveryID=%s", deliveryID)

	// 创建事件对象
	event := &models.GitHubEvent{
		Type:       eventType,
		DeliveryID: deliveryID,
		Payload:    body,
	}

	// 处理事件
	if err := h.eventProcessor.ProcessEvent(event); err != nil {
		log.Printf("处理事件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理事件失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":     "事件处理成功",
		"event_type":  eventType,
		"delivery_id": deliveryID,
	})
}

// verifySignature 验证GitHub webhook签名
func (h *WebhookHandler) verifySignature(signature string, payload []byte) bool {
	if h.webhookSecret == "" {
		log.Println("警告: webhook密钥未设置，跳过签名验证")
		return true
	}

	if signature == "" {
		log.Println("签名为空")
		return false
	}

	// GitHub签名格式: sha256=<hex>
	if !strings.HasPrefix(signature, "sha256=") {
		log.Println("签名格式错误")
		return false
	}

	// 提取十六进制签名
	expectedSignature := strings.TrimPrefix(signature, "sha256=")

	// 计算HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	calculatedSignature := hex.EncodeToString(mac.Sum(nil))

	// 比较签名
	return hmac.Equal([]byte(expectedSignature), []byte(calculatedSignature))
}

// verifySignatureConstantTime 使用常数时间比较的签名验证（推荐用于生产环境）
func (h *WebhookHandler) verifySignatureConstantTime(signature string, payload []byte) bool {
	if h.webhookSecret == "" {
		return true
	}

	if signature == "" {
		return false
	}

	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedSignature := strings.TrimPrefix(signature, "sha256=")
	expectedBytes, err := hex.DecodeString(expectedSignature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	calculatedBytes := mac.Sum(nil)

	return hmac.Equal(expectedBytes, calculatedBytes)
}
