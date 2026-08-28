package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/common"
	"blog-backend/dto"
	"blog-backend/services"
)

// ListArticles 前台：文章列表
// @Summary 前台文章列表
// @Tags 前台-文章
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "标题/摘要关键词"
// @Param category_id query int false "分类ID"
// @Param tag_id query int false "标签ID"
// @Success 200 {object} common.Response{data=common.PageResult{list=[]models.Article}} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/articles [get]
// @Id article_list
func ListArticles(c *gin.Context) {
	var q dto.ArticleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.InvalidParam(c, "查询参数错误")
		return
	}
	result, err := services.ListArticles(&q)
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, result)
}

// GetArticle 前台：文章详情
// @Summary 前台文章详情
// @Tags 前台-文章
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} common.Response{data=models.Article} "成功"
// @Failure 404 {object} common.Response "文章不存在"
// @Router /api/v1/articles/{id} [get]
// @Id article_get
func GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的文章 ID")
		return
	}
	article, err := services.GetArticle(id)
	if err != nil {
		common.NotFound(c, err.Error())
		return
	}
	common.Success(c, article)
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
// @Success 200 {object} common.Response{data=common.PageResult{list=[]models.Article}} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/articles [get]
// @Id admin_article_list
func AdminListArticles(c *gin.Context) {
	var q dto.ArticleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.InvalidParam(c, "查询参数错误")
		return
	}
	result, err := services.AdminListArticles(&q)
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, result)
}

// AdminGetArticle 后台：文章详情
// @Summary 后台文章详情
// @Tags 后台-文章
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章ID"
// @Success 200 {object} common.Response{data=models.Article} "成功"
// @Failure 404 {object} common.Response "文章不存在"
// @Router /api/v1/admin/articles/{id} [get]
// @Id admin_article_get
func AdminGetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的文章 ID")
		return
	}
	article, err := services.AdminGetArticle(id)
	if err != nil {
		common.NotFound(c, err.Error())
		return
	}
	common.Success(c, article)
}

// CreateArticle 后台：创建文章
// @Summary 创建文章
// @Tags 后台-文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.ArticleCreateRequest true "文章信息"
// @Success 200 {object} common.Response{data=models.Article} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/articles [post]
// @Id admin_article_create
func CreateArticle(c *gin.Context) {
	var req dto.ArticleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.InvalidParam(c, "标题不能为空")
		return
	}
	article, err := services.CreateArticle(&req)
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, article)
}

// UpdateArticle 后台：更新文章
// @Summary 更新文章
// @Tags 后台-文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章ID"
// @Param body body dto.ArticleUpdateRequest true "文章信息"
// @Success 200 {object} common.Response{data=models.Article} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/articles/{id} [put]
// @Id admin_article_update
func UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的文章 ID")
		return
	}
	var req dto.ArticleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.InvalidParam(c, "请求体格式错误")
		return
	}
	article, err := services.UpdateArticle(id, &req)
	if err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, article)
}

// DeleteArticle 后台：删除文章
// @Summary 删除文章
// @Tags 后台-文章
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章ID"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/articles/{id} [delete]
// @Id admin_article_delete
func DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的文章 ID")
		return
	}
	if err := services.DeleteArticle(id); err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, nil)
}

// Archive 前台：时间归档
// @Summary 按时间归档
// @Tags 前台-文章
// @Produce json
// @Success 200 {object} common.Response{data=map[string][]models.Article} "成功"
// @Router /api/v1/archive [get]
// @Id article_archive
func Archive(c *gin.Context) {
	result, err := services.Archive()
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, result)
}
