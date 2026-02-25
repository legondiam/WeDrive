package service

import (
	"WeDrive/internal/model"
	"WeDrive/internal/oss"
	"WeDrive/internal/repository"
	"WeDrive/pkg/utils/convert"
	"WeDrive/pkg/utils/hash"
	"context"
	"fmt"
	"mime/multipart"
	"path"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type FileService struct {
	fileRepo *repository.FileRepo
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
	ParentID  int64  `json:"parent_id"`
}

func NewFileService(fileRepo *repository.FileRepo, storage *oss.Storage, db *gorm.DB) *FileService {
	return &FileService{fileRepo: fileRepo, storage: storage, db: db}
}

// checkParentFolder 检查父文件夹是否合法
func (s *FileService) checkParentFolder(ctx context.Context, userID uint, parentID int64) error {
	// 根目录合法
	if parentID == 0 {
		return nil
	}
	// 查文件夹是否存在
	folder, err := s.fileRepo.GetFileByID(ctx, uint(parentID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("父文件夹不存在")
		}
		return errors.WithMessage(err, "检查父文件夹失败")
	}
	// 检查归属权
	if folder.UserId != userID {
		return errors.New("无权访问该目录")
	}
	// 检查是否为文件夹
	if !folder.IsFolder {
		return errors.New("目标不是文件夹")
	}
	return nil
}

// UploadFile 上传文件
func (s *FileService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, userID uint, parentID int64) error {
	// 检查父文件夹
	err := s.checkParentFolder(ctx, userID, parentID)
	if err != nil {
		return errors.WithMessage(err, "父文件夹不合法")
	}
	// 计算文件hash
	fileHash, err := hash.HashFile(fileHeader)
	if err != nil {
		return errors.WithMessage(err, "文件hash计算失败")
	}
	// 查询文件哈希
	fileStore, err := s.fileRepo.GetFileByHash(ctx, fileHash)
	// 秒传成功
	if err == nil {
		err = s.fileRepo.CreateUserFile(ctx, &model.UserFile{
			UserId:      userID,
			FileName:    fileHeader.Filename,
			FileStoreID: &fileStore.ID,
			ParentID:    parentID,
		})
		if err != nil {
			return errors.WithMessage(err, "秒传文件存储失败")
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.WithMessage(err, "文件查询异常")
	}

	// 秒传失败, 正常上传
	stream, err := fileHeader.Open()
	if err != nil {
		return errors.WithMessage(err, "文件打开失败")
	}
	defer stream.Close()

	// 拼接文件名
	ext := path.Ext(fileHeader.Filename)
	objectName := fmt.Sprintf("%s%s", fileHash, ext)
	// minio上传文件
	err = s.storage.UploadFile(ctx, objectName, stream, fileHeader.Size)
	if err != nil {
		return errors.WithMessage(err, "上传云储存失败")
	}
	// 若上传数据库失败，清理minio文件
	shouldCleanMinio := true
	defer func() {
		if shouldCleanMinio {
			_ = s.storage.DeleteFile(ctx, objectName)
		}
	}()
	// 开启数据库事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 插入文件元数据
		newFileStore := &model.FileStore{
			FileHash: fileHash,
			FileName: fileHeader.Filename,
			FileSize: fileHeader.Size,
			FileAddr: objectName,
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
		shouldCleanMinio = false
		return nil
	})
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(ctx context.Context, userID uint, userFileID uint) error {
	err := s.fileRepo.DeleteUserFile(ctx, userID, userFileID)
	if err != nil {
		if errors.Is(repository.ErrFileNotFound, err) {
			return err
		}
		return errors.WithMessage(err, "删除文件失败")
	}
	return nil
}

// GetUserFile 获取用户文件列表
func (s *FileService) GetUserFile(ctx context.Context, userID uint, parentID int64) ([]FileResp, error) {
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
			UpdatedAt: f.UpdatedAt.Format("2006-01-02"),
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
func (s *FileService) CreateFolder(ctx context.Context, userID uint, parentID int64, name string) error {
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
			return err
		}
		return errors.WithMessage(err, "恢复文件失败")
	}
	return nil
}
