package router

// RouterGroup 路由层聚合结构体
type RouterGroup struct {
	AuthRouter
	ArticleRouter
	CategoryRouter
	CommentRouter
	CoverRouter
}

// RouterGroupApp 路由层全局单例
var RouterGroupApp = new(RouterGroup)
