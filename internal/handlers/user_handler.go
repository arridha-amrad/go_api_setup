package handlers

import (
	"database/sql"
	"errors"
	"my-go-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct{ service services.IUserService }

func NewUserHandler(service services.IUserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUserById(c *gin.Context) {
	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user id"})
		return
	}
	user, err := h.service.GetUserById(c.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.service.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"errors": "Something went wrong"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *UserHandler) Update(c *gin.Context) {
	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user id"})
		return
	}
	value, exist := c.Get("validatedBody")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error"})
		return
	}
	existingUser, err := h.service.GetUserById(c.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
			return
		}
	}
	if v, ok := value.(map[string]any); ok {
		if username, exists := v["username"].(string); exists {
			existingUser.Username = username
		}
		if name, exists := v["name"].(string); exists {
			existingUser.Name = name
		}
		if email, exists := v["email"].(string); exists {
			existingUser.Email = email
		}
		if password, exists := v["password"].(string); exists {
			existingUser.Password = password
		}
		if role, exists := v["role"].(string); exists {
			existingUser.Role = role
		}
	}
	h.service.UpdateUser(c.Request.Context(), existingUser)
	c.JSON(http.StatusOK, gin.H{"user": existingUser})
}
