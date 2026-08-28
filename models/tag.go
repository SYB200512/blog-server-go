package models

import "time"

// Tag 标签模型
type Tag struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`

	// 非表字段：文章数量（查询时填充）
	ArticleCount int64 `gorm:"-" json:"article_count"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}
