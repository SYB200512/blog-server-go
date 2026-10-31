package response

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

// 业务状态码
const (
	CodeSuccess      = 0    // 成功
	CodeInvalidParam = 4000 // 参数错误
	CodeUnauthorized = 4001 // 未认证
	CodeForbidden    = 4003 // 无权限
	CodeNotFound     = 4004 // 资源不存在
	CodeServerError  = 5000 // 服务器内部错误
)

// Response 统一响应体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{Code: CodeSuccess, Message: "success", Data: data})
}

// Fail 失败响应（自定义状态码）
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: nil})
}

// ServerError 服务器内部错误
func ServerError(c *gin.Context, err error) {
	c.JSON(500, Response{Code: CodeServerError, Message: "服务器内部错误", Data: nil})
}

// FromError 按错误类型返回：
//   - DeadlineExceeded：请求预算用完，返回 504，提示超时；
//   - Canceled：客户端已断开（如关闭页面），返回 499 且不写响应体；
//   - 其它：按服务器内部错误返回 500。
func FromError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		Fail(c, 504, CodeServerError, "请求处理超时")
	case errors.Is(err, context.Canceled):
		c.AbortWithStatus(499)
	default:
		ServerError(c, err)
	}
}

// InvalidParam 参数错误
func InvalidParam(c *gin.Context, message string) {
	Fail(c, 400, CodeInvalidParam, message)
}

// Unauthorized 未认证
func Unauthorized(c *gin.Context, message string) {
	Fail(c, 401, CodeUnauthorized, message)
}

// Forbidden 无权限
func Forbidden(c *gin.Context) {
	Fail(c, 403, CodeForbidden, "无权限访问")
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Fail(c, 404, CodeNotFound, message)
}
