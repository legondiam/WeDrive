package cache

import "time"

type UserInfo struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	TotalSpace  int64      `json:"total_space"`
	UsedSpace   int64      `json:"used_space"`
	MemberLevel int8       `json:"member_level"`
	VipExpireAt *time.Time `json:"vip_expire_at"`
}

type FileListItem struct {
	ID        uint   `json:"id"`
	FileName  string `json:"file_name"`
	IsFolder  bool   `json:"is_folder"`
	FileSize  string `json:"file_size"`
	UpdatedAt string `json:"updated_at"`
	ParentID  uint   `json:"parent_id"`
}

type RecycleFileListItem struct {
	ID        uint   `json:"id"`
	FileName  string `json:"file_name"`
	IsFolder  bool   `json:"is_folder"`
	FileSize  string `json:"file_size"`
	DeletedAt string `json:"deleted_at"`
}

type UserFileMeta struct {
	FileStoreID uint   `json:"file_store_id"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	FileAddr    string `json:"file_addr"`
}

type FileIdentity struct {
	ID        uint   `json:"id"`
	HashType  string `json:"hash_type"`
	FileHash  string `json:"file_hash"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	FileAddr  string `json:"file_addr"`
	HeadHash  string `json:"head_hash"`
	MidHash   string `json:"mid_hash"`
	TailHash  string `json:"tail_hash"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ShareToken struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"user_id"`
	UserFileID uint       `json:"user_file_id"`
	ShareToken string     `json:"share_token"`
	KeyHash    string     `json:"key_hash"`
	ExpiresAt  *time.Time `json:"expires_at"`
}
