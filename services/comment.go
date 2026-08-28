package services

import (
	"errors"

	"blog-backend/common"
	"blog-backend/dao"
	"blog-backend/dto"
	"blog-backend/models"
)

// ListComments 后台评论列表
func ListComments(q *dto.CommentQuery) (*common.PageResult, error) {
	p := &common.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	comments, total, err := dao.ListComments(p, 0, q.Status)
	if err != nil {
		return nil, err
	}
	return &common.PageResult{List: comments, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// ListCommentsByArticle 前台文章评论列表
func ListCommentsByArticle(articleID uint64) ([]models.Comment, error) {
	return dao.ListCommentsByArticle(articleID)
}

// CreateComment 发表评论
func CreateComment(articleID uint64, userID *uint64, req *dto.CommentCreateRequest) (*models.Comment, error) {
	// 校验文章存在且已发布
	if _, err := dao.GetPublishedArticleByID(articleID); err != nil {
		return nil, errors.New("文章不存在")
	}
	comment := &models.Comment{
		ArticleID: articleID,
		UserID:    userID,
		Content:   req.Content,
		Email:     req.Email,
		ParentID:  req.ParentID,
	}
	if userID == nil {
		if req.Content == "" {
			return nil, errors.New("评论内容不能为空")
		}
	}
	if err := dao.CreateComment(comment); err != nil {
		return nil, err
	}
	return comment, nil
}

// ReviewComment 审核评论（status: 0 待审核，1 通过）
func ReviewComment(id uint64, status int8) error {
	return dao.UpdateCommentStatus(id, status)
}
