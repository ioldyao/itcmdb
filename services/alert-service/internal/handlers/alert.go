package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/alert-service/internal/models"
	"github.com/itcmdb/alert-service/internal/services"
	"github.com/itcmdb/shared/pkg/audit"
	"github.com/itcmdb/shared/pkg/response"
	"gorm.io/gorm"
)

// AlertHandler 告警处理器
type AlertHandler struct {
	db          *gorm.DB
	alertEngine *services.AlertEngine
	vmClient    *services.VictoriaMetricsClient
}

// NewAlertHandler 创建告警处理器
func NewAlertHandler(db *gorm.DB, alertEngine *services.AlertEngine, vmClient *services.VictoriaMetricsClient) *AlertHandler {
	return &AlertHandler{
		db:          db,
		alertEngine: alertEngine,
		vmClient:    vmClient,
	}
}

// GetAlerts 获取告警列表
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	var req models.AlertListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	// 验证和规范化分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页大小，防止性能问题
	}

	// 构建查询
	query := h.db.Model(&models.AlertInstance{})

	// 状态过滤
	if len(req.Status) > 0 {
		query = query.Where("status IN ?", req.Status)
	}

	// 严重程度过滤
	if len(req.Severity) > 0 {
		query = query.Where("severity IN ?", req.Severity)
	}

	// 分类过滤
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	// 时间范围过滤 - 使用正确的字段名 last_triggered
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err == nil {
			query = query.Where("last_triggered >= ?", startTime)
		}
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err == nil {
			query = query.Where("last_triggered <= ?", endTime)
		}
	}

	// 搜索关键词
	if req.SearchKeyword != "" {
		searchPattern := "%" + req.SearchKeyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询总数失败", err.Error()))
		return
	}

	// 排序 - 使用白名单防止SQL注入
	allowedSortFields := map[string]bool{
		"last_triggered":  true,
		"first_triggered": true,
		"severity":        true,
		"status":          true,
		"created_at":      true,
		"updated_at":      true,
	}

	sortField := "last_triggered" // 默认排序字段
	if req.SortField != "" && allowedSortFields[req.SortField] {
		sortField = req.SortField
	}

	direction := "DESC"
	if req.SortOrder == "asc" {
		direction = "ASC"
	}

	query = query.Order(sortField + " " + direction)

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 查询数据
	var alerts []models.AlertInstance
	if err := query.Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, response.Success(models.AlertListResponse{
		Total:  int(total),
		Alerts: alerts,
	}))
}

// GetAlertByID 获取单个告警详情
func (h *AlertHandler) GetAlertByID(c *gin.Context) {
	id := c.Param("id")
	alertID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的告警ID", err.Error()))
		return
	}

	var alert models.AlertInstance
	if err := h.db.Where("id = ?", alertID).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.Error("告警不存在", ""))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, response.Success(alert))
}

// AcknowledgeAlert 确认告警
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	id := c.Param("id")
	alertID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的告警ID", err.Error()))
		return
	}

	var req models.AcknowledgeAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	// 查询告警
	var alert models.AlertInstance
	if err := h.db.Where("id = ?", alertID).First(&alert).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("告警不存在", err.Error()))
		return
	}

	// 检查状态
	if alert.Status != "firing" {
		c.JSON(http.StatusBadRequest, response.Error("只能确认活跃告警", ""))
		return
	}

	// 使用事务更新状态和创建历史记录
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 更新告警状态
		now := time.Now()
		updates := map[string]interface{}{
			"status":          "acknowledged",
			"handler":         req.Handler,
			"handling_notes":  req.Notes,
			"acknowledged_at": &now,
		}

		if err := tx.Model(&alert).Updates(updates).Error; err != nil {
			return err
		}

		// 创建历史记录
		history := models.AlertHistory{
			AlertID:    alert.ID,
			EventType:  "acknowledged",
			OldStatus:  "firing",
			NewStatus:  "acknowledged",
			OperatedBy: req.Handler,
			Message:    "告警已确认",
			OperatedAt: now,
		}
		return tx.Create(&history).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("确认失败", err.Error()))
		return
	}

	auditID := uint(alertID)
	audit.LogSuccess(c, "acknowledge", "alert", &auditID, nil)

	c.JSON(http.StatusOK, response.Success(nil))
}

// CloseAlert 关闭告警
func (h *AlertHandler) CloseAlert(c *gin.Context) {
	id := c.Param("id")
	alertID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的告警ID", err.Error()))
		return
	}

	var req models.CloseAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	// 查询告警
	var alert models.AlertInstance
	if err := h.db.Where("id = ?", alertID).First(&alert).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("告警不存在", err.Error()))
		return
	}

	// 检查状态
	if alert.Status == "closed" {
		c.JSON(http.StatusBadRequest, response.Error("告警已关闭", ""))
		return
	}

	// 使用事务更新状态和创建历史记录
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 更新告警状态
		now := time.Now()
		updates := map[string]interface{}{
			"status":         "closed",
			"handler":        req.Handler,
			"handling_notes": req.Notes,
			"closed_at":      &now,
		}

		if err := tx.Model(&alert).Updates(updates).Error; err != nil {
			return err
		}

		// 创建历史记录
		history := models.AlertHistory{
			AlertID:    alert.ID,
			EventType:  "closed",
			OldStatus:  alert.Status,
			NewStatus:  "closed",
			OperatedBy: req.Handler,
			Message:    "告警已关闭",
			OperatedAt: now,
		}
		return tx.Create(&history).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("关闭失败", err.Error()))
		return
	}

	auditID := uint(alertID)
	audit.LogSuccess(c, "close", "alert", &auditID, nil)

	c.JSON(http.StatusOK, response.Success(nil))
}

