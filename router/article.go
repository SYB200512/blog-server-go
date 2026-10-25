package router

import (
	"github.com/gin-gonic/gin"

	"blog-backend/api"
)

// ArticleRouter 文章路由
type ArticleRouter struct{}

// InitPublicRouter 前台文章路由
func (r *ArticleRouter) InitPublicRouter(public *gin.RouterGroup) {
	public.GET("/articles", api.ApiGroupApp.ArticleApi.ListArticles)
	public.GET("/articles/:id", api.ApiGroupApp.ArticleApi.GetArticle)
	public.GET("/archive", api.ApiGroupApp.ArticleApi.Archive)
}

// InitAdminRouter 后台文章管理路由
func (r *ArticleRouter) InitAdminRouter(admin *gin.RouterGroup) {
	admin.GET("/articles", api.ApiGroupApp.ArticleApi.AdminListArticles)
	admin.GET("/articles/:id", api.ApiGroupApp.ArticleApi.AdminGetArticle)
	admin.POST("/articles", api.ApiGroupApp.ArticleApi.CreateArticle)
	admin.PUT("/articles/:id", api.ApiGroupApp.ArticleApi.UpdateArticle)
	admin.DELETE("/articles/:id", api.ApiGroupApp.ArticleApi.DeleteArticle)
}
