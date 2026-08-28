package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/common"
	"blog-backend/dto"
	"blog-backend/services"
)

// ListCategories 前台：分类列表
// @Summary 分类列表
// @Tags 前台-分类
// @Produce json
// @Success 200 {object} common.Response{data=[]models.Category} "成功"
// @Router /api/v1/categories [get]
// @Id category_list
func ListCategories(c *gin.Context) {
	categories, err := services.ListCategories()
	if err != nil {
		common.ServerError(c, err)
		return
	}
	common.Success(c, categories)
}

// CreateCategory 后台：创建分类
// @Summary 创建分类
// @Tags 后台-分类
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CategoryCreateRequest true "分类信息"
// @Success 200 {object} common.Response{data=models.Category} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/categories [post]
// @Id admin_category_create
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
// @Summary 更新分类
// @Tags 后台-分类
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Param body body dto.CategoryUpdateRequest true "分类信息"
// @Success 200 {object} common.Response{data=models.Category} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/categories/{id} [put]
// @Id admin_category_update
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
// @Summary 删除分类
// @Tags 后台-分类
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/admin/categories/{id} [delete]
// @Id admin_category_delete
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
