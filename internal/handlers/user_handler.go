package handlers

import (
	"my-go-api/internal/dto"
	"my-go-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	service  *services.UserService
	validate *validator.Validate
}

func NewUserHandler(
	service *services.UserService,
	validate *validator.Validate,
) *UserHandler {
	return &UserHandler{service: service, validate: validate}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.service.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"errors": "Something went wrong"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *UserHandler) Create(c *gin.Context) {
	value, exist := c.Get("validatedBody")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error"})
		return
	}

	body, ok := value.(dto.CreateUser)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid type for validatedBody"})
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}
