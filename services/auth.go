package services

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"blog-backend/common"
	"blog-backend/dao"
	"blog-backend/dto"
	"blog-backend/models"
)

// Login 用户登录
func Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := dao.GetUserByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	// 签发 JWT
	token, err := common.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.New("生成 token 失败")
	}
	return &dto.LoginResponse{
		Token: token,
		User:  ToUserInfo(user),
	}, nil
}

// GetProfile 获取用户信息
func GetProfile(userID uint64) (*dto.UserInfo, error) {
	user, err := dao.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return ToUserInfo(user), nil
}

// ToUserInfo 模型转 DTO
func ToUserInfo(user *models.User) *dto.UserInfo {
	return &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Email:    user.Email,
		Role:     user.Role,
	}
}
