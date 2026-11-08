package initialize

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"blog-backend/global"
	"blog-backend/utils/cache"
)

// InitRedis 创建缓存客户端并做启动探测。
//
// 与 InitGorm 的关键差异：探测失败只告警、不退出。
// MySQL 挂了服务没有任何意义，但 Redis 挂了博客仍可降级为直连数据库；
// go-redis 内部会自动重连，配合 cache 层的熔断半开探测，Redis 恢复后缓存自动生效。
func InitRedis() {
	cfg := global.Config.Redis
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Millisecond,
		// 不重试：缓存操作必须快进快出，重试会把单次抖动放大成倍数延迟
		MaxRetries: 0,
	})
	global.Redis = client

	// 启动探测：失败仅告警。进程继续运行，接口走 DB 降级路径
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		global.Log.Warn("Redis 启动探测失败，缓存层降级运行（Redis 恢复后自动生效）",
			zap.String("addr", cfg.Addr), zap.Error(err))
	} else {
		global.Log.Info("Redis 连接成功", zap.String("addr", cfg.Addr), zap.Int("db", cfg.DB))
	}

	// 缓存命中率统计日志（每 5 分钟汇总一次，用于上线后调 TTL）
	go cache.LogHitRateLoop(5 * time.Minute)
}
