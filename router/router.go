package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"blog-backend/controllers"
	_ "blog-backend/docs"
	"blog-backend/middleware"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 0, "message": "ok", "data": nil})
	})

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// ===== 认证 =====
		auth := api.Group("/auth")
		{
			auth.POST("/login", controllers.Login)
			auth.GET("/profile", middleware.JWTAuth(), controllers.GetProfile)
		}

		// ===== 前台（公开）=====
		api.GET("/articles", controllers.ListArticles)
		api.GET("/articles/:id", controllers.GetArticle)
		api.GET("/categories", controllers.ListCategories)
		api.GET("/tags", controllers.ListTags)
		api.GET("/archive", controllers.Archive)
		api.GET("/articles/:id/comments", controllers.ListCommentsByArticle)
		api.POST("/articles/:id/comments", controllers.CreateComment)

		// ===== 后台（需登录 + 管理员）=====
		admin := api.Group("/admin", middleware.JWTAuth(), middleware.AdminRequired())
		{
			// 文章管理
			admin.GET("/articles", controllers.AdminListArticles)
			admin.GET("/articles/:id", controllers.AdminGetArticle)
			admin.POST("/articles", controllers.CreateArticle)
			admin.PUT("/articles/:id", controllers.UpdateArticle)
			admin.DELETE("/articles/:id", controllers.DeleteArticle)

			// 分类管理
			admin.POST("/categories", controllers.CreateCategory)
			admin.PUT("/categories/:id", controllers.UpdateCategory)
			admin.DELETE("/categories/:id", controllers.DeleteCategory)

			// 标签管理
			admin.POST("/tags", controllers.CreateTag)
			admin.DELETE("/tags/:id", controllers.DeleteTag)

			// 评论管理
			admin.GET("/comments", controllers.AdminListComments)
			admin.PUT("/comments/:id", controllers.ReviewComment)
		}
	}
	return r
}
