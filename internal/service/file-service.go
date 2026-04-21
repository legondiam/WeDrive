package service

import (
	"WeDrive/internal/model"
	"WeDrive/internal/oss"
	"WeDrive/internal/repository"
	"WeDrive/pkg/logger"
	"WeDrive/pkg/utils/convert"
	"WeDrive/pkg/utils/hash"
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var ErrUserSpaceNotEnough = errors.New("用户空间不足")
var ErrFileNotFound = errors.New("文件不存在")
var ErrParentFolderInvalid = errors.New("父文件夹不合法")
var ErrInstantUploadUnavailable = errors.New("秒传条件失效")
var ErrUploadSessionInvalid = errors.New("上传会话无效")
var ErrChunkUploadIncomplete = errors.New("分块上传未完成")
var ErrChunkFileHashMismatch = errors.New("文件哈希不匹配")
var ErrUnsupportedHashType = errors.New("不支持的哈希规则")
var ErrUploadRequestInvalid = errors.New("上传请求无效")
var ErrUploadMethodInvalid = errors.New("上传方式不符合文件大小规则")
var ErrChunkAlreadyUploaded = errors.New("分块已上传完成")
var ErrChunkHashConflict = errors.New("分块哈希与历史记录冲突")
var ErrInstantProofRequired = errors.New("秒传需要所有权证明")
var ErrInstantProofInvalid = errors.New("秒传所有权证明无效")

const (
	uploadSessionStatusPending   = "pending"
	uploadSessionStatusCompleted = "completed"
	hashTypeFullSHA256           = "full_sha256_v1"
	merkleTypeChunkSHA256V1      = "merkle_sha256_5mb_v1"
	chunkUploadThreshold         = 16 << 20
	uploadStateExpire            = 24 * time.Hour
	partUploadURLExpire          = 15 * time.Minute
	instantPrepareExpire         = 5 * time.Minute
	instantProofTokenExpire      = 2 * time.Minute
	instantProofChallengeCount   = 2
)

type FileService struct {
	fileRepo    *repository.FileRepo
	uploadCache *repository.UploadCacheRepo
	userRepo    *repository.UserRepo
	storage     *oss.Storage
	db          *gorm.DB
}

type RecycleFileResp struct {
	ID        uint   `json:"id"`
	FileName  string `json:"file_name"`
	IsFolder  bool   `json:"is_folder"`
	FileSize  string `json:"file_size"`
	DeletedAt string `json:"deleted_at"`
}

type FileResp struct {
	ID        uint   `json:"id"`
	FileName  string `json:"file_name"`
	IsFolder  bool   `json:"is_folder"`
	FileSize  string `json:"file_size"`
	UpdatedAt string `json:"updated_at"`
	ParentID  uint   `json:"parent_id"`
}

type DownloadFileResp struct {
	URL      string `json:"url"`
	FileName string `json:"file_name"`
}

type InstantUploadResp struct {
	ID uint `json:"id"`
}

type QuickCheckReq struct {
	FileSize int64
	HeadHash string
	MidHash  string
	TailHash string
}

type ChunkUploadInitReq struct {
	HashType   string
	FileHash   string
	FileName   string
	FileSize   int64
	ParentID   uint
	ChunkSize  int64
	ChunkCount int
	HeadHash   string
	MidHash    string
	TailHash   string
}

type ChunkUploadInitResp struct {
	UploadID       uint  `json:"upload_id,omitempty"`
	UploadedChunks []int `json:"uploaded_chunks,omitempty"`
}

type SignPartReq struct {
	UploadID   uint
	PartNumber int
	ChunkHash  string
}

type SignedPartResp struct {
	PartNumber int               `json:"part_number"`
	UploadURL  string            `json:"upload_url"`
	Headers    map[string]string `json:"headers,omitempty"`
}

type ReportUploadedPartReq struct {
	UploadID   uint
	PartNumber int
	ETag       string
}

type PrepareInstantUploadReq struct {
	HashType   string
	FileHash   string
	FileName   string
	FileSize   int64
	ParentID   uint
	MerkleType string
	MerkleRoot string
}

type MerkleChallenge struct {
	PartNumber int `json:"part_number"`
}

type PrepareInstantUploadResp struct {
	PrepareID     string            `json:"prepare_id"`
	ProofRequired bool              `json:"proof_required"`
	MerkleType    string            `json:"merkle_type"`
	Challenges    []MerkleChallenge `json:"challenges"`
	Nonce         string            `json:"nonce"`
	ExpiresInSec  int               `json:"expires_in_sec"`
}

type MerkleProofItem struct {
	PartNumber    int
	LeafHash      string
	ChallengeHash string
	Proof         []hash.MerkleProofNode
}

type VerifyInstantUploadProofReq struct {
	PrepareID string
	Proofs    []MerkleProofItem
}

type VerifyInstantUploadProofResp struct {
	ProofToken   string `json:"proof_token"`
	ExpiresInSec int    `json:"expires_in_sec"`
}

type instantPrepareState struct {
	UserID        uint   `json:"user_id"`
	FileStoreID   uint   `json:"file_store_id"`
	ParentID      uint   `json:"parent_id"`
	FileName      string `json:"file_name"`
	FileHash      string `json:"file_hash"`
	HashType      string `json:"hash_type"`
	MerkleType    string `json:"merkle_type"`
	MerkleRoot    string `json:"merkle_root"`
	Nonce         string `json:"nonce"`
	ChallengePart []int  `json:"challenge_part"`
}

type instantProofTokenState struct {
	UserID      uint   `json:"user_id"`
	FileStoreID uint   `json:"file_store_id"`
	ParentID    uint   `json:"parent_id"`
	FileName    string `json:"file_name"`
	FileHash    string `json:"file_hash"`
	HashType    string `json:"hash_type"`
}

func NewFileService(fileRepo *repository.FileRepo, uploadCache *repository.UploadCacheRepo, userRepo *repository.UserRepo, storage *oss.Storage, db *gorm.DB) *FileService {
	return &FileService{fileRepo: fileRepo, uploadCache: uploadCache, storage: storage, db: db, userRepo: userRepo}
}

// isDuplicateKeyError 判断是否为重复键错误
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate entry") || strings.Contains(message, "unique constraint")
}

// validateHashType 验证哈希类型
func validateHashType(hashType string) error {
	if hashType != hashTypeFullSHA256 {
		return ErrUnsupportedHashType
	}
	return nil
}

