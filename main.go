package main

import (
	"fmt"
	"os"

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
	// 1. 加载配置：校验不通过直接退出，绝不带着空密钥/空 DSN 启动
	warnings, err := core.InitConf()
	if err != nil {
		fmt.Fprintln(os.Stderr, "配置校验失败：")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// 2. 初始化日志（此前 global.Log 不可用，所以错误只能写 stderr）
	core.InitLogger()
	// 3. 输出降级告警与生效配置摘要
	core.LogConfig(warnings)
	// 4. 连接数据库
	initialize.InitGorm()
	// 4.1 连接缓存：探测失败仅告警，服务降级为直连数据库继续运行
	initialize.InitRedis()
	// 5. 处理命令行参数（-admin 重置管理员密码，无参数则继续）
	flag.InitFlag()
	// 6. 初始化路由并启动服务器
	core.RunServer()
}
