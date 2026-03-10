package repository

import (
	"WeDrive/internal/model"
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ShareRepo struct {
	db *gorm.DB
}

func NewShareRepo(db *gorm.DB) *ShareRepo {
	return &ShareRepo{db: db}
}

// CreateShareFile 创建分享文件
func (r *ShareRepo) CreateShareFile(ctx context.Context, shareFile *model.ShareFile) error {
	return r.db.WithContext(ctx).Create(shareFile).Error
}

// GetShareFile 获取分享文件
func (r *ShareRepo) GetShareFile(ctx context.Context, token string) (*model.ShareFile, error) {
	var shareFile model.ShareFile
	err := r.db.WithContext(ctx).Where("share_token = ?", token).First(&shareFile).Error
	return &shareFile, errors.WithStack(err)
}
