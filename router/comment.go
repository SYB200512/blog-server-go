package router

import (
	"github.com/gin-gonic/gin"

	"blog-backend/api"
)

// CommentRouter 评论路由
type CommentRouter struct{}

// InitPublicRouter 前台评论路由
func (r *CommentRouter) InitPublicRouter(public *gin.RouterGroup) {
	public.GET("/articles/:id/comments", api.ApiGroupApp.CommentApi.ListCommentsByArticle)
	public.POST("/articles/:id/comments", api.ApiGroupApp.CommentApi.CreateComment)
}

// InitAdminRouter 后台评论管理路由
func (r *CommentRouter) InitAdminRouter(admin *gin.RouterGroup) {
	admin.GET("/comments", api.ApiGroupApp.CommentApi.AdminListComments)
	admin.PUT("/comments/:id", api.ApiGroupApp.CommentApi.ReviewComment)
}
