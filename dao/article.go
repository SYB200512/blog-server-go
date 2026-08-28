package dao

import (
	"blog-backend/common"
	"blog-backend/models"

	"gorm.io/gorm"
)

// CountArticles 统计文章数量
func CountArticles(status *int8) (int64, error) {
	var count int64
	query := DB.Model(&models.Article{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Count(&count).Error
	return count, err
}

// ListArticles 文章分页列表
// status 为空则不过滤状态；admin 表示是否后台查询（后台返回草稿）
func ListArticles(p *common.PageParams, keyword string, categoryID, tagID uint64, status *int8, admin bool) ([]models.Article, int64, error) {
	var articles []models.Article
	query := DB.Model(&models.Article{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ?", like, like)
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	// 前台只显示已发布，后台可按状态过滤
	if !admin {
		query = query.Where("status = 1")
	} else if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 标签筛选（需关联子查询）
	if tagID > 0 {
		query = query.Where("id IN (SELECT article_id FROM article_tags WHERE tag_id = ?)", tagID)
	}

	err := query.Preload("Category").Preload("Tags").
		Order("is_top DESC, created_at DESC").
		Offset(p.Offset()).Limit(p.PageSize).
		Find(&articles).Error
	return articles, total, err
}

// GetArticleByID 根据 ID 查询文章（含关联）
func GetArticleByID(id uint64) (*models.Article, error) {
	var article models.Article
	err := DB.Preload("Category").Preload("Tags").
		Where("id = ?", id).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetPublishedArticleByID 查询已发布文章
func GetPublishedArticleByID(id uint64) (*models.Article, error) {
	var article models.Article
	err := DB.Preload("Category").Preload("Tags").
		Where("id = ? AND status = 1", id).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// CreateArticle 创建文章
func CreateArticle(article *models.Article, tagIDs []uint64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(article).Error; err != nil {
			return err
		}
		if len(tagIDs) > 0 {
			var tags []models.Tag
			if err := tx.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
				return err
			}
			if err := tx.Model(article).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateArticle 更新文章及标签
func UpdateArticle(article *models.Article, tagIDs []uint64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(article).Error; err != nil {
			return err
		}
		if tagIDs != nil {
			var tags []models.Tag
			if len(tagIDs) > 0 {
				if err := tx.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(article).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteArticle 软删除文章
func DeleteArticle(id uint64) error {
	return DB.Delete(&models.Article{}, id).Error
}

// IncrementViews 浏览量 +1
func IncrementViews(id uint64) error {
	return DB.Model(&models.Article{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1")).Error
}

// ListTopArticles 置顶文章列表
func ListTopArticles() ([]models.Article, error) {
	var articles []models.Article
	err := DB.Where("status = 1 AND is_top = 1").
		Order("created_at DESC").Find(&articles).Error
	return articles, err
}

// ArchiveByMonth 按月归档已发布文章（返回格式：2026-08 等）
func ArchiveByMonth() (map[string][]models.Article, error) {
	var articles []models.Article
	err := DB.Where("status = 1").
		Order("created_at DESC").
		Find(&articles).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string][]models.Article)
	for _, a := range articles {
		key := a.CreatedAt.Format("2006-01")
		result[key] = append(result[key], a)
	}
	return result, nil
}
