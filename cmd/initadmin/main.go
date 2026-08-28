// initadmin 初始化管理员密码工具
// 用法: go run ./cmd/initadmin -password=你的密码
// 默认将 users 表中 username=admin 的密码重置为指定值（未指定则默认 admin123）
package main

import (
	"flag"
	"log"

	"golang.org/x/crypto/bcrypt"

	"blog-backend/config"
	"blog-backend/dao"
	"blog-backend/models"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	username := flag.String("username", "admin", "管理员用户名")
	password := flag.String("password", "admin123", "管理员密码")
	flag.Parse()

	// 加载配置并连接数据库
	cfg, err := config.InitConfig(*configPath)
	if err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	dao.Init(db)

	// 生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("生成密码哈希失败: %v", err)
	}

	// 更新或创建管理员
	var user models.User
	if err := dao.DB.Where("username = ?", *username).First(&user).Error; err != nil {
		// 不存在则创建
		user = models.User{
			Username: *username,
			Nickname: "管理员",
			Role:     1,
		}
	}
	user.Password = string(hash)
	if err := dao.DB.Save(&user).Error; err != nil {
		log.Fatalf("保存用户失败: %v", err)
	}
	log.Printf("管理员 [%s] 密码已重置，登录密码: %s", *username, *password)
}
