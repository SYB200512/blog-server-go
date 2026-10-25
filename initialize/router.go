package initialize

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "blog-backend/docs"
	"blog-backend/middleware"
	"blog-backend/router"
)

// InitRouter 初始化路由：挂载静态文件、创建公开/后台两个路由组
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.GinLogger(), middleware.GinRecovery(), middleware.CORS())

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 0, "message": "ok", "data": nil})
	})

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 公开路由组
	public := r.Group("/api/v1")
	// 后台管理路由组（JWT + 管理员双重校验）
	admin := r.Group("/api/v1/admin", middleware.JWTAuth(), middleware.AdminRequired())

	router.RouterGroupApp.AuthRouter.InitAuthRouter(public)
	router.RouterGroupApp.ArticleRouter.InitPublicRouter(public)
	router.RouterGroupApp.ArticleRouter.InitAdminRouter(admin)
	router.RouterGroupApp.CategoryRouter.InitPublicRouter(public)
	router.RouterGroupApp.CategoryRouter.InitAdminRouter(admin)
	router.RouterGroupApp.CommentRouter.InitPublicRouter(public)
	router.RouterGroupApp.CommentRouter.InitAdminRouter(admin)
	router.RouterGroupApp.CoverRouter.InitPublicRouter(public)
	return r
}
