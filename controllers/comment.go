package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/common"
	"blog-backend/dto"
	"blog-backend/services"
)

// ListCommentsByArticle 前台：文章评论列表
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
