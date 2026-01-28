package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/alert-service/internal/services"
)

// ReceiverHandler 接收人处理器
type ReceiverHandler struct {
	db *gorm.DB
}

// NewReceiverHandler 创建接收人处理器
func NewReceiverHandler(db *gorm.DB) *ReceiverHandler {
	return &ReceiverHandler{
		db: db,
	}
}

// ListReceivers 获取接收人列表
func (h *ReceiverHandler) ListReceivers(c *gin.Context) {
	var receivers []models.AlertReceiver
	var total int64

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	receiverType := c.Query("type")
	enabled := c.Query("enabled")

	query := h.db.Model(&models.AlertReceiver{})

	// 筛选条件
	if receiverType != "" {
		query = query.Where("type = ?", receiverType)
	}
	if enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Find(&receivers)

	c.JSON(http.StatusOK, models.ReceiverListResponse{
		Total:      int(total),
		Receivers: receivers,
	})
}

// GetReceiver 获取接收人详情
func (h *ReceiverHandler) GetReceiver(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	var receiver models.AlertReceiver
	if err := h.db.First(&receiver, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, receiver)
}

// CreateReceiver 创建接收人
func (h *ReceiverHandler) CreateReceiver(c *gin.Context) {
	var req models.CreateReceiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	receiver := models.AlertReceiver{
		Name:       req.Name,
		Type:       req.Type,
		WebhookURL: req.WebhookURL,
		AtMobiles:  req.AtMobiles,
		AtUserIDs:  req.AtUserIDs,
		Secret:     req.Secret,
		Config:     req.Config,
		Enabled:    true,
	}

	if err := h.db.Create(&receiver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, receiver)
}

// UpdateReceiver 更新接收人
func (h *ReceiverHandler) UpdateReceiver(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	var receiver models.AlertReceiver
	if err := h.db.First(&receiver, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var req models.UpdateReceiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != nil {
		receiver.Name = *req.Name
	}
	if req.WebhookURL != nil {
		receiver.WebhookURL = *req.WebhookURL
	}
	if req.AtMobiles != nil {
		receiver.AtMobiles = *req.AtMobiles
	}
	if req.AtUserIDs != nil {
		receiver.AtUserIDs = *req.AtUserIDs
	}
	if req.Secret != nil {
		receiver.Secret = *req.Secret
	}
	if req.Config != nil {
		receiver.Config = req.Config
	}
	if req.Enabled != nil {
		receiver.Enabled = *req.Enabled
	}

	if err := h.db.Save(&receiver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, receiver)
}

// DeleteReceiver 删除接收人
func (h *ReceiverHandler) DeleteReceiver(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	if err := h.db.Delete(&models.AlertReceiver{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Receiver deleted successfully"})
}

// TestReceiver 测试接收人
func (h *ReceiverHandler) TestReceiver(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	var receiver models.AlertReceiver
	if err := h.db.First(&receiver, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 创建通知服务
	notificationService := services.NewNotificationService()

	// 发送测试消息
	testTitle := "告警测试通知"
	testContent := "这是一条测试告警消息，如果您看到此消息，说明接收人配置正确！"
	err = notificationService.SendAlertNotification(
		receiver.Type,
		receiver.WebhookURL,
		receiver.Secret,
		"test-001",
		testTitle,
		testContent,
		"info",
		"test",
		nil,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Test notification sent successfully",
	})
}

// ReceiverGroupHandler 接收组处理器
type ReceiverGroupHandler struct {
	db *gorm.DB
}

// NewReceiverGroupHandler 创建接收组处理器
func NewReceiverGroupHandler(db *gorm.DB) *ReceiverGroupHandler {
	return &ReceiverGroupHandler{
		db: db,
	}
}

// ListReceiverGroups 获取接收组列表
func (h *ReceiverGroupHandler) ListReceiverGroups(c *gin.Context) {
	var groups []models.AlertReceiverGroup
	var total int64

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	enabled := c.Query("enabled")

	query := h.db.Model(&models.AlertReceiverGroup{})

	// 筛选条件
	if enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}

	// 计算总数
	query.Count(&total)

	// 分页查询（预加载接收人）
	offset := (page - 1) * pageSize
	query.Preload("Receivers").Offset(offset).Limit(pageSize).Find(&groups)

	c.JSON(http.StatusOK, models.ReceiverGroupListResponse{
		Total:  int(total),
		Groups: groups,
	})
}

// GetReceiverGroup 获取接收组详情
func (h *ReceiverGroupHandler) GetReceiverGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group models.AlertReceiverGroup
	if err := h.db.Preload("Receivers").First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, group)
}

// CreateReceiverGroup 创建接收组
func (h *ReceiverGroupHandler) CreateReceiverGroup(c *gin.Context) {
	var req models.CreateReceiverGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := models.AlertReceiverGroup{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     true,
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建接收组
	if err := tx.Create(&group).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 关联接收人
	if len(req.ReceiverIDs) > 0 {
		var members []models.AlertReceiverGroupMember
		for _, receiverID := range req.ReceiverIDs {
			members = append(members, models.AlertReceiverGroupMember{
				GroupID:    group.ID,
				ReceiverID: receiverID,
			})
		}
		if err := tx.Create(&members).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载包含接收人的数据
	h.db.Preload("Receivers").First(&group, group.ID)

	c.JSON(http.StatusCreated, group)
}

// UpdateReceiverGroup 更新接收组
func (h *ReceiverGroupHandler) UpdateReceiverGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group models.AlertReceiverGroup
	if err := h.db.First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var req models.UpdateReceiverGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新字段
	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	if req.Enabled != nil {
		group.Enabled = *req.Enabled
	}

	if err := tx.Save(&group).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新接收人关联
	if req.ReceiverIDs != nil {
		// 删除旧的关联
		tx.Where("group_id = ?", group.ID).Delete(&models.AlertReceiverGroupMember{})

		// 创建新的关联
		if len(req.ReceiverIDs) > 0 {
			var members []models.AlertReceiverGroupMember
			for _, receiverID := range req.ReceiverIDs {
				members = append(members, models.AlertReceiverGroupMember{
					GroupID:    group.ID,
					ReceiverID: receiverID,
				})
			}
			if err := tx.Create(&members).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载包含接收人的数据
	h.db.Preload("Receivers").First(&group, group.ID)

	c.JSON(http.StatusOK, group)
}

// DeleteReceiverGroup 删除接收组
func (h *ReceiverGroupHandler) DeleteReceiverGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	if err := h.db.Delete(&models.AlertReceiverGroup{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// AlertReceiverGroupMember 接收组成员关联表（用于gorm）
type AlertReceiverGroupMember struct {
	ID         int `gorm:"primaryKey"`
	GroupID    int
	ReceiverID int
	CreatedAt  string
}
