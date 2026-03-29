package service

import (
	"WeDrive/internal/model"
	"WeDrive/internal/oss"
	"WeDrive/internal/repository"
	"WeDrive/pkg/logger"
	"WeDrive/pkg/utils/convert"
	"WeDrive/pkg/utils/hash"
	"context"
	"fmt"
	"mime/multipart"
	"path"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var ErrUserSpaceNotEnough = errors.New("用户空间不足")
var ErrFileNotFound = errors.New("文件不存在")
var ErrParentFolderInvalid = errors.New("父文件夹不合法")
var ErrInstantUploadUnavailable = errors.New("秒传条件失效")

type FileService struct {
	fileRepo *repository.FileRepo
	userRepo *repository.UserRepo
	storage  *oss.Storage
	db       *gorm.DB
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

func NewFileService(fileRepo *repository.FileRepo, userRepo *repository.UserRepo, storage *oss.Storage, db *gorm.DB) *FileService {
	return &FileService{fileRepo: fileRepo, storage: storage, db: db, userRepo: userRepo}
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

// InstantUpload 命中秒传后的确认落库
func (s *FileService) InstantUpload(ctx context.Context, userID uint, parentID uint, fileName string, fileHash string, fileSize int64) (uint, error) {
	if err := s.checkParentFolder(ctx, userID, parentID); err != nil {
		return 0, errors.WithMessage(err, "父文件夹不合法")
	}
	var uploadedID uint
	err := s.db.Transaction(func(tx *gorm.DB) error {
		//校验哈希
		fileStore, err := s.fileRepo.GetFileByHashForUpdate(ctx, fileHash, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInstantUploadUnavailable
			}
			return errors.WithMessage(err, "文件查询异常")
		}
		//校验文件大小
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

	logger.S.Infof("前端秒传确认成功, fileHash: %s, userID: %d, parentID: %d", fileHash, userID, parentID)
	return uploadedID, nil
}

// UploadFile 上传文件
func (s *FileService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, userID uint, parentID uint) (uint, error) {
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
	// 计算文件hash
	fileHashes, err := hash.HashFileWithSamples(fileHeader)
	if err != nil {
		return 0, errors.WithMessage(err, "文件hash计算失败")
	}
	fileHash := fileHashes.Full
	// 查询文件哈希
	_, err = s.fileRepo.GetFileByHash(ctx, fileHash)
	// 秒传成功
	if err == nil {
		var uploadedID uint
		err = s.db.Transaction(func(tx *gorm.DB) error {
			lockedStore, lockErr := s.fileRepo.GetFileByHashForUpdate(ctx, fileHash, tx)
			if lockErr != nil {
				if errors.Is(lockErr, gorm.ErrRecordNotFound) {
					return ErrInstantUploadUnavailable
				}
				return errors.WithMessage(lockErr, "文件查询异常")
			}
			uploadedID, err = s.createInstantUploadRecord(ctx, tx, userID, parentID, fileHeader.Filename, fileHeader.Size, lockedStore)
			return err
		})
		if !errors.Is(err, ErrInstantUploadUnavailable) {
			logger.S.Infof("上传成功, fileHash: %s, userID: %d, parentID: %d", fileHash, userID, parentID)
			return uploadedID, err
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.WithMessage(err, "文件查询异常")
	}

	// 秒传失败, 正常上传
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
			FileHash: fileHash,
			FileName: fileHeader.Filename,
			FileSize: fileHeader.Size,
			FileAddr: objectName,
			HeadHash: fileHashes.Head,
			MidHash:  fileHashes.Mid,
			TailHash: fileHashes.Tail,
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
