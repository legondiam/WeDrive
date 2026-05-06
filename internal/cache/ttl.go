package cache

import "time"

const (
	UserInfoTTL     = 30 * time.Minute
	FileListTTL     = 5 * time.Minute
	FileMetaTTL     = 15 * time.Minute
	FileIdentityTTL = time.Hour
	FileSampleTTL   = time.Hour
	ShareTokenTTL   = 30 * time.Minute
)
