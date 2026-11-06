package initialize

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"blog-backend/global"
	"blog-backend/model/database"
)

// InitGorm 连接 MySQL 并注入全局 DB
func InitGorm() {
	// TranslateError：让 GORM 把 MySQL 1062 等驱动错误翻译成
	// gorm.ErrDuplicatedKey，dao 层不必硬编码错误码
	db, err := gorm.Open(mysql.Open(global.Config.Database.DSN()), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		global.Log.Fatal("初始化数据库失败", zap.Error(err))
	}
	sqlDB, err := db.DB()
	if err != nil {
		global.Log.Fatal("获取数据库连接池失败", zap.Error(err))
	}
	// P0：连接池基础参数
	sqlDB.SetMaxOpenConns(global.Config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(global.Config.Database.MaxIdleConns)

	// P0：连接最长存活时间，避免拿到被 MySQL wait_timeout 或中间件断开的连接
	if v := global.Config.Database.ConnMaxLifetime; v > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(v) * time.Minute)
	}
	// P0：连接最长空闲时间，高峰期过后主动回收
	if v := global.Config.Database.ConnMaxIdleTime; v > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(v) * time.Minute)
	}

	// P0：启动时做一次探测，不要等第一个请求才发现连不上
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		global.Log.Fatal("数据库启动探测失败", zap.Error(err))
	}

	global.DB = db

	// P0：定期打印连接池状态，方便观察 WaitCount 是否持续增长
	go logDBStats(sqlDB)

	// 图片台账表：新增表由 GORM 自动建表，已有表不受影响
	if err := db.AutoMigrate(&database.Image{}); err != nil {
		global.Log.Fatal("自动迁移图片表失败", zap.Error(err))
	}
	global.Log.Info("数据库连接成功")
}

// logDBStats 每 30 秒打印一次连接池状态。
// 如果 wait_count 持续上涨、wait_duration 变大，先查慢 SQL，再考虑扩池。
func logDBStats(sqlDB *sql.DB) {
	//ticker 每 30 秒触发一次
	ticker := time.NewTicker(30 * time.Second)

	defer ticker.Stop()
	for range ticker.C {
		s := sqlDB.Stats()
		global.Log.Info("数据库连接池状态",
			zap.Int("open", s.OpenConnections),
			zap.Int("in_use", s.InUse),
			zap.Int("idle", s.Idle),
			zap.Int("max_open", s.MaxOpenConnections),
			zap.Int64("wait_count", s.WaitCount),
			zap.Duration("wait_duration", s.WaitDuration),
		)
	}
}
