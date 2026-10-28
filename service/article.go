package service

import (
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/request"
	"blog-backend/model/response"
)

// ArticleService 文章业务
type ArticleService struct{}

// ListArticles 前台文章列表
func (s *ArticleService) ListArticles(q *request.ArticleQuery) (*response.PageResult, error) {
	p := &request.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	articles, total, err := dao.ListArticles(p, q.Keyword, q.CategoryID, q.TagID, nil, false)
	if err != nil {
		global.Log.Error("查询前台文章列表失败", zap.Error(err))
		return nil, err
	}
	return &response.PageResult{List: articles, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// AdminListArticles 后台文章列表
func (s *ArticleService) AdminListArticles(q *request.ArticleQuery) (*response.PageResult, error) {
	p := &request.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	articles, total, err := dao.ListArticles(p, q.Keyword, q.CategoryID, q.TagID, q.Status, true)
	if err != nil {
		global.Log.Error("查询后台文章列表失败", zap.Error(err))
		return nil, err
	}
	return &response.PageResult{List: articles, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// GetArticle 前台文章详情（浏览量 +1）
func (s *ArticleService) GetArticle(id uint64) (*database.Article, error) {
	article, err := dao.GetPublishedArticleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.Log.Warn("前台文章不存在", zap.Uint64("article_id", id))
			return nil, errors.New("文章不存在")
		}
		global.Log.Error("查询前台文章详情失败", zap.Uint64("article_id", id), zap.Error(err))
		return nil, err
	}
	// 异步/忽略错误地增加浏览量
	_ = dao.IncrementViews(id)
	article.Views++
	return article, nil
}

// AdminGetArticle 后台文章详情
func (s *ArticleService) AdminGetArticle(id uint64) (*database.Article, error) {
	article, err := dao.GetArticleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.Log.Warn("后台文章不存在", zap.Uint64("article_id", id))
			return nil, errors.New("文章不存在")
		}
		global.Log.Error("查询后台文章详情失败", zap.Uint64("article_id", id), zap.Error(err))
		return nil, err
	}
	return article, nil
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(req *request.ArticleCreateRequest) (*database.Article, error) {
	article := &database.Article{
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		Cover:      req.Cover,
		CategoryID: req.CategoryID,
		Status:     req.Status,
		IsTop:      req.IsTop,
	}
	// 默认发布
	if article.Status != 0 {
		article.Status = 1
	}
	if err := dao.CreateArticle(article, req.TagIDs); err != nil {
		global.Log.Error("创建文章失败", zap.String("title", article.Title), zap.Error(err))
		return nil, err
	}
	// 回写图片台账分类：封面 / 正文插图
	ServiceGroupApp.ImageService.MarkArticleImages(article.Cover, article.Content)
	// 发布且未指定封面时，异步生成封面并回写（不阻塞发布）
	if article.Status == 1 && article.Cover == "" {
		s.generateCoverAsync(article.ID, article.Title, article.Summary)
	}
	return article, nil
}

// generateCoverAsync 后台协程生成封面，成功后回写 cover 字段
func (s *ArticleService) generateCoverAsync(id uint64, title, summary string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Log.Error("生成封面协程 panic", zap.Any("error", r), zap.Uint64("article_id", id))
			}
		}()
		cover, err := generateArticleCover(title, summary)
		if err != nil {
			global.Log.Error("自动生成封面失败（异步）", zap.Uint64("article_id", id), zap.Error(err))
			return
		}
		if cover == "" {
			return
		}
		if err := dao.UpdateArticleCover(id, cover); err != nil {
			global.Log.Error("回写封面失败", zap.Uint64("article_id", id), zap.Error(err))
		} else {
			global.Log.Info("文章封面已生成", zap.Uint64("article_id", id), zap.String("cover", cover))
		}
	}()
}

// UpdateArticle 更新文章
func (s *ArticleService) UpdateArticle(id uint64, req *request.ArticleUpdateRequest) (*database.Article, error) {
	article, err := dao.GetArticleByID(id)
	if err != nil {
		global.Log.Warn("更新文章时文章不存在", zap.Uint64("article_id", id))
		return nil, errors.New("文章不存在")
	}
	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Summary != "" {
		article.Summary = req.Summary
	}
	if req.Content != "" {
		article.Content = req.Content
	}
	if req.Cover != "" {
		article.Cover = req.Cover
	}
	if req.CategoryID != nil {
		article.CategoryID = req.CategoryID
	}
	if req.Status != nil {
		article.Status = *req.Status
	}
	if req.IsTop != nil {
		article.IsTop = *req.IsTop
	}
	// 已发布且仍未设置封面时，异步生成封面并回写（不阻塞更新）
	if article.Status == 1 && article.Cover == "" {
		s.generateCoverAsync(article.ID, article.Title, article.Summary)
	}
	if err := dao.UpdateArticle(article, req.TagIDs); err != nil {
		global.Log.Error("更新文章失败", zap.Uint64("article_id", id), zap.Error(err))
		return nil, err
	}
	// 回写图片台账分类：封面 / 正文插图
	ServiceGroupApp.ImageService.MarkArticleImages(article.Cover, article.Content)
	return article, nil
}

// DeleteArticle 删除文章（软删除）
func (s *ArticleService) DeleteArticle(id uint64) error {
	if _, err := dao.GetArticleByID(id); err != nil {
		global.Log.Warn("删除文章时文章不存在", zap.Uint64("article_id", id))
		return errors.New("文章不存在")
	}
	if err := dao.DeleteArticle(id); err != nil {
		global.Log.Error("删除文章失败", zap.Uint64("article_id", id), zap.Error(err))
		return err
	}
	return nil
}

// Archive 时间归档
func (s *ArticleService) Archive() (map[string][]database.Article, error) {
	result, err := dao.ArchiveByMonth()
	if err != nil {
		global.Log.Error("查询时间归档失败", zap.Error(err))
		return nil, err
	}
	return result, nil
}
