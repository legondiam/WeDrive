package model

import "gorm.io/gorm"

// UploadSession 分块上传会话
type UploadSession struct {
	gorm.Model
	UserID          uint   `gorm:"index:idx_upload_session_lookup,priority:1"`
	ParentID        uint   `gorm:"index:idx_upload_session_lookup,priority:2"`
	FileHash        string `gorm:"type:varchar(128);index:idx_upload_session_lookup,priority:3"`
	FileName        string `gorm:"type:varchar(255)"`
	FileSize        int64
	ChunkSize       int64
	ChunkCount      int
	ObjectName      string `gorm:"type:varchar(512)"`
	StorageUploadID string `gorm:"type:varchar(255)"`
	UploadedChunks  string `gorm:"type:longtext"`
	Status          string `gorm:"type:varchar(32);index"`
}
