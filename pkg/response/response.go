package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// 通用
	CodeOK              = 0
	CodeInvalidParam    = 1001 // JSON/form/path/query 参数不合法
	CodeMissingFile     = 1002 // 缺少上传文件
	CodeInvalidParentID = 1003 // parent_id 无效
	CodeInvalidFileID   = 1004 // 文件ID无效

	// 用户模块
	CodeUserExisted         = 2001 // 用户已存在
	CodeAccountOrPassword   = 2002 // 用户名或密码错误
	CodeRefreshTokenMissing = 2003 // refreshToken不存在/无效

	// 文件模块
	CodeUserSpaceNotEnough = 3001 // 用户空间不足
	CodeFileNotFound       = 3002 // 文件不存在

	// 分享模块
	CodeShareNotFound   = 4001 // 分享不存在
	CodeShareExpired    = 4002 // 分享已过期
	CodeShareInvalidKey = 4003 // 分享密钥不正确

	// 服务端
	CodeInternal = 5000 // 未分类服务端错误
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` //为空时不返回
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code: CodeOK,
		Msg:  "success",
		Data: data,
	})
}

// BusinessError 业务错误响应
func BusinessError(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusBadRequest, Body{
		Code: code,
		Msg:  msg,
	})
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Body{
		Code: CodeInternal,
		Msg:  msg,
	})
}
