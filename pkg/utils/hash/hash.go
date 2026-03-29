package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"io"
	"mime/multipart"
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(hash), nil
}

// CheckPassword 检查密码是否正确
func CheckPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) { //密码不匹配
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return true, nil
}

// HashFile 计算文件的sha256值
func HashFile(fileHeader *multipart.FileHeader) (string, error) {
	stream, err := fileHeader.Open()
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer stream.Close()
	hash := sha256.New()
	if _, err = io.Copy(hash, stream); err != nil {
		return "", errors.WithStack(err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
