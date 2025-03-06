package dto

import (
	"errors"
	"fmt"
	"my-go-api/internal/validation"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserValidationMiddleware struct {
	validate *validator.Validate
}

func RegisterUserValidationMiddleware(validate *validator.Validate) *UserValidationMiddleware {
	return &UserValidationMiddleware{validate: validate}
}

type CreateUser struct {
	Name     string `json:"name" validate:"required,min=5"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=5"`
	Password string `json:"password" validate:"required,strongPassword"`
}

func (m *UserValidationMiddleware) CreateUser(c *gin.Context) {
	var input CreateUser
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		c.Abort()
		return
	}
	if err := m.validate.Struct(input); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		var msgErrors = make(map[string]string)
		if errors.As(err, &validationErrors) {
			for _, e := range validationErrors {
				message := validation.Messages[e.Tag()]
				if e.Param() != "" {
					msgErrors[strings.ToLower(e.Field())] = fmt.Sprintf(message, e.Param())
				} else {
					msgErrors[strings.ToLower(e.Field())] = message
				}
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": msgErrors})
		c.Abort()
		return
	}
	c.Set("validatedBody", input)
	c.Next()
}
