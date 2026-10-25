package router

import (
	"github.com/gin-gonic/gin"

	"blog-backend/api"
)

// CategoryRouter 分类/标签路由
type CategoryRouter struct{}

// InitPublicRouter 前台分类/标签路由
func (r *CategoryRouter) InitPublicRouter(public *gin.RouterGroup) {
	public.GET("/categories", api.ApiGroupApp.CategoryApi.ListCategories)
	public.GET("/tags", api.ApiGroupApp.CategoryApi.ListTags)
}

// InitAdminRouter 后台分类/标签管理路由
func (r *CategoryRouter) InitAdminRouter(admin *gin.RouterGroup) {
	admin.POST("/categories", api.ApiGroupApp.CategoryApi.CreateCategory)
	admin.PUT("/categories/:id", api.ApiGroupApp.CategoryApi.UpdateCategory)
	admin.DELETE("/categories/:id", api.ApiGroupApp.CategoryApi.DeleteCategory)

	admin.POST("/tags", api.ApiGroupApp.CategoryApi.CreateTag)
	admin.DELETE("/tags/:id", api.ApiGroupApp.CategoryApi.DeleteTag)
}