// checksumBase64FromHex 将十六进制SHA-256转换为base64
func checksumBase64FromHex(chunkHash string) (string, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(chunkHash))
	if err != nil || len(raw) == 0 {
		return "", ErrUploadRequestInvalid
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

func isMultipartUploadNotFound(err error) bool {
	if err == nil {
		return false
	}
	resp := minio.ToErrorResponse(errors.Cause(err))
	return resp.Code == minio.NoSuchUpload
}

// collectPartHashes 收集分块哈希
func collectPartHashes(parts []repository.UploadPartState, chunkCount int) ([]hash.ChunkHash, error) {
	if len(parts) != chunkCount {
		return nil, ErrChunkUploadIncomplete
	}
	chunkHashes := make([]hash.ChunkHash, 0, len(parts))
	for index, part := range parts {
		expected := index + 1
		if part.PartNumber != expected || part.Value == "" {
			return nil, ErrChunkUploadIncomplete
		}
		chunkHashes = append(chunkHashes, hash.ChunkHash{
			PartNumber: part.PartNumber,
			Hash:       part.Value,
		})
	}
	return chunkHashes, nil
}

// buildMerkleMetaFromChunkHashes 计算默克尔树root和叶子数
func buildMerkleMetaFromChunkHashes(parts []hash.ChunkHash) (string, int, error) {
	root, leafCount, err := hash.MerkleRootFromChunkHashes(parts)
	if err != nil {
		return "", 0, errors.WithMessage(err, "计算Root失败")
	}
	return root, leafCount, nil
}

// buildMerkleMetaFromFileHeader 计算文件的默克尔树
func (s *FileService) buildMerkleMetaFromFileHeader(fileHeader *multipart.FileHeader) (string, int, error) {
	chunkHashes, err := hash.ChunkHashesFromFileHeader(fileHeader, hash.ChunkIdentitySize)
	if err != nil {
		return "", 0, errors.WithMessage(err, "计算分块哈希失败")
	}
	return buildMerkleMetaFromChunkHashes(chunkHashes)
}

// ensureMerkleMetadata 保存Merkle元数据
func (s *FileService) ensureMerkleMetadata(ctx context.Context, tx *gorm.DB, fileStore *model.FileStore, merkleRoot string, leafCount int) error {
	if fileStore == nil || merkleRoot == "" || leafCount <= 0 {
		return nil
	}
	if fileStore.MerkleRoot == merkleRoot && fileStore.MerkleType == merkleTypeChunkSHA256V1 && fileStore.MerkleChunkSize == hash.ChunkIdentitySize && fileStore.MerkleLeafCount == leafCount {
		return nil
	}
	fileStore.MerkleType = merkleTypeChunkSHA256V1
	fileStore.MerkleRoot = merkleRoot
	fileStore.MerkleChunkSize = hash.ChunkIdentitySize
	fileStore.MerkleLeafCount = leafCount
	if err := s.fileRepo.SaveFileStore(ctx, fileStore, tx); err != nil {
		return errors.WithMessage(err, "保存Merkle元数据失败")
	}
	return nil
}

func generateRandomToken() (string, error) {
	return uuid.NewString(), nil
}

// randomPartNumbers 随机挑选分块编号
func randomPartNumbers(total int, count int) ([]int, error) {
	if total <= 0 {
		return nil, ErrInstantProofInvalid
	}
	if count > total {
		count = total
	}
	selected := make(map[int]struct{}, count)
	results := make([]int, 0, count)
	for len(results) < count {
		n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(total)))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		partNumber := int(n.Int64()) + 1
		if _, ok := selected[partNumber]; ok {
			continue
		}
		selected[partNumber] = struct{}{}
		results = append(results, partNumber)
	}
	return results, nil
}

func objectRangeForPart(fileSize int64, chunkSize int64, partNumber int) (int64, int64, error) {
	if fileSize <= 0 || chunkSize <= 0 || partNumber <= 0 {
		return 0, 0, ErrInstantProofInvalid
	}
	offset := int64(partNumber-1) * chunkSize
	if offset >= fileSize {
		return 0, 0, ErrInstantProofInvalid
	}
	length := chunkSize
	if offset+length > fileSize {
		length = fileSize - offset
	}
	return offset, length, nil
}

// checkParentFolder 检查父文件夹是否合法
func (s *FileService) checkParentFolder(ctx context.Context, userID uint, parentID uint) error {
	// 根目录合法
	if parentID == 0 {
		return nil
	}
	// 查文件夹是否存在
	folder, err := s.fileRepo.GetFileByID(ctx, parentID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WithMessage(ErrParentFolderInvalid, "父文件夹不存在")
		}
		return errors.WithMessage(err, "检查父文件夹失败")
	}
	// 检查归属权
	if folder.UserId != userID {
		return errors.WithMessage(ErrParentFolderInvalid, "无权访问该目录")
	}
	// 检查是否为文件夹
	if !folder.IsFolder {
		return errors.WithMessage(ErrParentFolderInvalid, "目标不是文件夹")
	}
	return nil
}

// collectSubtreeIDs 收集子树ID
func (s *FileService) collectSubtreeIDs(ctx context.Context, userID uint, parentID uint, ids *[]uint, visited map[uint]struct{}) error {
	// 获取子文件列表
	children, err := s.fileRepo.GetUserFileByParentID(ctx, userID, parentID)
	if err != nil {
		return errors.WithMessage(err, "获取子文件列表失败")
	}
	// 遍历子文件
	for _, child := range children {
		// 如果子文件已访问，跳过
		if _, ok := visited[child.ID]; ok {
			continue
		}
		// 标记子文件已访问
		visited[child.ID] = struct{}{}

		*ids = append(*ids, child.ID)
		// 如果子文件是文件夹，递归收集子树ID
		if child.IsFolder {
			err := s.collectSubtreeIDs(ctx, userID, child.ID, ids, visited)
			if err != nil {
				return errors.WithMessage(err, "收集子树ID失败")
			}
		}
	}
	return nil
}

// checkUserMember 检查用户会员状态
func (s *FileService) checkUserMember(user *model.User) string {
	if user == nil {
		return "free"
	}
	if user.MemberLevel == 0 {
		return "free"
	}
	if user.VipExpireAt != nil && user.VipExpireAt.Before(time.Now()) {
		return "free"
	}
	return "vip"
}

// createInstantUploadRecord 创建秒传记录
func (s *FileService) createInstantUploadRecord(ctx context.Context, tx *gorm.DB, userID uint, parentID uint, fileName string, fileSize int64, fileStore *model.FileStore) (uint, error) {
	newUserFile := &model.UserFile{
		UserId:      userID,
		FileName:    fileName,
		FileStoreID: &fileStore.ID,
		ParentID:    parentID,
	}
	if err := s.fileRepo.CreateUserFile(ctx, newUserFile, tx); err != nil {
		return 0, errors.WithMessage(err, "秒传文件存储失败")
	}
	if err := s.userRepo.UpdateUserSpace(ctx, userID, fileSize, tx); err != nil {
		return 0, errors.WithMessage(err, "更新用户空间失败")
	}
	return newUserFile.ID, nil
}

