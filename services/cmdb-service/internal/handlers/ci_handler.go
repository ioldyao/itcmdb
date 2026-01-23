package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/cmdb-service/internal/service"
	"github.com/itcmdb/shared/pkg/response"
)

type CIHandler struct {
	ciService service.CIService
}

func NewCIHandler(ciService service.CIService) *CIHandler {
	return &CIHandler{ciService: ciService}
}

// GetCITypes 获取CI类型列表
func (h *CIHandler) GetCITypes(c *gin.Context) {
	types, err := h.ciService.GetCITypes()
	if err != nil {
		c.JSON(500, response.Error("Failed to get CI types", err.Error()))
		return
	}
	c.JSON(200, response.Success(types))
}

// GetCIInstances 获取CI实例列表
func (h *CIHandler) GetCIInstances(c *gin.Context) {
	ciTypeID, _ := strconv.Atoi(c.Query("ci_type_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	instances, total, err := h.ciService.GetCIInstances(uint(ciTypeID), filters, page, pageSize)
	if err != nil {
		c.JSON(500, response.Error("Failed to get CI instances", err.Error()))
		return
	}

	c.JSON(200, response.SuccessWithPagination(instances, total, page, pageSize))
}

// GetCIInstance 获取单个CI实例
func (h *CIHandler) GetCIInstance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, response.Error("Invalid CI ID", err.Error()))
		return
	}

	instance, err := h.ciService.GetCIInstanceByID(uint(id))
	if err != nil {
		c.JSON(404, response.Error("CI instance not found", err.Error()))
		return
	}

	c.JSON(200, response.Success(instance))
}

// CreateCIInstance 创建CI实例
func (h *CIHandler) CreateCIInstance(c *gin.Context) {
	var req service.CreateCIInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("Invalid request", err.Error()))
		return
	}

	// 从JWT中获取用户ID
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		uid = 1 // fallback
	}

	instance, err := h.ciService.CreateCIInstance(&req, uid)
	if err != nil {
		c.JSON(400, response.Error("Failed to create CI instance", err.Error()))
		return
	}

	c.JSON(201, response.Success(instance))
}

// UpdateCIInstance 更新CI实例
func (h *CIHandler) UpdateCIInstance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, response.Error("Invalid CI ID", err.Error()))
		return
	}

	var req service.UpdateCIInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("Invalid request", err.Error()))
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		uid = 1
	}

	instance, err := h.ciService.UpdateCIInstance(uint(id), &req, uid)
	if err != nil {
		c.JSON(400, response.Error("Failed to update CI instance", err.Error()))
		return
	}

	c.JSON(200, response.Success(instance))
}

// DeleteCIInstance 删除CI实例
func (h *CIHandler) DeleteCIInstance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, response.Error("Invalid CI ID", err.Error()))
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		uid = 1
	}

	if err := h.ciService.DeleteCIInstance(uint(id), uid); err != nil {
		c.JSON(400, response.Error("Failed to delete CI instance", err.Error()))
		return
	}

	c.JSON(200, response.Success(nil))
}

// GetCIRelations 获取CI关系
func (h *CIHandler) GetCIRelations(c *gin.Context) {
	ciID, err := strconv.ParseUint(c.Query("ci_id"), 10, 32)
	if err != nil {
		c.JSON(400, response.Error("Invalid CI ID", err.Error()))
		return
	}

	relations, err := h.ciService.GetCIRelations(uint(ciID))
	if err != nil {
		c.JSON(500, response.Error("Failed to get CI relations", err.Error()))
		return
	}

	c.JSON(200, response.Success(relations))
}

// CreateCIRelation 创建CI关系
func (h *CIHandler) CreateCIRelation(c *gin.Context) {
	var req service.CreateCIRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error("Invalid request", err.Error()))
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		uid = 1
	}

	relation, err := h.ciService.CreateCIRelation(&req, uid)
	if err != nil {
		c.JSON(400, response.Error("Failed to create CI relation", err.Error()))
		return
	}

	c.JSON(201, response.Success(relation))
}

// GetCIHistory 获取CI变更历史
func (h *CIHandler) GetCIHistory(c *gin.Context) {
	ciID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, response.Error("Invalid CI ID", err.Error()))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	history, err := h.ciService.GetCIHistory(uint(ciID), limit)
	if err != nil {
		c.JSON(500, response.Error("Failed to get CI history", err.Error()))
		return
	}

	c.JSON(200, response.Success(history))
}