// GetAlertHistory 获取告警历史
func (h *AlertHandler) GetAlertHistory(c *gin.Context) {
	id := c.Param("id")
	alertID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无效的告警ID", err.Error()))
		return
	}

	// 查询历史记录
	var history []models.AlertHistory
	if err := h.db.Where("alert_id = ?", alertID).
		Order("operated_at DESC").
		Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(history))
}

// GetAlertStatistics 获取告警统计
func (h *AlertHandler) GetAlertStatistics(c *gin.Context) {
	// 使用单个聚合查询优化性能，避免N+1查询
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	if err := h.db.Model(&models.AlertInstance{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	// 构建响应数据
	stats := struct {
		Total        int64 `json:"total"`
		Firing       int64 `json:"firing"`
		Acknowledged int64 `json:"acknowledged"`
		Resolved     int64 `json:"resolved"`
		Closed       int64 `json:"closed"`
	}{}

	// 从聚合结果中提取数据
	for _, stat := range statusStats {
		stats.Total += stat.Count
		switch stat.Status {
		case "firing":
			stats.Firing = stat.Count
		case "acknowledged":
			stats.Acknowledged = stat.Count
		case "resolved":
			stats.Resolved = stat.Count
		case "closed":
			stats.Closed = stat.Count
		}
	}

	// 严重程度统计
	var severityStats []struct {
		Severity string `json:"severity"`
		Count    int64  `json:"count"`
	}

	if err := h.db.Model(&models.AlertInstance{}).
		Select("severity, count(*) as count").
		Group("severity").
		Scan(&severityStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"stats":          stats,
		"severity_stats": severityStats,
	}))
}

// GetAlertAnalytics 获取告警分析数据
func (h *AlertHandler) GetAlertAnalytics(c *gin.Context) {
	var req models.AlertAnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	// 解析时间范围
	startTime, endTime, err := services.ParseTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("时间范围解析失败", err.Error()))
		return
	}

	// 查询时间范围内的告警 - 使用正确的字段名
	var alerts []models.AlertInstance
	query := h.db.Where("last_triggered >= ? AND last_triggered <= ?", startTime, endTime)

	if err := query.Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("查询失败", err.Error()))
		return
	}

	// 构建响应
	responseData := models.AlertAnalyticsResponse{
		Dimensions: []models.AnalyticsDimension{},
		TimeSeries: models.TimeSeriesData{
			Dates: []string{},
			Series: []models.TimeSeriesSeries{
				{Name: "firing", Data: []int{}},
				{Name: "acknowledged", Data: []int{}},
				{Name: "resolved", Data: []int{}},
				{Name: "closed", Data: []int{}},
			},
		},
	}

	// TODO: 实现按维度分组统计
	// TODO: 实现时间序列数据生成

	c.JSON(http.StatusOK, response.Success(responseData))
}

// BatchAcknowledgeAlerts 批量确认告警
func (h *AlertHandler) BatchAcknowledgeAlerts(c *gin.Context) {
	var req struct {
		AlertIDs []int  `json:"alert_ids" binding:"required"`
		Handler  *int   `json:"handler" binding:"required"`
		Notes    string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":          "acknowledged",
		"handler":         req.Handler,
		"handling_notes":  req.Notes,
		"acknowledged_at": &now,
	}

	// 批量更新
	result := h.db.Model(&models.AlertInstance{}).
		Where("id IN ? AND status = ?", req.AlertIDs, "firing").
		Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("批量确认失败", result.Error.Error()))
		return
	}

	audit.LogSuccess(c, "batch_acknowledge", "alert", nil, map[string]interface{}{
		"alert_ids":    req.AlertIDs,
		"affected_rows": result.RowsAffected,
	})

	c.JSON(http.StatusOK, response.Success(gin.H{
		"affected_rows": result.RowsAffected,
	}))
}

// BatchCloseAlerts 批量关闭告警
func (h *AlertHandler) BatchCloseAlerts(c *gin.Context) {
	var req struct {
		AlertIDs []int  `json:"alert_ids" binding:"required"`
		Handler  *int   `json:"handler" binding:"required"`
		Notes    string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("参数错误", err.Error()))
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":    "closed",
		"handler":   req.Handler,
		"handling_notes": req.Notes,
		"closed_at": &now,
	}

	// 批量更新
	result := h.db.Model(&models.AlertInstance{}).
		Where("id IN ? AND status != ?", req.AlertIDs, "closed").
		Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("批量关闭失败", result.Error.Error()))
		return
	}

	audit.LogSuccess(c, "batch_close", "alert", nil, map[string]interface{}{
		"alert_ids":    req.AlertIDs,
		"affected_rows": result.RowsAffected,
	})

	c.JSON(http.StatusOK, response.Success(gin.H{
		"affected_rows": result.RowsAffected,
	}))
}