// InitChunkUpload 初始化分块上传
func (s *FileService) InitChunkUpload(ctx context.Context, userID uint, req ChunkUploadInitReq) (ChunkUploadInitResp, error) {
	if err := validateHashType(req.HashType); err != nil {
		return ChunkUploadInitResp{}, err
	}
	if req.FileHash == "" || req.FileName == "" || req.FileSize <= 0 || req.ChunkSize != hash.ChunkIdentitySize || req.ChunkCount <= 0 {
		return ChunkUploadInitResp{}, ErrUploadRequestInvalid
	}
	if req.FileSize < chunkUploadThreshold {
		return ChunkUploadInitResp{}, ErrUploadMethodInvalid
	}
	if err := s.checkParentFolder(ctx, userID, req.ParentID); err != nil {
		return ChunkUploadInitResp{}, errors.WithMessage(err, "父文件夹不合法")
	}
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return ChunkUploadInitResp{}, errors.WithMessage(err, "获取用户信息失败")
	}
	if user.UsedSpace+req.FileSize > user.TotalSpace {
		return ChunkUploadInitResp{}, ErrUserSpaceNotEnough
	}

	//查询是否已存在上传会话
	session, err := s.fileRepo.GetPendingUploadSession(ctx, userID, req.ParentID, req.HashType, req.FileHash)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return ChunkUploadInitResp{}, errors.WithMessage(err, "查询上传会话失败")
	}
	//已存在未完成的上传会话
	if err == nil {
		//该上传会话可用
		if session.FileSize == req.FileSize && session.ChunkSize == req.ChunkSize && session.ChunkCount == req.ChunkCount && session.ObjectName != "" && session.StorageUploadID != "" {
			if touchErr := s.fileRepo.TouchUploadSession(ctx, session.ID); touchErr != nil {
				return ChunkUploadInitResp{}, errors.WithMessage(touchErr, "刷新上传会话活跃时间失败")
			}
			uploadedParts, cacheErr := s.uploadCache.ListUploadedParts(ctx, session.ID)
			if cacheErr != nil {
				return ChunkUploadInitResp{}, errors.WithMessage(cacheErr, "查询已上传分块失败")
			}
			//返回已上传分块
			return ChunkUploadInitResp{
				UploadID:       session.ID,
				UploadedChunks: uploadedParts,
			}, nil
		}
	}

	//不存在可用的上传会话

	//生成oss对象名
	objectName := fmt.Sprintf("multipart/%s%s", uuid.NewString(), path.Ext(req.FileName))
	uploadID, err := s.storage.NewMultipartUpload(ctx, objectName)
	if err != nil {
		return ChunkUploadInitResp{}, errors.WithMessage(err, "初始化对象分块上传失败")
	}
	//创建新的上传会话
	newSession := &model.UploadSession{
		UserID:          userID,
		ParentID:        req.ParentID,
		HashType:        req.HashType,
		FileHash:        req.FileHash,
		FileName:        req.FileName,
		FileSize:        req.FileSize,
		ChunkSize:       req.ChunkSize,
		ChunkCount:      req.ChunkCount,
		HeadHash:        req.HeadHash,
		MidHash:         req.MidHash,
		TailHash:        req.TailHash,
		ObjectName:      objectName,
		StorageUploadID: uploadID,
		Status:          uploadSessionStatusPending,
	}
	//创建上传会话记录
	if err := s.fileRepo.CreateUploadSession(ctx, newSession); err != nil {
		//出错时清理上传会话
		_ = s.storage.AbortMultipartUpload(ctx, objectName, uploadID)
		return ChunkUploadInitResp{}, errors.WithMessage(err, "创建上传会话失败")
	}

	return ChunkUploadInitResp{
		UploadID:       newSession.ID,
		UploadedChunks: []int{},
	}, nil
}

// SignPartUpload 为分块上传签名
func (s *FileService) SignPartUpload(ctx context.Context, userID uint, req SignPartReq) (SignedPartResp, error) {
	if req.UploadID == 0 || req.PartNumber <= 0 || strings.TrimSpace(req.ChunkHash) == "" {
		return SignedPartResp{}, ErrUploadRequestInvalid
	}
	checksumSHA256Base64, err := checksumBase64FromHex(req.ChunkHash)
	if err != nil {
		return SignedPartResp{}, err
	}
	//查询上传会话
	session, err := s.fileRepo.GetUploadSessionByID(ctx, req.UploadID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SignedPartResp{}, ErrUploadSessionInvalid
		}
		return SignedPartResp{}, errors.WithMessage(err, "查询上传会话失败")
	}
	//会话失效或分块序号不合法
	if session.Status != uploadSessionStatusPending || req.PartNumber > session.ChunkCount {
		return SignedPartResp{}, ErrUploadSessionInvalid
	}
	///检查分块是否已上传
	uploadedETag, etagExists, err := s.uploadCache.GetPartETag(ctx, session.ID, req.PartNumber)
	if err != nil {
		return SignedPartResp{}, errors.WithMessage(err, "读取分块ETag失败")
	}
	if etagExists && strings.TrimSpace(uploadedETag) != "" {
		return SignedPartResp{}, ErrChunkAlreadyUploaded
	}
	//检查分块是否变化
	existingHash, hashExists, err := s.uploadCache.GetPartHash(ctx, session.ID, req.PartNumber)
	if err != nil {
		return SignedPartResp{}, errors.WithMessage(err, "读取分块哈希失败")
	}
	if hashExists && existingHash != req.ChunkHash {
		return SignedPartResp{}, ErrChunkHashConflict
	}
	//保存分块哈希
	if err := s.uploadCache.SetPartHash(ctx, session.ID, req.PartNumber, req.ChunkHash, uploadStateExpire); err != nil {
		return SignedPartResp{}, errors.WithMessage(err, "保存分块哈希失败")
	}
	//生成上传URL
	if err := s.fileRepo.TouchUploadSession(ctx, session.ID); err != nil {
		return SignedPartResp{}, errors.WithMessage(err, "刷新上传会话活跃时间失败")
	}
	uploadURL, headers, err := s.storage.PresignUploadPart(ctx, session.ObjectName, session.StorageUploadID, req.PartNumber, checksumSHA256Base64, partUploadURLExpire)
	if err != nil {
		return SignedPartResp{}, errors.WithMessage(err, "生成分块上传地址失败")
	}
	return SignedPartResp{
		PartNumber: req.PartNumber,
		UploadURL:  uploadURL,
		Headers:    headers,
	}, nil
}

// ReportUploadedPart 回报已上传分块
func (s *FileService) ReportUploadedPart(ctx context.Context, userID uint, req ReportUploadedPartReq) error {
	if req.UploadID == 0 || req.PartNumber <= 0 || strings.TrimSpace(req.ETag) == "" {
		return ErrUploadRequestInvalid
	}
	//查询上传会话
	session, err := s.fileRepo.GetUploadSessionByID(ctx, req.UploadID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUploadSessionInvalid
		}
		return errors.WithMessage(err, "查询上传会话失败")
	}
	//会话失效或分块序号不合法
	if session.Status != uploadSessionStatusPending || req.PartNumber > session.ChunkCount {
		return ErrUploadSessionInvalid
	}
	//保存分块ETag
	if err := s.uploadCache.SetPartETag(ctx, session.ID, req.PartNumber, req.ETag, uploadStateExpire); err != nil {
		return errors.WithMessage(err, "保存分块ETag失败")
	}
	return nil
}

