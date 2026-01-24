package repository

import (
	"github.com/itcmdb/audit-service/internal/models"
	"gorm.io/gorm"
)

type AuditRepository interface {
	CreateBatch(logs []models.AuditLog) error
	GetLogs(offset, limit int, userID *uint, action, resource, startTime, endTime *string) ([]models.AuditLog, int64, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) CreateBatch(logs []models.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.Create(&logs).Error
}

func (r *auditRepository) GetLogs(offset, limit int, userID *uint, action, resource, startTime, endTime *string) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{})

	// 筛选条件
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != nil {
		query = query.Where("action = ?", *action)
	}
	if resource != nil {
		query = query.Where("resource = ?", *resource)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
