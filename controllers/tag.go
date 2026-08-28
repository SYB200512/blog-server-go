package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/common"
	"blog-backend/dto"
	"blog-backend/services"
)

// ListTags 前台：标签列表
func ListTags(c *gin.Context) {
	tags, err := services.ListTags()
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, tags)
}

// CreateTag 后台：创建标签
func CreateTag(c *gin.Context) {
	var req dto.TagCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.InvalidParam(c, "标签名不能为空")
		return
	}
	tag, err := services.CreateTag(&req)
	if err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, tag)
}

// DeleteTag 后台：删除标签
func DeleteTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的标签 ID")
		return
	}
	if err := services.DeleteTag(id); err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, nil)
}
