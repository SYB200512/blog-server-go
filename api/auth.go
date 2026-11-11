package api

import (
	"github.com/gin-gonic/gin"

	"blog-backend/model/request"
	"blog-backend/model/response"
	"blog-backend/utils"
)

// AuthApi 认证接口
type AuthApi struct{}

// Login 用户登录
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body request.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=response.LoginResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/auth/login [post]
// @Id auth_login
func (a *AuthApi) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "用户名和密码不能为空")
		return
	}
	result, err := authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	// refresh token 只写入 HttpOnly Cookie，响应体里不出现
	utils.SetRefreshToken(c, result.RefreshToken, int(result.RefreshTTL.Seconds()))
	response.Success(c, &response.LoginResponse{
		AccessToken:          result.AccessToken,
		AccessTokenExpiresAt: result.ExpiresAt.UnixMilli(),
		User:                 result.User,
	})
}

// Logout 退出登录
// @Summary 退出登录
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "成功"
// @Router /api/v1/auth/logout [post]
// @Id auth_logout
func (a *AuthApi) Logout(c *gin.Context) {
	authService.Logout(c.Request.Context(), utils.GetRefreshToken(c))
	utils.ClearRefreshToken(c)
	response.Success(c, nil)
}

// GetProfile 获取当前用户信息
// @Summary 获取当前用户信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=response.UserInfo} "成功"
// @Failure 401 {object} response.Response "未认证"
// @Router /api/v1/auth/profile [get]
// @Id auth_profile
func (a *AuthApi) GetProfile(c *gin.Context) {
	userID := c.GetUint64("userID")
	info, err := authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, info)
}
