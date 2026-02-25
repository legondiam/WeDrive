package repository

import (
	"WeDrive/internal/model"
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// CreateUser 创建用户
func (r *UserRepo) CreateUser(ctx context.Context, user *model.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetUserByName 根据用户名获取用户
func (r *UserRepo) GetUserByName(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

// GetUserInfo 获取用户信息
func (r *UserRepo) GetUserInfo(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

// UpdateUserSpace 更新用户空间
func (r *UserRepo) UpdateUserSpace(ctx context.Context, userID uint, delta int64, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("used_space", gorm.Expr("used_space + ?", delta)).Error
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
