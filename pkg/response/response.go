package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeOK              = 0
	CodeInvalidParam    = 1001
	CodeMissingFile     = 1002
	CodeInvalidParentID = 1003
	CodeInvalidFileID   = 1004
	CodeUnauthorized    = 1005
	CodeForbidden       = 1006

	CodeUserExisted         = 2001
	CodeAccountOrPassword   = 2002
	CodeRefreshTokenMissing = 2003

	CodeUserSpaceNotEnough      = 3001
	CodeFileNotFound            = 3002
	CodeInstantUnavailable      = 3003
	CodeUploadSessionInvalid    = 3004
	CodeChunkUploadIncomplete   = 3005
	CodeChunkFileHashMismatch   = 3006
	CodeUploadMethodInvalid     = 3007
	CodeChunkAlreadyUploaded    = 3008
	CodeChunkHashConflict       = 3009
	CodeInstantProofRequired    = 3010
	CodeInstantProofInvalid     = 3011
	CodeInstantPrepareInvalid   = 3012
	CodeInstantProofMismatch    = 3013
	CodeRateLimited             = 3014
	CodeTooManyPendingUploads   = 3015
	CodeUploadSessionProcessing = 3016

	CodeShareNotFound   = 4001
	CodeShareExpired    = 4002
	CodeShareInvalidKey = 4003

	CodeInternal = 5000
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code: CodeOK,
		Msg:  "success",
		Data: data,
	})
}

func BusinessError(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusBadRequest, Body{
		Code: code,
		Msg:  msg,
	})
}

func RateLimited(c *gin.Context, msg string) {
	c.JSON(http.StatusTooManyRequests, Body{
		Code: CodeRateLimited,
		Msg:  msg,
	})
}

func ServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Body{
		Code: CodeInternal,
		Msg:  msg,
	})
}
