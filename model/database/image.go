package database

import (
	"time"

	"blog-backend/model/appTypes"
)

// Image 图片台账：记录所有上传/生成的图片及其存储方式
type Image struct {
	ID        uint64                 `gorm:"primaryKey" json:"id"`
	Name      string                 `gorm:"size:255" json:"name"`       // 文件名（对象键）
	URL       string                 `gorm:"size:500;unique" json:"url"` // 访问地址
	Category  appTypes.ImageCategory `gorm:"default:0" json:"category"`  // 用途分类
	Storage   appTypes.Storage       `gorm:"default:0" json:"storage"`   // 存储方式
	Size      int64                  `gorm:"default:0" json:"size"`      // 文件大小（字节）
	CreatedAt time.Time              `json:"created_at"`
}

// TableName 指定表名
func (Image) TableName() string {
	return "images"
}
