package router

import (
	"github.com/gin-gonic/gin"

	"blog-backend/api"
)

// CoverRouter 封面资源路由
type CoverRouter struct{}

// InitPublicRouter 前台封面图片路由
func (r *CoverRouter) InitPublicRouter(public *gin.RouterGroup) {
	public.GET("/covers/:filename", api.ApiGroupApp.CoverApi.GetCover)
}
