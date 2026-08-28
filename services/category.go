package services

import (
	"errors"

	"gorm.io/gorm"

	"blog-backend/dao"
	"blog-backend/dto"
	"blog-backend/models"
)

// ListCategories 分类列表
func ListCategories() ([]models.Category, error) {
	return dao.ListCategories()
}

// CreateCategory 创建分类
func CreateCategory(req *dto.CategoryCreateRequest) (*models.Category, error) {
	if _, err := dao.GetCategoryByName(req.Name); err == nil {
		return nil, errors.New("分类已存在")
	}
	category := &models.Category{Name: req.Name, Description: req.Description}
	if err := dao.CreateCategory(category); err != nil {
		return nil, err
	}
	return category, nil
}

// UpdateCategory 更新分类
func UpdateCategory(id uint64, req *dto.CategoryUpdateRequest) (*models.Category, error) {
	category, err := dao.GetCategoryByID(id)
	if err != nil {
		return nil, errors.New("分类不存在")
	}
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if err := dao.UpdateCategory(category); err != nil {
		return nil, err
	}
	return category, nil
}

// DeleteCategory 删除分类
func DeleteCategory(id uint64) error {
	if _, err := dao.GetCategoryByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分类不存在")
		}
		return err
	}
	return dao.DeleteCategory(id)
}

// ListTags 标签列表
func ListTags() ([]models.Tag, error) {
	return dao.ListTags()
}

// CreateTag 创建标签
func CreateTag(req *dto.TagCreateRequest) (*models.Tag, error) {
	tag := &models.Tag{Name: req.Name}
	if err := dao.CreateTag(tag); err != nil {
		return nil, errors.New("标签已存在或创建失败")
	}
	return tag, nil
}

// DeleteTag 删除标签
func DeleteTag(id uint64) error {
	if _, err := dao.GetTagByID(id); err != nil {
		return errors.New("标签不存在")
	}
	return dao.DeleteTag(id)
}
