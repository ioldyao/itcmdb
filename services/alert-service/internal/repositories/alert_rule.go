package repositories

import (
	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// AlertRuleRepository 告警规则仓储
type AlertRuleRepository struct {
	db *gorm.DB
}

// NewAlertRuleRepository 创建告警规则仓储
func NewAlertRuleRepository(db *gorm.DB) *AlertRuleRepository {
	return &AlertRuleRepository{db: db}
}

// FindByID 根据ID查找
func (r *AlertRuleRepository) FindByID(id int) (*models.AlertRule, error) {
	var rule models.AlertRule
	err := r.db.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// FindAll 查找所有
func (r *AlertRuleRepository) FindAll() ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := r.db.Where("deleted_at IS NULL").Order("id ASC").Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// FindEnabled 查找所有启用的规则
func (r *AlertRuleRepository) FindEnabled() ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := r.db.Where("enabled = ? AND deleted_at IS NULL", true).
		Where("(silenced_until IS NULL OR silenced_until < NOW())").
		Order("id ASC").
		Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// Create 创建
func (r *AlertRuleRepository) Create(rule *models.AlertRule) error {
	return r.db.Create(rule).Error
}

// Update 更新
func (r *AlertRuleRepository) Update(rule *models.AlertRule) error {
	return r.db.Save(rule).Error
}

// Delete 软删除
func (r *AlertRuleRepository) Delete(id int) error {
	return r.db.Model(&models.AlertRule{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()")).Error
}
