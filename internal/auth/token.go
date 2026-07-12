package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authHeader, ok := headers["Authorization"]
	if !ok {
		return "", errors.New("no auth header")
	}

	parts := strings.Split(authHeader[0], " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid auth header")
	}

	return parts[1], nil
}

func MakeRefreshToken() string {
	//generate 32 bytes random string
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("Failed to generate refresh token: ", err)
		return ""
	}
	return hex.EncodeToString(b)
}