// CompleteChunkUpload 完成分块上传
func (s *FileService) CompleteChunkUpload(ctx context.Context, userID uint, sessionID uint) (uint, error) {
	//查询上传会话
	session, err := s.fileRepo.GetUploadSessionByID(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUploadSessionInvalid
		}
		return 0, errors.WithMessage(err, "查询上传会话失败")
	}
	if session.Status != uploadSessionStatusPending {
		return 0, ErrUploadSessionInvalid
	}
	//校验哈希类型
	if err := validateHashType(session.HashType); err != nil {
		return 0, err
	}

	if err := s.checkParentFolder(ctx, userID, session.ParentID); err != nil {
		return 0, errors.WithMessage(err, "父文件夹不合法")
	}
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return 0, errors.WithMessage(err, "获取用户信息失败")
	}
	if user.UsedSpace+session.FileSize > user.TotalSpace {
		return 0, ErrUserSpaceNotEnough
	}
	//读取分块ETag
	partEtags, err := s.uploadCache.ListPartETags(ctx, session.ID)
	if err != nil {
		return 0, errors.WithMessage(err, "读取分块ETag失败")
	}
	if len(partEtags) != session.ChunkCount {
		return 0, ErrChunkUploadIncomplete
	}
	partHashes, err := s.uploadCache.ListPartHashes(ctx, session.ID)
	if err != nil {
		return 0, errors.WithMessage(err, "读取分块哈希失败")
	}
	chunkHashes, err := collectPartHashes(partHashes, session.ChunkCount)
	if err != nil {
		return 0, err
	}
	merkleRoot, merkleLeafCount, err := buildMerkleMetaFromChunkHashes(chunkHashes)
	if err != nil {
		return 0, err
	}
	//组装分块合并请求参数
	completeParts := make([]oss.CompletePart, 0, len(partEtags))
	for index, part := range partEtags {
		expected := index + 1
		if part.PartNumber != expected || part.Value == "" {
			return 0, ErrChunkUploadIncomplete
		}
		completeParts = append(completeParts, oss.CompletePart{
			PartNumber: part.PartNumber,
			ETag:       part.Value,
		})
	}
	if err := s.storage.CompleteMultipartUpload(ctx, session.ObjectName, session.StorageUploadID, completeParts); err != nil {
		return 0, errors.WithMessage(err, "完成对象分块上传失败")
	}

	shouldCleanMinio := true
	defer func() {
		if shouldCleanMinio {
			_ = s.storage.DeleteFile(ctx, session.ObjectName)
		}
		_ = s.uploadCache.DeleteUploadState(ctx, session.ID)
	}()

	var uploadedID uint
	err = s.db.Transaction(func(tx *gorm.DB) error {
		lockedSession, lockErr := s.fileRepo.GetUploadSessionByIDForUpdate(ctx, sessionID, userID, tx)
		if lockErr != nil {
			if errors.Is(lockErr, gorm.ErrRecordNotFound) {
				return ErrUploadSessionInvalid
			}
			return errors.WithMessage(lockErr, "锁定上传会话失败")
		}
		if lockedSession.Status != uploadSessionStatusPending {
			return ErrUploadSessionInvalid
		}

		//确保去重
		fileStore, findErr := s.fileRepo.GetFileByIdentityForUpdate(ctx, session.HashType, session.FileHash, tx)
		switch {
		case findErr == nil:
			if err := s.ensureMerkleMetadata(ctx, tx, fileStore, merkleRoot, merkleLeafCount); err != nil {
				return err
			}
			//去重命中，创建秒传记录
			uploadedID, err = s.createInstantUploadRecord(ctx, tx, userID, session.ParentID, session.FileName, fileStore.FileSize, fileStore)
			if err != nil {
				return err
			}
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			//去重未命中，创建文件元数据
			newFileStore := &model.FileStore{
				HashType:        session.HashType,
				FileHash:        session.FileHash,
				FileName:        session.FileName,
				FileSize:        session.FileSize,
				FileAddr:        session.ObjectName,
				HeadHash:        session.HeadHash,
				MidHash:         session.MidHash,
				TailHash:        session.TailHash,
				MerkleType:      merkleTypeChunkSHA256V1,
				MerkleRoot:      merkleRoot,
				MerkleChunkSize: hash.ChunkIdentitySize,
				MerkleLeafCount: merkleLeafCount,
			}
			if createErr := s.fileRepo.CreateFileStore(ctx, newFileStore, tx); createErr != nil {
				if !isDuplicateKeyError(createErr) {
					return errors.WithMessage(createErr, "文件元数据存储失败")
				}
				//再次查询以去重
				fileStore, createErr = s.fileRepo.GetFileByIdentityForUpdate(ctx, session.HashType, session.FileHash, tx)
				if createErr != nil {
					return errors.WithMessage(createErr, "查询并发写入的文件失败")
				}
				if err := s.ensureMerkleMetadata(ctx, tx, fileStore, merkleRoot, merkleLeafCount); err != nil {
					return err
				}
				uploadedID, err = s.createInstantUploadRecord(ctx, tx, userID, session.ParentID, session.FileName, fileStore.FileSize, fileStore)
				if err != nil {
					return err
				}
				break
			}
			//创建用户文件记录
			newUserFile := &model.UserFile{
				UserId:      userID,
				FileStoreID: &newFileStore.ID,
				FileName:    session.FileName,
				ParentID:    session.ParentID,
			}
			if err := s.fileRepo.CreateUserFile(ctx, newUserFile, tx); err != nil {
				return errors.WithMessage(err, "用户文件数据存储失败")
			}
			uploadedID = newUserFile.ID
			if err := s.userRepo.UpdateUserSpace(ctx, userID, session.FileSize, tx); err != nil {
				return errors.WithMessage(err, "更新用户空间失败")
			}
		default:
			return errors.WithMessage(findErr, "查询文件失败")
		}

		lockedSession.Status = uploadSessionStatusCompleted
		if err := s.fileRepo.SaveUploadSession(ctx, lockedSession, tx); err != nil {
			return errors.WithMessage(err, "更新上传会话状态失败")
		}
		if err := s.fileRepo.DeleteUploadSession(ctx, lockedSession.ID, tx); err != nil {
			return errors.WithMessage(err, "清理上传会话失败")
		}
		return nil
	})
	if err != nil {
		_ = s.fileRepo.DeleteUploadSession(ctx, session.ID)
		return 0, errors.WithMessage(err, "完成分块上传失败")
	}

	shouldCleanMinio = false
	logger.S.Infof("直传分块上传完成, uploadID: %d, hashType: %s, fileHash: %s, userID: %d", sessionID, session.HashType, session.FileHash, userID)
	return uploadedID, nil
}

