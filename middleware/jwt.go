package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"blog-backend/common"
)

// JWTAuth JWT 认证中间件
// 校验通过后，将 userID / username / role 写入上下文
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		// 支持 "Bearer <token>" 或裸 token
		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
		claims, err := common.ParseToken(tokenString)
		if err != nil {
			common.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminRequired 管理员权限校验（需在 JWTAuth 之后使用）
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			common.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		if role.(int8) != 1 {
			common.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
