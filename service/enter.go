package service

// ServiceGroup 服务层聚合结构体
type ServiceGroup struct {
	ArticleService
	AuthService
	CategoryService
	CommentService
}

// ServiceGroupApp 服务层全局单例
var ServiceGroupApp = new(ServiceGroup)
