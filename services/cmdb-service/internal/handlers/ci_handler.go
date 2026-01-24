package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/cmdb-service/internal/service"
	"github.com/itcmdb/shared/pkg/audit"
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
		audit.LogError(c, "create", "ci_instances", nil, err.Error(), req)
		c.JSON(400, response.Error("Failed to create CI instance", err.Error()))
		return
	}

	audit.LogSuccess(c, "create", "ci_instances", &instance.ID, map[string]interface{}{
		"ci_type_id": req.CITypeID,
		"ci_name":    req.Name,
		"status":     instance.Status,
	})
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

	instanceID := uint(id)
	instance, err := h.ciService.UpdateCIInstance(instanceID, &req, uid)
	if err != nil {
		audit.LogError(c, "update", "ci_instances", &instanceID, err.Error(), req)
		c.JSON(400, response.Error("Failed to update CI instance", err.Error()))
		return
	}

	audit.LogSuccess(c, "update", "ci_instances", &instance.ID, map[string]interface{}{
		"ci_name": req.Name,
		"status":  instance.Status,
	})
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

	instanceID := uint(id)
	if err := h.ciService.DeleteCIInstance(instanceID, uid); err != nil {
		audit.LogError(c, "delete", "ci_instances", &instanceID, err.Error(), nil)
		c.JSON(400, response.Error("Failed to delete CI instance", err.Error()))
		return
	}

	audit.LogSuccess(c, "delete", "ci_instances", &instanceID, map[string]interface{}{
		"ci_instance_id": instanceID,
	})
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

// ExportCIInstances 导出CI实例
func (h *CIHandler) ExportCIInstances(c *gin.Context) {
	ciTypeID, _ := strconv.Atoi(c.Query("ci_type_id"))

	data, err := h.ciService.ExportCIInstances(uint(ciTypeID))
	if err != nil {
		c.JSON(500, response.Error("Failed to export CI instances", err.Error()))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=ci_instances.json")
	c.Data(200, "application/json", data)
}

// ImportCIInstances 导入CI实例
func (h *CIHandler) ImportCIInstances(c *gin.Context) {
	ciTypeID, _ := strconv.Atoi(c.Query("ci_type_id"))
	if ciTypeID == 0 {
		c.JSON(400, response.Error("Invalid request", "ci_type_id is required"))
		return
	}

	// 读取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, response.Error("Invalid request", "file is required"))
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(500, response.Error("Failed to open file", err.Error()))
		return
	}
	defer src.Close()

	// 读取文件内容
	data := make([]byte, file.Size)
	if _, err := src.Read(data); err != nil {
		c.JSON(500, response.Error("Failed to read file", err.Error()))
		return
	}

	// 获取用户ID
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		uid = 1
	}

	// 导入数据
	result, err := h.ciService.ImportCIInstances(uint(ciTypeID), data, uid)
	if err != nil {
		c.JSON(400, response.Error("Failed to import CI instances", err.Error()))
		return
	}

	c.JSON(200, response.Success(result))
}
