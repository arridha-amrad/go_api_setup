package validation

import (
	"my-go-api/pkg/utils"

	"github.com/go-playground/validator/v10"
)

func Init() *validator.Validate {
	validate := validator.New()

	// Register custom validations
	err := validate.RegisterValidation("strongPassword", utils.PasswordValidator)
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
