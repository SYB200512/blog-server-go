package dao

import (
	"context"

	"blog-backend/global"
	"blog-backend/model/database"
)

// ListCategories 分类列表（含文章数量）
func ListCategories(ctx context.Context) ([]database.Category, error) {
	var categories []database.Category
	err := global.DB.WithContext(ctx).Model(&database.Category{}).
		Select("categories.*, COUNT(a.id) AS article_count").
		Joins("LEFT JOIN articles a ON a.category_id = categories.id AND a.deleted_at IS NULL AND a.status = 1").
		Group("categories.id").
		Order("categories.id ASC").
		Scan(&categories).Error
	return categories, err
}

// GetCategoryByID 根据 ID 查询分类
func GetCategoryByID(ctx context.Context, id uint64) (*database.Category, error) {
	var category database.Category
	err := global.DB.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetCategoryByName 根据名称查询分类
func GetCategoryByName(ctx context.Context, name string) (*database.Category, error) {
	var category database.Category
	err := global.DB.WithContext(ctx).Where("name = ?", name).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// CreateCategory 创建分类
func CreateCategory(ctx context.Context, category *database.Category) error {
	return global.DB.WithContext(ctx).Create(category).Error
}

// UpdateCategory 更新分类
func UpdateCategory(ctx context.Context, category *database.Category) error {
	return global.DB.WithContext(ctx).Save(category).Error
}

// DeleteCategory 删除分类
func DeleteCategory(ctx context.Context, id uint64) error {
	return global.DB.WithContext(ctx).Delete(&database.Category{}, id).Error
}

// ListTags 标签列表（含文章数量）
func ListTags(ctx context.Context) ([]database.Tag, error) {
	var tags []database.Tag
	err := global.DB.WithContext(ctx).Model(&database.Tag{}).
		Select("tags.*, COUNT(at.article_id) AS article_count").
		Joins("LEFT JOIN article_tags at ON at.tag_id = tags.id AND at.article_id IN (SELECT id FROM articles WHERE status = 1 AND deleted_at IS NULL)").
		Group("tags.id").
		Order("tags.id ASC").
		Scan(&tags).Error
	return tags, err
}

// GetTagByID 根据 ID 查询标签
func GetTagByID(ctx context.Context, id uint64) (*database.Tag, error) {
	var tag database.Tag
	err := global.DB.WithContext(ctx).First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// CreateTag 创建标签
func CreateTag(ctx context.Context, tag *database.Tag) error {
	return global.DB.WithContext(ctx).Create(tag).Error
}

// DeleteTag 删除标签
func DeleteTag(ctx context.Context, id uint64) error {
	return global.DB.WithContext(ctx).Delete(&database.Tag{}, id).Error
}
