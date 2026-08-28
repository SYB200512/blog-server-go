package main

import (
	"fmt"
	"log"

	"blog-backend/config"
	"blog-backend/dao"
	"blog-backend/router"
)

// @title 博客系统 API
// @version 1.0
// @description 个人博客系统的后端 HTTP API
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. 加载配置
	cfg, err := config.InitConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 2. 初始化数据库
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	dao.Init(db)
	log.Println("数据库连接成功")

	// 3. 初始化路由
	r := router.InitRouter()

	// 4. 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("博客后端启动成功，监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
