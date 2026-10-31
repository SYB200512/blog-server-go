package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/request"
	"blog-backend/model/response"
)

// CommentService 评论业务
type CommentService struct{}

// ListComments 后台评论列表
func (s *CommentService) ListComments(ctx context.Context, q *request.CommentQuery) (*response.PageResult, error) {
	p := &request.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	comments, total, err := dao.ListComments(ctx, p, 0, q.Status)
	if err != nil {
		global.Log.Error("查询后台评论列表失败", zap.Error(err))
		return nil, err
	}
	return &response.PageResult{List: comments, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// ListCommentsByArticle 前台文章评论列表
func (s *CommentService) ListCommentsByArticle(ctx context.Context, articleID uint64) ([]database.Comment, error) {
	comments, err := dao.ListCommentsByArticle(ctx, articleID)
	if err != nil {
		global.Log.Error("查询文章评论列表失败", zap.Uint64("article_id", articleID), zap.Error(err))
		return nil, err
	}
	return comments, nil
}

// CreateComment 发表评论
func (s *CommentService) CreateComment(ctx context.Context, articleID uint64, userID *uint64, req *request.CommentCreateRequest) (*database.Comment, error) {
	// 校验文章存在且已发布
	if _, err := dao.GetPublishedArticleByID(ctx, articleID); err != nil {
		global.Log.Warn("发表评论时文章不存在", zap.Uint64("article_id", articleID))
		return nil, errors.New("文章不存在")
	}
	comment := &database.Comment{
		ArticleID: articleID,
		UserID:    userID,
		Content:   req.Content,
		Email:     req.Email,
		ParentID:  req.ParentID,
	}
	if userID == nil {
		if req.Content == "" {
			global.Log.Warn("游客评论内容为空", zap.Uint64("article_id", articleID))
			return nil, errors.New("评论内容不能为空")
		}
	}
	if err := dao.CreateComment(ctx, comment); err != nil {
		global.Log.Error("创建评论失败", zap.Uint64("article_id", articleID), zap.Error(err))
		return nil, err
	}
	return comment, nil
}

// ReviewComment 审核评论（status: 0 待审核，1 通过）
func (s *CommentService) ReviewComment(ctx context.Context, id uint64, status int8) error {
	if err := dao.UpdateCommentStatus(ctx, id, status); err != nil {
		global.Log.Error("审核评论失败", zap.Uint64("comment_id", id), zap.Int8("status", status), zap.Error(err))
		return err
	}
	return nil
}
