package dao

import "blog-backend/models"

// ListCategories 分类列表（含文章数量）
func ListCategories() ([]models.Category, error) {
	var categories []models.Category
	err := DB.Model(&models.Category{}).
		Select("categories.*, COUNT(a.id) AS article_count").
		Joins("LEFT JOIN articles a ON a.category_id = categories.id AND a.deleted_at IS NULL").
		Group("categories.id").
		Order("categories.id ASC").
		Scan(&categories).Error
	return categories, err
}

// GetCategoryByID 根据 ID 查询分类
func GetCategoryByID(id uint64) (*models.Category, error) {
	var category models.Category
	err := DB.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetCategoryByName 根据名称查询分类
func GetCategoryByName(name string) (*models.Category, error) {
	var category models.Category
	err := DB.Where("name = ?", name).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// CreateCategory 创建分类
func CreateCategory(category *models.Category) error {
	return DB.Create(category).Error
}

// UpdateCategory 更新分类
func UpdateCategory(category *models.Category) error {
	return DB.Save(category).Error
}

// DeleteCategory 删除分类
func DeleteCategory(id uint64) error {
	return DB.Delete(&models.Category{}, id).Error
}

// ListTags 标签列表（含文章数量）
func ListTags() ([]models.Tag, error) {
	var tags []models.Tag
	err := DB.Model(&models.Tag{}).
		Select("tags.*, COUNT(at.article_id) AS article_count").
		Joins("LEFT JOIN article_tags at ON at.tag_id = tags.id").
		Group("tags.id").
		Order("tags.id ASC").
		Scan(&tags).Error
	return tags, err
}

// GetTagByID 根据 ID 查询标签
func GetTagByID(id uint64) (*models.Tag, error) {
	var tag models.Tag
	err := DB.First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// CreateTag 创建标签
func CreateTag(tag *models.Tag) error {
	return DB.Create(tag).Error
}

// DeleteTag 删除标签
func DeleteTag(id uint64) error {
	return DB.Delete(&models.Tag{}, id).Error
}
