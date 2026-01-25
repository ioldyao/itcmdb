package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/itcmdb/cmdb-service/internal/models"
	"github.com/itcmdb/cmdb-service/internal/repository"
	kafkapkg "github.com/itcmdb/shared/pkg/kafka"
)

type CIService interface {
	// CI Types
	GetCITypes() ([]models.CIType, error)
	GetCITypeByID(id uint) (*models.CIType, error)

	// CI Instances
	GetCIInstances(ciTypeID uint, filters map[string]interface{}, page, pageSize int) ([]models.CIInstance, int64, error)
	GetCIInstanceByID(id uint) (*models.CIInstance, error)
	CreateCIInstance(req *CreateCIInstanceRequest, userID uint) (*models.CIInstance, error)
	UpdateCIInstance(id uint, req *UpdateCIInstanceRequest, userID uint) (*models.CIInstance, error)
	DeleteCIInstance(id uint, userID uint) error

	// CI Relations
	GetCIRelations(ciID uint) ([]models.CIRelation, error)
	CreateCIRelation(req *CreateCIRelationRequest, userID uint) (*models.CIRelation, error)

	// CI History
	GetCIHistory(ciID uint, limit int) ([]models.CIHistory, error)

	// Import/Export
	ExportCIInstances(ciTypeID uint) ([]byte, error)
	ImportCIInstances(ciTypeID uint, data []byte, userID uint) (*ImportResult, error)
}

type ciService struct {
	repo repository.CIRepository
}

func NewCIService(repo repository.CIRepository) CIService {
	return &ciService{repo: repo}
}

// Request DTOs
type CreateCIInstanceRequest struct {
	CITypeID   uint                   `json:"ci_type_id" binding:"required"`
	Name       string                 `json:"name" binding:"required"`
	Status     string                 `json:"status"`
	Attributes map[string]interface{} `json:"attributes"`
	Tags       map[string]interface{} `json:"tags"`
}

type UpdateCIInstanceRequest struct {
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	Attributes map[string]interface{} `json:"attributes"`
	Tags       map[string]interface{} `json:"tags"`
}

type CreateCIRelationRequest struct {
	ParentID     uint   `json:"parent_id" binding:"required"`
	ChildID      uint   `json:"child_id" binding:"required"`
	RelationType string `json:"relation_type" binding:"required"`
	Description  string `json:"description"`
}

type ImportResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

// CI Types

func (s *ciService) GetCITypes() ([]models.CIType, error) {
	return s.repo.GetCITypes()
}

func (s *ciService) GetCITypeByID(id uint) (*models.CIType, error) {
	return s.repo.GetCITypeByID(id)
}

// CI Instances

func (s *ciService) GetCIInstances(ciTypeID uint, filters map[string]interface{}, page, pageSize int) ([]models.CIInstance, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.GetCIInstances(ciTypeID, filters, page, pageSize)
}

func (s *ciService) GetCIInstanceByID(id uint) (*models.CIInstance, error) {
	return s.repo.GetCIInstanceByID(id)
}

func (s *ciService) CreateCIInstance(req *CreateCIInstanceRequest, userID uint) (*models.CIInstance, error) {
	// 验证 CI 类型是否存在
	ciType, err := s.repo.GetCITypeByID(req.CITypeID)
	if err != nil {
		return nil, errors.New("CI type not found")
	}
	if !ciType.IsActive {
		return nil, errors.New("CI type is not active")
	}

	// 验证必填属性
	if err := s.validateAttributes(req.Attributes, ciType.Attributes); err != nil {
		return nil, err
	}

	instance := &models.CIInstance{
		CITypeID:   req.CITypeID,
		Name:       req.Name,
		Status:     req.Status,
		Attributes: req.Attributes,
		Tags:       req.Tags,
		CreatedBy:  userID,
		UpdatedBy:  userID,
	}

	if instance.Status == "" {
		instance.Status = "active"
	}

	if err := s.repo.CreateCIInstance(instance); err != nil {
		return nil, err
	}

	// 记录历史
	s.recordHistory(instance.ID, userID, "create", "", "", "")

	// 发布CI创建事件
	go func() {
		eventData := map[string]interface{}{
			"ci_id":      instance.ID,
			"ci_type_id": instance.CITypeID,
			"ci_name":    instance.Name,
			"action":     "create",
			"changed_by": userID,
			"new_value": map[string]interface{}{
				"name":       instance.Name,
				"status":     instance.Status,
				"attributes": instance.Attributes,
			},
		}
		kafkapkg.PublishCIEvent(kafkapkg.EventCIChanged, eventData)
	}()

	return instance, nil
}

