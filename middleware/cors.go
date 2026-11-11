package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, x-access-token")
			c.Header("Access-Control-Allow-Credentials", "true")
			// 无感续签的新 token 通过响应头下发，跨域时必须显式暴露，否则浏览器读不到
			c.Header("Access-Control-Expose-Headers", "new-access-token, new-access-expires-at")
			c.Header("Access-Control-Max-Age", "86400")
		}
		// 预检请求直接返回
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
