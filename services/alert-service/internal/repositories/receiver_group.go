package repositories

import (
	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// AlertReceiverGroupRepository 接收人组仓储
type AlertReceiverGroupRepository struct {
	db *gorm.DB
}

// NewAlertReceiverGroupRepository 创建接收人组仓储
func NewAlertReceiverGroupRepository(db *gorm.DB) *AlertReceiverGroupRepository {
	return &AlertReceiverGroupRepository{db: db}
}

// FindByID 根据ID查找
func (r *AlertReceiverGroupRepository) FindByID(id int) (*models.AlertReceiverGroup, error) {
	var group models.AlertReceiverGroup
	err := r.db.Preload("Members.Receiver").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// FindAll 查找所有
func (r *AlertReceiverGroupRepository) FindAll() ([]models.AlertReceiverGroup, error) {
	var groups []models.AlertReceiverGroup
	err := r.db.Preload("Members.Receiver").Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// FindEnabled 查找所有启用的接收人组
func (r *AlertReceiverGroupRepository) FindEnabled() ([]models.AlertReceiverGroup, error) {
	var groups []models.AlertReceiverGroup
	err := r.db.Preload("Members.Receiver").Where("enabled = ?", true).Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// Create 创建
func (r *AlertReceiverGroupRepository) Create(group *models.AlertReceiverGroup) error {
	return r.db.Create(group).Error
}

// Update 更新
func (r *AlertReceiverGroupRepository) Update(group *models.AlertReceiverGroup) error {
	return r.db.Save(group).Error
}

// Delete 删除
func (r *AlertReceiverGroupRepository) Delete(id int) error {
	return r.db.Delete(&models.AlertReceiverGroup{}, id).Error
}
