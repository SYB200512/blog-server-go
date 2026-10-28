package initialize

import (
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"blog-backend/global"
	"blog-backend/model/database"
)

// InitGorm 连接 MySQL 并注入全局 DB
func InitGorm() {
	db, err := gorm.Open(mysql.Open(global.Config.Database.DSN()), &gorm.Config{})
	if err != nil {
		global.Log.Fatal("初始化数据库失败", zap.Error(err))
	}
	sqlDB, err := db.DB()
	if err != nil {
		global.Log.Fatal("获取数据库连接池失败", zap.Error(err))
	}
	sqlDB.SetMaxOpenConns(global.Config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(global.Config.Database.MaxIdleConns)
	global.DB = db
	// 图片台账表：新增表由 GORM 自动建表，已有表不受影响
	if err := db.AutoMigrate(&database.Image{}); err != nil {
		global.Log.Fatal("自动迁移图片表失败", zap.Error(err))
	}
	global.Log.Info("数据库连接成功")
}
