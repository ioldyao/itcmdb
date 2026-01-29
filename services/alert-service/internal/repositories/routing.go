package repositories

import (
	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// AlertRoutingRuleRepository 路由规则仓储
type AlertRoutingRuleRepository struct {
	db *gorm.DB
}

// NewAlertRoutingRuleRepository 创建路由规则仓储
func NewAlertRoutingRuleRepository(db *gorm.DB) *AlertRoutingRuleRepository {
	return &AlertRoutingRuleRepository{db: db}
}

// FindByID 根据ID查找
func (r *AlertRoutingRuleRepository) FindByID(id int) (*models.AlertRoutingRule, error) {
	var rule models.AlertRoutingRule
	err := r.db.Preload("ReceiverGroup").First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// FindAll 查找所有
func (r *AlertRoutingRuleRepository) FindAll() ([]models.AlertRoutingRule, error) {
	var rules []models.AlertRoutingRule
	err := r.db.Preload("ReceiverGroup").Order("priority ASC").Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// FindEnabled 查找所有启用的规则
func (r *AlertRoutingRuleRepository) FindEnabled() ([]models.AlertRoutingRule, error) {
	var rules []models.AlertRoutingRule
	err := r.db.Preload("ReceiverGroup").Where("enabled = ?", true).Order("priority ASC").Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// Create 创建
func (r *AlertRoutingRuleRepository) Create(rule *models.AlertRoutingRule) error {
	return r.db.Create(rule).Error
}

// Update 更新
func (r *AlertRoutingRuleRepository) Update(rule *models.AlertRoutingRule) error {
	return r.db.Save(rule).Error
}

// Delete 删除
func (r *AlertRoutingRuleRepository) Delete(id int) error {
	return r.db.Delete(&models.AlertRoutingRule{}, id).Error
}
