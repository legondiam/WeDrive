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
