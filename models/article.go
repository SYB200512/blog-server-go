package models

import (
	"time"
)

// Article 文章模型
type Article struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	Title      string     `gorm:"size:200;not null" json:"title"`
	Summary    string     `gorm:"size:500" json:"summary"`
	Content    string     `gorm:"type:longtext" json:"content"`
	Cover      string     `gorm:"size:255" json:"cover"`
	CategoryID *uint64    `gorm:"index" json:"category_id"`
	Status     int8       `gorm:"default:1" json:"status"` // 0 草稿，1 已发布
	Views      int        `gorm:"default:0" json:"views"`
	LikeCount  int        `gorm:"default:0" json:"like_count"`
	IsTop      int8       `gorm:"default:0" json:"is_top"` // 0 否，1 是
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `gorm:"index" json:"-"`

	// 关联（查询时填充）
	Category   *Category `gorm:"foreignKey:CategoryID" json:"category"`
	Tags       []Tag     `gorm:"many2many:article_tags;" json:"tags"`
	CommentNum int64     `gorm:"-" json:"comment_num"`
}

// TableName 指定表名
func (Article) TableName() string {
	return "articles"
}
