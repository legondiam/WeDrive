package repository

import (
	"WeDrive/internal/model"
	"context"
	"time"

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

// ListActiveShareTokenAfterID 分页查询仍有效的分享 token。
func (r *ShareRepo) ListActiveShareTokenAfterID(ctx context.Context, lastID uint, limit int, now time.Time) ([]model.ShareFile, error) {
	if limit <= 0 {
		limit = 1000
	}
	var list []model.ShareFile
	err := r.db.WithContext(ctx).
		Select("id", "share_token").
		Where("id > ?", lastID).
		Where("expires_at IS NULL OR expires_at > ?", now).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, errors.WithStack(err)
}
