package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/errs"
	"blog-backend/model/request"
)

// CategoryService 分类/标签业务
type CategoryService struct{}

// ListCategories 分类列表
func (s *CategoryService) ListCategories(ctx context.Context) ([]database.Category, error) {
	categories, err := dao.ListCategories(ctx)
	if err != nil {
		global.Log.Error("查询分类列表失败", zap.Error(err))
		return nil, err
	}
	return categories, nil
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(ctx context.Context, req *request.CategoryCreateRequest) (*database.Category, error) {
	if _, err := dao.GetCategoryByName(ctx, req.Name); err == nil {
		global.Log.Warn("创建分类失败：分类已存在", zap.String("name", req.Name))
		return nil, errs.New(errs.ErrDuplicated, "分类已存在")
	} else if !errors.Is(err, errs.ErrNotFound) {
		// 只有“确实不存在”才继续创建；查询故障必须向上抛，不能被当成可创建
		global.Log.Error("创建分类前查询失败", zap.String("name", req.Name), zap.Error(err))
		return nil, err
	}
	category := &database.Category{Name: req.Name, Description: req.Description}
	if err := dao.CreateCategory(ctx, category); err != nil {
		global.Log.Error("创建分类失败", zap.String("name", req.Name), zap.Error(err))
		if errors.Is(err, errs.ErrDuplicated) {
			return nil, errs.Wrap(errs.ErrDuplicated, "分类已存在", err)
		}
		return nil, err
	}
	return category, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(ctx context.Context, id uint64, req *request.CategoryUpdateRequest) (*database.Category, error) {
	category, err := dao.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			global.Log.Warn("更新分类时分类不存在", zap.Uint64("category_id", id))
			return nil, errs.Wrap(errs.ErrNotFound, "分类不存在", err)
		}
		global.Log.Error("更新分类时查询失败", zap.Uint64("category_id", id), zap.Error(err))
		return nil, err
	}
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if err := dao.UpdateCategory(ctx, category); err != nil {
		global.Log.Error("更新分类失败", zap.Uint64("category_id", id), zap.Error(err))
		return nil, err
	}
	return category, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if _, err := dao.GetCategoryByID(ctx, id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			global.Log.Warn("删除分类时分类不存在", zap.Uint64("category_id", id))
			return errs.Wrap(errs.ErrNotFound, "分类不存在", err)
		}
		global.Log.Error("删除分类时查询失败", zap.Uint64("category_id", id), zap.Error(err))
		return err
	}
	if err := dao.DeleteCategory(ctx, id); err != nil {
		global.Log.Error("删除分类失败", zap.Uint64("category_id", id), zap.Error(err))
		return err
	}
	return nil
}

// ListTags 标签列表
func (s *CategoryService) ListTags(ctx context.Context) ([]database.Tag, error) {
	tags, err := dao.ListTags(ctx)
	if err != nil {
		global.Log.Error("查询标签列表失败", zap.Error(err))
		return nil, err
	}
	return tags, nil
}

// CreateTag 创建标签
func (s *CategoryService) CreateTag(ctx context.Context, req *request.TagCreateRequest) (*database.Tag, error) {
	tag := &database.Tag{Name: req.Name}
	if err := dao.CreateTag(ctx, tag); err != nil {
		global.Log.Error("创建标签失败", zap.String("name", req.Name), zap.Error(err))
		if errors.Is(err, errs.ErrDuplicated) {
			return nil, errs.Wrap(errs.ErrDuplicated, "标签已存在", err)
		}
		return nil, err
	}
	return tag, nil
}

// DeleteTag 删除标签
func (s *CategoryService) DeleteTag(ctx context.Context, id uint64) error {
	if _, err := dao.GetTagByID(ctx, id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			global.Log.Warn("删除标签时标签不存在", zap.Uint64("tag_id", id))
			return errs.Wrap(errs.ErrNotFound, "标签不存在", err)
		}
		global.Log.Error("删除标签时查询失败", zap.Uint64("tag_id", id), zap.Error(err))
		return err
	}
	if err := dao.DeleteTag(ctx, id); err != nil {
		global.Log.Error("删除标签失败", zap.Uint64("tag_id", id), zap.Error(err))
		return err
	}
	return nil
}
