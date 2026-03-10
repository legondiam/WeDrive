package jwts

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateToken(n int) (string, error) {
	token := make([]byte, n)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token), nil

}
