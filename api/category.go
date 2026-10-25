package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/model/request"
	"blog-backend/model/response"
)

// CategoryApi 分类/标签接口
type CategoryApi struct{}

// ListCategories 前台：分类列表
// @Summary 分类列表
// @Tags 前台-分类
// @Produce json
// @Success 200 {object} response.Response{data=[]database.Category} "成功"
// @Router /api/v1/categories [get]
// @Id category_list
func (a *CategoryApi) ListCategories(c *gin.Context) {
	categories, err := categoryService.ListCategories()
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.Success(c, categories)
}

// CreateCategory 后台：创建分类
// @Summary 创建分类
// @Tags 后台-分类
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.CategoryCreateRequest true "分类信息"
// @Success 200 {object} response.Response{data=database.Category} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/categories [post]
// @Id admin_category_create
func (a *CategoryApi) CreateCategory(c *gin.Context) {
	var req request.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "分类名不能为空")
		return
	}
	category, err := categoryService.CreateCategory(&req)
	if err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, category)
}

// UpdateCategory 后台：更新分类
// @Summary 更新分类
// @Tags 后台-分类
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Param body body request.CategoryUpdateRequest true "分类信息"
// @Success 200 {object} response.Response{data=database.Category} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/categories/{id} [put]
// @Id admin_category_update
func (a *CategoryApi) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的分类 ID")
		return
	}
	var req request.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体格式错误")
		return
	}
	category, err := categoryService.UpdateCategory(id, &req)
	if err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, category)
}

// DeleteCategory 后台：删除分类
// @Summary 删除分类
// @Tags 后台-分类
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/categories/{id} [delete]
// @Id admin_category_delete
func (a *CategoryApi) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的分类 ID")
		return
	}
	if err := categoryService.DeleteCategory(id); err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, nil)
}

// ListTags 前台：标签列表
// @Summary 标签列表
// @Tags 前台-标签
// @Produce json
// @Success 200 {object} response.Response{data=[]database.Tag} "成功"
// @Router /api/v1/tags [get]
// @Id tag_list
func (a *CategoryApi) ListTags(c *gin.Context) {
	tags, err := categoryService.ListTags()
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.Success(c, tags)
}

// CreateTag 后台：创建标签
// @Summary 创建标签
// @Tags 后台-标签
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.TagCreateRequest true "标签信息"
// @Success 200 {object} response.Response{data=database.Tag} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/tags [post]
// @Id admin_tag_create
func (a *CategoryApi) CreateTag(c *gin.Context) {
	var req request.TagCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "标签名不能为空")
		return
	}
	tag, err := categoryService.CreateTag(&req)
	if err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, tag)
}

// DeleteTag 后台：删除标签
// @Summary 删除标签
// @Tags 后台-标签
// @Produce json
// @Security BearerAuth
// @Param id path int true "标签ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/tags/{id} [delete]
// @Id admin_tag_delete
func (a *CategoryApi) DeleteTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的标签 ID")
		return
	}
	if err := categoryService.DeleteTag(id); err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, nil)
}
