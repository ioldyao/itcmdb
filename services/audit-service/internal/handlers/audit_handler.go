package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/audit-service/internal/repository"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/response"
)

type AuditHandler struct {
	repo repository.AuditRepository
}

func NewAuditHandler(repo repository.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

// GetAuditLogs 获取审计日志列表
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	// 获取当前用户信息
	currentUserID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(401, response.Error("Unauthorized", "authentication required"))
		return
	}

	// 检查用户是否有 audit:view 权限（查看所有日志）
	roles, _ := auth.GetRoles(c)
	hasAuditViewPermission := false
	for _, role := range roles {
		if role == "audit:view" || role == "*:*" {
			hasAuditViewPermission = true
			break
		}
	}

	// 解析筛选条件
	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := strconv.ParseUint(userIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			userID = &uid
		}
	}

	// 如果用户没有 audit:view 权限，强制只查看自己的日志
	if !hasAuditViewPermission {
		uid := uint(currentUserID)
		userID = &uid
	}

	var action, resource, startTime, endTime *string
	if val := c.Query("action"); val != "" {
		action = &val
	}
	if val := c.Query("resource"); val != "" {
		resource = &val
	}
	if val := c.Query("start_time"); val != "" {
		startTime = &val
	}
	if val := c.Query("end_time"); val != "" {
		endTime = &val
	}

	logs, total, err := h.repo.GetLogs(offset, pageSize, userID, action, resource, startTime, endTime)
	if err != nil {
		c.JSON(500, response.Error("Internal Error", "failed to get audit logs"))
		return
	}

	c.JSON(200, response.Success(gin.H{
		"logs": logs,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	}))
}

// GetAuditStats 获取审计统计
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	// 获取当前用户信息
	currentUserID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(401, response.Error("Unauthorized", "authentication required"))
		return
	}

	// 检查用户是否有 audit:view 权限（查看所有日志统计）
	roles, _ := auth.GetRoles(c)
	hasAuditViewPermission := false
	for _, role := range roles {
		if role == "audit:view" || role == "*:*" {
			hasAuditViewPermission = true
			break
		}
	}

	// 解析时间范围
	var startTime, endTime *string
	if val := c.Query("start_time"); val != "" {
		startTime = &val
	}
	if val := c.Query("end_time"); val != "" {
		endTime = &val
	}

	// 如果用户没有 audit:view 权限，只统计自己的日志
	var userID *uint
	if !hasAuditViewPermission {
		uid := uint(currentUserID)
		userID = &uid
	}

	total, byAction, byResource, err := h.repo.GetStats(startTime, endTime, userID)
	if err != nil {
		c.JSON(500, response.Error("Internal Error", "failed to get audit stats"))
		return
	}

	c.JSON(200, response.Success(gin.H{
		"total_logs":  total,
		"by_action":   byAction,
		"by_resource": byResource,
	}))
}
