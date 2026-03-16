package router

import (
	"WeDrive/internal/api"
	"WeDrive/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(userHandler *api.UserHandler, fileHandler *api.FileHandler, shareHandler *api.ShareHandler) *gin.Engine {
	r := gin.Default()
	publicGroup := r.Group("/api/v1")
	{
		publicGroup.POST("/user/register", userHandler.Register)
		publicGroup.POST("/user/login", userHandler.Login)
		publicGroup.POST("/user/refresh", userHandler.Refresh)

		publicGroup.GET("/share/download", shareHandler.GetShareDownloadURL)

	}
	privateGroup := publicGroup.Group("/")
	privateGroup.Use(middleware.AuthMiddleware())
	{
		privateGroup.POST("/file/upload", fileHandler.Upload)
		privateGroup.POST("/file/upload-folder", fileHandler.CreateFolder)
		privateGroup.GET("/file/list", fileHandler.GetUserFile)
		privateGroup.DELETE("/file/delete/:ID", fileHandler.Delete)
		privateGroup.DELETE("/file/permanent-delete/:ID", fileHandler.PermanentlyDelete)
		privateGroup.GET("/file/recycle", fileHandler.ListRecycleBin)
		privateGroup.POST("/file/restore/:ID", fileHandler.Restore)

		privateGroup.GET("/user/info", userHandler.GetUserInfo)
		privateGroup.GET("/file/download/:ID", fileHandler.GetDownloadURL)

		privateGroup.POST("/share/create", shareHandler.CreateShareFile)

	}

	adminGroup := privateGroup.Group("/admin")
	adminGroup.Use(middleware.AdminMiddleware())
	{
		adminGroup.POST("/user/update-member", userHandler.UpdateUserMember)
	}
	return r
}
