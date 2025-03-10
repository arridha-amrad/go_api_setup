package validation

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

func Init() *validator.Validate {
	validate := validator.New()

	// Register custom validations
	err := validate.RegisterValidation("strongPassword", ValidatePassword)
	if err != nil {
		panic(err) // Handle error during initialization
	}
	return validate
}

var Messages = map[string]string{
	"email":          "Invalid email",
	"min":            "Too short. A minimum of %s characters is required",
	"required":       "This field is required",
	"strongPassword": "A minimum of 5 characters including an uppercase letter, a lowercase letter, and a number is required",
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
