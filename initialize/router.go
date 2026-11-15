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

// InitRouter 初始化路由：挂载中间件、静态挂载静态文件、创建公开/后台两个路由组
func InitRouter() *gin.Engine {
	// 创建路由引擎
	r := gin.New()

	// P0：JSON 请求体 1 MB；multipart 按图片配置留 2 MB 余量（配置上限 + 表单开销）
	jsonLimit := int64(1 << 20)
	uploadLimit := int64(global.Config.Storage.MaxSize+2) * 1024 * 1024

	// RequestContext 必须在 Logger 之前，保证日志能带上 trace id
	r.Use(middleware.RequestContext(), middleware.GinLogger(), middleware.GinRecovery(), middleware.CORS(), middleware.BodyLimit(jsonLimit, uploadLimit))

	// 健康检查：主要用于部署流水线，确认服务是否正常运行
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 0, "message": "ok", "data": nil})
	})

	// Swagger 文档 ：线上关闭，避免公开接口清单；本地直接访问 127.0.0.1:8080/swagger/index.html
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

	// 初始化认证路由
	router.RouterGroupApp.AuthRouter.InitAuthRouter(public)
	// 初始化文章路由
	router.RouterGroupApp.ArticleRouter.InitPublicRouter(public)
	// 初始化文章路由（管理员权限）
	router.RouterGroupApp.ArticleRouter.InitAdminRouter(admin)

	// 初始化分类路由
	router.RouterGroupApp.CategoryRouter.InitPublicRouter(public)
	// 初始化分类路由（管理员权限）
	router.RouterGroupApp.CategoryRouter.InitAdminRouter(admin)

	// 初始化评论路由
	router.RouterGroupApp.CommentRouter.InitPublicRouter(public)
	// 初始化评论路由（管理员权限）
	router.RouterGroupApp.CommentRouter.InitAdminRouter(admin)

	// 初始化封面路由
	router.RouterGroupApp.CoverRouter.InitPublicRouter(public)

	// 初始化图片路由
	router.RouterGroupApp.ImageRouter.InitPublicRouter(public)
	// 初始化图片路由（管理员权限）
	router.RouterGroupApp.ImageRouter.InitAdminRouter(admin)
	return r
}
