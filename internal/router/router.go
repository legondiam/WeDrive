package router

import (
	"WeDrive/internal/api"
	"WeDrive/internal/middleware"

	"time"

	"github.com/gin-gonic/gin"
)

func NewRouter(userHandler *api.UserHandler, fileHandler *api.FileHandler, shareHandler *api.ShareHandler) *gin.Engine {
	r := gin.Default()
	publicGroup := r.Group("/api/v1")
	publicGroup.Use(middleware.TimeoutMiddleware(3 * time.Second))
	{
		publicGroup.POST("/user/register", userHandler.Register)
		publicGroup.POST("/user/login", userHandler.Login)
		publicGroup.POST("/user/refresh", userHandler.Refresh)

		publicGroup.POST("/share/download", shareHandler.GetShareDownloadURL)

	}
	privateGroup := publicGroup.Group("/")
	privateGroup.Use(middleware.AuthMiddleware())
	{

		privateGroup.POST("/file/upload-folder", fileHandler.CreateFolder)
		privateGroup.POST("/file/quick-check", fileHandler.QuickCheck)
		privateGroup.POST("/file/instant-upload", fileHandler.InstantUpload)
		privateGroup.POST("/file/upload/init", fileHandler.InitChunkUpload)
		privateGroup.POST("/file/upload/sign-part", fileHandler.SignPartUpload)
		privateGroup.POST("/file/upload/report-part", fileHandler.ReportUploadedPart)
		privateGroup.POST("/file/upload/complete", fileHandler.CompleteChunkUpload)
		privateGroup.GET("/file/list", fileHandler.GetUserFile)
		privateGroup.DELETE("/file/delete/:ID", fileHandler.Delete)
		privateGroup.POST("/file/batch-delete", fileHandler.BatchDelete)
		privateGroup.DELETE("/file/permanent-delete/:ID", fileHandler.PermanentlyDelete)
		privateGroup.GET("/file/recycle", fileHandler.ListRecycleBin)
		privateGroup.POST("/file/restore/:ID", fileHandler.Restore)

		privateGroup.GET("/user/info", userHandler.GetUserInfo)
		privateGroup.GET("/file/download/:ID", fileHandler.GetDownloadURL)

		privateGroup.POST("/share/create", shareHandler.CreateShareFile)

	}

	timeoutGroup := r.Group("/api/v1")
	timeoutGroup.Use(middleware.TimeoutMiddleware(120 * time.Second))
	timeoutGroup.Use(middleware.AuthMiddleware())
	{
		timeoutGroup.POST("/file/upload", fileHandler.Upload)
	}

	adminGroup := privateGroup.Group("/admin")
	adminGroup.Use(middleware.AdminMiddleware())
	{
		adminGroup.POST("/user/update-member", userHandler.UpdateUserMember)
	}
	return r
}
