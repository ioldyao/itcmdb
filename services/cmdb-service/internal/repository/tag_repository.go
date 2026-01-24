package repository

import (
	"github.com/itcmdb/cmdb-service/internal/models"
	"gorm.io/gorm"
)

type TagRepository interface {
	// 标签分类管理
	GetTagCategories() ([]models.TagCategory, error)
	GetTagCategoryByID(id uint) (*models.TagCategory, error)
	CreateTagCategory(category *models.TagCategory) error
	UpdateTagCategory(category *models.TagCategory) error
	DeleteTagCategory(id uint) error

	// 标签管理
	GetTags(categoryID *uint) ([]models.Tag, error)
	GetTagByID(id uint) (*models.Tag, error)
	GetTagByName(categoryID uint, name string) (*models.Tag, error)
	CreateTag(tag *models.Tag) error
	UpdateTag(tag *models.Tag) error
	DeleteTag(id uint) error
	IncreaseTagUsage(id uint) error
	DecreaseTagUsage(id uint) error

	// CI实例标签操作
	GetCITags(ciID uint) ([]models.CITag, error)
	AssignTag(ciID uint, tagID uint, taggedBy uint) error
	RemoveTag(ciID uint, tagID uint) error
	GetCIsByTag(tagID uint, page, pageSize int) ([]uint, int64, error)

	// 批量操作
	BatchAssignTags(ciIDs []uint, tagID uint, taggedBy uint) error
	BatchRemoveTags(ciIDs []uint, tagID uint) error

	// 标签统计
	GetTagStats() ([]map[string]interface{}, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// ==================== 标签分类管理 ====================

func (r *tagRepository) GetTagCategories() ([]models.TagCategory, error) {
	var categories []models.TagCategory
	err := r.db.Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

func (r *tagRepository) GetTagCategoryByID(id uint) (*models.TagCategory, error) {
	var category models.TagCategory
	err := r.db.Preload("Tags").First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *tagRepository) CreateTagCategory(category *models.TagCategory) error {
	return r.db.Create(category).Error
}

func (r *tagRepository) UpdateTagCategory(category *models.TagCategory) error {
	return r.db.Save(category).Error
}

func (r *tagRepository) DeleteTagCategory(id uint) error {
	return r.db.Delete(&models.TagCategory{}, id).Error
}

// ==================== 标签管理 ====================

func (r *tagRepository) GetTags(categoryID *uint) ([]models.Tag, error) {
	var tags []models.Tag
	query := r.db.Preload("Category").Where("is_active = ?", true)
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	err := query.Order("id ASC").Find(&tags).Error
	return tags, err
}

func (r *tagRepository) GetTagByID(id uint) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.Preload("Category").First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) GetTagByName(categoryID uint, name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.Where("category_id = ? AND name = ?", categoryID, name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) CreateTag(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

func (r *tagRepository) UpdateTag(tag *models.Tag) error {
	return r.db.Save(tag).Error
}

func (r *tagRepository) DeleteTag(id uint) error {
	return r.db.Delete(&models.Tag{}, id).Error
}

func (r *tagRepository) IncreaseTagUsage(id uint) error {
	return r.db.Model(&models.Tag{}).Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (r *tagRepository) DecreaseTagUsage(id uint) error {
	return r.db.Model(&models.Tag{}).Where("id = ? AND usage_count > 0", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count - 1")).Error
}

// ==================== CI实例标签操作 ====================

func (r *tagRepository) GetCITags(ciID uint) ([]models.CITag, error) {
	var tags []models.CITag
	err := r.db.Preload("Tag").Preload("Tag.Category").Where("ci_id = ?", ciID).Find(&tags).Error
	return tags, err
}

func (r *tagRepository) AssignTag(ciID uint, tagID uint, taggedBy uint) error {
	// 检查是否已存在
	var count int64
	if err := r.db.Model(&models.CITag{}).
		Where("ci_id = ? AND tag_id = ?", ciID, tagID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // 已存在，不需要重复添加
	}

	ciTag := &models.CITag{
		CIID:     ciID,
		TagID:    tagID,
		TaggedBy: &taggedBy,
	}

	if err := r.db.Create(ciTag).Error; err != nil {
		return err
	}

	// 记录历史
	history := &models.TagHistory{
		CIID:   &ciID,
		TagID:  &tagID,
		Action: "added",
		UserID: &taggedBy,
	}
	r.db.Create(history)

	// 增加使用计数
	return r.IncreaseTagUsage(tagID)
}

func (r *tagRepository) RemoveTag(ciID uint, tagID uint) error {
	if err := r.db.Where("ci_id = ? AND tag_id = ?", ciID, tagID).Delete(&models.CITag{}).Error; err != nil {
		return err
	}

	// 减少使用计数
	return r.DecreaseTagUsage(tagID)
}

func (r *tagRepository) GetCIsByTag(tagID uint, page, pageSize int) ([]uint, int64, error) {
	var ciIDs []uint
	var total int64

	// 计算总数
	if err := r.db.Model(&models.CITag{}).Where("tag_id = ?", tagID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Model(&models.CITag{}).Where("tag_id = ?", tagID).
		Limit(pageSize).Offset(offset).
		Pluck("ci_id", &ciIDs).Error; err != nil {
		return nil, 0, err
	}

	return ciIDs, total, nil
}

// ==================== 批量操作 ====================

func (r *tagRepository) BatchAssignTags(ciIDs []uint, tagID uint, taggedBy uint) error {
	var ciTags []models.CITag
	for _, ciID := range ciIDs {
		ciTags = append(ciTags, models.CITag{
			CIID:     ciID,
			TagID:    tagID,
			TaggedBy: &taggedBy,
		})
	}

	if err := r.db.Create(&ciTags).Error; err != nil {
		return err
	}

	// 批量增加使用计数
	increment := int64(len(ciIDs))
	return r.db.Model(&models.Tag{}).Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", increment)).Error
}

func (r *tagRepository) BatchRemoveTags(ciIDs []uint, tagID uint) error {
	if err := r.db.Where("tag_id = ? AND ci_id IN ?", tagID, ciIDs).Delete(&models.CITag{}).Error; err != nil {
		return err
	}

	// 批量减少使用计数
	decrement := int64(len(ciIDs))
	return r.db.Model(&models.Tag{}).Where("id = ? AND usage_count >= ?", tagID, decrement).
		UpdateColumn("usage_count", gorm.Expr("usage_count - ?", decrement)).Error
}

// ==================== 标签统计 ====================

func (r *tagRepository) GetTagStats() ([]map[string]interface{}, error) {
	var stats []map[string]interface{}

	rows, err := r.db.Model(&models.Tag{}).
		Select("tags.id, tags.name, tags.display_name, tags.color, tags.category_id, tag_categories.name as category_name, COUNT(ci_tags.id) as usage_count").
		Joins("LEFT JOIN ci_tags ON ci_tags.tag_id = tags.id").
		Joins("LEFT JOIN tag_categories ON tag_categories.id = tags.category_id").
		Group("tags.id, tags.name, tags.display_name, tags.color, tags.category_id, tag_categories.name").
		Order("usage_count DESC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat struct {
			ID           uint
			Name         string
			DisplayName  string
			Color        string
			CategoryID   *uint
			CategoryName string
			UsageCount   int64
		}
		if err := rows.Scan(&stat.ID, &stat.Name, &stat.DisplayName, &stat.Color, &stat.CategoryID, &stat.CategoryName, &stat.UsageCount); err != nil {
			return nil, err
		}

		stats = append(stats, map[string]interface{}{
			"id":           stat.ID,
			"name":         stat.Name,
			"display_name": stat.DisplayName,
			"color":        stat.Color,
			"category_id":  stat.CategoryID,
			"category_name": stat.CategoryName,
			"usage_count":  stat.UsageCount,
		})
	}

	return stats, rows.Err()
}
