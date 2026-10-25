package service

import (
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/request"
	"blog-backend/model/response"
	"blog-backend/utils"
)

// AuthService 认证业务
type AuthService struct{}

// Login 用户登录
func (s *AuthService) Login(req *request.LoginRequest) (*response.LoginResponse, error) {
	user, err := dao.GetUserByUsername(req.Username)
	if err != nil {
		// 用户名不存在或查询失败，统一返回"用户名或密码错误"，避免泄露用户是否存在
		global.Log.Warn("登录失败：用户查询错误", zap.String("username", req.Username), zap.Error(err))
		return nil, errors.New("用户名或密码错误")
	}
	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		global.Log.Warn("登录失败：密码错误", zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}
	// 签发 JWT
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		global.Log.Error("生成 token 失败", zap.String("username", req.Username), zap.Error(err))
		return nil, errors.New("生成 token 失败")
	}
	return &response.LoginResponse{
		Token: token,
		User:  toUserInfo(user),
	}, nil
}

// GetProfile 获取用户信息
func (s *AuthService) GetProfile(userID uint64) (*response.UserInfo, error) {
	user, err := dao.GetUserByID(userID)
	if err != nil {
		global.Log.Warn("获取用户信息失败", zap.Uint64("user_id", userID), zap.Error(err))
		return nil, errors.New("用户不存在")
	}
	return toUserInfo(user), nil
}

// toUserInfo 模型转响应结构体
func toUserInfo(user *database.User) *response.UserInfo {
	return &response.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Email:    user.Email,
		Role:     user.Role,
	}
}
