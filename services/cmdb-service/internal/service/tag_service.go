package service

import (
	"errors"
	"github.com/itcmdb/cmdb-service/internal/models"
	"github.com/itcmdb/cmdb-service/internal/repository"
)

type TagService interface {
	// 标签分类管理
	GetTagCategories() ([]models.TagCategory, error)
	CreateTagCategory(req *CreateTagCategoryRequest) (*models.TagCategory, error)
	UpdateTagCategory(id uint, req *UpdateTagCategoryRequest) error
	DeleteTagCategory(id uint) error

	// 标签管理
	GetTags(categoryID *uint) ([]models.Tag, error)
	GetTagByID(id uint) (*models.Tag, error)
	CreateTag(req *CreateTagRequest) (*models.Tag, error)
	UpdateTag(id uint, req *UpdateTagRequest) error
	DeleteTag(id uint) error
	GetTagStats() ([]map[string]interface{}, error)

	// CI实例标签操作
	GetCITags(ciID uint) ([]models.CITag, error)
	AssignTag(ciID uint, tagID uint, userID uint) error
	RemoveTag(ciID uint, tagID uint) error

	// 批量操作
	BatchAssignTags(ciIDs []uint, tagID uint, userID uint) error
	BatchRemoveTags(ciIDs []uint, tagID uint) error
}

type tagService struct {
	tagRepo repository.TagRepository
	ciRepo  repository.CIRepository
}

func NewTagService(tagRepo repository.TagRepository, ciRepo repository.CIRepository) TagService {
	return &tagService{
		tagRepo: tagRepo,
		ciRepo:  ciRepo,
	}
}

type CreateTagCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateTagCategoryRequest struct {
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	SortOrder   *int    `json:"sort_order"`
}

type CreateTagRequest struct {
	CategoryID  *uint   `json:"category_id"`
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

type UpdateTagRequest struct {
	DisplayName *string `json:"display_name"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

// ==================== 标签分类管理 ====================

func (s *tagService) GetTagCategories() ([]models.TagCategory, error) {
	return s.tagRepo.GetTagCategories()
}

func (s *tagService) CreateTagCategory(req *CreateTagCategoryRequest) (*models.TagCategory, error) {
	category := &models.TagCategory{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		IsSystem:    false,
	}

	if err := s.tagRepo.CreateTagCategory(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *tagService) UpdateTagCategory(id uint, req *UpdateTagCategoryRequest) error {
	category, err := s.tagRepo.GetTagCategoryByID(id)
	if err != nil {
		return err
	}

	if category.IsSystem {
		return errors.New("系统预设标签分类不可修改")
	}

	if req.DisplayName != nil {
		category.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.Color != nil {
		category.Color = *req.Color
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	return s.tagRepo.UpdateTagCategory(category)
}

func (s *tagService) DeleteTagCategory(id uint) error {
	category, err := s.tagRepo.GetTagCategoryByID(id)
	if err != nil {
		return err
	}

	if category.IsSystem {
		return errors.New("系统预设标签分类不可删除")
	}

	return s.tagRepo.DeleteTagCategory(id)
}

// ==================== 标签管理 ====================

func (s *tagService) GetTags(categoryID *uint) ([]models.Tag, error) {
	return s.tagRepo.GetTags(categoryID)
}

func (s *tagService) GetTagByID(id uint) (*models.Tag, error) {
	return s.tagRepo.GetTagByID(id)
}

func (s *tagService) CreateTag(req *CreateTagRequest) (*models.Tag, error) {
	tag := &models.Tag{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Color:       req.Color,
		Description: req.Description,
		IsActive:    true,
	}

	if err := s.tagRepo.CreateTag(tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *tagService) UpdateTag(id uint, req *UpdateTagRequest) error {
	tag, err := s.tagRepo.GetTagByID(id)
	if err != nil {
		return err
	}

	if req.DisplayName != nil {
		tag.DisplayName = *req.DisplayName
	}
	if req.Color != nil {
		tag.Color = *req.Color
	}
	if req.Description != nil {
		tag.Description = *req.Description
	}
	if req.IsActive != nil {
		tag.IsActive = *req.IsActive
	}

	return s.tagRepo.UpdateTag(tag)
}

func (s *tagService) DeleteTag(id uint) error {
	return s.tagRepo.DeleteTag(id)
}

func (s *tagService) GetTagStats() ([]map[string]interface{}, error) {
	return s.tagRepo.GetTagStats()
}

// ==================== CI实例标签操作 ====================

func (s *tagService) GetCITags(ciID uint) ([]models.CITag, error) {
	// 验证CI存在
	_, err := s.ciRepo.GetCIInstanceByID(ciID)
	if err != nil {
		return nil, errors.New("CI实例不存在")
	}

	return s.tagRepo.GetCITags(ciID)
}

func (s *tagService) AssignTag(ciID uint, tagID uint, userID uint) error {
	// 验证CI存在
	_, err := s.ciRepo.GetCIInstanceByID(ciID)
	if err != nil {
		return errors.New("CI实例不存在")
	}

	// 验证标签存在
	_, err = s.tagRepo.GetTagByID(tagID)
	if err != nil {
		return errors.New("标签不存在")
	}

	return s.tagRepo.AssignTag(ciID, tagID, userID)
}

func (s *tagService) RemoveTag(ciID uint, tagID uint) error {
	// 验证CI存在
	_, err := s.ciRepo.GetCIInstanceByID(ciID)
	if err != nil {
		return errors.New("CI实例不存在")
	}

	return s.tagRepo.RemoveTag(ciID, tagID)
}

// ==================== 批量操作 ====================

func (s *tagService) BatchAssignTags(ciIDs []uint, tagID uint, userID uint) error {
	if len(ciIDs) == 0 {
		return errors.New("CI ID列表不能为空")
	}

	// 验证标签存在
	_, err := s.tagRepo.GetTagByID(tagID)
	if err != nil {
		return errors.New("标签不存在")
	}

	// 验证所有CI存在
	for _, ciID := range ciIDs {
		_, err := s.ciRepo.GetCIInstanceByID(ciID)
		if err != nil {
			return errors.New("CI实例不存在")
		}
	}

	return s.tagRepo.BatchAssignTags(ciIDs, tagID, userID)
}

func (s *tagService) BatchRemoveTags(ciIDs []uint, tagID uint) error {
	if len(ciIDs) == 0 {
		return errors.New("CI ID列表不能为空")
	}

	return s.tagRepo.BatchRemoveTags(ciIDs, tagID)
}
