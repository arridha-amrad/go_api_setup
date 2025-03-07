package utils

import (
	"fmt"
	"unicode"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	// Generate a hashed version of the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashedPassword, password string) error {
	// Compare the hashed password with the plaintext password
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func ValidatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 5 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}

		// If all conditions are met, no need to continue looping
		if hasUpper && hasLower && hasDigit {
			return true
		}
	}

	return false
}
