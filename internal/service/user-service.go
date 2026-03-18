package service

import (
	"WeDrive/internal/model"
	"WeDrive/internal/repository"
	"WeDrive/pkg/logger"
	"WeDrive/pkg/utils/convert"
	"WeDrive/pkg/utils/hash"
	"WeDrive/pkg/utils/jwts"
	"context"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

var ErrUserExisted = errors.New("用户已存在")
var ErrAccountOrPassword = errors.New("账号或密码错误")
var ErrTokenNotFound = errors.New("token不存在")

type UserService struct {
	userRepo      *repository.UserRepo
	usercacheRepo *repository.UserCacheRepo
}

type UserInfoResp struct {
	TotalSpace string
	UsedSpace  string
}

func NewUserService(userrepo *repository.UserRepo, usercacherepo *repository.UserCacheRepo) *UserService {
	return &UserService{userRepo: userrepo, usercacheRepo: usercacherepo}
}

// Register 注册
func (s *UserService) Register(ctx context.Context, username string, password string) error {
	password, err := hash.HashPassword(password)
	if err != nil {
		return errors.WithMessage(err, "密码加密失败")
	}
	user := &model.User{
		Username: username,
		Password: password,
	}
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return errors.WithMessage(err, "用户名已存在")
		}
		return errors.WithMessage(ErrUserExisted, "数据库创建用户失败")
	}
	return nil
}

// Login 登录
func (s *UserService) Login(ctx context.Context, username string, password string) (string, string, error) {
	//查找用户
	user, err := s.userRepo.GetUserByName(ctx, username)
	if err != nil {
		return "", "", errors.WithMessage(ErrAccountOrPassword, "账号或密码错误")
	}
	//校验密码
	ok, err := hash.CheckPassword(password, user.Password)
	if err != nil {
		return "", "", errors.WithMessage(err, "密码校验失败")
	}
	if !ok {
		return "", "", errors.WithMessage(ErrAccountOrPassword, "账号或密码错误")
	}
	//生成token
	accessToken, err := jwts.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return "", "", errors.WithMessage(err, "生成accessToken失败")
	}
	refreshToken, tokenID, err := jwts.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return "", "", errors.WithMessage(err, "生成refreshToken失败")
	}
	//缓存refreshtoken
	err = s.usercacheRepo.SetRefreshToken(ctx, user.ID, tokenID, 7*24*time.Hour)
	if err != nil {
		return "", "", errors.WithMessage(err, "缓存refreshToken失败")
	}
	//fmt.Println("refreshtoken:", refreshToken, "\ntokenID:", tokenID, "\nuserid:", user.ID)
	return accessToken, refreshToken, nil
}

// RefreshToken 刷新token
func (s *UserService) RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error) {
	//校验token
	claims, err := jwts.ValidateToken(oldRefreshToken)
	if err != nil {
		return "", "", errors.WithMessage(err, "token校验失败")
	}
	oldTokenID := claims.RegisteredClaims.ID
	userID := claims.UserID
	//校验token是否在缓存中
	ok, err := s.usercacheRepo.GetRefreshToken(ctx, userID, oldTokenID)
	if err != nil {
		return "", "", errors.WithMessage(err, "校验token失败")
	}
	if !ok {
		return "", "", ErrTokenNotFound
	}

	//生成token
	accessToken, err := jwts.GenerateAccessToken(userID, claims.Username)
	if err != nil {
		return "", "", errors.WithMessage(err, "生成accessToken失败")
	}
	newRefreshToken, newTokenID, err := jwts.GenerateRefreshToken(userID, claims.Username)
	if err != nil {
		return "", "", errors.WithMessage(err, "生成refreshToken失败")
	}
	//缓存refreshToken
	err = s.usercacheRepo.SetRefreshToken(ctx, userID, newTokenID, 7*24*time.Hour)
	if err != nil {
		return "", "", errors.WithMessage(err, "缓存refreshToken失败")
	}
	err = s.usercacheRepo.DeleteRefreshToken(ctx, oldTokenID)
	if err != nil {
		logger.S.Warnf("删除refreshToken失败:%v", err)
		return "", "", errors.WithMessage(err, "删除refreshToken失败")
	}
	return accessToken, newRefreshToken, nil
}

// Logout 退出登录
func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := jwts.ValidateToken(refreshToken)
	if err != nil {
		return nil
	}
	tokenID := claims.RegisteredClaims.ID
	if tokenID == "" {
		return nil
	}
	err = s.usercacheRepo.DeleteRefreshToken(ctx, tokenID)
	if err != nil {
		return errors.WithMessage(err, "删除refreshToken失败")
	}
	return nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userID uint) (*UserInfoResp, error) {
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, errors.WithMessage(err, "获取用户信息失败")
	}
	//格式化返回数据
	userInfoResp := &UserInfoResp{
		TotalSpace: convert.FormatFileSize(user.TotalSpace),
		UsedSpace:  convert.FormatFileSize(user.UsedSpace),
	}
	return userInfoResp, nil
}

// UpdateUserMember 更新用户会员状态
func (s *UserService) UpdateUserMember(ctx context.Context, userID uint, memberLevel int8, vipMonths int) error {
	//获取用户会员信息
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return errors.WithMessage(err, "获取用户会员信息失败")
	}
	//计算会员到期时间
	baseTime := time.Now()
	if user.VipExpireAt != nil && user.VipExpireAt.After(baseTime) {
		baseTime = *user.VipExpireAt
	}
	vipExpireAt := baseTime.AddDate(0, vipMonths, 0)
	//更新用户会员状态
	err = s.userRepo.UpdateUserMember(ctx, userID, memberLevel, &vipExpireAt)
	if err != nil {
		return errors.WithMessage(err, "更新用户会员状态失败")
	}
	return nil
}
