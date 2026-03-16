package middleware

import (
	"WeDrive/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}
		uid, ok := v.(uint)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息无效"})
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
			c.JSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}
