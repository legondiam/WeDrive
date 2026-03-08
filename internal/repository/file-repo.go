package repository

import (
	"WeDrive/internal/model"
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
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
func (r *FileRepo) GetFileStoreByID(ctx context.Context, ID uint, userID uint) (*model.FileStore, error) {
	var fileStore model.FileStore
	err := r.db.WithContext(ctx).
		Table("file_stores").
		Select("file_stores.*").
		Joins("JOIN user_files ON user_files.file_store_id = file_stores.id").
		Where("user_files.id = ?", ID).
		Where("user_files.user_id = ?", userID).
		Where("user_files.is_folder = ?", false).
		Where("user_files.deleted_at IS NULL").
		Where("file_stores.deleted_at IS NULL").
		First(&fileStore).Error

	return &fileStore, errors.WithStack(err)
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

// CountUserFileByStoreID 统计未被软删除的用户文件引用数量
func (r *FileRepo) CountUserFileByStoreID(ctx context.Context, fileStoreID uint, tx ...*gorm.DB) (int64, error) {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	var count int64
	err := db.WithContext(ctx).
		Model(&model.UserFile{}).
		Where("file_store_id = ? AND deleted_at IS NULL", fileStoreID).
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
