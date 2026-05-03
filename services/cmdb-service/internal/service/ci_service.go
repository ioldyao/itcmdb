package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

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
		CreatedBy:  ptrUint(userID),
		UpdatedBy:  ptrUint(userID),
	}

	if instance.Status == "" {
		instance.Status = "active"
	}

	// 使用Context传递用户ID，让GORM Hooks自动记录历史
	ctx := context.WithValue(context.Background(), models.UserIDKey, userID)
	if err := s.repo.CreateCIInstanceWithContext(ctx, instance); err != nil {
		return nil, err
	}

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

	// 保存旧数据用于Kafka事件
	oldData, _ := json.Marshal(instance)

	// 更新字段
	if req.Name != "" {
		instance.Name = req.Name
	}

	if req.Status != "" {
		instance.Status = req.Status
	}

	if req.Attributes != nil {
		// 验证属性
		if err := s.validateAttributes(req.Attributes, instance.CIType.Attributes); err != nil {
			return nil, err
		}
		instance.Attributes = req.Attributes
	}

	if req.Tags != nil {
		instance.Tags = req.Tags
	}

	instance.UpdatedBy = ptrUint(userID)

	// 使用Context传递用户ID，让GORM Hooks自动记录历史
	ctx := context.WithValue(context.Background(), models.UserIDKey, userID)
	if err := s.repo.UpdateCIInstanceWithContext(ctx, instance); err != nil {
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

	// 使用Context传递用户ID，让GORM Hooks自动记录历史
	ctx := context.WithValue(context.Background(), models.UserIDKey, userID)
	if err := s.repo.DeleteCIInstanceWithContext(ctx, id); err != nil {
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
			result.Errors = append(result.Errors, "Row "+strconv.Itoa(i+1)+": missing or invalid name")
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
			result.Errors = append(result.Errors, "Row "+strconv.Itoa(i+1)+": "+err.Error())
			continue
		}

		// 创建CI实例
		instance := &models.CIInstance{
			CITypeID:   ciTypeID,
			Name:       name,
			Status:     status,
			Attributes: attributes,
			Tags:       tags,
			CreatedBy:  ptrUint(userID),
			UpdatedBy:  ptrUint(userID),
		}

		if err := s.repo.CreateCIInstance(instance); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "Row "+strconv.Itoa(i+1)+": "+err.Error())
			continue
		}

		// 批量导入跳过历史记录
		result.Success++
	}

	return result, nil
}
