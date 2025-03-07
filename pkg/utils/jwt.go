// pkg/utils/jwt_utils.go
package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Define a custom type for TokenType
type TokenType string

// Declare valid values (like a TypeScript union)
const (
	LinkToken    TokenType = "link"
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

var secretKey string

func SetTokenSecretKey(key string) {
	secretKey = key
}

func GenerateToken(userId string, tokenId string, tokenType TokenType) (string, error) {
	claims := jwt.MapClaims{
		"tokenId": tokenId,
		"type":    string(tokenType),
		"user_id": userId,
		"exp":     time.Now().Add(24 * time.Hour).UnixMilli(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}

	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
