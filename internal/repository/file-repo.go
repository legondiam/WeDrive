package repository

import (
	"WeDrive/internal/model"
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrFileNotFound = errors.New("文件不存在或已恢复")

type FileRepo struct {
	db *gorm.DB
}

func NewFileRepo(db *gorm.DB) *FileRepo {
	return &FileRepo{
		db: db,
	}
}

// CreateUploadSession 创建分块上传会话
func (r *FileRepo) CreateUploadSession(ctx context.Context, session *model.UploadSession, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(session).Error
}

// SaveUploadSession 保存分块上传会话
func (r *FileRepo) SaveUploadSession(ctx context.Context, session *model.UploadSession, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Save(session).Error
}

// GetPendingUploadSession 获取待完成的分块上传会话
func (r *FileRepo) GetPendingUploadSession(ctx context.Context, userID uint, parentID uint, fileHash string) (*model.UploadSession, error) {
	var session model.UploadSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND parent_id = ? AND file_hash = ? AND status = ?", userID, parentID, fileHash, "pending").
		Order("id DESC").
		First(&session).Error
	return &session, errors.WithStack(err)
}

// GetUploadSessionByID 获取分块上传会话
func (r *FileRepo) GetUploadSessionByID(ctx context.Context, sessionID uint, userID uint, tx ...*gorm.DB) (*model.UploadSession, error) {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var session model.UploadSession
	err := db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error
	return &session, errors.WithStack(err)
}

// GetUploadSessionByIDForUpdate 获取分块上传会话并加锁
func (r *FileRepo) GetUploadSessionByIDForUpdate(ctx context.Context, sessionID uint, userID uint, tx *gorm.DB) (*model.UploadSession, error) {
	var session model.UploadSession
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error
	return &session, errors.WithStack(err)
}

// DeleteUploadSession 删除分块上传会话
func (r *FileRepo) DeleteUploadSession(ctx context.Context, sessionID uint, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Delete(&model.UploadSession{}, sessionID).Error
}

// CreateFileStore 插入文件元数据
func (r *FileRepo) CreateFileStore(ctx context.Context, file *model.FileStore, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(file).Error
}

// CreateUserFile 插入用户文件记录
func (r *FileRepo) CreateUserFile(ctx context.Context, userFile *model.UserFile, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(userFile).Error
}

// GetFileByHash 根据文件hash获取文件
func (r *FileRepo) GetFileByHash(ctx context.Context, hash string) (*model.FileStore, error) {
	var fileStore model.FileStore
	err := r.db.WithContext(ctx).Where("file_hash = ?", hash).First(&fileStore).Error
	return &fileStore, errors.WithStack(err)
}

// GetFileBySample 根据抽样哈希获取文件
func (r *FileRepo) GetFileBySample(ctx context.Context, fileSize int64, headHash string, midHash string, tailHash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.FileStore{}).
		Where("file_size = ? AND head_hash = ? AND mid_hash = ? AND tail_hash = ?", fileSize, headHash, midHash, tailHash).
		Count(&count).Error
	if err != nil {
		return false, errors.WithStack(err)
	}
	return count > 0, nil
}

// GetFileByHashForUpdate 根据文件hash获取文件并加锁
func (r *FileRepo) GetFileByHashForUpdate(ctx context.Context, hash string, tx *gorm.DB) (*model.FileStore, error) {
	var fileStore model.FileStore
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("file_hash = ?", hash).
		First(&fileStore).Error
	return &fileStore, errors.WithStack(err)
}

// GetFileStoreByIDForUpdate 根据文件池ID获取文件并加锁
func (r *FileRepo) GetFileStoreByIDForUpdate(ctx context.Context, fileStoreID uint, tx *gorm.DB) (*model.FileStore, error) {
	var fileStore model.FileStore
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&fileStore, fileStoreID).Error
	return &fileStore, errors.WithStack(err)
}

// GetFileByID 根据文件ID获取用户文件
func (r *FileRepo) GetFileByID(ctx context.Context, ID uint, userID uint) (*model.UserFile, error) {
	var file model.UserFile
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", ID, userID).First(&file).Error
	return &file, errors.WithStack(err)
}

// GetFileStoreByID 根据文件ID获取文件池文件
//
//	func (r *FileRepo) GetFileStoreByID(ctx context.Context, ID uint,userID uint) (*model.UserFile, error) {
//		var userFile model.UserFile
//		err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", ID, userID).Preload("FileStore").First(&userFile).Error
//		return &userFile, errors.WithStack(err)
//	}
func (r *FileRepo) GetFileStoreByID(ctx context.Context, UserFileID uint, userID uint) (*model.FileStore, string, error) {
	var row struct {
		model.FileStore
		FileName string
	}
	err := r.db.WithContext(ctx).
		Table("file_stores").
		Select("file_stores.*,user_files.file_name").
		Joins("JOIN user_files ON user_files.file_store_id = file_stores.id").
		Where("user_files.id = ?", UserFileID).
		Where("user_files.user_id = ?", userID).
		Where("user_files.is_folder = ?", false).
		Where("user_files.deleted_at IS NULL").
		Where("file_stores.deleted_at IS NULL").
		First(&row).Error

	return &row.FileStore, row.FileName, errors.WithStack(err)
}

