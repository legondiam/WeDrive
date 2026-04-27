package model

import (
	"time"

	"gorm.io/gorm"
)

// UploadSession 分块上传会话
type UploadSession struct {
	gorm.Model
	UserID          uint   `gorm:"index:idx_upload_session_lookup,priority:1"`
	ParentID        uint   `gorm:"index:idx_upload_session_lookup,priority:2"`
	HashType        string `gorm:"type:varchar(64);index:idx_upload_session_lookup,priority:3"`
	FileHash        string `gorm:"type:varchar(128);index:idx_upload_session_lookup,priority:4"`
	FileName        string `gorm:"type:varchar(255)"`
	FileSize        int64
	ChunkSize       int64
	ChunkCount      int
	HeadHash        string `gorm:"type:varchar(128)"`
	MidHash         string `gorm:"type:varchar(128)"`
	TailHash        string `gorm:"type:varchar(128)"`
	ObjectName      string `gorm:"type:varchar(512)"`
	StorageUploadID string `gorm:"type:varchar(255)"`
	Status          string `gorm:"type:varchar(32);index"`
	UserFileID      uint
	CompletedAt     *time.Time
}
