package service

import (
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/request"
)

// CategoryService 分类/标签业务
type CategoryService struct{}

// ListCategories 分类列表
func (s *CategoryService) ListCategories() ([]database.Category, error) {
	categories, err := dao.ListCategories()
	if err != nil {
		global.Log.Error("查询分类列表失败", zap.Error(err))
		return nil, err
	}
	return categories, nil
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(req *request.CategoryCreateRequest) (*database.Category, error) {
	if _, err := dao.GetCategoryByName(req.Name); err == nil {
		global.Log.Warn("创建分类失败：分类已存在", zap.String("name", req.Name))
		return nil, errors.New("分类已存在")
	}
	category := &database.Category{Name: req.Name, Description: req.Description}
	if err := dao.CreateCategory(category); err != nil {
		global.Log.Error("创建分类失败", zap.String("name", req.Name), zap.Error(err))
		return nil, err
	}
	return category, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(id uint64, req *request.CategoryUpdateRequest) (*database.Category, error) {
	category, err := dao.GetCategoryByID(id)
	if err != nil {
		global.Log.Warn("更新分类时分类不存在", zap.Uint64("category_id", id))
		return nil, errors.New("分类不存在")
	}
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if err := dao.UpdateCategory(category); err != nil {
		global.Log.Error("更新分类失败", zap.Uint64("category_id", id), zap.Error(err))
		return nil, err
	}
	return category, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(id uint64) error {
	if _, err := dao.GetCategoryByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.Log.Warn("删除分类时分类不存在", zap.Uint64("category_id", id))
			return errors.New("分类不存在")
		}
		global.Log.Error("删除分类时查询失败", zap.Uint64("category_id", id), zap.Error(err))
		return err
	}
	if err := dao.DeleteCategory(id); err != nil {
		global.Log.Error("删除分类失败", zap.Uint64("category_id", id), zap.Error(err))
		return err
	}
	return nil
}

// ListTags 标签列表
func (s *CategoryService) ListTags() ([]database.Tag, error) {
	tags, err := dao.ListTags()
	if err != nil {
		global.Log.Error("查询标签列表失败", zap.Error(err))
		return nil, err
	}
	return tags, nil
}

// CreateTag 创建标签
func (s *CategoryService) CreateTag(req *request.TagCreateRequest) (*database.Tag, error) {
	tag := &database.Tag{Name: req.Name}
	if err := dao.CreateTag(tag); err != nil {
		global.Log.Error("创建标签失败", zap.String("name", req.Name), zap.Error(err))
		return nil, errors.New("标签已存在或创建失败")
	}
	return tag, nil
}

// DeleteTag 删除标签
func (s *CategoryService) DeleteTag(id uint64) error {
	if _, err := dao.GetTagByID(id); err != nil {
		global.Log.Warn("删除标签时标签不存在", zap.Uint64("tag_id", id))
		return errors.New("标签不存在")
	}
	if err := dao.DeleteTag(id); err != nil {
		global.Log.Error("删除标签失败", zap.Uint64("tag_id", id), zap.Error(err))
		return err
	}
	return nil
}
