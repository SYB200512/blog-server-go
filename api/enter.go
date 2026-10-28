package api

import "blog-backend/service"

// ApiGroup 接口层聚合结构体
type ApiGroup struct {
	ArticleApi
	AuthApi
	CategoryApi
	CommentApi
	CoverApi
	ImageApi
}

// ApiGroupApp 接口层全局单例
var ApiGroupApp = new(ApiGroup)

// 各 API 持有的 service 引用（层间解耦）
var (
	articleService  = service.ServiceGroupApp.ArticleService
	authService     = service.ServiceGroupApp.AuthService
	categoryService = service.ServiceGroupApp.CategoryService
	commentService  = service.ServiceGroupApp.CommentService
	imageService    = service.ServiceGroupApp.ImageService
)
