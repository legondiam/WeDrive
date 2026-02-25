package api

import (
	"WeDrive/internal/config"
	"WeDrive/internal/service"
	"WeDrive/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	//解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "传入数据有误"})
		return
	}
	//注册
	err := h.userService.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExisted) {
			c.JSON(400, gin.H{"error": "用户已存在"})
			logger.S.Infof("用户已存在：%+v", err)
			return
		}
		c.JSON(500, gin.H{"error": "服务器错误"})
		logger.S.Errorf("服务器错误：%+v", err)
		return
	}
	c.JSON(200, gin.H{"message": "注册成功"})
	logger.S.Info("用户注册成功。", "username:", req.Username)
}
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	//解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "传入数据有误"})
		return
	}
	//登录
	accessToken, refreshToken, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrAccountOrPassword) {
			c.JSON(400, gin.H{"error": "用户名或密码错误"})
			logger.S.Infof("用户名或密码错误：%+v", err)
			return
		}
		c.JSON(500, gin.H{"error": "服务器错误"})
		logger.S.Errorf("服务器错误：%+v", err)
		return
	}
	//将refreshToken存入cookie
	maxAge := int(config.GlobalConf.Jwt.RefreshTokenExpiration.Seconds())
	c.SetCookie("refreshToken", refreshToken, maxAge, "/", "localhost", false, true)
	c.JSON(200, gin.H{"message": "登录成功", "accessToken": accessToken})
	logger.S.Info("用户登录成功。", "username:", req.Username)
}

func (h *UserHandler) Refresh(c *gin.Context) {
	//获取旧的refreshToken
	oldRefreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(400, gin.H{"error": "refreshToken不存在"})
		return
	}
	//刷新token
	accessToken, refreshToken, err := h.userService.RefreshToken(c.Request.Context(), oldRefreshToken)
	if err != nil {
		//清除cookie
		c.SetCookie("refreshToken", "", -1, "/", "localhost", false, true)
		if errors.Is(err, service.ErrTokenNotFound) {
			c.JSON(400, gin.H{"error": "refreshToken不存在"})
			return
		}
		c.JSON(500, gin.H{"error": "服务器错误"})
		logger.S.Errorf("服务器错误：%+v", err)
		return
	}
	//将新的refreshToken存入cookie
	maxAge := int(config.GlobalConf.Jwt.RefreshTokenExpiration.Seconds())
	c.SetCookie("refreshToken", refreshToken, maxAge, "/", "localhost", false, true)
	c.JSON(200, gin.H{"message": "刷新成功", "accessToken": accessToken})
}
