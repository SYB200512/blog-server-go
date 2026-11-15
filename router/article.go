package router

import (
	"github.com/gin-gonic/gin"

	"blog-backend/api"
)

// ArticleRouter 文章路由
type ArticleRouter struct{}

// InitPublicRouter 前台文章路由
func (r *ArticleRouter) InitPublicRouter(public *gin.RouterGroup) {
	//文章列表
	public.GET("/articles", api.ApiGroupApp.ArticleApi.ListArticles)
	//文章详情
	public.GET("/articles/:id", api.ApiGroupApp.ArticleApi.GetArticle)
	//文章归档
	public.GET("/archive", api.ApiGroupApp.ArticleApi.Archive)
}

// InitAdminRouter 后台文章管理路由
func (r *ArticleRouter) InitAdminRouter(admin *gin.RouterGroup) {
	//文章列表
	admin.GET("/articles", api.ApiGroupApp.ArticleApi.AdminListArticles)
	//文章详情
	admin.GET("/articles/:id", api.ApiGroupApp.ArticleApi.AdminGetArticle)
	//创建文章
	admin.POST("/articles", api.ApiGroupApp.ArticleApi.CreateArticle)
	//更新文章
	admin.PUT("/articles/:id", api.ApiGroupApp.ArticleApi.UpdateArticle)
	//删除文章
	admin.DELETE("/articles/:id", api.ApiGroupApp.ArticleApi.DeleteArticle)
}
