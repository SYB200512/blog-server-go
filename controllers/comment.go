package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/common"
	"blog-backend/dto"
	"blog-backend/services"
)

// ListCommentsByArticle 前台：文章评论列表
// @Summary 文章评论列表
// @Tags 前台-评论
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} common.Response{data=[]models.Comment} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/articles/{id}/comments [get]
// @Id comment_list
func ListCommentsByArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的文章 ID")
		return
	}
	comments, err := services.ListCommentsByArticle(articleID)
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, comments)
}

// CreateComment 前台：发表评论
// @Summary 发表评论
// @Tags 前台-评论
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param body body dto.CommentCreateRequest true "评论信息"
// @Success 200 {object} common.Response{data=models.Comment} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/articles/{id}/comments [post]
// @Id comment_create
func CreateComment(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的文章 ID")
		return
	}
	var req dto.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.InvalidParam(c, "评论内容不能为空")
		return
	}
	// 已登录用户与游客均可评论
	var userID *uint64
	if uid, exists := c.Get("userID"); exists {
		id := uid.(uint64)
		userID = &id
	}
	comment, err := services.CreateComment(articleID, userID, &req)
	if err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, comment)
}

// AdminListComments 后台：评论管理列表
// @Summary 后台评论列表
// @Tags 后台-评论
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query int false "状态: 0 待审核, 1 已通过"
// @Success 200 {object} common.Response{data=[]models.Comment} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/comments [get]
// @Id admin_comment_list
func AdminListComments(c *gin.Context) {
	var q dto.CommentQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.InvalidParam(c, "查询参数错误")
		return
	}
	result, err := services.ListComments(&q)
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, result)
}

// ReviewComment 后台：审核评论
// @Summary 审核评论
// @Tags 后台-评论
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Param body body dto.ReviewCommentRequest true "审核信息"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/comments/{id} [put]
// @Id admin_comment_review
func ReviewComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的评论 ID")
		return
	}
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.InvalidParam(c, "请求体格式错误")
		return
	}
	if err := services.ReviewComment(id, req.Status); err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, nil)
}
