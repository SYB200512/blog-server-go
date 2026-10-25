package dao

import (
	"blog-backend/global"
	"blog-backend/model/database"
)

// GetUserByUsername 根据用户名查询用户
func GetUserByUsername(username string) (*database.User, error) {
	var user database.User
	err := global.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据 ID 查询用户
func GetUserByID(id uint64) (*database.User, error) {
	var user database.User
	err := global.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func CreateUser(user *database.User) error {
	return global.DB.Create(user).Error
}
