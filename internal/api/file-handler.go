package api

import (
	"WeDrive/internal/service"
	"WeDrive/pkg/logger"
	"WeDrive/pkg/response"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

type instantUploadReq struct {
	FileHash string `json:"file_hash" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
	FileSize int64  `json:"file_size"`
	ParentID uint   `json:"parent_id"`
}

// Upload 上传文件
func (h *FileHandler) Upload(c *gin.Context) {
	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BusinessError(c, response.CodeMissingFile, "请上传文件")
		return
	}
	// 获取父文件夹ID
	parentIDString := c.PostForm("parent_id")
	var parentID uint
	if parentIDString != "" {
		var parsed uint64
		parsed, err = strconv.ParseUint(parentIDString, 10, 64)
		if err != nil {
			response.BusinessError(c, response.CodeInvalidParentID, "parent_id无效")
			return
		}
		parentID = uint(parsed)
	}
	// 获取用户ID
	userID, _ := c.Get("userID")
	// 上传文件

	uploadedID, err := h.fileService.UploadFile(c.Request.Context(), fileHeader, userID.(uint), parentID)
	if err != nil {
		if errors.Is(err, service.ErrParentFolderInvalid) {
			response.BusinessError(c, response.CodeInvalidParentID, "parent_id不合法")
			return
		}
		if errors.Is(err, service.ErrUserSpaceNotEnough) {
			response.BusinessError(c, response.CodeUserSpaceNotEnough, "用户空间不足")
			return
		}
		response.ServerError(c, "上传失败")
		logger.S.Errorf("文件上传失败：%v", err)
		return
	}
	response.Success(c, gin.H{"id": uploadedID})
	logger.S.Infof("文件上传成功。文件ID: %v", uploadedID)
}

// InstantUpload 秒传
func (h *FileHandler) InstantUpload(c *gin.Context) {
	var req instantUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	userID, _ := c.Get("userID")
	uploadedID, err := h.fileService.InstantUpload(c.Request.Context(), userID.(uint), req.ParentID, req.FileName, req.FileHash, req.FileSize)
	if err != nil {
		if errors.Is(err, service.ErrParentFolderInvalid) {
			response.BusinessError(c, response.CodeInvalidParentID, "parent_id不合法")
			return
		}
		if errors.Is(err, service.ErrUserSpaceNotEnough) {
			response.BusinessError(c, response.CodeUserSpaceNotEnough, "用户空间不足")
			return
		}
		if errors.Is(err, service.ErrInstantUploadUnavailable) {
			response.BusinessError(c, response.CodeInstantUnavailable, "不允许秒传")
			return
		}
		response.ServerError(c, "秒传失败")
		logger.S.Errorf("秒传失败：%v", err)
		return
	}
	response.Success(c, gin.H{"instant": true, "id": uploadedID})
}

// CreateFolder 创建文件夹
func (h *FileHandler) CreateFolder(c *gin.Context) {
	type CreateFolderReq struct {
		Name     string `json:"name" binding:"required"`
		ParentID uint   `json:"parent_id"`
	}
	var req CreateFolderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	// 默认父目录为根目录
	parentID := req.ParentID

	userID, _ := c.Get("userID")
	if err := h.fileService.CreateFolder(c.Request.Context(), userID.(uint), parentID, req.Name); err != nil {
		response.ServerError(c, "创建文件夹失败")
		logger.S.Errorf("创建文件夹失败：%v", err)
		return
	}

	response.Success(c, nil)
}

// Delete 删除文件
func (h *FileHandler) Delete(c *gin.Context) {
	IDString := c.Param("ID")
	ID, err := strconv.ParseUint(IDString, 10, 64)
	if err != nil {
		response.BusinessError(c, response.CodeInvalidFileID, "无效的文件ID")
		logger.S.Infof("无效的文件id:%v", err)
		return
	}
	userID, _ := c.Get("userID")
	err = h.fileService.DeleteFile(c.Request.Context(), userID.(uint), uint(ID))
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			response.BusinessError(c, response.CodeFileNotFound, "文件不存在")
			return
		}
		response.ServerError(c, "删除失败")
		logger.S.Errorf("文件删除失败：%v", err)
		return
	}
	response.Success(c, nil)
}

// BatchDelete 批量删除文件
func (h *FileHandler) BatchDelete(c *gin.Context) {
	type BatchDeleteReq struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	var req BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}

	userID, _ := c.Get("userID")
	err := h.fileService.BatchDeleteFile(c.Request.Context(), userID.(uint), req.IDs)
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			response.BusinessError(c, response.CodeFileNotFound, "文件不存在")
			return
		}
		response.ServerError(c, "批量删除失败")
		logger.S.Errorf("批量删除文件失败：%v", err)
		return
	}

	response.Success(c, gin.H{"count": len(req.IDs)})
}

// GetUserFile 获取用户文件列表
func (h *FileHandler) GetUserFile(c *gin.Context) {
	parentIDString := c.Query("parent_id")
	parentID, err := strconv.ParseUint(parentIDString, 10, 64)
	if err != nil {
		response.BusinessError(c, response.CodeInvalidParentID, "parent_id无效")
		return
	}
	userID, _ := c.Get("userID")
	list, err := h.fileService.GetUserFile(c.Request.Context(), userID.(uint), uint(parentID))
	if err != nil {
		response.ServerError(c, "获取用户文件列表失败")
		logger.S.Errorf("获取用户文件列表失败：%v", err)
		return
	}
	response.Success(c, list)
}

// ListRecycleBin 回收站列表
func (h *FileHandler) ListRecycleBin(c *gin.Context) {
	userID, _ := c.Get("userID")
	list, err := h.fileService.ListRecycleBin(c.Request.Context(), userID.(uint))
	if err != nil {
		response.ServerError(c, "获取回收站列表失败")
		logger.S.Errorf("获取回收站列表失败：%v", err)
		return
	}
	response.Success(c, list)
}

// Restore 从回收站恢复文件
func (h *FileHandler) Restore(c *gin.Context) {
	IDString := c.Param("ID")
	ID, err := strconv.ParseUint(IDString, 10, 64)
	if err != nil {
		response.BusinessError(c, response.CodeInvalidFileID, "无效的文件ID")
		logger.S.Infof("无效的文件id:%v", err)
		return
	}
	userID, _ := c.Get("userID")
	err = h.fileService.RestoreUserFile(c.Request.Context(), userID.(uint), uint(ID))
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			response.BusinessError(c, response.CodeFileNotFound, "文件不存在")
			return
		}
		response.ServerError(c, "恢复失败")
		logger.S.Errorf("恢复文件失败：%v", err)
		return
	}
	response.Success(c, nil)
}

// PermanentlyDelete 永久删除回收站中的文件/文件夹
func (h *FileHandler) PermanentlyDelete(c *gin.Context) {
	IDString := c.Param("ID")
	ID, err := strconv.ParseUint(IDString, 10, 64)
	if err != nil {
		response.BusinessError(c, response.CodeInvalidFileID, "无效的文件ID")
		logger.S.Infof("无效的文件id:%v", err)
		return
	}

	userID, _ := c.Get("userID")
	err = h.fileService.PermanentlyDeleteFile(c.Request.Context(), userID.(uint), uint(ID))
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			response.BusinessError(c, response.CodeFileNotFound, "文件不存在")
			return
		}
		response.ServerError(c, "永久删除失败")
		logger.S.Errorf("永久删除文件失败：%v", err)
		return
	}

	response.Success(c, nil)
}

// GetDownloadURL 获取下载URL
func (h *FileHandler) GetDownloadURL(c *gin.Context) {
	IDString := c.Param("ID")
	userID, _ := c.Get("userID")
	ID, err := strconv.ParseUint(IDString, 10, 64)
	if err != nil {
		response.BusinessError(c, response.CodeInvalidFileID, "无效的文件ID")
		logger.S.Infof("无效的文件id:%v", err)
		return
	}
	downloadFileResp, err := h.fileService.GetDownloadURL(c.Request.Context(), userID.(uint), uint(ID))
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			response.BusinessError(c, response.CodeFileNotFound, "文件不存在")
			return
		}
		response.ServerError(c, "获取下载URL失败")
		logger.S.Errorf("获取下载URL失败：%v", err)
		return
	}
	response.Success(c, downloadFileResp)
}