// QuickCheck 抽样哈希快速判断是否可能秒传
func (s *FileService) QuickCheck(ctx context.Context, userID uint, req QuickCheckReq) (bool, error) {
	if req.FileSize < 0 || req.HeadHash == "" || req.MidHash == "" || req.TailHash == "" {
		return false, nil
	}
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return false, errors.WithMessage(err, "获取用户信息失败")
	}
	if user.UsedSpace+req.FileSize > user.TotalSpace {
		return false, ErrUserSpaceNotEnough
	}
	exists, err := s.fileRepo.GetFileBySample(ctx, req.FileSize, req.HeadHash, req.MidHash, req.TailHash)
	if err != nil {
		return false, errors.WithMessage(err, "抽样哈希查询失败")
	}
	return exists, nil
}

// PrepareInstantUpload 为秒传准备默克尔树所有权挑战
func (s *FileService) PrepareInstantUpload(ctx context.Context, userID uint, req PrepareInstantUploadReq) (PrepareInstantUploadResp, error) {
	if err := validateHashType(req.HashType); err != nil {
		return PrepareInstantUploadResp{}, err
	}
	if err := s.checkParentFolder(ctx, userID, req.ParentID); err != nil {
		return PrepareInstantUploadResp{}, errors.WithMessage(err, "父文件夹不合法")
	}
	if strings.TrimSpace(req.FileHash) == "" || strings.TrimSpace(req.FileName) == "" || strings.TrimSpace(req.MerkleRoot) == "" || strings.TrimSpace(req.MerkleType) == "" {
		return PrepareInstantUploadResp{}, ErrUploadRequestInvalid
	}
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return PrepareInstantUploadResp{}, errors.WithMessage(err, "获取用户信息失败")
	}
	fileStore, err := s.fileRepo.GetFileByIdentity(ctx, req.HashType, req.FileHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PrepareInstantUploadResp{}, ErrInstantUploadUnavailable
		}
		return PrepareInstantUploadResp{}, errors.WithMessage(err, "查询文件失败")
	}
	if req.FileSize > 0 && fileStore.FileSize != req.FileSize {
		return PrepareInstantUploadResp{}, ErrInstantUploadUnavailable
	}
	if fileStore.MerkleRoot == "" || fileStore.MerkleType == "" || fileStore.MerkleChunkSize <= 0 || fileStore.MerkleLeafCount <= 0 {
		return PrepareInstantUploadResp{}, ErrInstantUploadUnavailable
	}
	if fileStore.MerkleType != req.MerkleType || fileStore.MerkleRoot != req.MerkleRoot {
		return PrepareInstantUploadResp{}, ErrInstantUploadUnavailable
	}
	if user.UsedSpace+fileStore.FileSize > user.TotalSpace {
		return PrepareInstantUploadResp{}, ErrUserSpaceNotEnough
	}

	challengeParts, err := randomPartNumbers(fileStore.MerkleLeafCount, instantProofChallengeCount)
	if err != nil {
		return PrepareInstantUploadResp{}, errors.WithMessage(err, "生成秒传挑战失败")
	}
	prepareID, err := generateRandomToken()
	if err != nil {
		return PrepareInstantUploadResp{}, errors.WithMessage(err, "生成秒传挑战失败")
	}
	nonce, err := generateRandomToken()
	if err != nil {
		return PrepareInstantUploadResp{}, errors.WithMessage(err, "生成秒传挑战失败")
	}
	state := instantPrepareState{
		UserID:        userID,
		FileStoreID:   fileStore.ID,
		ParentID:      req.ParentID,
		FileName:      req.FileName,
		FileHash:      req.FileHash,
		HashType:      req.HashType,
		MerkleType:    fileStore.MerkleType,
		MerkleRoot:    fileStore.MerkleRoot,
		Nonce:         nonce,
		ChallengePart: challengeParts,
	}
	if err := s.uploadCache.SetInstantPrepare(ctx, prepareID, state, instantPrepareExpire); err != nil {
		return PrepareInstantUploadResp{}, errors.WithMessage(err, "保存秒传挑战失败")
	}
	resp := PrepareInstantUploadResp{
		PrepareID:     prepareID,
		ProofRequired: true,
		MerkleType:    fileStore.MerkleType,
		Nonce:         nonce,
		ExpiresInSec:  int(instantPrepareExpire / time.Second),
		Challenges:    make([]MerkleChallenge, 0, len(challengeParts)),
	}
	for _, partNumber := range challengeParts {
		resp.Challenges = append(resp.Challenges, MerkleChallenge{PartNumber: partNumber})
	}
	return resp, nil
}

