package core

import (
	"fmt"

	"go.uber.org/zap"

	"blog-backend/global"
	"blog-backend/initialize"
)

// RunServer 初始化路由并启动 HTTP 服务器
func RunServer() {
	r := initialize.InitRouter()
	addr := fmt.Sprintf(":%d", global.Config.Server.Port)
	global.Log.Info("博客后端启动成功，监听", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		global.Log.Fatal("服务启动失败", zap.Error(err))
	}
}