// GetUserFileByParentID 根据父文件夹ID获取用户文件列表
func (r *FileRepo) GetUserFileByParentID(ctx context.Context, userID uint, parentID uint) ([]model.UserFile, error) {
	var list []model.UserFile
	err := r.db.WithContext(ctx).
		Select("id,parent_id,is_folder").
		Where("user_id = ? AND parent_id = ?", userID, parentID).
		Find(&list).Error
	return list, errors.WithStack(err)
}

// DeleteUserFile 删除用户文件记录
func (r *FileRepo) DeleteUserFile(ctx context.Context, userID uint, userFileID uint) error {
	result := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, userFileID).Delete(&model.UserFile{})
	if result.RowsAffected == 0 {
		return ErrFileNotFound
	}
	return errors.WithStack(result.Error)
}

// DeleteUserFileByIDs 批量删除用户文件记录
func (r *FileRepo) DeleteUserFileByIDs(ctx context.Context, userID uint, ids []uint) error {
	result := r.db.WithContext(ctx).Where("user_id = ? AND id IN (?)", userID, ids).Delete(&model.UserFile{})
	if result.Error != nil {
		return errors.WithStack(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFileNotFound
	}
	return nil
}

// GetUserFile 获取用户文件列表
func (r *FileRepo) GetUserFile(ctx context.Context, userID uint, parentID uint) ([]model.UserFile, error) {
	var list []model.UserFile
	err := r.db.WithContext(ctx).
		Select("id,file_name,is_folder,updated_at,file_store_id").
		Where("user_id = ? AND parent_id = ?", userID, parentID).
		Preload("FileStore", func(db *gorm.DB) *gorm.DB { return db.Select("id,file_size") }).
		Order("is_folder DESC, updated_at DESC").
		Find(&list).Error
	return list, errors.WithStack(err)
}

// ListRecycleBin 获取回收站中的用户文件
func (r *FileRepo) ListRecycleBin(ctx context.Context, userID uint) ([]model.UserFile, error) {
	var list []model.UserFile
	err := r.db.WithContext(ctx).Unscoped().
		Select("id,file_name,is_folder,deleted_at,file_store_id").
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Preload("FileStore", func(db *gorm.DB) *gorm.DB { return db.Select("id,file_size") }).
		Order("deleted_at DESC").
		Find(&list).Error
	return list, errors.WithStack(err)
}

// RestoreUserFile 从回收站恢复用户文件
func (r *FileRepo) RestoreUserFile(ctx context.Context, userID uint, ID uint) error {
	result := r.db.WithContext(ctx).Unscoped().
		Model(&model.UserFile{}).
		Where("user_id = ? AND id = ? AND deleted_at IS NOT NULL", userID, ID).
		Update("deleted_at", gorm.Expr("NULL"))
	if result.RowsAffected == 0 {
		return ErrFileNotFound
	}
	return errors.WithStack(result.Error)
}

// GetDeletedUserFileByID 获取回收站中的单个用户文件（已软删除）
func (r *FileRepo) GetDeletedUserFileByID(ctx context.Context, userID uint, userFileID uint, tx ...*gorm.DB) (*model.UserFile, error) {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	var file model.UserFile
	err := db.WithContext(ctx).Unscoped().
		Where("user_id = ? AND id = ? AND deleted_at IS NOT NULL", userID, userFileID).
		Preload("FileStore").
		First(&file).Error

	return &file, errors.WithStack(err)
}

// HardDeleteUserFile 永久删除用户文件记录
func (r *FileRepo) HardDeleteUserFile(ctx context.Context, userID uint, userFileID uint, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	result := db.WithContext(ctx).Unscoped().
		Where("user_id = ? AND id = ?", userID, userFileID).
		Delete(&model.UserFile{})
	if result.Error != nil {
		return errors.WithStack(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFileNotFound
	}
	return nil
}

// CountAllUserFileByStoreID 统计文件池记录的全部引用数量（包含回收站中的软删除记录）
func (r *FileRepo) CountAllUserFileByStoreID(ctx context.Context, fileStoreID uint, tx ...*gorm.DB) (int64, error) {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	var count int64
	err := db.WithContext(ctx).Unscoped().
		Model(&model.UserFile{}).
		Where("file_store_id = ?", fileStoreID).
		Count(&count).Error
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return count, nil
}

// HardDeleteFileStore 永久删除文件池中的文件元数据
func (r *FileRepo) HardDeleteFileStore(ctx context.Context, fileStoreID uint, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	result := db.WithContext(ctx).Unscoped().
		Delete(&model.FileStore{}, fileStoreID)
	if result.Error != nil {
		return errors.WithStack(result.Error)
	}
	return nil
}