// VerifyInstantUploadProof 校验挑战分块的默克尔证明并签发一次性凭证
func (s *FileService) VerifyInstantUploadProof(ctx context.Context, userID uint, req VerifyInstantUploadProofReq) (VerifyInstantUploadProofResp, error) {
	if strings.TrimSpace(req.PrepareID) == "" || len(req.Proofs) == 0 {
		return VerifyInstantUploadProofResp{}, ErrUploadRequestInvalid
	}
	var state instantPrepareState
	found, err := s.uploadCache.GetInstantPrepare(ctx, req.PrepareID, &state)
	if err != nil {
		return VerifyInstantUploadProofResp{}, errors.WithMessage(err, "读取秒传挑战失败")
	}
	if !found || state.UserID != userID {
		return VerifyInstantUploadProofResp{}, ErrInstantProofInvalid
	}
	fileStore, err := s.fileRepo.GetFileByIdentity(ctx, state.HashType, state.FileHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VerifyInstantUploadProofResp{}, ErrInstantUploadUnavailable
		}
		return VerifyInstantUploadProofResp{}, errors.WithMessage(err, "查询文件失败")
	}
	if fileStore.ID != state.FileStoreID || fileStore.MerkleRoot != state.MerkleRoot || fileStore.MerkleType != state.MerkleType {
		return VerifyInstantUploadProofResp{}, ErrInstantProofInvalid
	}
	if len(req.Proofs) != len(state.ChallengePart) {
		return VerifyInstantUploadProofResp{}, ErrInstantProofInvalid
	}
	expectedParts := make(map[int]struct{}, len(state.ChallengePart))
	for _, partNumber := range state.ChallengePart {
		expectedParts[partNumber] = struct{}{}
	}
	for _, proof := range req.Proofs {
		if _, ok := expectedParts[proof.PartNumber]; !ok {
			return VerifyInstantUploadProofResp{}, ErrInstantProofInvalid
		}
		offset, length, rangeErr := objectRangeForPart(fileStore.FileSize, fileStore.MerkleChunkSize, proof.PartNumber)
		if rangeErr != nil {
			return VerifyInstantUploadProofResp{}, rangeErr
		}
		chunkBytes, readErr := s.storage.ReadFileRange(ctx, fileStore.FileAddr, offset, length)
		if readErr != nil {
			return VerifyInstantUploadProofResp{}, errors.WithMessage(readErr, "读取挑战分块失败")
		}
		expectedLeafHash := hash.HashBytesHex(chunkBytes)
		if proof.LeafHash != expectedLeafHash {
			return VerifyInstantUploadProofResp{}, ErrInstantProofInvalid
		}
		expectedChallengeHash := hash.ChallengeChunkHash(state.Nonce, chunkBytes)
		if proof.ChallengeHash != expectedChallengeHash {
			return VerifyInstantUploadProofResp{}, ErrInstantProofInvalid
		}
		ok, verifyErr := hash.VerifyMerkleProof(proof.LeafHash, fileStore.MerkleRoot, proof.Proof)
		if verifyErr != nil {
			return VerifyInstantUploadProofResp{}, errors.WithMessage(verifyErr, "验证 Merkle Proof 失败")
		}
		if !ok {
			return VerifyInstantUploadProofResp{}, ErrInstantProofInvalid
		}
		delete(expectedParts, proof.PartNumber)
	}
	if len(expectedParts) != 0 {
		return VerifyInstantUploadProofResp{}, ErrInstantProofInvalid
	}

	proofToken, err := generateRandomToken()
	if err != nil {
		return VerifyInstantUploadProofResp{}, errors.WithMessage(err, "生成秒传凭证失败")
	}
	tokenState := instantProofTokenState{
		UserID:      userID,
		FileStoreID: fileStore.ID,
		ParentID:    state.ParentID,
		FileName:    state.FileName,
		FileHash:    state.FileHash,
		HashType:    state.HashType,
	}
	if err := s.uploadCache.SetInstantProofToken(ctx, proofToken, tokenState, instantProofTokenExpire); err != nil {
		return VerifyInstantUploadProofResp{}, errors.WithMessage(err, "保存秒传凭证失败")
	}
	_ = s.uploadCache.DeleteInstantPrepare(ctx, req.PrepareID)
	return VerifyInstantUploadProofResp{
		ProofToken:   proofToken,
		ExpiresInSec: int(instantProofTokenExpire / time.Second),
	}, nil
}

// InstantUpload 命中秒传后的确认落库
func (s *FileService) InstantUpload(ctx context.Context, userID uint, parentID uint, hashType string, fileName string, fileHash string, fileSize int64, prepareID string, proofToken string) (uint, error) {
	if err := validateHashType(hashType); err != nil {
		return 0, err
	}
	if err := s.checkParentFolder(ctx, userID, parentID); err != nil {
		return 0, errors.WithMessage(err, "父文件夹不合法")
	}
	if strings.TrimSpace(prepareID) == "" || strings.TrimSpace(proofToken) == "" {
		return 0, ErrInstantProofRequired
	}

	var tokenState instantProofTokenState
	found, err := s.uploadCache.GetInstantProofToken(ctx, proofToken, &tokenState)
	if err != nil {
		return 0, errors.WithMessage(err, "读取秒传凭证失败")
	}
	if !found {
		return 0, ErrInstantProofInvalid
	}

	var prepareState instantPrepareState
	prepareFound, err := s.uploadCache.GetInstantPrepare(ctx, prepareID, &prepareState)
	if err != nil {
		return 0, errors.WithMessage(err, "读取秒传挑战失败")
	}
	if prepareFound {
		return 0, ErrInstantProofInvalid
	}
	if tokenState.UserID != userID || tokenState.ParentID != parentID || tokenState.HashType != hashType || tokenState.FileHash != fileHash {
		return 0, ErrInstantProofInvalid
	}

	var uploadedID uint
	err = s.db.Transaction(func(tx *gorm.DB) error {
		fileStore, loadErr := s.fileRepo.GetFileStoreByIDForUpdate(ctx, tokenState.FileStoreID, tx)
		if loadErr != nil {
			if errors.Is(loadErr, gorm.ErrRecordNotFound) {
				return ErrInstantUploadUnavailable
			}
			return errors.WithMessage(loadErr, "文件查询异常")
		}
		if fileStore.HashType != hashType || fileStore.FileHash != fileHash {
			return ErrInstantProofInvalid
		}
		if fileSize > 0 && fileStore.FileSize != fileSize {
			return ErrInstantUploadUnavailable
		}
		user, userErr := s.userRepo.GetUserInfo(ctx, userID)
		if userErr != nil {
			return errors.WithMessage(userErr, "获取用户信息失败")
		}
		if user.UsedSpace+fileStore.FileSize > user.TotalSpace {
			return ErrUserSpaceNotEnough
		}
		uploadedID, err = s.createInstantUploadRecord(ctx, tx, userID, parentID, fileName, fileStore.FileSize, fileStore)
		return errors.WithMessage(err, "秒传文件存储失败")
	})
	if err != nil {
		return 0, errors.WithMessage(err, "秒传失败")
	}

	_ = s.uploadCache.DeleteInstantProofToken(ctx, proofToken)
	logger.S.Infof("前端秒传确认成功, hashType: %s, fileHash: %s, userID: %d, parentID: %d", hashType, fileHash, userID, parentID)
	return uploadedID, nil
}

