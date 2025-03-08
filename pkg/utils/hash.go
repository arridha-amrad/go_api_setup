package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateRandomBytes(size int) (string, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HashWithSHA256(randomStr string) string {
	hash := sha256.Sum256([]byte(randomStr))
	return hex.EncodeToString(hash[:])
}
