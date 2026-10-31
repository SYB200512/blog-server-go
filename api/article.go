package api

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/model/request"
	"blog-backend/model/response"
)

// ArticleApi 文章接口
type ArticleApi struct{}

// ListArticles 前台：文章列表
// @Summary 前台文章列表
// @Tags 前台-文章
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "标题/摘要关键词"
// @Param category_id query int false "分类ID"
// @Param tag_id query int false "标签ID"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]database.Article}} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/articles [get]
// @Id article_list
func (a *ArticleApi) ListArticles(c *gin.Context) {
	var q request.ArticleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.InvalidParam(c, "查询参数错误")
		return
	}
	result, err := articleService.ListArticles(c.Request.Context(), &q)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, result)
}

// GetArticle 前台：文章详情
// @Summary 前台文章详情
// @Tags 前台-文章
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response{data=database.Article} "成功"
// @Failure 404 {object} response.Response "文章不存在"
// @Router /api/v1/articles/{id} [get]
// @Id article_get
func (a *ArticleApi) GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的文章 ID")
		return
	}
	// 传请求 ctx：客户端断开或请求超时会中断下方的 DB 查询
	article, err := articleService.GetArticle(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			response.FromError(c, err)
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, article)
}

// AdminListArticles 后台：文章列表
// @Summary 后台文章列表
// @Tags 后台-文章
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "标题/摘要关键词"
// @Param category_id query int false "分类ID"
// @Param tag_id query int false "标签ID"
// @Param status query int false "状态: 0 草稿, 1 已发布"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]database.Article}} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/articles [get]
// @Id admin_article_list
func (a *ArticleApi) AdminListArticles(c *gin.Context) {
	var q request.ArticleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.InvalidParam(c, "查询参数错误")
		return
	}
	result, err := articleService.AdminListArticles(c.Request.Context(), &q)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, result)
}

// AdminGetArticle 后台：文章详情
// @Summary 后台文章详情
// @Tags 后台-文章
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response{data=database.Article} "成功"
// @Failure 404 {object} response.Response "文章不存在"
// @Router /api/v1/admin/articles/{id} [get]
// @Id admin_article_get
func (a *ArticleApi) AdminGetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的文章 ID")
		return
	}
	article, err := articleService.AdminGetArticle(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			response.FromError(c, err)
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, article)
}

// CreateArticle 后台：创建文章
// @Summary 创建文章
// @Tags 后台-文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.ArticleCreateRequest true "文章信息"
// @Success 200 {object} response.Response{data=database.Article} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/articles [post]
// @Id admin_article_create
func (a *ArticleApi) CreateArticle(c *gin.Context) {
	var req request.ArticleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "标题不能为空")
		return
	}
	article, err := articleService.CreateArticle(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, article)
}

// UpdateArticle 后台：更新文章
// @Summary 更新文章
// @Tags 后台-文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章ID"
// @Param body body request.ArticleUpdateRequest true "文章信息"
// @Success 200 {object} response.Response{data=database.Article} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/articles/{id} [put]
// @Id admin_article_update
func (a *ArticleApi) UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的文章 ID")
		return
	}
	var req request.ArticleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体格式错误")
		return
	}
	article, err := articleService.UpdateArticle(c.Request.Context(), id, &req)
	if err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, article)
}

// DeleteArticle 后台：删除文章
// @Summary 删除文章
// @Tags 后台-文章
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/articles/{id} [delete]
// @Id admin_article_delete
func (a *ArticleApi) DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的文章 ID")
		return
	}
	if err := articleService.DeleteArticle(c.Request.Context(), id); err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, nil)
}

// Archive 前台：时间归档
// @Summary 按时间归档
// @Tags 前台-文章
// @Produce json
// @Success 200 {object} response.Response{data=map[string][]database.Article} "成功"
// @Router /api/v1/archive [get]
// @Id article_archive
func (a *ArticleApi) Archive(c *gin.Context) {
	result, err := articleService.Archive(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, result)
}
