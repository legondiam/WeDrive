package model

import (
	"gorm.io/gorm"
)

// FileStore 文件池
type FileStore struct {
	gorm.Model
	FileHash string `gorm:"type:varchar(128);uniqueIndex"`
	FileName string `gorm:"type:varchar(255)"`
	FileSize int64  `gorm:"index:idx_file_store_sample,priority:1"`
	FileAddr string `gorm:"type:varchar(512)"`
	HeadHash string `gorm:"type:varchar(128);index:idx_file_store_sample,priority:2"`
	MidHash  string `gorm:"type:varchar(128);index:idx_file_store_sample,priority:3"`
	TailHash string `gorm:"type:varchar(128);index:idx_file_store_sample,priority:4"`
}

// UserFile 用户文件
type UserFile struct {
	gorm.Model
	UserId      uint      `gorm:"index"`
	FileStoreID *uint     `gorm:"index"`
	FileName    string    `gorm:"type:varchar(255)"`
	ParentID    uint      `gorm:"index;default:0"`
	IsFolder    bool      `gorm:"default:false"`
	FileStore   FileStore `gorm:"foreignKey:FileStoreID"`
}
