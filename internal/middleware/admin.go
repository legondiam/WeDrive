package middleware

import (
	"WeDrive/internal/config"
	"WeDrive/pkg/response"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("userID")
		if !ok {
			response.BusinessError(c, response.CodeUnauthorized, "未登录")
			c.Abort()
			return
		}
		uid, ok := v.(uint)
		if !ok {
			response.BusinessError(c, response.CodeUnauthorized, "用户信息无效")
			c.Abort()
			return
		}

		isAdmin := false
		for _, id := range config.GlobalConf.Admin.UserIDs {
			if id == uid {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			response.BusinessError(c, response.CodeForbidden, "无管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
