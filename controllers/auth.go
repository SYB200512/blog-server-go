package controllers

import (
	"github.com/gin-gonic/gin"

	"blog-backend/common"
	"blog-backend/dto"
	"blog-backend/services"
)

// Login 用户登录
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "登录信息"
// @Success 200 {object} common.Response{data=dto.LoginResponse} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Router /api/v1/auth/login [post]
// @Id auth_login
func Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.InvalidParam(c, "用户名和密码不能为空")
		return
	}
	resp, err := services.Login(&req)
	if err != nil {
		common.Fail(c, 400, common.CodeInvalidParam, err.Error())
		return
	}
	common.Success(c, resp)
}

// GetProfile 获取当前用户信息
// @Summary 获取当前用户信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.Response{data=dto.UserInfo} "成功"
// @Failure 401 {object} common.Response "未认证"
// @Router /api/v1/auth/profile [get]
// @Id auth_profile
func GetProfile(c *gin.Context) {
	userID := c.GetUint64("userID")
	info, err := services.GetProfile(userID)
	if err != nil {
		common.Fail(c, 400, common.CodeNotFound, err.Error())
		return
	}
	common.Success(c, info)
}
