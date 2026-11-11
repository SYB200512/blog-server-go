package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/response"
	"blog-backend/service"
	"blog-backend/utils"
)

var authService = service.ServiceGroupApp.AuthService

// JWTAuth JWT 认证中间件。
//
// 校验 access token 通过后，将 userID / username / role 写入上下文；
// access token 过期时，用 HttpOnly Cookie 中的 refresh token 无感续签，
// 新 token 通过响应头 new-access-token / new-access-expires-at 返回给前端。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		j := utils.NewJWT()
		accessToken := utils.GetAccessToken(c)

		claims, err := j.ParseAccessToken(accessToken)
		if err == nil {
			setAuthContext(c, claims.UserID, claims.Username, claims.Role)
			c.Next()
			return
		}

		// 签名错误、算法不符等非法 token 不触发续签，避免被伪造 token 反复打数据库
		if accessToken != "" && !utils.IsTokenExpired(err) {
			global.Log.Warn("access token 校验失败", zap.String("path", c.Request.URL.Path), zap.Error(err))
			utils.ClearRefreshToken(c)
			response.Unauthorized(c, "登录状态异常，请重新登录")
			c.Abort()
			return
		}

		// access token 缺失或已过期：尝试用 refresh cookie 续签
		refreshToken := utils.GetRefreshToken(c)
		if refreshToken == "" {
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		refreshClaims, err := j.ParseRefreshToken(refreshToken)
		if err != nil {
			global.Log.Warn("refresh token 校验失败", zap.String("path", c.Request.URL.Path), zap.Error(err))
			utils.ClearRefreshToken(c)
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		ctx := c.Request.Context()
		if !authService.ValidateSession(ctx, refreshClaims.UserID, refreshClaims.SessionID, refreshToken) {
			global.Log.Warn("refresh 会话不存在或已注销", zap.Uint64("user_id", refreshClaims.UserID))
			utils.ClearRefreshToken(c)
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		user, err := dao.GetUserByID(ctx, refreshClaims.UserID)
		if err != nil {
			global.Log.Warn("续签失败：用户查询错误", zap.Uint64("user_id", refreshClaims.UserID), zap.Error(err))
			utils.ClearRefreshToken(c)
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}

		newToken, expiresAt, err := j.CreateAccessToken(user.ID, user.Username, user.Role)
		if err != nil {
			global.Log.Error("续签失败：生成 access token 错误", zap.Uint64("user_id", user.ID), zap.Error(err))
			response.FromError(c, err)
			c.Abort()
			return
		}
		c.Header("new-access-token", newToken)
		c.Header("new-access-expires-at", strconv.FormatInt(expiresAt.UnixMilli(), 10))
		setAuthContext(c, user.ID, user.Username, user.Role)
		c.Next()
	}
}

func setAuthContext(c *gin.Context, userID uint64, username string, role int8) {
	c.Set("userID", userID)
	c.Set("username", username)
	c.Set("role", role)
}

// AdminRequired 管理员权限校验（需在 JWTAuth 之后使用）
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		if role.(int8) != 1 {
			response.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