func (s *ciService) UpdateCIInstance(id uint, req *UpdateCIInstanceRequest, userID uint) (*models.CIInstance, error) {
	instance, err := s.repo.GetCIInstanceByID(id)
	if err != nil {
		return nil, errors.New("CI instance not found")
	}

	// 记录变更
	oldData, _ := json.Marshal(instance)

	if req.Name != "" && req.Name != instance.Name {
		s.recordHistory(id, userID, "update", "name", instance.Name, req.Name)
		instance.Name = req.Name
	}

	if req.Status != "" && req.Status != instance.Status {
		s.recordHistory(id, userID, "update", "status", instance.Status, req.Status)
		instance.Status = req.Status
	}

	if req.Attributes != nil {
		// 验证属性
		if err := s.validateAttributes(req.Attributes, instance.CIType.Attributes); err != nil {
			return nil, err
		}

		// 比较新旧 attributes 是否真的发生了变化
		oldAttrs, _ := json.Marshal(instance.Attributes)
		newAttrs, _ := json.Marshal(req.Attributes)

		// 只有在 attributes 真正发生变化时才记录历史
		if string(oldAttrs) != string(newAttrs) {
			// 计算变化的字段数量
			changedFields := s.getChangedAttributes(instance.Attributes, req.Attributes)

			if len(changedFields) > 0 {
				// 记录变化的字段（简化显示，不记录完整 JSON）
				fieldList := strings.Join(changedFields, ", ")
				s.recordHistory(id, userID, "update", "attributes", "-", fieldList)
			}
		}

		instance.Attributes = req.Attributes
	}

	if req.Tags != nil {
		instance.Tags = req.Tags
	}

	instance.UpdatedBy = userID

	if err := s.repo.UpdateCIInstance(instance); err != nil {
		return nil, err
	}

	// 发布CI更新事件
	go func() {
		var oldValue, newValue map[string]interface{}
		json.Unmarshal(oldData, &oldValue)

		newData, _ := json.Marshal(instance)
		json.Unmarshal(newData, &newValue)

		eventData := map[string]interface{}{
			"ci_id":      instance.ID,
			"ci_type_id": instance.CITypeID,
			"ci_name":    instance.Name,
			"action":     "update",
			"changed_by": userID,
			"old_value":  oldValue,
			"new_value":  newValue,
		}
		kafkapkg.PublishCIEvent(kafkapkg.EventCIChanged, eventData)
	}()

	return instance, nil
}

func (s *ciService) DeleteCIInstance(id uint, userID uint) error {
	instance, err := s.repo.GetCIInstanceByID(id)
	if err != nil {
		return errors.New("CI instance not found")
	}

	// 记录删除历史
	s.recordHistory(id, userID, "delete", "", instance.Name, "")

	if err := s.repo.DeleteCIInstance(id); err != nil {
		return err
	}

	// 发布CI删除事件
	go func() {
		eventData := map[string]interface{}{
			"ci_id":      instance.ID,
			"ci_type_id": instance.CITypeID,
			"ci_name":    instance.Name,
			"action":     "delete",
			"changed_by": userID,
		}
		kafkapkg.PublishCIEvent(kafkapkg.EventCIDeleted, eventData)
	}()

	return nil
}

// CI Relations

func (s *ciService) GetCIRelations(ciID uint) ([]models.CIRelation, error) {
	return s.repo.GetCIRelations(ciID)
}

