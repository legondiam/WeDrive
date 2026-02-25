package model

import "gorm.io/gorm"

// FileStore 文件池
type FileStore struct {
	gorm.Model
	FileHash string `gorm:"type:varchar(128);uniqueIndex"`
	FileName string `gorm:"type:varchar(255)"`
	FileSize int64
	FileAddr string `gorm:"type:varchar(512)"`
}

// UserFile 用户文件
type UserFile struct {
	gorm.Model
	UserId      uint      `gorm:"index"`
	FileStoreID *uint     `gorm:"index"`
	FileName    string    `gorm:"type:varchar(255)"`
	ParentID    int64     `gorm:"index;default:0"`
	IsFolder    bool      `gorm:"default:false"`
	FileStore   FileStore `gorm:"foreignKey:FileStoreID"`
}
