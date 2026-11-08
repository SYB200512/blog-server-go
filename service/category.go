package service

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/errs"
	"blog-backend/model/request"
	"blog-backend/utils/cache"
)

// categoryListTTL 分类/标签列表缓存 TTL：长 TTL + 写失效双保险
const categoryListTTL = 30 * time.Minute

// CategoryService 分类/标签业务
type CategoryService struct {
	// group 合并并发回源：同一 key 的缓存击穿流量只查一次 DB
	group singleflight.Group
}

// ListCategories 分类列表：整包缓存 + singleflight 回源
func (s *CategoryService) ListCategories(ctx context.Context) ([]database.Category, error) {
	key := cache.CategoryListKey()

	var cached []database.Category
	if cache.Get(ctx, key, &cached) == cache.Hit {
		return cached, nil
	}

	v, err, _ := s.group.Do(key, func() (any, error) {
		// 二次检查：等待 singleflight 期间可能已被其他请求回填
		var c []database.Category
		if cache.Get(ctx, key, &c) == cache.Hit {
			return c, nil
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		categories, err := dao.ListCategories(dbCtx)
		if err != nil {
			return nil, err
		}
		cache.Set(ctx, key, categories, cache.WithJitter(categoryListTTL))
		return categories, nil
	})
	if err != nil {
		global.Log.Error("查询分类列表失败", zap.Error(err))
		return nil, err
	}
	return v.([]database.Category), nil
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
	// 事务提交成功后失效列表缓存
	cache.Del(ctx, cache.CategoryListKey())
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
	// 事务提交成功后失效列表缓存
	cache.Del(ctx, cache.CategoryListKey())
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
	cache.Del(ctx, cache.CategoryListKey())
	return nil
}

// ListTags 标签列表：整包缓存 + singleflight 回源
func (s *CategoryService) ListTags(ctx context.Context) ([]database.Tag, error) {
	key := cache.TagListKey()

	var cached []database.Tag
	if cache.Get(ctx, key, &cached) == cache.Hit {
		return cached, nil
	}

	v, err, _ := s.group.Do(key, func() (any, error) {
		// 二次检查
		var t []database.Tag
		if cache.Get(ctx, key, &t) == cache.Hit {
			return t, nil
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		tags, err := dao.ListTags(dbCtx)
		if err != nil {
			return nil, err
		}
		cache.Set(ctx, key, tags, cache.WithJitter(categoryListTTL))
		return tags, nil
	})
	if err != nil {
		global.Log.Error("查询标签列表失败", zap.Error(err))
		return nil, err
	}
	return v.([]database.Tag), nil
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
	cache.Del(ctx, cache.TagListKey())
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
	cache.Del(ctx, cache.TagListKey())
	return nil
}