func (s *ciService) CreateCIRelation(req *CreateCIRelationRequest, userID uint) (*models.CIRelation, error) {
	// 验证父子CI是否存在
	if _, err := s.repo.GetCIInstanceByID(req.ParentID); err != nil {
		return nil, errors.New("parent CI not found")
	}
	if _, err := s.repo.GetCIInstanceByID(req.ChildID); err != nil {
		return nil, errors.New("child CI not found")
	}

	relation := &models.CIRelation{
		ParentID:     req.ParentID,
		ChildID:      req.ChildID,
		RelationType: req.RelationType,
		Description:  req.Description,
		CreatedBy:    userID,
	}

	if err := s.repo.CreateCIRelation(relation); err != nil {
		return nil, err
	}

	// 发布CI关系变更事件
	go func() {
		eventData := map[string]interface{}{
			"relation_id":   relation.ID,
			"parent_id":     relation.ParentID,
			"child_id":      relation.ChildID,
			"relation_type": relation.RelationType,
			"action":        "create",
			"changed_by":    userID,
		}
		kafkapkg.PublishCIEvent(kafkapkg.EventCIRelationshipChanged, eventData)
	}()

	return relation, nil
}

// CI History

func (s *ciService) GetCIHistory(ciID uint, limit int) ([]models.CIHistory, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.repo.GetCIHistory(ciID, limit)
}

// Helper functions

// getChangedAttributes 比较两个 attributes 并返回变化的字段列表
func (s *ciService) getChangedAttributes(oldAttrs, newAttrs map[string]interface{}) []string {
	changedFields := []string{}

	// 忽略的字段（不记录为变更）
	ignoredFields := map[string]bool{
		"last_hardware_report": true, // 最后上报时间，每次上报都会更新
	}

	// 实时变化的子字段（比较时忽略）
	ignoreSubFields := map[string][]string{
		"optical_modules_info": {"temperature"}, // 光模块温度是实时数据
	}

	// 检查新 attrs 中有哪些字段与旧 attrs 不同
	for key, newValue := range newAttrs {
		// 跳过忽略的字段
		if ignoredFields[key] {
			continue
		}

		oldValue, exists := oldAttrs[key]
		if !exists {
			// 新增的字段
			changedFields = append(changedFields, key+"(新增)")
			continue
		}

		// 比较值是否相同
		oldJSON, _ := json.Marshal(oldValue)
		newJSON, _ := json.Marshal(newValue)

		// 如果是直接比较不同，检查是否因为有实时数据字段导致
		if string(oldJSON) != string(newJSON) {
			// 如果该字段有需要忽略的子字段（如 optical_modules_info 的 temperature）
			if subFieldsToIgnore, ok := ignoreSubFields[key]; ok {
				// 清理实时数据后再比较
				cleanedOld := s.cleanRealtimeData(oldValue, subFieldsToIgnore)
				cleanedNew := s.cleanRealtimeData(newValue, subFieldsToIgnore)
				cleanedOldJSON, _ := json.Marshal(cleanedOld)
				cleanedNewJSON, _ := json.Marshal(cleanedNew)

				// 只有在清理后仍然不同才认为是真正的变化
				if string(cleanedOldJSON) != string(cleanedNewJSON) {
					changedFields = append(changedFields, key)
				}
			} else {
				// 没有需要忽略的子字段，直接记录变化
				changedFields = append(changedFields, key)
			}
		}
	}

	// 检查是否有字段被删除（同样忽略特定字段）
	for key := range oldAttrs {
		if ignoredFields[key] {
			continue
		}
		if _, exists := newAttrs[key]; !exists {
			changedFields = append(changedFields, key+"(删除)")
		}
	}

	return changedFields
}

