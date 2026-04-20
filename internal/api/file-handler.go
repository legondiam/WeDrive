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
	HashType string `json:"hash_type" binding:"required"`
	FileHash string `json:"file_hash" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
	FileSize int64  `json:"file_size"`
	ParentID uint   `json:"parent_id"`
}

type quickCheckReq struct {
	FileSize int64  `json:"file_size" binding:"required"`
	HeadHash string `json:"head_hash" binding:"required"`
	MidHash  string `json:"mid_hash" binding:"required"`
	TailHash string `json:"tail_hash" binding:"required"`
}

type initChunkUploadReq struct {
	HashType   string `json:"hash_type" binding:"required"`
	FileHash   string `json:"file_hash" binding:"required"`
	FileName   string `json:"file_name" binding:"required"`
	FileSize   int64  `json:"file_size" binding:"required"`
	ParentID   uint   `json:"parent_id"`
	ChunkSize  int64  `json:"chunk_size" binding:"required"`
	ChunkCount int    `json:"chunk_count" binding:"required"`
	HeadHash   string `json:"head_hash" binding:"required"`
	MidHash    string `json:"mid_hash" binding:"required"`
	TailHash   string `json:"tail_hash" binding:"required"`
}

type signPartReq struct {
	UploadID   uint   `json:"upload_id" binding:"required"`
	PartNumber int    `json:"part_number" binding:"required"`
	ChunkHash  string `json:"chunk_hash" binding:"required"`
}

type reportUploadedPartReq struct {
	UploadID   uint   `json:"upload_id" binding:"required"`
	PartNumber int    `json:"part_number" binding:"required"`
	ETag       string `json:"etag" binding:"required"`
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
		if errors.Is(err, service.ErrUploadMethodInvalid) {
			response.BusinessError(c, response.CodeUploadMethodInvalid, "文件过大，请使用分块上传")
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

// QuickCheck 抽样哈希快速筛选秒传候选
func (h *FileHandler) QuickCheck(c *gin.Context) {
	var req quickCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	userID, _ := c.Get("userID")
	matched, err := h.fileService.QuickCheck(c.Request.Context(), userID.(uint), service.QuickCheckReq{
		FileSize: req.FileSize,
		HeadHash: req.HeadHash,
		MidHash:  req.MidHash,
		TailHash: req.TailHash,
	})
	if err != nil {
		if errors.Is(err, service.ErrUserSpaceNotEnough) {
			response.BusinessError(c, response.CodeUserSpaceNotEnough, "用户空间不足")
			return
		}
		response.ServerError(c, "快速校验失败")
		logger.S.Errorf("快速校验失败：%v", err)
		return
	}
	response.Success(c, matched)
}

// InstantUpload 秒传
func (h *FileHandler) InstantUpload(c *gin.Context) {
	var req instantUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	userID, _ := c.Get("userID")
	uploadedID, err := h.fileService.InstantUpload(c.Request.Context(), userID.(uint), req.ParentID, req.HashType, req.FileName, req.FileHash, req.FileSize)
	if err != nil {
		if errors.Is(err, service.ErrParentFolderInvalid) {
			response.BusinessError(c, response.CodeInvalidParentID, "parent_id不合法")
			return
		}
		if errors.Is(err, service.ErrUnsupportedHashType) {
			response.BusinessError(c, response.CodeInvalidParam, "hash_type不支持")
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

// InitChunkUpload 初始化分块上传
func (h *FileHandler) InitChunkUpload(c *gin.Context) {
	var req initChunkUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	userID, _ := c.Get("userID")
	resp, err := h.fileService.InitChunkUpload(c.Request.Context(), userID.(uint), service.ChunkUploadInitReq{
		HashType:   req.HashType,
		FileHash:   req.FileHash,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		ParentID:   req.ParentID,
		ChunkSize:  req.ChunkSize,
		ChunkCount: req.ChunkCount,
		HeadHash:   req.HeadHash,
		MidHash:    req.MidHash,
		TailHash:   req.TailHash,
	})
	if err != nil {
		if errors.Is(err, service.ErrUploadRequestInvalid) {
			response.BusinessError(c, response.CodeInvalidParam, "上传请求无效")
			return
		}
		if errors.Is(err, service.ErrUploadMethodInvalid) {
			response.BusinessError(c, response.CodeUploadMethodInvalid, "小文件请使用普通上传")
			return
		}
		if errors.Is(err, service.ErrParentFolderInvalid) {
			response.BusinessError(c, response.CodeInvalidParentID, "parent_id不合法")
			return
		}
		if errors.Is(err, service.ErrUnsupportedHashType) {
			response.BusinessError(c, response.CodeInvalidParam, "hash_type不支持")
			return
		}
		if errors.Is(err, service.ErrUserSpaceNotEnough) {
			response.BusinessError(c, response.CodeUserSpaceNotEnough, "用户空间不足")
			return
		}
		response.ServerError(c, "初始化分块上传失败")
		logger.S.Errorf("初始化分块上传失败：%v", err)
		return
	}
	response.Success(c, resp)
}

// SignPartUpload 为分块上传签名
func (h *FileHandler) SignPartUpload(c *gin.Context) {
	var req signPartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	userID, _ := c.Get("userID")
	resp, err := h.fileService.SignPartUpload(c.Request.Context(), userID.(uint), service.SignPartReq{
		UploadID:   req.UploadID,
		PartNumber: req.PartNumber,
		ChunkHash:  req.ChunkHash,
	})
	if err != nil {
		if errors.Is(err, service.ErrUploadRequestInvalid) {
			response.BusinessError(c, response.CodeInvalidParam, "上传请求无效")
			return
		}
		if errors.Is(err, service.ErrChunkAlreadyUploaded) {
			response.BusinessError(c, response.CodeChunkAlreadyUploaded, "分块已上传完成")
			return
		}
		if errors.Is(err, service.ErrChunkHashConflict) {
			response.BusinessError(c, response.CodeChunkHashConflict, "分块哈希与历史记录冲突，请重新初始化上传")
			return
		}
		if errors.Is(err, service.ErrUploadSessionInvalid) {
			response.BusinessError(c, response.CodeUploadSessionInvalid, "上传会话无效")
			return
		}
		response.ServerError(c, "生成分块上传地址失败")
		logger.S.Errorf("生成分块上传地址失败：%v", err)
		return
	}
	response.Success(c, resp)
}

// ReportUploadedPart 回报已上传分块
func (h *FileHandler) ReportUploadedPart(c *gin.Context) {
	var req reportUploadedPartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	userID, _ := c.Get("userID")
	if err := h.fileService.ReportUploadedPart(c.Request.Context(), userID.(uint), service.ReportUploadedPartReq{
		UploadID:   req.UploadID,
		PartNumber: req.PartNumber,
		ETag:       req.ETag,
	}); err != nil {
		if errors.Is(err, service.ErrUploadRequestInvalid) {
			response.BusinessError(c, response.CodeInvalidParam, "上传请求无效")
			return
		}
		if errors.Is(err, service.ErrUploadSessionInvalid) {
			response.BusinessError(c, response.CodeUploadSessionInvalid, "上传会话无效")
			return
		}
		response.ServerError(c, "保存分块上传状态失败")
		logger.S.Errorf("保存分块上传状态失败：%v", err)
		return
	}
	response.Success(c, nil)
}

// CompleteChunkUpload 完成分块上传
func (h *FileHandler) CompleteChunkUpload(c *gin.Context) {
	type completeChunkUploadReq struct {
		UploadID uint `json:"upload_id" binding:"required"`
	}
	var req completeChunkUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, response.CodeInvalidParam, "参数无效")
		return
	}
	userID, _ := c.Get("userID")
	uploadedID, err := h.fileService.CompleteChunkUpload(c.Request.Context(), userID.(uint), req.UploadID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUploadRequestInvalid):
			response.BusinessError(c, response.CodeInvalidParam, "上传请求无效")
		case errors.Is(err, service.ErrUserSpaceNotEnough):
			response.BusinessError(c, response.CodeUserSpaceNotEnough, "用户空间不足")
		case errors.Is(err, service.ErrUploadSessionInvalid):
			response.BusinessError(c, response.CodeUploadSessionInvalid, "上传会话无效")
		case errors.Is(err, service.ErrChunkUploadIncomplete):
			response.BusinessError(c, response.CodeChunkUploadIncomplete, "仍有分块未上传完成")
		case errors.Is(err, service.ErrChunkFileHashMismatch):
			response.BusinessError(c, response.CodeChunkFileHashMismatch, "文件校验失败，请重新上传")
		case errors.Is(err, service.ErrParentFolderInvalid):
			response.BusinessError(c, response.CodeInvalidParentID, "parent_id不合法")
		default:
			response.ServerError(c, "完成分块上传失败")
			logger.S.Errorf("完成分块上传失败：%v", err)
		}
		return
	}
	response.Success(c, gin.H{"id": uploadedID})
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
