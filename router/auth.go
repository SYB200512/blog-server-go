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
		auth.POST("/login", api.ApiGroupApp.AuthApi.Login)
		auth.GET("/profile", middleware.JWTAuth(), api.ApiGroupApp.AuthApi.GetProfile)
	}
}
