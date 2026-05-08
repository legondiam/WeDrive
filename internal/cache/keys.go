package cache

import "fmt"

const keyPrefix = "wedrive:v1"

func UserInfoKey(userID uint) string {
	return fmt.Sprintf("%s:user:info:%d", keyPrefix, userID)
}

func UserFileListKey(userID uint, parentID uint) string {
	return fmt.Sprintf("%s:user:file:list:%d:%d", keyPrefix, userID, parentID)
}

func RecycleBinListKey(userID uint) string {
	return fmt.Sprintf("%s:user:recycle:list:%d", keyPrefix, userID)
}

func DownloadFileMetaKey(userID uint, userFileID uint) string {
	return fmt.Sprintf("%s:user:file:meta:%d:%d", keyPrefix, userID, userFileID)
}

func FileIdentityKey(hashType string, fileHash string) string {
	return fmt.Sprintf("%s:file:identity:%s:%s", keyPrefix, hashType, fileHash)
}

func FileSampleKey(fileSize int64, headHash string, midHash string, tailHash string) string {
	return fmt.Sprintf("%s:file:sample:%d:%s:%s:%s", keyPrefix, fileSize, headHash, midHash, tailHash)
}

func ShareTokenKey(token string) string {
	return fmt.Sprintf("%s:share:token:%s", keyPrefix, token)
}
