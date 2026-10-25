package main

import (
	"blog-backend/core"
	"blog-backend/flag"
	"blog-backend/initialize"
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
	core.InitConf()
	// 2. 初始化日志（zap + lumberjack 滚动落盘）
	core.InitLogger()
	// 3. 连接数据库
	initialize.InitGorm()
	// 4. 处理命令行参数（-admin 重置管理员密码，无参数则继续）
	flag.InitFlag()
	// 5. 初始化路由并启动服务器
	core.RunServer()
}
