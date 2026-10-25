package global

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"blog-backend/config"
)

// 全局单例：配置、数据库连接与日志
var (
	// Config 全局配置（由 core.InitConf 加载）
	Config *config.Config
	// DB 全局数据库实例（由 initialize.InitGorm 初始化）
	DB *gorm.DB
	// Log 全局日志实例（由 core.InitLogger 初始化）
	Log *zap.Logger
)
