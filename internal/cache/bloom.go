package cache

const (
	BloomFileIdentity = "file_identity"
	BloomShareToken   = "share_token"
)

// BloomKey 返回布隆过滤器位图 key。
func BloomKey(name string) string {
	return "bf:" + name
}

// BloomReadyKey 返回布隆过滤器 ready 标记 key。
func BloomReadyKey(name string) string {
	return "bf:" + name + ":ready"
}

// FileIdentityBloomItem 返回文件身份布隆过滤器元素。
func FileIdentityBloomItem(hashType string, fileHash string) string {
	return hashType + ":" + fileHash
}