// UploadFile 普通上传文件
func (s *FileService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, userID uint, parentID uint) (uint, error) {
	if fileHeader.Size >= chunkUploadThreshold {
		return 0, ErrUploadMethodInvalid
	}
	// 检查父文件夹
	err := s.checkParentFolder(ctx, userID, parentID)
	if err != nil {
		return 0, errors.WithMessage(err, "父文件夹不合法")
	}
	// 检查用户空间
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return 0, errors.WithMessage(err, "获取用户信息失败")
	}
	if user.UsedSpace+fileHeader.Size > user.TotalSpace {
		return 0, ErrUserSpaceNotEnough
	}
	// 计算身份哈希与抽样哈希
	fileHashes, err := hash.HashFileWithSamples(fileHeader)
	if err != nil {
		return 0, errors.WithMessage(err, "文件hash计算失败")
	}
	fileHash := fileHashes.Full
	merkleRoot, merkleLeafCount, err := s.buildMerkleMetaFromFileHeader(fileHeader)
	if err != nil {
		return 0, err
	}
	// 查询文件身份
	_, err = s.fileRepo.GetFileByIdentity(ctx, hashTypeFullSHA256, fileHash)
	// 去重上传成功
	if err == nil {
		var uploadedID uint
		err = s.db.Transaction(func(tx *gorm.DB) error {
			lockedStore, lockErr := s.fileRepo.GetFileByIdentityForUpdate(ctx, hashTypeFullSHA256, fileHash, tx)
			if lockErr != nil {
				if errors.Is(lockErr, gorm.ErrRecordNotFound) {
					return ErrInstantUploadUnavailable
				}
				return errors.WithMessage(lockErr, "文件查询异常")
			}
			if err := s.ensureMerkleMetadata(ctx, tx, lockedStore, merkleRoot, merkleLeafCount); err != nil {
				return err
			}
			uploadedID, err = s.createInstantUploadRecord(ctx, tx, userID, parentID, fileHeader.Filename, fileHeader.Size, lockedStore)
			return err
		})
		if !errors.Is(err, ErrInstantUploadUnavailable) {
			logger.S.Infof("去重上传成功, fileHash: %s, userID: %d, parentID: %d", fileHash, userID, parentID)
			return uploadedID, err
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.WithMessage(err, "文件查询异常")
	}

	// 正常上传
	stream, err := fileHeader.Open()
	if err != nil {
		return 0, errors.WithMessage(err, "文件打开失败")
	}
	defer stream.Close()

	// 拼接文件名
	ext := path.Ext(fileHeader.Filename)
	objectName := fmt.Sprintf("%s%s", fileHash, ext)
	// minio上传文件
	err = s.storage.UploadFile(ctx, objectName, stream, fileHeader.Size)
	if err != nil {
		return 0, errors.WithMessage(err, "上传云储存失败")
	}
	// 若上传数据库失败，清理minio文件
	shouldCleanMinio := true
	defer func() {
		if shouldCleanMinio {
			_ = s.storage.DeleteFile(ctx, objectName)
		}
	}()
	// 开启数据库事务
	var uploadedID uint
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 插入文件元数据
		newFileStore := &model.FileStore{
			HashType:        hashTypeFullSHA256,
			FileHash:        fileHash,
			FileName:        fileHeader.Filename,
			FileSize:        fileHeader.Size,
			FileAddr:        objectName,
			HeadHash:        fileHashes.Head,
			MidHash:         fileHashes.Mid,
			TailHash:        fileHashes.Tail,
			MerkleType:      merkleTypeChunkSHA256V1,
			MerkleRoot:      merkleRoot,
			MerkleChunkSize: hash.ChunkIdentitySize,
			MerkleLeafCount: merkleLeafCount,
		}
		err = s.fileRepo.CreateFileStore(ctx, newFileStore, tx)
		if err != nil {
			return errors.WithMessage(err, "文件元数据存储失败")
		}
		// 插入用户文件数据
		newUserFile := &model.UserFile{
			UserId:      userID,
			FileStoreID: &newFileStore.ID,
			FileName:    fileHeader.Filename,
			ParentID:    parentID,
		}
		err = s.fileRepo.CreateUserFile(ctx, newUserFile, tx)
		if err != nil {
			return errors.WithMessage(err, "用户文件数据存储失败")
		}
		uploadedID = newUserFile.ID
		// 更新用户空间
		err = s.userRepo.UpdateUserSpace(ctx, userID, fileHeader.Size, tx)
		if err != nil {
			return errors.WithMessage(err, "更新用户空间失败")
		}
		shouldCleanMinio = false
		return nil
	})
	return uploadedID, err
}

// DeleteFile 删除文件/文件夹
func (s *FileService) DeleteFile(ctx context.Context, userID uint, userFileID uint) error {
	// 获取文件
	root, err := s.fileRepo.GetFileByID(ctx, userFileID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFileNotFound
		}
		return errors.WithMessage(err, "获取文件失败")
	}
	// 删除文件夹
	if root.IsFolder {
		ids := make([]uint, 0, 16)
		ids = append(ids, userFileID) // 添加根文件ID

		visited := make(map[uint]struct{}, 16)
		visited[userFileID] = struct{}{} // 标记根文件已访问
		// 收集子树ID
		err := s.collectSubtreeIDs(ctx, userID, userFileID, &ids, visited)
		if err != nil {
			return errors.WithMessage(err, "收集子树ID失败")
		}
		// 删除子文件
		err = s.fileRepo.DeleteUserFileByIDs(ctx, userID, ids)
		if err != nil {
			if errors.Is(repository.ErrFileNotFound, err) {
				return ErrFileNotFound
			}
			return errors.WithMessage(err, "删除子文件失败")
		}
		return nil
	}
	// 删除文件
	err = s.fileRepo.DeleteUserFile(ctx, userID, userFileID)
	if err != nil {
		if errors.Is(repository.ErrFileNotFound, err) {
			return ErrFileNotFound
		}
		return errors.WithMessage(err, "删除文件失败")
	}
	return nil
}

// BatchDeleteFile 批量删除文件/文件夹
func (s *FileService) BatchDeleteFile(ctx context.Context, userID uint, userFileIDs []uint) error {
	if len(userFileIDs) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(userFileIDs))
	visited := make(map[uint]struct{}, len(userFileIDs))

	for _, userFileID := range userFileIDs {
		if _, ok := visited[userFileID]; ok {
			continue
		}

		root, err := s.fileRepo.GetFileByID(ctx, userFileID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFileNotFound
			}
			return errors.WithMessage(err, "获取文件失败")
		}

		visited[userFileID] = struct{}{}
		ids = append(ids, userFileID)

		if root.IsFolder {
			if err := s.collectSubtreeIDs(ctx, userID, userFileID, &ids, visited); err != nil {
				return errors.WithMessage(err, "收集子树ID失败")
			}
		}
	}

	err := s.fileRepo.DeleteUserFileByIDs(ctx, userID, ids)
	if err != nil {
		if errors.Is(repository.ErrFileNotFound, err) {
			return ErrFileNotFound
		}
		return errors.WithMessage(err, "批量删除文件失败")
	}

	return nil
}

