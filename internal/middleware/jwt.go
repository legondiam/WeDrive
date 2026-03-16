package middleware

import (
	"WeDrive/pkg/utils/jwts"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//获取请求头的authorization字段
		authorization := c.GetHeader("Authorization")

		if authorization == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少Authorization",
			})
			c.Abort()
			return
		}
		//解析authorization字段
		parts := strings.SplitN(authorization, " ", 2)
		//authorization格式不正确
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization格式不正确",
			})
			c.Abort()
			return
		}
		tokenString := parts[1]

		//验证token
		claims, err := jwts.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "token已过期",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token无效",
			})
			c.Abort()
			return
		}

		//将claims中的信息存入context中
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		//继续处理请求
		c.Next()
	}
}
