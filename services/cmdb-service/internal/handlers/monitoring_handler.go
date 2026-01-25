package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/cmdb-service/internal/service"
	"github.com/itcmdb/shared/pkg/response"
)

type MonitoringHandler struct {
	monitoringService service.MonitoringService
}

func NewMonitoringHandler(monitoringService service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
	}
}

// GetContainerStats 获取容器监控数据
// @Summary 获取容器监控数据
// @Description 通过cAdvisor获取容器的实时监控数据（CPU、内存、网络、磁盘等）
// @Tags Monitoring
// @Param id path int true "CI实例ID"
// @Success 200 {object} response.Response{data=cadvisor.ContainerStats}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/monitoring/containers/{id}/stats [get]
func (h *MonitoringHandler) GetContainerStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, response.Error("Invalid CI ID", err.Error()))
		return
	}

	stats, err := h.monitoringService.GetContainerStats(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(500, response.Error("Failed to get container stats", err.Error()))
		return
	}

	c.JSON(200, response.Success(stats))
}

// HealthCheckCAdvisor 检查cAdvisor服务健康状态
// @Summary 检查cAdvisor健康状态
// @Description 检查指定cAdvisor端点是否可用
// @Tags Monitoring
// @Param endpoint query string true "cAdvisor端点URL"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/monitoring/cadvisor/health [get]
func (h *MonitoringHandler) HealthCheckCAdvisor(c *gin.Context) {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		c.JSON(400, response.Error("Invalid request", "endpoint parameter is required"))
		return
	}

	if err := h.monitoringService.HealthCheckCAdvisor(c.Request.Context(), endpoint); err != nil {
		c.JSON(500, response.Error("cAdvisor health check failed", err.Error()))
		return
	}

	c.JSON(200, response.Success(map[string]interface{}{
		"status":   "healthy",
		"endpoint": endpoint,
	}))
}

// HealthCheckVictoriaMetrics 检查VictoriaMetrics服务健康状态
// @Summary 检查VictoriaMetrics健康状态
// @Description 检查VictoriaMetrics服务是否可用
// @Tags Monitoring
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/monitoring/victoriametrics/health [get]
func (h *MonitoringHandler) HealthCheckVictoriaMetrics(c *gin.Context) {
	if err := h.monitoringService.HealthCheckVictoriaMetrics(c.Request.Context()); err != nil {
		c.JSON(500, response.Error("VictoriaMetrics health check failed", err.Error()))
		return
	}

	c.JSON(200, response.Success(map[string]interface{}{
		"status": "healthy",
	}))
}
