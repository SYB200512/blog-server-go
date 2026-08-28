package services

import (
	"errors"

	"gorm.io/gorm"

	"blog-backend/common"
	"blog-backend/dao"
	"blog-backend/dto"
	"blog-backend/models"
)

// ListArticles 前台文章列表
func ListArticles(q *dto.ArticleQuery) (*common.PageResult, error) {
	p := &common.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	articles, total, err := dao.ListArticles(p, q.Keyword, q.CategoryID, q.TagID, nil, false)
	if err != nil {
		return nil, err
	}
	return &common.PageResult{List: articles, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// AdminListArticles 后台文章列表
func AdminListArticles(q *dto.ArticleQuery) (*common.PageResult, error) {
	p := &common.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	articles, total, err := dao.ListArticles(p, q.Keyword, q.CategoryID, q.TagID, q.Status, true)
	if err != nil {
		return nil, err
	}
	return &common.PageResult{List: articles, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// GetArticle 前台文章详情（浏览量 +1）
func GetArticle(id uint64) (*models.Article, error) {
	article, err := dao.GetPublishedArticleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}
	// 异步/忽略错误地增加浏览量
	_ = dao.IncrementViews(id)
	article.Views++
	return article, nil
}

// AdminGetArticle 后台文章详情
func AdminGetArticle(id uint64) (*models.Article, error) {
	article, err := dao.GetArticleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}
	return article, nil
}

// CreateArticle 创建文章
func CreateArticle(req *dto.ArticleCreateRequest) (*models.Article, error) {
	article := &models.Article{
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
		return nil, err
	}
	return article, nil
}

// UpdateArticle 更新文章
func UpdateArticle(id uint64, req *dto.ArticleUpdateRequest) (*models.Article, error) {
	article, err := dao.GetArticleByID(id)
	if err != nil {
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
	if err := dao.UpdateArticle(article, req.TagIDs); err != nil {
		return nil, err
	}
	return article, nil
}

// DeleteArticle 删除文章（软删除）
func DeleteArticle(id uint64) error {
	if _, err := dao.GetArticleByID(id); err != nil {
		return errors.New("文章不存在")
	}
	return dao.DeleteArticle(id)
}

// Archive 时间归档
func Archive() (map[string][]models.Article, error) {
	return dao.ArchiveByMonth()
}
