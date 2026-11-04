package dao

import (
	"context"

	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/request"

	"gorm.io/gorm"
)

// CountArticles 统计文章数量
func CountArticles(ctx context.Context, status *int8) (int64, error) {
	var count int64
	query := global.DB.WithContext(ctx).Model(&database.Article{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Count(&count).Error
	return count, err
}

// ListArticles 文章分页列表
// status 为空则不过滤状态；admin 表示是否后台查询（后台返回草稿）。
// WithContext(ctx)：SQL 继承请求的取消与超时，客户端断开时会中断查询。
func ListArticles(ctx context.Context, p *request.PageParams, keyword string, categoryID, tagID uint64, status *int8, admin bool) ([]database.Article, int64, error) {
	var articles []database.Article
	query := global.DB.WithContext(ctx).Model(&database.Article{})

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

	// P0：标签条件必须在 Count 之前，否则总数不含标签筛选，分页会错位
	if tagID > 0 {
		// EXISTS 比 IN 子查询更稳定，配合 article_tags(tag_id, article_id) 索引
		query = query.Where(
			"EXISTS (SELECT 1 FROM article_tags AS at WHERE at.article_id = articles.id AND at.tag_id = ?)",
			tagID,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// P0：列表不返回完整 longtext；保留 400 字是为了让摘要为空的卡片仍能生成摘要
	err := query.
		Select("id, title, summary, cover, category_id, status, views, like_count, is_top, created_at, updated_at, LEFT(content, 400) AS content").
		Preload("Category").Preload("Tags").
		Order("is_top DESC, created_at DESC").
		Offset(p.Offset()).Limit(p.PageSize).
		Find(&articles).Error
	return articles, total, err
}

// GetArticleByID 根据 ID 查询文章（含关联）
func GetArticleByID(ctx context.Context, id uint64) (*database.Article, error) {
	var article database.Article
	err := global.DB.WithContext(ctx).Preload("Category").Preload("Tags").
		Where("id = ?", id).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetPublishedArticleByID 查询已发布文章
func GetPublishedArticleByID(ctx context.Context, id uint64) (*database.Article, error) {
	var article database.Article
	err := global.DB.WithContext(ctx).Preload("Category").Preload("Tags").
		Where("id = ? AND status = 1", id).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// CreateArticle 创建文章
func CreateArticle(ctx context.Context, article *database.Article, tagIDs []uint64) error {
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(article).Error; err != nil {
			return err
		}
		if len(tagIDs) > 0 {
			var tags []database.Tag
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
func UpdateArticle(ctx context.Context, article *database.Article, tagIDs []uint64) error {
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(article).Error; err != nil {
			return err
		}
		if tagIDs != nil {
			var tags []database.Tag
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
func DeleteArticle(ctx context.Context, id uint64) error {
	return global.DB.WithContext(ctx).Delete(&database.Article{}, id).Error
}

// UpdateArticleCover 异步生成封面完成后，仅回写 cover 字段
func UpdateArticleCover(ctx context.Context, id uint64, cover string) error {
	return global.DB.WithContext(ctx).Model(&database.Article{}).Where("id = ?", id).
		UpdateColumn("cover", cover).Error
}

// IncrementViews 浏览量 +1
func IncrementViews(ctx context.Context, id uint64) error {
	return global.DB.WithContext(ctx).Model(&database.Article{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1")).Error
}

// ListTopArticles 置顶文章列表
func ListTopArticles(ctx context.Context) ([]database.Article, error) {
	var articles []database.Article
	err := global.DB.WithContext(ctx).Where("status = 1 AND is_top = 1").
		Order("created_at DESC").Find(&articles).Error
	return articles, err
}

// ArchiveByMonth 按月归档已发布文章（返回格式：2026-08 等）
// P0：时间线只用到 id/title/cover/created_at，不再把整篇正文加载进内存。
func ArchiveByMonth(ctx context.Context) (map[string][]database.Article, error) {
	var articles []database.Article
	err := global.DB.WithContext(ctx).
		Select("id, title, cover, created_at").
		Where("status = 1").
		Order("created_at DESC").
		Find(&articles).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string][]database.Article)
	for _, a := range articles {
		key := a.CreatedAt.Format("2006-01")
		result[key] = append(result[key], a)
	}
	return result, nil
}
