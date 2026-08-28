package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/common"
	"blog-backend/dto"
	"blog-backend/services"
)

// ListCategories 前台：分类列表
func ListCategories(c *gin.Context) {
	categories, err := services.ListCategories()
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, categories)
}

// CreateCategory 后台：创建分类
func CreateCategory(c *gin.Context) {
	var req dto.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.InvalidParam(c, "分类名不能为空")
		return
	}
	category, err := services.CreateCategory(&req)
	if err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, category)
}

// UpdateCategory 后台：更新分类
func UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的分类 ID")
		return
	}
	var req dto.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.InvalidParam(c, "请求体格式错误")
		return
	}
	category, err := services.UpdateCategory(id, &req)
	if err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, category)
}

// DeleteCategory 后台：删除分类
func DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.InvalidParam(c, "无效的分类 ID")
		return
	}
	if err := services.DeleteCategory(id); err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, nil)
}
