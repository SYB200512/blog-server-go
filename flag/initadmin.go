package flag

import (
	"errors"
	"flag"
	"os"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"blog-backend/global"
	"blog-backend/model/database"
)

// InitFlag 处理命令行参数（无相关参数时直接返回，继续启动服务）
// 用法: go run . -admin -username=admin -password=xxx
func InitFlag() {
	admin := flag.Bool("admin", false, "重置管理员密码")
	username := flag.String("username", "admin", "管理员用户名")
	password := flag.String("password", "admin123", "管理员密码")
	flag.Parse()

	if *admin {
		resetAdminPassword(*username, *password)
		os.Exit(0)
	}
}

// resetAdminPassword 更新或创建管理员账号
func resetAdminPassword(username, password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		global.Log.Fatal("生成密码哈希失败", zap.Error(err))
	}

	var user database.User
	if err := global.DB.Where("username = ?", username).First(&user).Error; err != nil {
		// 真实查询失败也记录日志，避免 DB 异常时被误当成"用户不存在"
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			global.Log.Warn("查询管理员失败，将尝试新建", zap.String("username", username), zap.Error(err))
		}
		// 不存在则创建
		user = database.User{
			Username: username,
			Nickname: "管理员",
			Role:     1,
		}
	}
	user.Password = string(hash)
	if err := global.DB.Save(&user).Error; err != nil {
		global.Log.Fatal("保存用户失败", zap.Error(err))
	}
	global.Log.Info("管理员密码已重置", zap.String("username", username), zap.String("password", password))
}
