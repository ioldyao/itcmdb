package repository

import (
	"context"

	"github.com/itcmdb/cmdb-service/internal/models"
	"gorm.io/gorm"
)

type CIRepository interface {
	// CI Types
	GetCITypes() ([]models.CIType, error)
	GetCITypeByID(id uint) (*models.CIType, error)
	GetCITypeByName(name string) (*models.CIType, error)

	// CI Instances
	GetCIInstances(ciTypeID uint, filters map[string]interface{}, page, pageSize int) ([]models.CIInstance, int64, error)
	GetCIInstanceByID(id uint) (*models.CIInstance, error)
	CreateCIInstance(instance *models.CIInstance) error
	CreateCIInstanceWithContext(ctx context.Context, instance *models.CIInstance) error
	UpdateCIInstance(instance *models.CIInstance) error
	UpdateCIInstanceWithContext(ctx context.Context, instance *models.CIInstance) error
	DeleteCIInstance(id uint) error
	DeleteCIInstanceWithContext(ctx context.Context, id uint) error

	// CI Relations
	GetCIRelations(ciID uint) ([]models.CIRelation, error)
	CreateCIRelation(relation *models.CIRelation) error
	DeleteCIRelation(id uint) error

	// CI History
	GetCIHistory(ciID uint, limit int) ([]models.CIHistory, error)
	CreateCIHistory(history *models.CIHistory) error
}

type ciRepository struct {
	db *gorm.DB
}

func NewCIRepository(db *gorm.DB) CIRepository {
	return &ciRepository{db: db}
}

// CI Types

func (r *ciRepository) GetCITypes() ([]models.CIType, error) {
	var types []models.CIType
	err := r.db.Where("is_active = ?", true).
		Preload("Attributes", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("id ASC").
		Find(&types).Error
	return types, err
}

func (r *ciRepository) GetCITypeByID(id uint) (*models.CIType, error) {
	var ciType models.CIType
	err := r.db.Preload("Attributes", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&ciType, id).Error
	if err != nil {
		return nil, err
	}
	return &ciType, nil
}

func (r *ciRepository) GetCITypeByName(name string) (*models.CIType, error) {
	var ciType models.CIType
	err := r.db.Where("name = ?", name).
		Preload("Attributes").
		First(&ciType).Error
	if err != nil {
		return nil, err
	}
	return &ciType, nil
}

// CI Instances

func (r *ciRepository) GetCIInstances(ciTypeID uint, filters map[string]interface{}, page, pageSize int) ([]models.CIInstance, int64, error) {
	var instances []models.CIInstance
	var total int64

	query := r.db.Model(&models.CIInstance{})

	if ciTypeID > 0 {
		query = query.Where("ci_type_id = ?", ciTypeID)
	}

	// 应用过滤器
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	// 处理 JSONB attributes 字段的查询
	if systemSerial, ok := filters["system_serial"].(string); ok && systemSerial != "" {
		query = query.Where("attributes->>'system_serial' = ?", systemSerial)
	}
	if hostname, ok := filters["hostname"].(string); ok && hostname != "" {
		query = query.Where("attributes->>'hostname' = ?", hostname)
	}

	// 获取总数
	query.Count(&total)

	// 分页
	offset := (page - 1) * pageSize
	err := query.Preload("CIType").
		Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&instances).Error

	return instances, total, err
}

func (r *ciRepository) GetCIInstanceByID(id uint) (*models.CIInstance, error) {
	var instance models.CIInstance
	err := r.db.Preload("CIType").
		Preload("CIType.Attributes").
		First(&instance, id).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *ciRepository) CreateCIInstance(instance *models.CIInstance) error {
	return r.db.Create(instance).Error
}

func (r *ciRepository) CreateCIInstanceWithContext(ctx context.Context, instance *models.CIInstance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

func (r *ciRepository) UpdateCIInstance(instance *models.CIInstance) error {
	return r.db.Save(instance).Error
}

func (r *ciRepository) UpdateCIInstanceWithContext(ctx context.Context, instance *models.CIInstance) error {
	return r.db.WithContext(ctx).Save(instance).Error
}

func (r *ciRepository) DeleteCIInstance(id uint) error {
	return r.db.Delete(&models.CIInstance{}, id).Error
}

func (r *ciRepository) DeleteCIInstanceWithContext(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.CIInstance{}, id).Error
}

// CI Relations

func (r *ciRepository) GetCIRelations(ciID uint) ([]models.CIRelation, error) {
	var relations []models.CIRelation
	err := r.db.Where("parent_id = ? OR child_id = ?", ciID, ciID).
		Preload("Parent").
		Preload("Child").
		Find(&relations).Error
	return relations, err
}

func (r *ciRepository) CreateCIRelation(relation *models.CIRelation) error {
	return r.db.Create(relation).Error
}

func (r *ciRepository) DeleteCIRelation(id uint) error {
	return r.db.Delete(&models.CIRelation{}, id).Error
}

// CI History

func (r *ciRepository) GetCIHistory(ciID uint, limit int) ([]models.CIHistory, error) {
	var history []models.CIHistory
	query := r.db.Where("ci_id = ?", ciID).
		Order("changed_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&history).Error
	return history, err
}

func (r *ciRepository) CreateCIHistory(history *models.CIHistory) error {
	return r.db.Create(history).Error
}
