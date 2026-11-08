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

// 各 API 持有的 service 引用（层间解耦）。
// 必须用指针：ArticleService / CategoryService 内含 singleflight.Group（带锁），
// 按值拷贝会复制锁，go vet copylocks 会报错
var (
	articleService  = &service.ServiceGroupApp.ArticleService
	authService     = &service.ServiceGroupApp.AuthService
	categoryService = &service.ServiceGroupApp.CategoryService
	commentService  = &service.ServiceGroupApp.CommentService
	imageService    = &service.ServiceGroupApp.ImageService
)