// GetUserFile 获取用户文件列表
func (s *FileService) GetUserFile(ctx context.Context, userID uint, parentID uint) ([]FileResp, error) {
	// 检查父文件夹
	err := s.checkParentFolder(ctx, userID, parentID)
	if err != nil {
		return nil, errors.WithMessage(err, "父文件夹不合法")
	}
	// 查询用户文件
	list, err := s.fileRepo.GetUserFile(ctx, userID, parentID)
	if err != nil {
		return nil, errors.WithMessage(err, "查询用户文件失败")
	}
	// 格式化返回数据
	resp := make([]FileResp, 0, len(list))
	for _, f := range list {
		item := FileResp{
			ID:        f.ID,
			FileName:  f.FileName,
			IsFolder:  f.IsFolder,
			UpdatedAt: f.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		// 判断是否为文件夹
		if f.IsFolder {
			item.FileSize = "0"
		} else {
			item.FileSize = convert.FormatFileSize(f.FileStore.FileSize)
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// CreateFolder 创建文件夹
func (s *FileService) CreateFolder(ctx context.Context, userID uint, parentID uint, name string) error {
	// 检查父文件夹是否合法
	if err := s.checkParentFolder(ctx, userID, parentID); err != nil {
		return errors.WithMessage(err, "父文件夹不合法")
	}

	// 构造用户文件夹记录
	newFolder := &model.UserFile{
		UserId:   userID,
		FileName: name,
		ParentID: parentID,
		IsFolder: true,
	}

	// 写入目录记录
	if err := s.fileRepo.CreateUserFile(ctx, newFolder); err != nil {
		return errors.WithMessage(err, "创建文件夹失败")
	}

	return nil
}

// ListRecycleBin 查询回收站
func (s *FileService) ListRecycleBin(ctx context.Context, userID uint) ([]RecycleFileResp, error) {
	// 查询回收站
	list, err := s.fileRepo.ListRecycleBin(ctx, userID)
	if err != nil {
		return nil, errors.WithMessage(err, "查询回收站失败")
	}
	// 格式化返回数据
	resp := make([]RecycleFileResp, 0, len(list))
	for _, f := range list {
		item := RecycleFileResp{
			ID:        f.ID,
			FileName:  f.FileName,
			IsFolder:  f.IsFolder,
			DeletedAt: f.DeletedAt.Time.Format("2006-01-02 15:04:05"),
		}
		// 判断是否为文件夹
		if f.IsFolder {
			item.FileSize = "0"
		} else {
			item.FileSize = convert.FormatFileSize(f.FileStore.FileSize)
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// RestoreUserFile 恢复文件
func (s *FileService) RestoreUserFile(ctx context.Context, userID uint, ID uint) error {
	err := s.fileRepo.RestoreUserFile(ctx, userID, ID)
	if err != nil {
		if errors.Is(repository.ErrFileNotFound, err) {
			return ErrFileNotFound
		}
		return errors.WithMessage(err, "恢复文件失败")
	}
	return nil
}

// PermanentlyDeleteFile 永久删除回收站中的文件/文件夹
func (s *FileService) PermanentlyDeleteFile(ctx context.Context, userID uint, userFileID uint) error {
	// 仅允许对当前用户回收站中的文件进行永久删除
	file, err := s.fileRepo.GetDeletedUserFileByID(ctx, userID, userFileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFileNotFound
		}
		return errors.WithMessage(err, "查询回收站文件失败")
	}

	// 文件夹不占空间，也没有文件池记录，直接硬删除用户文件记录
	if file.IsFolder || file.FileStoreID == nil {
		return s.fileRepo.HardDeleteUserFile(ctx, userID, userFileID)
	}

	// 普通文件：需要释放用户空间，并在无人引用时删除文件池和 MinIO 对象
	fileSize := file.FileStore.FileSize
	storeID := *file.FileStoreID
	objectName := file.FileStore.FileAddr

	// 开启事务:删除用户文件记录、扣减用户空间、检查剩余引用并删除文件池记录
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 永久删除用户文件记录
		if err := s.fileRepo.HardDeleteUserFile(ctx, userID, userFileID, tx); err != nil {
			return errors.WithMessage(err, "永久删除用户文件失败")
		}

		// 扣减用户空间
		if err := s.userRepo.UpdateUserSpace(ctx, userID, -fileSize, tx); err != nil {
			return errors.WithMessage(err, "更新用户空间失败")
		}

		if _, err := s.fileRepo.GetFileStoreByIDForUpdate(ctx, storeID, tx); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return errors.WithMessage(err, "锁定文件池记录失败")
		}

		// 检查是否还有其他引用
		count, err := s.fileRepo.CountAllUserFileByStoreID(ctx, storeID, tx)
		if err != nil {
			return errors.WithMessage(err, "统计文件引用数量失败")
		}
		if count == 0 {
			// 无其他引用时，删除文件池记录
			if err := s.fileRepo.HardDeleteFileStore(ctx, storeID, tx); err != nil {
				return errors.WithMessage(err, "删除文件池记录失败")
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 事务提交后，若已无任何引用，则尝试删除 MinIO 对象
	count, err := s.fileRepo.CountAllUserFileByStoreID(ctx, storeID)
	if err == nil && count == 0 {
		if delErr := s.storage.DeleteFile(ctx, objectName); delErr != nil {
			logger.S.Errorf("删除MinIO对象失败：%v", delErr)
		}
	}

	return nil
}

// GetDownloadURL 获取下载URL
func (s *FileService) GetDownloadURL(ctx context.Context, userID uint, userFileID uint) (DownloadFileResp, error) {
	file, fileName, err := s.fileRepo.GetFileStoreByID(ctx, userFileID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DownloadFileResp{}, ErrFileNotFound
		}
		return DownloadFileResp{}, errors.WithMessage(err, "获取文件失败")
	}
	//检查用户会员状态
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return DownloadFileResp{}, errors.WithMessage(err, "获取用户信息失败")
	}
	tier := s.checkUserMember(user)

	url, err := s.storage.DownloadFile(ctx, file.FileAddr, fileName, 15*time.Minute, tier)
	if err != nil {
		return DownloadFileResp{}, errors.WithMessage(err, "下载URL获取失败")
	}
	return DownloadFileResp{URL: url, FileName: fileName}, nil
}

// CleanupExpiredUploadSessions 清理超时未完成的分块上传会话
func (s *FileService) CleanupExpiredUploadSessions(ctx context.Context, expireBefore time.Time, limit int) (int, error) {
	//获取已超时的会话
	sessions, err := s.fileRepo.ListExpiredPendingUploadSessions(ctx, expireBefore, limit)
	if err != nil {
		return 0, errors.WithMessage(err, "查询超时上传会话失败")
	}
	cleaned := 0
	for _, session := range sessions {
		if err := s.cleanupExpiredUploadSession(ctx, session); err != nil {
			logger.S.Errorf("清理僵尸分块上传失败, uploadID: %d, objectName: %s, err: %+v", session.ID, session.ObjectName, err)
			continue
		}
		cleaned++
	}
	return cleaned, nil
}

// cleanupExpiredUploadSession 清理某条僵尸会话
func (s *FileService) cleanupExpiredUploadSession(ctx context.Context, session model.UploadSession) error {
	if session.Status != uploadSessionStatusPending {
		return nil
	}
	if session.ObjectName != "" && session.StorageUploadID != "" {
		err := s.storage.AbortMultipartUpload(ctx, session.ObjectName, session.StorageUploadID)
		if err != nil && !isMultipartUploadNotFound(err) {
			return errors.WithMessage(err, "终止 MinIO 分块上传失败")
		}
	}
	if err := s.uploadCache.DeleteUploadState(ctx, session.ID); err != nil {
		return errors.WithMessage(err, "删除 Redis 上传状态失败")
	}
	if err := s.fileRepo.DeleteUploadSession(ctx, session.ID); err != nil {
		return errors.WithMessage(err, "删除上传会话失败")
	}
	logger.S.Infof("清理僵尸分块上传成功, uploadID: %d, objectName: %s", session.ID, session.ObjectName)
	return nil
}
