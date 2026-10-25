package core

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"blog-backend/global"
)

// InitLogger 初始化全局日志实例 global.Log。
// 使用 zap + lumberjack：日志按大小/数量/天数滚动写入 config.yaml 中 log.director 目录，
// debug 级别时同时输出到控制台，方便查看。
func InitLogger() {
	cfg := global.Config.Log

	// 按配置的级别构建 zap 日志级别（默认 info）
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // 级别大写: INFO / ERROR
		EncodeTime:     zapcore.ISO8601TimeEncoder,  // 时间格式 2006-01-02T15:04:05.000+0800
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// lumberjack 负责按大小切割、按数量/天数清理旧日志
	writer := &lumberjack.Logger{
		Filename:   cfg.Director + "/app.log",
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		LocalTime:  true,
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		level,
	)

	var cores []zapcore.Core
	cores = append(cores, fileCore)

	// debug 模式下额外输出到控制台，方便本地开发
	if cfg.Level == "debug" {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	global.Log = zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(), // 日志中带上调用位置的 文件名:行号
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// 记录初始化完成，便于在日志文件头部确认
	global.Log.Info("日志初始化完成", zap.String("level", cfg.Level), zap.String("director", cfg.Director))
}
