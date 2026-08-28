package models

import "time"

// Category 分类模型
type Category struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`

	// 非表字段：文章数量（查询时填充）
	ArticleCount int64 `gorm:"-" json:"article_count"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}
