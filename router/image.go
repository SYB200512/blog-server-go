package router

import (
	"github.com/gin-gonic/gin"

	"blog-backend/api"
)

// ImageRouter 图片管理路由
type ImageRouter struct{}

// InitPublicRouter 公开图片路由（图片墙）
func (r *ImageRouter) InitPublicRouter(public *gin.RouterGroup) {
	public.GET("/photos", api.ApiGroupApp.ImageApi.ListPhotos)
}

// InitAdminRouter 后台图片管理路由
func (r *ImageRouter) InitAdminRouter(admin *gin.RouterGroup) {
	images := admin.Group("/images")
	{
		images.POST("/upload", api.ApiGroupApp.ImageApi.UploadImage)
		images.GET("", api.ApiGroupApp.ImageApi.ListImages)
		images.DELETE("", api.ApiGroupApp.ImageApi.DeleteImages)
		images.PUT("/category", api.ApiGroupApp.ImageApi.UpdateCategory)
	}
}
