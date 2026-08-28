package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/common"
	"blog-backend/dto"
	"blog-backend/services"
)

// ListTags 前台：标签列表
// @Summary 标签列表
// @Tags 前台-标签
// @Produce json
// @Success 200 {object} common.Response{data=[]models.Tag} "成功"
// @Router /api/v1/tags [get]
// @Id tag_list
func ListTags(c *gin.Context) {
	tags, err := services.ListTags()
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, tags)
}

// CreateTag 后台：创建标签
// @Summary 创建标签
// @Tags 后台-标签
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.TagCreateRequest true "标签信息"
// @Success 200 {object} common.Response{data=models.Tag} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/tags [post]
// @Id admin_tag_create
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
// @Summary 删除标签
// @Tags 后台-标签
// @Produce json
// @Security BearerAuth
// @Param id path int true "标签ID"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/tags/{id} [delete]
// @Id admin_tag_delete
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
