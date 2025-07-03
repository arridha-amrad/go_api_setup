package handlers

import (
	"errors"
	"fmt"
	"log"
	"my-go-api/internal/constants"
	"my-go-api/internal/dto"
	"my-go-api/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IAuthHandler interface {
	Register(c *gin.Context)
	Logout(c *gin.Context)
	RefreshToken(c *gin.Context)
	GetAuth(c *gin.Context)
	Login(c *gin.Context)
}

type authHandler struct {
	as services.IAuthService
	us services.IUserService
}

type Cookie struct {
	token    string
	userId   uuid.UUID
	deviceId uuid.UUID
}

func NewAuthHandler(service services.IAuthService, us services.IUserService) IAuthHandler {
	return &authHandler{as: service, us: us}
}

func getCookies(c *gin.Context) (*Cookie, error) {
	cookieRefToken, err := c.Cookie(constants.COOKIE_REFRESH_TOKEN)
	if err != nil {
		return nil, errors.New("refresh token cookie is not exists")
	}
	cookieDeviceId, err := c.Cookie(constants.COOKIE_DEVICE_ID)
	if err != nil {
		return nil, errors.New("device id token cookie is not exists")
	}
	cookieUserId, err := c.Cookie(constants.COOKIE_USER_ID)
	if err != nil {
		return nil, errors.New("user id token cookie is not exists")
	}
	userId, err := uuid.Parse(cookieUserId)
	if err != nil {
		log.Println("failed to parse to uuid")
		return nil, errors.New("failed to parse to uuid")
	}
	deviceId, err := uuid.Parse(cookieDeviceId)
	if err != nil {
		return nil, errors.New("failed to parse to uuid")
	}
	return &Cookie{
		token:    cookieRefToken,
		userId:   userId,
		deviceId: deviceId,
	}, nil
}

func (h *authHandler) Register(c *gin.Context) {
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
	user, err := h.as.CreateUser(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"errors": err.Error()})
		return
	}
	token, err := h.as.GenerateToken(user.ID)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Something went wrong"})
		return
	}
	if err := h.as.SendVerificationEmail(user.Name, user.Email, token); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Something went wrong"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("An email has been sent to %s. Please follow the instruction to verify your account.", user.Email)},
	)
}

func (h *authHandler) Logout(c *gin.Context) {
	cookies, err := getCookies(c)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	err = h.as.DeleteRefreshToken(c.Request.Context(), cookies.userId, cookies.deviceId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	c.SetCookie(constants.COOKIE_REFRESH_TOKEN, "", -1, "/", "", false, false)
	c.SetCookie(constants.COOKIE_DEVICE_ID, "", -1, "/", "", false, false)
	c.SetCookie(constants.COOKIE_USER_ID, "", -1, "/", "", false, false)
	c.JSON(http.StatusOK, gin.H{"message": "Logout"})
}

func (h *authHandler) RefreshToken(c *gin.Context) {
	cookies, err := getCookies(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err := h.as.VerifyRefreshToken(c.Request.Context(), cookies.userId, cookies.deviceId, cookies.token); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tokenAcc, err := h.as.GenerateToken(cookies.userId)
	if err != nil {
		log.Println("failed to generate a token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	if err := h.as.DeleteRefreshToken(c.Request.Context(), cookies.userId, cookies.deviceId); err != nil {
		log.Println("failed to delete the old token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	newRefreshToken, hashToken, err := h.as.GenerateRefreshToken()
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	if err := h.as.StoreRefreshToken(c.Request.Context(), cookies.userId, cookies.deviceId, hashToken); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	c.SetCookie(constants.COOKIE_REFRESH_TOKEN, newRefreshToken, 3600*24*365, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"token": "Bearer " + tokenAcc})
}

func (h *authHandler) GetAuth(c *gin.Context) {
	value, exist := c.Get("authenticatedUserId")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validated body not exists"})
		return
	}
	userId, ok := value.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.us.GetUserById(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *authHandler) Login(c *gin.Context) {
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
	existingUser, err := h.as.GetUserByIdentity(c.Request.Context(), body.Identity)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isMatch := h.as.VerifyPassword(existingUser.Password, body.Password); !isMatch {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}
	jti := uuid.New()
	tokenAcc, err := h.as.GenerateToken(existingUser.ID, jti)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	deviceId := uuid.New()
	newRefreshToken, hashToken, err := h.as.GenerateRefreshToken()
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	if err := h.as.StoreRefreshToken(c.Request.Context(), jti, existingUser.ID, deviceId, hashToken); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	c.SetCookie(constants.COOKIE_REFRESH_TOKEN, newRefreshToken, 3600*24*365, "/", "", false, true)
	c.SetCookie(constants.COOKIE_DEVICE_ID, deviceId.String(), 3600*24*365, "/", "", false, false)
	c.SetCookie(constants.COOKIE_USER_ID, existingUser.ID.String(), 3600*24*365, "/", "", false, false)
	c.JSON(http.StatusOK, gin.H{
		"user":  existingUser,
		"token": "Bearer " + tokenAcc,
	})
}
