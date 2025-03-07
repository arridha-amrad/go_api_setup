package middleware

import (
	"errors"
	"fmt"
	"log"
	"my-go-api/internal/dto"
	"my-go-api/internal/validation"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type middleware struct {
	validate *validator.Validate
}

func RegisterValidationMiddleware(validate *validator.Validate) *middleware {
	return &middleware{validate: validate}
}

func (m *middleware) runValidation(c *gin.Context, input any) {
	log.Println(input)
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		c.Abort()
		return
	}
	if err := m.validate.Struct(input); err != nil {
		var validationErrors validator.ValidationErrors
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
}

func (m *middleware) Login(c *gin.Context) {
	var input dto.Login
	m.runValidation(c, &input)
	c.Set("validatedBody", input)
	c.Next()
}

func (m *middleware) CreateUser(c *gin.Context) {
	var input dto.CreateUser
	m.runValidation(c, &input)
	c.Set("validatedBody", input)
	c.Next()
}
