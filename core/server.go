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
	"blog-backend/service"
)

// RunServer 初始化路由并启动 HTTP 服务器，收到退出信号后优雅停机
func RunServer() {
	// 初始化路由：根据配置创建路由实例
	r := initialize.InitRouter()

	// 从全局配置中获取 HTTP 服务器端口
	addr := fmt.Sprintf(":%d", global.Config.Server.Port)

	// P0：启动独立的本机 pprof 服务，不占用业务端口
	StartPprof(global.Config.Server.PprofAddr)

	// 创建 HTTP 服务器实例：后续操作通过该实例进行
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second, // 防止慢速请求头攻击
		// P0：读整个请求（含 body）的上限，必须覆盖上传预算，防止慢速请求占用资源
		ReadTimeout: time.Duration(global.Config.Server.ReadTimeout) * time.Second,
		// P0：写响应的上限，防止慢速响应占用资源
		WriteTimeout: time.Duration(global.Config.Server.WriteTimeout) * time.Second,
		// P0：keep-alive 空闲连接回收，防止连接占用资源
		IdleTimeout: time.Duration(global.Config.Server.IdleTimeout) * time.Second,
		// P0：请求头上限，防止超大 Header 占用内存
		MaxHeaderBytes: global.Config.Server.MaxHeaderBytes,
	}

	global.Log.Info("博客后端启动成功，监听", zap.String("addr", addr))

	// 监听放到后台 goroutine，主 goroutine 留给信号处理
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			global.Log.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	// 收到 Ctrl+C / k8s SIGTERM 时取消 ctx
	//NotifyContext 会返回一个 context.Context，收到信号后会取消该 context:监听的信号：ctrl+c / k8s SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// 取消信号监听：确保在程序退出前不会收到信号中断
	defer stop()

	<-ctx.Done()

	// 优雅停机：等待所有在途请求处理完，再关闭服务器
	// 停机预算：<=0 时兜底 15 秒
	timeout := 15 * time.Second
	if global.Config.Server.ShutdownTimeout > 0 {
		timeout = time.Duration(global.Config.Server.ShutdownTimeout) * time.Second
	}
	// Shutdown 会停止接收新连接，并等待在途请求处理完；超时则强制退出
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// shutdownCtx 用于等待在途请求处理完，超时则强制退出
	//调用Shutdown方法关闭服务器,停止接收新连接,等待在途请求处理完,超时则强制退出
	if err := srv.Shutdown(shutdownCtx); err != nil {
		global.Log.Error("服务关闭失败", zap.Error(err))
	}

	// 停机收尾：先把内存中尚未落库的浏览量回写 MySQL，再关缓存连接
	flushCtx, flushCancel := context.WithTimeout(context.Background(), 5*time.Second)
	service.FlushArticleViews(flushCtx)
	// 取消刷新任务：确保在程序退出前不会收到刷新信号
	flushCancel()
	if global.Redis != nil {
		_ = global.Redis.Close()
	}
	global.Log.Info("服务已退出")
}
