package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"blog-backend/model/response"
)

// BodyLimit 请求体大小限制中间件。
// JSON 与 multipart 使用不同的上限：JSON 防止超大文本拖垮内存，multipart 给图片上传留余量。
// Content-Length 已知时先快速拒绝；未知（chunked）时用 MaxBytesReader 在读取过程中兜底。
func BodyLimit(jsonLimit, uploadLimit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		var limit int64
		switch {
		case strings.HasPrefix(c.ContentType(), "application/json"):
			limit = jsonLimit
		case strings.HasPrefix(c.ContentType(), "multipart/form-data"):
			limit = uploadLimit
		default:
			// GET、静态资源等没有请求体，直接放行
			c.Next()
			return
		}

		// 已知长度时直接拒绝，避免把超大 body 读进内存
		if c.Request.ContentLength > limit {
			c.Abort()
			response.Fail(c, http.StatusRequestEntityTooLarge, response.CodeInvalidParam, "请求体过大")
			return
		}
		// 未知长度时限制读取上限，超过会返回 *http.MaxBytesError
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
