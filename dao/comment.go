package dao

import (
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/request"
)

// ListComments 评论分页列表（后台管理用）
func ListComments(p *request.PageParams, articleID uint64, status *int8) ([]database.Comment, int64, error) {
	var comments []database.Comment
	query := global.DB.Model(&database.Comment{})

	if articleID > 0 {
		query = query.Where("article_id = ?", articleID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").
		Order("created_at DESC").
		Offset(p.Offset()).Limit(p.PageSize).
		Find(&comments).Error
	return comments, total, err
}

// ListCommentsByArticle 前台：查询文章的已通过评论（顶层 + 回复）
func ListCommentsByArticle(articleID uint64) ([]database.Comment, error) {
	var comments []database.Comment
	err := global.DB.Preload("User").
		Where("article_id = ? AND status = 1", articleID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// CreateComment 创建评论
func CreateComment(comment *database.Comment) error {
	return global.DB.Create(comment).Error
}

// UpdateCommentStatus 更新评论状态（审核通过/删除）
func UpdateCommentStatus(id uint64, status int8) error {
	return global.DB.Model(&database.Comment{}).Where("id = ?", id).
		Update("status", status).Error
}

// CountComments 评论总数
func CountComments(status *int8) (int64, error) {
	var count int64
	query := global.DB.Model(&database.Comment{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Count(&count).Error
	return count, err
}
