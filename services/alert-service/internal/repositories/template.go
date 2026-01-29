package repositories

import (
	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// AlertNotificationTemplateRepository 通知模板仓储
type AlertNotificationTemplateRepository struct {
	db *gorm.DB
}

// NewAlertNotificationTemplateRepository 创建通知模板仓储
func NewAlertNotificationTemplateRepository(db *gorm.DB) *AlertNotificationTemplateRepository {
	return &AlertNotificationTemplateRepository{db: db}
}

// FindByID 根据ID查找
func (r *AlertNotificationTemplateRepository) FindByID(id int) (*models.AlertNotificationTemplate, error) {
	var tpl models.AlertNotificationTemplate
	err := r.db.First(&tpl, id).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// FindAll 查找所有
func (r *AlertNotificationTemplateRepository) FindAll() ([]models.AlertNotificationTemplate, error) {
	var tpls []models.AlertNotificationTemplate
	err := r.db.Order("template_type ASC").Find(&tpls).Error
	if err != nil {
		return nil, err
	}
	return tpls, nil
}

// FindByType 根据类型查找
func (r *AlertNotificationTemplateRepository) FindByType(templateType string) ([]models.AlertNotificationTemplate, error) {
	var tpls []models.AlertNotificationTemplate
	err := r.db.Where("template_type = ?", templateType).Order("is_default DESC, id ASC").Find(&tpls).Error
	if err != nil {
		return nil, err
	}
	return tpls, nil
}

// FindDefaultByType 查找指定类型的默认模板
func (r *AlertNotificationTemplateRepository) FindDefaultByType(templateType string) (*models.AlertNotificationTemplate, error) {
	var tpl models.AlertNotificationTemplate
	err := r.db.Where("template_type = ? AND is_default = ?", templateType, true).First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// Create 创建
func (r *AlertNotificationTemplateRepository) Create(tpl *models.AlertNotificationTemplate) error {
	return r.db.Create(tpl).Error
}

// Update 更新
func (r *AlertNotificationTemplateRepository) Update(tpl *models.AlertNotificationTemplate) error {
	return r.db.Save(tpl).Error
}

// Delete 删除
func (r *AlertNotificationTemplateRepository) Delete(id int) error {
	return r.db.Delete(&models.AlertNotificationTemplate{}, id).Error
}
