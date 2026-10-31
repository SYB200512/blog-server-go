package initialize

import (
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "blog-backend/docs"
	"blog-backend/global"
	"blog-backend/middleware"
	"blog-backend/router"
)

// InitRouter 初始化路由：挂载静态文件、创建公开/后台两个路由组
func InitRouter() *gin.Engine {
	r := gin.New()
	// RequestContext 必须在 Logger 之前，保证日志能带上 trace id
	r.Use(middleware.RequestContext(), middleware.GinLogger(), middleware.GinRecovery(), middleware.CORS())

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 0, "message": "ok", "data": nil})
	})

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 请求处理预算：普通接口 10 秒（可配），上传接口在 Timeout 内部放宽到 60 秒；
	// 配置缺失或非法时兜底 10 秒
	budget := time.Duration(global.Config.Server.RequestTimeout) * time.Second
	if budget <= 0 {
		budget = 10 * time.Second
	}

	// 公开路由组
	public := r.Group("/api/v1")
	public.Use(middleware.Timeout(budget))
	// 后台管理路由组（JWT + 管理员双重校验）
	admin := r.Group("/api/v1/admin", middleware.JWTAuth(), middleware.AdminRequired())
	admin.Use(middleware.Timeout(budget))

	router.RouterGroupApp.AuthRouter.InitAuthRouter(public)
	router.RouterGroupApp.ArticleRouter.InitPublicRouter(public)
	router.RouterGroupApp.ArticleRouter.InitAdminRouter(admin)
	router.RouterGroupApp.CategoryRouter.InitPublicRouter(public)
	router.RouterGroupApp.CategoryRouter.InitAdminRouter(admin)
	router.RouterGroupApp.CommentRouter.InitPublicRouter(public)
	router.RouterGroupApp.CommentRouter.InitAdminRouter(admin)
	router.RouterGroupApp.CoverRouter.InitPublicRouter(public)
	router.RouterGroupApp.ImageRouter.InitPublicRouter(public)
	router.RouterGroupApp.ImageRouter.InitAdminRouter(admin)
	return r
}
