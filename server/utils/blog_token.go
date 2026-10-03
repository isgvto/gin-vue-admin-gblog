package utils

import (
	"errors"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
)

const (
	BlogAdminPrefix  = "admin:"
	BlogAccessPrefix = "blog:"
)

type BlogTokenClaims struct {
	Subject string `json:"subject"`
	jwt.RegisteredClaims
}

func CreateBlogToken(subject string, expiresIn time.Duration) (string, error) {
	claims := BlogTokenClaims{
		Subject: subject,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    global.GVA_CONFIG.JWT.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(global.GVA_CONFIG.JWT.SigningKey))
}

func ParseBlogToken(raw string) (*BlogTokenClaims, error) {
	tokenStr := NormalizeAuthToken(raw)
	if tokenStr == "" {
		return nil, errors.New("empty token")
	}
	token, err := jwt.ParseWithClaims(tokenStr, &BlogTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.GVA_CONFIG.JWT.SigningKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*BlogTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func NormalizeAuthToken(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return strings.TrimSpace(raw[7:])
	}
	return raw
}
