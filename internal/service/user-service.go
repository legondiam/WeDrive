package service

import (
	"WeDrive/internal/cache"
	"WeDrive/internal/model"
	"WeDrive/internal/mq"
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
	userRepo       *repository.UserRepo
	usercacheRepo  *repository.UserCacheRepo
	cachePublisher *mq.CacheInvalidationPublisher
}

type UserInfoResp struct {
	Username     string
	TotalSpace   string
	UsedSpace    string
	IsMember     bool
	MemberStatus string
}

func NewUserService(userrepo *repository.UserRepo, usercacherepo *repository.UserCacheRepo, cachePublisher *mq.CacheInvalidationPublisher) *UserService {
	return &UserService{userRepo: userrepo, usercacheRepo: usercacherepo, cachePublisher: cachePublisher}
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
	user, err := s.userRepo.GetUserByName(ctx, username)
	if err != nil {
		return "", "", errors.WithMessage(ErrAccountOrPassword, "账号或密码错误")
	}

	ok, err := hash.CheckPassword(password, user.Password)
	if err != nil {
		return "", "", errors.WithMessage(err, "密码校验失败")
	}
	if !ok {
		return "", "", errors.WithMessage(ErrAccountOrPassword, "账号或密码错误")
	}

	accessToken, err := jwts.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return "", "", errors.WithMessage(err, "生成accessToken失败")
	}
	refreshToken, tokenID, err := jwts.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return "", "", errors.WithMessage(err, "生成refreshToken失败")
	}
	if err = s.usercacheRepo.SetRefreshToken(ctx, user.ID, tokenID, 7*24*time.Hour); err != nil {
		return "", "", errors.WithMessage(err, "缓存refreshToken失败")
	}
	return accessToken, refreshToken, nil
}

// RefreshToken 刷新token
func (s *UserService) RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error) {
	claims, err := jwts.ValidateToken(oldRefreshToken)
	if err != nil {
		return "", "", errors.WithMessage(err, "token校验失败")
	}
	oldTokenID := claims.RegisteredClaims.ID
	userID := claims.UserID

	ok, err := s.usercacheRepo.GetRefreshToken(ctx, userID, oldTokenID)
	if err != nil {
		return "", "", errors.WithMessage(err, "校验token失败")
	}
	if !ok {
		return "", "", ErrTokenNotFound
	}

	accessToken, err := jwts.GenerateAccessToken(userID, claims.Username)
	if err != nil {
		return "", "", errors.WithMessage(err, "生成accessToken失败")
	}
	newRefreshToken, newTokenID, err := jwts.GenerateRefreshToken(userID, claims.Username)
	if err != nil {
		return "", "", errors.WithMessage(err, "生成refreshToken失败")
	}
	if err = s.usercacheRepo.SetRefreshToken(ctx, userID, newTokenID, 7*24*time.Hour); err != nil {
		return "", "", errors.WithMessage(err, "缓存refreshToken失败")
	}
	if err = s.usercacheRepo.DeleteRefreshToken(ctx, oldTokenID); err != nil {
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
	if err = s.usercacheRepo.DeleteRefreshToken(ctx, tokenID); err != nil {
		return errors.WithMessage(err, "删除refreshToken失败")
	}
	return nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userID uint) (*UserInfoResp, error) {
	cachedUser, ok, err := s.usercacheRepo.GetUserInfo(ctx, userID)
	if err != nil {
		logger.S.Warnf("读取用户信息缓存失败:%v", err)
	}
	if ok {
		return userInfoRespFromCache(cachedUser), nil
	}

	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, errors.WithMessage(err, "获取用户信息失败")
	}
	cacheUser := userInfoCacheFromModel(user)
	if err := s.usercacheRepo.SetUserInfo(ctx, cacheUser); err != nil {
		logger.S.Warnf("写入用户信息缓存失败:%v", err)
	}
	return userInfoRespFromCache(&cacheUser), nil
}

// UpdateUserMember 更新用户会员状态
func (s *UserService) UpdateUserMember(ctx context.Context, userID uint, memberLevel int8, vipMonths int) error {
	user, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return errors.WithMessage(err, "获取用户会员信息失败")
	}

	baseTime := time.Now()
	if user.VipExpireAt != nil && user.VipExpireAt.After(baseTime) {
		baseTime = *user.VipExpireAt
	}
	vipExpireAt := baseTime.AddDate(0, vipMonths, 0)
	if err := s.usercacheRepo.DeleteUserInfo(ctx, userID); err != nil {
		logger.S.Warnf("删除用户信息缓存失败:%v", err)
	}
	if err = s.userRepo.UpdateUserMember(ctx, userID, memberLevel, &vipExpireAt); err != nil {
		return errors.WithMessage(err, "更新用户会员状态失败")
	}
	cache.DelayedDelete(cache.DelayedDeleteDelay, func(ctx context.Context) error {
		err := s.usercacheRepo.DeleteUserInfo(ctx, userID)
		if err != nil && s.cachePublisher != nil {
			if publishErr := s.cachePublisher.PublishUserInfoRetry(context.Background(), userID); publishErr != nil {
				logger.S.Warnf("发送用户信息缓存删除重试消息失败:%v", publishErr)
			}
		}
		return err
	})
	return nil
}

// userInfoCacheFromModel 将用户模型转换为缓存结构
func userInfoCacheFromModel(user *model.User) cache.UserInfo {
	return cache.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		TotalSpace:  user.TotalSpace,
		UsedSpace:   user.UsedSpace,
		MemberLevel: user.MemberLevel,
		VipExpireAt: user.VipExpireAt,
	}
}

// userInfoRespFromCache 将用户缓存转换为响应结构
func userInfoRespFromCache(user *cache.UserInfo) *UserInfoResp {
	now := time.Now()
	isMember := user.MemberLevel > 0 && user.VipExpireAt != nil && user.VipExpireAt.After(now)
	return &UserInfoResp{
		Username:     user.Username,
		TotalSpace:   convert.FormatFileSize(user.TotalSpace),
		UsedSpace:    convert.FormatFileSize(user.UsedSpace),
		IsMember:     isMember,
		MemberStatus: map[bool]string{true: "会员", false: "非会员"}[isMember],
	}
}
