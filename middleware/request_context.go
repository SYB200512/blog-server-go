package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"blog-backend/utils/traceid"
)

// uploadPathPrefix 上传接口路径，需要比普通接口更宽的预算
const uploadPathPrefix = "/api/v1/admin/images/upload"

// uploadTimeout 上传接口超时预算
const uploadTimeout = 60 * time.Second

// RequestContext 为每个请求注入 trace id（不带超时）。
// 必须挂在 GinLogger 之前，否则日志里取不到 trace id。
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// trace id 属于“请求域元数据”：只放这一个值，不放业务参数
		ctx := traceid.With(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Timeout 为请求设置超时预算：到点后 ctx 被取消，下游的 DB 查询、
// HTTP 调用等支持 ctx 的操作会随之中断。
// 注意：取消是协作式的——如果下游不检查 ctx.Done()，它仍会跑完。
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d <= 0 {
			d = 10 * time.Second // 配置缺失时的兜底值
		}
		budget := d
		// 上传大文件通常超过 10s，按路径前缀单独放宽到 60s
		if strings.HasPrefix(c.Request.URL.Path, uploadPathPrefix) {
			budget = uploadTimeout
		}
		// 从请求 ctx 派生：客户端断开（父取消）或自身到点都会关闭 Done
		ctx, cancel := context.WithTimeout(c.Request.Context(), budget)
		// 请求结束必须 cancel：释放定时器，并从父 ctx 的子节点列表中摘除自己
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
