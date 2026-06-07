// Package model defines data structures.
package model

import "time"

type LinkMap struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	LongURL    string    `gorm:"column:long_url;type:text;not null"`
	ShortCode  string    `gorm:"column:short_code;type:varchar(16);uniqueIndex"`
	ClickCount int64     `gorm:"column:click_count;default:0"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (LinkMap) TableName() string {
	return "t_link_map"
}
