package models

import "time"

// Comment 评论模型
type Comment struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ArticleID uint64    `gorm:"not null;index" json:"article_id"`
	UserID    *uint64   `json:"user_id"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Email     string    `gorm:"size:100" json:"email"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ParentID  uint64    `gorm:"default:0;index" json:"parent_id"` // 0 = 顶层评论
	Status    int8      `gorm:"default:1" json:"status"`          // 0 待审核，1 已通过
	CreatedAt time.Time `json:"created_at"`

	// 关联（查询时填充）
	User *User `gorm:"foreignKey:UserID" json:"user"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}