// cleanRealtimeData 清理实时数据字段（用于比较）
func (s *ciService) cleanRealtimeData(data interface{}, fieldsToClean []string) interface{} {
	// 如果是数组，清理每个元素的指定字段
	if arr, ok := data.([]interface{}); ok {
		result := []interface{}{}
		for _, item := range arr {
			if itemMap, ok := item.(map[string]interface{}); ok {
				cleaned := make(map[string]interface{})
				for k, v := range itemMap {
					// 跳过需要清理的字段
					shouldClean := false
					for _, fieldToClean := range fieldsToClean {
						if k == fieldToClean {
							shouldClean = true
							break
						}
					}
					if !shouldClean {
						cleaned[k] = v
					}
				}
				result = append(result, cleaned)
			} else {
				result = append(result, item)
			}
		}
		return result
	}

	return data
}

func (s *ciService) validateAttributes(attrs map[string]interface{}, definitions []models.CIAttribute) error {
	// 检查必填属性
	for _, def := range definitions {
		if def.IsRequired {
			if _, exists := attrs[def.Name]; !exists {
				return errors.New("required attribute missing: " + def.DisplayName)
			}
		}
	}
	return nil
}

func (s *ciService) recordHistory(ciID, userID uint, action, fieldName, oldValue, newValue string) {
	// JSON-encode values if they're not already valid JSON
	oldValueJSON := s.ensureJSON(oldValue)
	newValueJSON := s.ensureJSON(newValue)

	history := &models.CIHistory{
		CIID:      ciID,
		ChangedBy: userID,
		Action:    action,
		FieldName: fieldName,
		OldValue:  oldValueJSON,
		NewValue:  newValueJSON,
		ChangedAt: time.Now(),
	}
	s.repo.CreateCIHistory(history)
}

// ensureJSON checks if a string is valid JSON, if not, wraps it as a JSON string
func (s *ciService) ensureJSON(value string) string {
	if value == "" {
		return "null"
	}

	// Check if it's already valid JSON
	var js json.RawMessage
	if err := json.Unmarshal([]byte(value), &js); err == nil {
		return value
	}

	// Not valid JSON, encode it as a JSON string
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

// Import/Export

func (s *ciService) ExportCIInstances(ciTypeID uint) ([]byte, error) {
	// 获取所有CI实例
	instances, _, err := s.repo.GetCIInstances(ciTypeID, nil, 1, 10000)
	if err != nil {
		return nil, err
	}

	// 转换为JSON格式
	data, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *ciService) ImportCIInstances(ciTypeID uint, data []byte, userID uint) (*ImportResult, error) {
	result := &ImportResult{
		Success: 0,
		Failed:  0,
		Errors:  []string{},
	}

	// 解析JSON数据
	var instances []map[string]interface{}
	if err := json.Unmarshal(data, &instances); err != nil {
		result.Errors = append(result.Errors, "Invalid JSON format: "+err.Error())
		return result, err
	}

	// 验证CI类型
	ciType, err := s.repo.GetCITypeByID(ciTypeID)
	if err != nil {
		result.Errors = append(result.Errors, "CI type not found")
		return result, err
	}

	// 逐个导入CI实例
	for i, instData := range instances {
		name, ok := instData["name"].(string)
		if !ok || name == "" {
			result.Failed++
			result.Errors = append(result.Errors, "Row "+string(rune(i+1))+": missing or invalid name")
			continue
		}

		status, _ := instData["status"].(string)
		if status == "" {
			status = "active"
		}

		attributes, _ := instData["attributes"].(map[string]interface{})
		tags, _ := instData["tags"].(map[string]interface{})

		// 验证必填属性
		if err := s.validateAttributes(attributes, ciType.Attributes); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "Row "+string(rune(i+1))+": "+err.Error())
			continue
		}

		// 创建CI实例
		instance := &models.CIInstance{
			CITypeID:   ciTypeID,
			Name:       name,
			Status:     status,
			Attributes: attributes,
			Tags:       tags,
			CreatedBy:  userID,
			UpdatedBy:  userID,
		}

		if err := s.repo.CreateCIInstance(instance); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "Row "+string(rune(i+1))+": "+err.Error())
			continue
		}

		// 记录历史
		s.recordHistory(instance.ID, userID, "create", "", "", "")
		result.Success++
	}

	return result, nil
}
