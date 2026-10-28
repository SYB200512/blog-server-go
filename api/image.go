package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-backend/model/appTypes"
	"blog-backend/model/request"
	"blog-backend/model/response"
)

// ImageApi 图片管理接口
type ImageApi struct{}

// UploadImage 后台：上传图片
// @Summary 上传图片
// @Tags 后台-图片
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param image formData file true "图片文件"
// @Param category formData string false "分类：未使用/封面/插图/头像/系统"
// @Success 200 {object} response.Response{data=response.ImageUploadResult} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/images/upload [post]
// @Id admin_image_upload
func (a *ImageApi) UploadImage(c *gin.Context) {
	_, header, err := c.Request.FormFile("image")
	if err != nil {
		response.InvalidParam(c, "请选择要上传的图片")
		return
	}
	category := appTypes.Unused
	if s := c.PostForm("category"); s != "" {
		category = appTypes.ToImageCategory(s)
	}

	result, err := imageService.UploadImage(header, category)
	if err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, result)
}

// ListPhotos 公开：图片墙
// @Summary 图片墙图片列表
// @Tags 图片
// @Produce json
// @Param limit query int false "返回数量" default(60)
// @Success 200 {object} response.Response{data=[]database.Image} "成功"
// @Router /api/v1/photos [get]
// @Id photo_list
func (a *ImageApi) ListPhotos(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "60"))
	list, err := imageService.ListPhotos(limit)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.Success(c, list)
}

// ListImages 后台：图片列表
// @Summary 图片列表
// @Tags 后台-图片
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param category query string false "分类中文名"
// @Param keyword query string false "文件名关键词"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]database.Image}} "成功"
// @Router /api/v1/admin/images [get]
// @Id admin_image_list
func (a *ImageApi) ListImages(c *gin.Context) {
	var q request.ImageListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.InvalidParam(c, "查询参数错误")
		return
	}
	result, err := imageService.ListImages(&q)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateCategory 后台：批量修改图片分类
// @Summary 批量修改图片分类
// @Tags 后台-图片
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.ImageCategoryRequest true "ID 列表与分类"
// @Success 200 {object} response.Response "成功"
// @Router /api/v1/admin/images/category [put]
// @Id admin_image_category
func (a *ImageApi) UpdateCategory(c *gin.Context) {
	var req request.ImageCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误")
		return
	}
	if err := imageService.UpdateCategory(req.IDs, appTypes.ToImageCategory(req.Category)); err != nil {
		response.ServerError(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteImages 后台：批量删除图片
// @Summary 批量删除图片
// @Tags 后台-图片
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.ImageDeleteRequest true "图片 ID 列表"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/admin/images [delete]
// @Id admin_image_delete
func (a *ImageApi) DeleteImages(c *gin.Context) {
	var req request.ImageDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请选择要删除的图片")
		return
	}
	if err := imageService.DeleteImages(req.IDs); err != nil {
		response.ServerError(c, err)
		return
	}
	response.Success(c, nil)
}
