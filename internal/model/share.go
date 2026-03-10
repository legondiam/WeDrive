package model

import (
	"time"

	"gorm.io/gorm"
)

// ShareFile 分享文件
type ShareFile struct {
	gorm.Model
	UserID     uint       `gorm:"index"`
	UserFileID uint       `gorm:"index"`
	ShareToken string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	KeyHash    string     `gorm:"type:varchar(255);default:''"`
	ExpiresAt  *time.Time `gorm:"index"`
}
