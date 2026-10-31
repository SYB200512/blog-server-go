package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"blog-backend/global"
	"blog-backend/initialize"
)

// RunServer 初始化路由并启动 HTTP 服务器，收到退出信号后优雅停机
func RunServer() {
	r := initialize.InitRouter()
	addr := fmt.Sprintf(":%d", global.Config.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second, // 防止慢速请求头攻击
	}

	global.Log.Info("博客后端启动成功，监听", zap.String("addr", addr))

	// 监听放到后台 goroutine，主 goroutine 留给信号处理
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			global.Log.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	// 收到 Ctrl+C / k8s SIGTERM 时取消 ctx
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	// 停机预算：<=0 时兜底 15 秒
	timeout := 15 * time.Second
	if global.Config.Server.ShutdownTimeout > 0 {
		timeout = time.Duration(global.Config.Server.ShutdownTimeout) * time.Second
	}
	// Shutdown 会停止接收新连接，并等待在途请求处理完；超时则强制退出
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		global.Log.Error("服务关闭失败", zap.Error(err))
	}
	global.Log.Info("服务已退出")
}
