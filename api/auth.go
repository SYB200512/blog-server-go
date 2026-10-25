package api

import (
	"github.com/gin-gonic/gin"

	"blog-backend/model/request"
	"blog-backend/model/response"
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
	resp, err := authService.Login(&req)
	if err != nil {
		response.Fail(c, 400, response.CodeInvalidParam, err.Error())
		return
	}
	response.Success(c, resp)
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
	info, err := authService.GetProfile(userID)
	if err != nil {
		response.Fail(c, 400, response.CodeNotFound, err.Error())
		return
	}
	response.Success(c, info)
}
