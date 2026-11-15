package router

import (
	"github.com/gin-gonic/gin"

	"blog-backend/api"
	"blog-backend/middleware"
)

// AuthRouter 认证路由
type AuthRouter struct{}

// InitAuthRouter 注册认证相关路由
func (r *AuthRouter) InitAuthRouter(public *gin.RouterGroup) {
	auth := public.Group("/auth")
	{
		// 登录
		auth.POST("/login", api.ApiGroupApp.AuthApi.Login)
		// 退出登录
		auth.POST("/logout", api.ApiGroupApp.AuthApi.Logout)
		// 获取用户信息，
		auth.GET("/profile", middleware.JWTAuth(), api.ApiGroupApp.AuthApi.GetProfile)
	}
}
