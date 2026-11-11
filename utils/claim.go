package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blog-backend/global"
)

const (
	// AccessTokenHeader access token 放在自定义请求头里，避免与第三方 Bearer 方案混淆。
	AccessTokenHeader = "x-access-token"
	// RefreshTokenCookie refresh token 只放在 HttpOnly Cookie，JS 无法读取。
	RefreshTokenCookie = "x-refresh-token"
)

// GetAccessToken 从请求头读取 access token。
func GetAccessToken(c *gin.Context) string {
	return c.GetHeader(AccessTokenHeader)
}

// GetRefreshToken 从 Cookie 读取 refresh token。
func GetRefreshToken(c *gin.Context) string {
	token, err := c.Cookie(RefreshTokenCookie)
	if err != nil {
		return ""
	}
	return token
}

// SetRefreshToken 写入 HttpOnly refresh cookie。
// maxAge 单位为秒，传负数表示立即过期。
func SetRefreshToken(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(RefreshTokenCookie, token, maxAge, "/", "", global.Config.JWT.CookieSecure, true)
}

// ClearRefreshToken 清除 refresh cookie。
func ClearRefreshToken(c *gin.Context) {
	SetRefreshToken(c, "", -1)
}
