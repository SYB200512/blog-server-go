package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"blog-backend/global"
	"blog-backend/utils/traceid"
)

// GinLogger 请求访问日志中间件：记录方法、路径、状态码、耗时与请求中的错误。
// 替代 gin.Logger()，日志统一走 global.Log 落盘。swagger 文档请求不计入，避免刷屏。
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		// 记录查询参数：用于日志分析，避免泄露敏感信息
		// 注意：查询参数可能包含密码、敏感信息等，建议在生产环境过滤或加密
		query := c.Request.URL.RawQuery

		c.Next()

		// 跳过 swagger 文档静态资源与健康检查：前者无业务价值，后者被 systemd/监控反复 ping，纯刷量
		if strings.HasPrefix(path, "/swagger") || path == "/api/health" {
			return
		}
		// 记录请求访问日志：包含 trace id、方法、路径、状态码、耗时、IP、查询参数（如果有）
		status := c.Writer.Status()
		fields := []zap.Field{
			zap.String("trace_id", traceid.From(c.Request.Context())),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.String("ip", c.ClientIP()),
			zap.Duration("cost", time.Since(start)),
		}
		// 记录查询参数：用于日志分析，避免泄露敏感信息
		if query != "" {
			fields = append(fields, zap.String("query", query))
		}
		// 请求处理过程中产生的错误（如业务 panic 前记录的 c.Error）
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}
		// 记录请求处理结果：成功或失败
		if status >= 500 {
			global.Log.Error("请求处理失败", fields...)
		} else {
			global.Log.Info("请求", fields...)
		}
	}
}

// GinRecovery panic 恢复中间件：捕获请求处理中的 panic，记录日志并返回 500。
// 替代 gin.Recovery()，避免服务因单个请求异常而崩溃。
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				global.Log.Error("请求 panic 已恢复",
					zap.Any("error", r),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
				)
				// 写回 500 错误响应：通知客户端服务端内部错误，中止后续处理
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
