package jwts

import (
	"WeDrive/internal/config"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type CustomClaims struct {
	UserID               uint   `json:"userid"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // exp, iat 等标准字段
}

// GenerateAccessToken 生成token
func GenerateAccessToken(userID uint, username string) (string, error) {
	//获取jwt配置
	jwtconfig := config.GlobalConf.Jwt
	//设置过期时间
	//fmt.Printf("从配置读取到的过期时间: %v\n", jwtconfig.AccessTokenExpiration)
	expirationTime := time.Now().Add(jwtconfig.AccessTokenExpiration)
	claims := &CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	//生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//签名完整jwt令牌
	tokenString, err := token.SignedString([]byte(jwtconfig.SecretKey))
	if err != nil {
		return "", errors.WithStack(err)
	}
	return tokenString, nil
}

// GenerateRefreshToken 生成refreshToken
func GenerateRefreshToken(userID uint, username string) (string, string, error) {
	//获取jwt配置
	jwtconfig := config.GlobalConf.Jwt
	//设置过期时间
	expirationTime := time.Now().Add(jwtconfig.RefreshTokenExpiration)
	//生成唯一id
	tokenid := uuid.New().String()
	//声明refreshToken
	claims := &CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenid,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtconfig.SecretKey))
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	return tokenString, tokenid, nil
}

// ValidateToken 验证token
func ValidateToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}
	//解析token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		//检查签名算法
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.GlobalConf.Jwt.SecretKey), nil
	})
	//解析失败
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// ExtractClaims 从token中提取claims
func ExtractClaims(token *jwt.Token) (*CustomClaims, error) {
	claims, ok := token.Claims.(*CustomClaims)
	if ok && token.Valid {
		return claims, nil
	} else {
		return nil, jwt.ErrSignatureInvalid
	}
}
