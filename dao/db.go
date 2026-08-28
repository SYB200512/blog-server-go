package dao

import "gorm.io/gorm"

// DB 全局数据库实例
var DB *gorm.DB

// Init 初始化数据库实例
func Init(db *gorm.DB) {
	DB = db
}
