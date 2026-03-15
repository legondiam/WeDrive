package api

import (
	"WeDrive/internal/config"
	"WeDrive/internal/service"
	"WeDrive/pkg/logger"
	"WeDrive/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register 注册
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	//解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "传入数据有误")
		return
	}
	//注册
	err := h.userService.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExisted) {
			response.BusinessError(c, response.CodeUserExisted, "用户已存在")
			logger.S.Infof("用户已存在：%+v", err)
			return
		}
		response.ServerError(c, "注册失败")
		logger.S.Errorf("服务器错误：%+v", err)
		return
	}
	response.Success(c, nil)
	logger.S.Info("用户注册成功。", "username:", req.Username)
}
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	//解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "传入数据有误")
		return
	}
	//登录
	accessToken, refreshToken, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrAccountOrPassword) {
			response.BusinessError(c, response.CodeAccountOrPassword, "用户名或密码错误")
			logger.S.Infof("用户名或密码错误：%+v", err)
			return
		}
		response.ServerError(c, "登录失败")
		logger.S.Errorf("服务器错误：%+v", err)
		return
	}
	//将refreshToken存入cookie
	maxAge := int(config.GlobalConf.Jwt.RefreshTokenExpiration.Seconds())
	c.SetCookie("refreshToken", refreshToken, maxAge, "/", "localhost", false, true)
	response.Success(c, gin.H{"accessToken": accessToken})
	logger.S.Info("用户登录成功。", "username:", req.Username)
}

// Refresh 刷新token
func (h *UserHandler) Refresh(c *gin.Context) {
	//获取旧的refreshToken
	oldRefreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		response.BusinessError(c, response.CodeRefreshTokenMissing, "refreshToken不存在")
		return
	}
	//刷新token
	accessToken, refreshToken, err := h.userService.RefreshToken(c.Request.Context(), oldRefreshToken)
	if err != nil {
		//清除cookie
		c.SetCookie("refreshToken", "", -1, "/", "localhost", false, true)
		if errors.Is(err, service.ErrTokenNotFound) {
			response.BusinessError(c, response.CodeRefreshTokenMissing, "refreshToken不存在")
			return
		}
		response.ServerError(c, "刷新token失败")
		logger.S.Errorf("服务器错误：%+v", err)
		return
	}
	//将新的refreshToken存入cookie
	maxAge := int(config.GlobalConf.Jwt.RefreshTokenExpiration.Seconds())
	c.SetCookie("refreshToken", refreshToken, maxAge, "/", "localhost", false, true)
	response.Success(c, gin.H{"accessToken": accessToken})
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, _ := c.Get("userID")
	userinfo, err := h.userService.GetUserInfo(c.Request.Context(), userID.(uint))
	if err != nil {
		response.ServerError(c, "获取用户信息失败")
		logger.S.Errorf("获取用户信息失败：%+v", err)
		return
	}
	response.Success(c, userinfo)
}

// UpdateUserMember 更新用户会员状态
func (h *UserHandler) UpdateUserMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req struct {
		MemberLevel   int8   `json:"member_level" binding:"required"`
		VipExpireTime string `json:"vip_expire_time" binding:"required"`
	}
	//解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "传入数据有误")
		return
	}
	if req.MemberLevel < 0 || req.MemberLevel > 1 {
		response.BusinessError(c, response.CodeInvalidParam, "会员等级无效")
		return
	}
	var vipExpireAt *time.Time
	switch req.VipExpireTime {
	case "1":
		t := time.Now().AddDate(0, 1, 0)
		vipExpireAt = &t
	case "3":
		t := time.Now().AddDate(0, 3, 0)
		vipExpireAt = &t
	case "12":
		t := time.Now().AddDate(1, 0, 0)
		vipExpireAt = &t
	default:
		response.BusinessError(c, response.CodeInvalidParam, "会员时长无效")
		return
	}
	//更新用户会员状态
	err := h.userService.UpdateUserMember(c.Request.Context(), userID.(uint), req.MemberLevel, vipExpireAt)
	if err != nil {
		response.ServerError(c, "更新用户会员状态失败")
		logger.S.Errorf("更新用户会员状态失败：%+v", err)
		return
	}
	response.Success(c, nil)
	logger.S.Infof("更新用户会员状态成功，用户ID：%d，会员等级：%d，会员到期时间：%v", userID.(uint), req.MemberLevel, vipExpireAt)
}
