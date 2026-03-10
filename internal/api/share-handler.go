package api

import (
	"WeDrive/internal/service"
	"WeDrive/pkg/logger"
	"WeDrive/pkg/response"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

type ShareHandler struct {
	shareService *service.ShareService
}

func NewShareHandler(shareService *service.ShareService) *ShareHandler {
	return &ShareHandler{shareService: shareService}
}

// CreateShareFile 创建分享文件
func (h *ShareHandler) CreateShareFile(c *gin.Context) {
	//获取用户ID
	userID, _ := c.Get("userID")
	//获取请求参数
	var req struct {
		UserFileID uint   `json:"user_file_id" binding:"required"`
		Key        string `json:"key"`
		Expiretime string `json:"expiretime" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	if req.Key != "" && len(req.Key) != 4 {
		response.BusinessError(c, response.CodeInvalidParam, "密钥格式不正确")
		return
	}
	//解析过期时间
	var expiretime *time.Time
	switch req.Expiretime {
	case "1":
		t := time.Now().AddDate(0, 0, 1)
		expiretime = &t
	case "7":
		t := time.Now().AddDate(0, 0, 7)
		expiretime = &t
	case "30":
		t := time.Now().AddDate(0, 0, 30)
		expiretime = &t
	case "permanent":
		expiretime = nil
	default:
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	//创建分享文件
	shareToken, err := h.shareService.CreateShareFile(c.Request.Context(), userID.(uint), req.UserFileID, req.Key, expiretime)
	if err != nil {
		if errors.Is(err, service.ErrFolderCannotShare) {
			response.BusinessError(c, response.CodeInvalidFileID, "文件夹不能分享")
			return
		}
		response.ServerError(c, "创建分享文件失败")
		logger.S.Errorf("创建分享文件失败：%v", err)
		return
	}
	response.Success(c, gin.H{"shareToken": shareToken})
	logger.S.Info("创建分享文件成功")
}

// GetShareDownloadURL 获取分享文件下载URL
func (h *ShareHandler) GetShareDownloadURL(c *gin.Context) {
	//获取请求参数
	var req struct {
		ShareToken string `json:"share_token" binding:"required"`
		Key        string `json:"key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	//获取分享文件下载URL
	downloadURL, err := h.shareService.GetShareDownloadURL(c.Request.Context(), req.ShareToken, req.Key)
	switch {
	case err == nil:
		response.Success(c, downloadURL)
	case errors.Is(err, service.ErrShareNotFound):
		response.BusinessError(c, response.CodeShareNotFound, "分享不存在")
	case errors.Is(err, service.ErrShareExpired):
		response.BusinessError(c, response.CodeShareExpired, "分享已过期")
	case errors.Is(err, service.ErrShareInvalidKey):
		response.BusinessError(c, response.CodeShareInvalidKey, "密钥不正确")
	default:
		response.ServerError(c, "获取分享文件下载URL失败")
		logger.S.Errorf("获取分享文件下载URL失败：%v", err)
		return
	}
	logger.S.Info("获取分享文件下载URL成功")
}
