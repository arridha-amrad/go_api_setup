package handlers

import (
	"my-go-api/internal/dto"
	"my-go-api/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	service     *services.AuthService
	userService *services.UserService
}

func NewAuthHandler(service *services.AuthService, userService *services.UserService) *AuthHandler {
	return &AuthHandler{service: service, userService: userService}
}

func (h AuthHandler) GetAuth(c *gin.Context) {
	value, exist := c.Get("authenticatedUserId")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validated body not exists"})
		return
	}
	userId, err := uuid.Parse(value.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.userService.GetUserById(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h AuthHandler) Login(c *gin.Context) {
	value, exist := c.Get("validatedBody")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validated body not exists"})
		return
	}

	body, ok := value.(dto.Login)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid type for validated body"})
		return
	}

	existingUser, err := h.service.GetUserByIdentity(c.Request.Context(), body.Identity)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if isMatch := h.service.VerifyPassword(existingUser.Password, body.Password); !isMatch {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	tokenAcc, err := h.service.GenerateToken(existingUser.ID.String(), "access")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tokenRef, err := h.service.GenerateToken(existingUser.ID.String(), "refresh")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("refresh-token", "Bearer "+tokenRef, 3600*24*365, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"user":  existingUser,
		"token": "Bearer " + tokenAcc,
	})
}
