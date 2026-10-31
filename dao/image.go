package dao

import (
	"context"

	"blog-backend/global"
	"blog-backend/model/appTypes"
	"blog-backend/model/database"
	"blog-backend/model/request"
)

// CreateImage 新增图片台账记录
func CreateImage(ctx context.Context, image *database.Image) error {
	return global.DB.WithContext(ctx).Create(image).Error
}

// ListImages 分页查询图片（可按分类、文件名筛选）
func ListImages(ctx context.Context, p *request.PageParams, category *appTypes.ImageCategory, keyword string) ([]database.Image, int64, error) {
	var images []database.Image
	query := global.DB.WithContext(ctx).Model(&database.Image{})
	if category != nil {
		query = query.Where("category = ?", *category)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset(p.Offset()).Limit(p.PageSize).Find(&images).Error
	return images, total, err
}

// GetImagesByIDs 按 ID 批量查询
func GetImagesByIDs(ctx context.Context, ids []uint64) ([]database.Image, error) {
	var images []database.Image
	err := global.DB.WithContext(ctx).Where("id IN ?", ids).Find(&images).Error
	return images, err
}

// DeleteImagesByIDs 按 ID 批量删除台账记录
func DeleteImagesByIDs(ctx context.Context, ids []uint64) error {
	return global.DB.WithContext(ctx).Where("id IN ?", ids).Delete(&database.Image{}).Error
}

// UpdateImageCategoryByURLs 按 URL 批量更新分类
func UpdateImageCategoryByURLs(ctx context.Context, urls []string, category appTypes.ImageCategory) error {
	if len(urls) == 0 {
		return nil
	}
	return global.DB.WithContext(ctx).Model(&database.Image{}).Where("url IN ?", urls).
		Update("category", category).Error
}

// UpdateImageCategoryByID 更新单张图片分类
func UpdateImageCategoryByID(ctx context.Context, id uint64, category appTypes.ImageCategory) error {
	return global.DB.WithContext(ctx).Model(&database.Image{}).Where("id = ?", id).
		Update("category", category).Error
}

// UpdateImageCategoryByIDs 批量更新图片分类
func UpdateImageCategoryByIDs(ctx context.Context, ids []uint64, category appTypes.ImageCategory) error {
	if len(ids) == 0 {
		return nil
	}
	return global.DB.WithContext(ctx).Model(&database.Image{}).Where("id IN ?", ids).
		Update("category", category).Error
}

// ListImagesByCategory 按分类取图片（公开接口用，按上传时间倒序）
func ListImagesByCategory(ctx context.Context, category appTypes.ImageCategory, limit int) ([]database.Image, error) {
	var images []database.Image
	err := global.DB.WithContext(ctx).Where("category = ?", category).
		Order("id DESC").Limit(limit).Find(&images).Error
	return images, err
}
