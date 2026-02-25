package api

import (
	"WeDrive/internal/repository"
	"WeDrive/internal/service"
	"WeDrive/pkg/logger"
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

func (h *FileHandler) Upload(c *gin.Context) {
	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "请上传文件"})
		return
	}
	// 获取父文件夹ID
	parentIDString := c.PostForm("parent_id")
	var parentID int64
	if parentIDString != "" {
		parentID, err = strconv.ParseInt(parentIDString, 10, 64)
		if err != nil {
			c.JSON(400, gin.H{"error": "parent_id无效"})
			return
		}
	}
	// 获取用户ID
	userID, _ := c.Get("userID")
	// 上传文件
	err = h.fileService.UploadFile(c.Request.Context(), fileHeader, userID.(uint), parentID)
	if err != nil {
		c.JSON(500, gin.H{"error": "上传失败"})
		logger.S.Errorf("文件上传失败：%v", err)
		return
	}
	c.JSON(200, gin.H{"message": "上传成功"})
	logger.S.Info("文件上传成功")
}

// CreateFolder 创建文件夹
func (h *FileHandler) CreateFolder(c *gin.Context) {
	type CreateFolderReq struct {
		Name     string `json:"name" binding:"required"`
		ParentID int64  `json:"parent_id"`
	}
	var req CreateFolderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数无效"})
		return
	}
	// 默认父目录为根目录
	parentID := req.ParentID

	userID, _ := c.Get("userID")
	if err := h.fileService.CreateFolder(c.Request.Context(), userID.(uint), parentID, req.Name); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logger.S.Errorf("创建文件夹失败：%v", err)
		return
	}

	c.JSON(200, gin.H{"message": "创建成功"})
}

// Delete 删除文件
func (h *FileHandler) Delete(c *gin.Context) {
	IDString := c.Param("ID")
	ID, err := strconv.ParseInt(IDString, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		logger.S.Infof("无效的文件id:%v", err)
		return
	}
	userID, _ := c.Get("userID")
	err = h.fileService.DeleteFile(c.Request.Context(), userID.(uint), uint(ID))
	if err != nil {
		if errors.Is(repository.ErrFileNotFound, err) {
			c.JSON(400, gin.H{"error": "文件不存在"})
			return
		}
		c.JSON(500, gin.H{"error": "删除失败"})
		logger.S.Errorf("文件删除失败：%v", err)
		return
	}
	c.JSON(200, gin.H{"message": "删除成功"})
}

// GetUserFile 获取用户文件列表
func (h *FileHandler) GetUserFile(c *gin.Context) {
	parentIDString := c.Query("parent_id")
	parentID, err := strconv.ParseInt(parentIDString, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "parent_id无效"})
		return
	}
	userID, _ := c.Get("userID")
	list, err := h.fileService.GetUserFile(c.Request.Context(), userID.(uint), parentID)
	if err != nil {
		c.JSON(500, gin.H{"error": "获取用户文件列表失败"})
		logger.S.Errorf("获取用户文件列表失败：%v", err)
		return
	}
	c.JSON(200, gin.H{"data": list})
}

// ListRecycleBin 回收站列表
func (h *FileHandler) ListRecycleBin(c *gin.Context) {
	userID, _ := c.Get("userID")
	list, err := h.fileService.ListRecycleBin(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"error": "获取回收站列表失败"})
		logger.S.Errorf("获取回收站列表失败：%v", err)
		return
	}
	c.JSON(200, gin.H{"data": list})
}

// Restore 从回收站恢复文件
func (h *FileHandler) Restore(c *gin.Context) {
	IDString := c.Param("ID")
	ID, err := strconv.ParseInt(IDString, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		logger.S.Infof("无效的文件id:%v", err)
		return
	}
	userID, _ := c.Get("userID")
	err = h.fileService.RestoreUserFile(c.Request.Context(), userID.(uint), uint(ID))
	if err != nil {
		if errors.Is(repository.ErrFileNotFound, err) {
			c.JSON(400, gin.H{"error": "文件不存在"})
			return
		}
		c.JSON(500, gin.H{"error": "恢复失败"})
		logger.S.Errorf("恢复文件失败：%v", err)
		return
	}
	c.JSON(200, gin.H{"message": "恢复成功"})
}
