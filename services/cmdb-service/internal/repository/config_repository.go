package repository

import (
	"github.com/itcmdb/cmdb-service/internal/models"
	"gorm.io/gorm"
)

type ConfigRepository interface {
	GetAllConfigs() ([]models.SystemConfig, error)
	GetConfigsByCategory(category string) ([]models.SystemConfig, error)
	GetConfig(category, key string) (*models.SystemConfig, error)
	GetConfigByID(id uint) (*models.SystemConfig, error)
	CreateConfig(config *models.SystemConfig) error
	UpdateConfig(config *models.SystemConfig) error
	DeleteConfig(id uint) error
	UpsertConfig(config *models.SystemConfig) error
}

type configRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}

func (r *configRepository) GetAllConfigs() ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	err := r.db.Where("is_active = ?", true).Order("category ASC, key ASC").Find(&configs).Error
	return configs, err
}

func (r *configRepository) GetConfigsByCategory(category string) ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	err := r.db.Where("category = ? AND is_active = ?", category, true).Order("key ASC").Find(&configs).Error
	return configs, err
}

func (r *configRepository) GetConfig(category, key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	err := r.db.Where("category = ? AND key = ? AND is_active = ?", category, key, true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *configRepository) GetConfigByID(id uint) (*models.SystemConfig, error) {
	var config models.SystemConfig
	err := r.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *configRepository) CreateConfig(config *models.SystemConfig) error {
	return r.db.Create(config).Error
}

func (r *configRepository) UpdateConfig(config *models.SystemConfig) error {
	return r.db.Save(config).Error
}

func (r *configRepository) DeleteConfig(id uint) error {
	return r.db.Delete(&models.SystemConfig{}, id).Error
}

// UpsertConfig 创建或更新配置
func (r *configRepository) UpsertConfig(config *models.SystemConfig) error {
	var existing models.SystemConfig
	err := r.db.Where("category = ? AND key = ?", config.Category, config.Key).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		return r.db.Create(config).Error
	} else if err != nil {
		return err
	}

	// 存在，更新记录
	config.ID = existing.ID
	return r.db.Save(config).Error
}
