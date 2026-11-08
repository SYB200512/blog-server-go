package global

import (
	"github.com/redis/go-redis/v9"
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
	// Redis 全局缓存客户端（由 initialize.InitRedis 初始化）
	// 即使启动探测失败也会赋值：go-redis 自动重连，缓存层靠 fail-open + 熔断降级
	Redis *redis.Client
)
