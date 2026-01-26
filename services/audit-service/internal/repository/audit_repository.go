package repository

import (
	"github.com/itcmdb/audit-service/internal/models"
	"gorm.io/gorm"
)

type AuditRepository interface {
	CreateBatch(logs []models.AuditLog) error
	GetLogs(offset, limit int, userID *uint, action, resource, startTime, endTime *string) ([]models.AuditLog, int64, error)
	GetStats(startTime, endTime *string, userID *uint) (int64, map[string]int64, map[string]int64, error)
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

func (r *auditRepository) GetStats(startTime, endTime *string, userID *uint) (int64, map[string]int64, map[string]int64, error) {
	var total int64
	byAction := make(map[string]int64)
	byResource := make(map[string]int64)

	query := r.db.Model(&models.AuditLog{})

	// 用户筛选
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 时间范围筛选
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return 0, nil, nil, err
	}

	// 按操作类型统计
	var actionStats []struct {
		Action string
		Count  int64
	}
	if err := query.Select("action, COUNT(*) as count").Group("action").Scan(&actionStats).Error; err != nil {
		return 0, nil, nil, err
	}
	for _, stat := range actionStats {
		byAction[stat.Action] = stat.Count
	}

	// 按资源类型统计
	var resourceStats []struct {
		Resource string
		Count    int64
	}
	if err := query.Select("resource, COUNT(*) as count").Group("resource").Scan(&resourceStats).Error; err != nil {
		return 0, nil, nil, err
	}
	for _, stat := range resourceStats {
		byResource[stat.Resource] = stat.Count
	}

	return total, byAction, byResource, nil
}
