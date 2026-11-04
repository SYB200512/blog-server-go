package core

import (
	"net/http"
	_ "net/http/pprof" // 注册 /debug/pprof/* 到 http.DefaultServeMux

	"go.uber.org/zap"

	"blog-backend/global"
)

// StartPprof 启动独立的 pprof 服务。
// 默认只监听 127.0.0.1:6060，公网访问不到；需要看 profile 时用 SSH 隧道转发到本机。
// addr 为空或 "off" 时不启动，方便在不需要排障的环境关闭。
func StartPprof(addr string) {
	if addr == "" || addr == "off" {
		return
	}
	go func() {
		global.Log.Info("pprof 已启动", zap.String("addr", addr))
		// nil 表示使用 http.DefaultServeMux，pprof 的 handler 已注册在这里
		if err := http.ListenAndServe(addr, nil); err != nil {
			global.Log.Error("pprof 服务退出", zap.Error(err))
		}
	}()
}
