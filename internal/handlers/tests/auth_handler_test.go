package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"my-go-api/internal/constants"
	"my-go-api/internal/dto"
	"my-go-api/internal/handlers"
	"my-go-api/internal/mocks/mock_services"
	"my-go-api/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mock_services.NewMockIAuthService(ctrl)
	mockUserService := mock_services.NewMockIUserService(ctrl)
	authHandler := handlers.NewAuthHandler(mockAuthService, mockUserService)

	t.Run("should return 500 if cookies are missing", func(t *testing.T) {
		router := gin.Default()
		router.GET("/logout", authHandler.Logout)

		req, _ := http.NewRequest(http.MethodGet, "/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Something went wrong"}`, w.Body.String())
	})

	t.Run("should return 500 if DeleteRefreshToken fails", func(t *testing.T) {
		router := gin.Default()
		router.GET("/logout", func(c *gin.Context) {
			c.Set(constants.COOKIE_REFRESH_TOKEN, "valid-token")
			c.Set(constants.COOKIE_DEVICE_ID, uuid.New().String())
			c.Set(constants.COOKIE_USER_ID, uuid.New().String())

			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_REFRESH_TOKEN, Value: "valid-token"})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_DEVICE_ID, Value: uuid.New().String()})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_USER_ID, Value: uuid.New().String()})

			authHandler.Logout(c)
		})

		mockAuthService.EXPECT().DeleteRefreshToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed to delete token"))

		req, _ := http.NewRequest(http.MethodGet, "/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Something went wrong"}`, w.Body.String())
	})

	t.Run("should return 200 and clear cookies on successful logout", func(t *testing.T) {
		userId := uuid.New()
		deviceId := uuid.New()

		router := gin.Default()
		router.GET("/logout", func(c *gin.Context) {
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_REFRESH_TOKEN, Value: "valid-token"})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_DEVICE_ID, Value: deviceId.String()})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_USER_ID, Value: userId.String()})

			authHandler.Logout(c)
		})

		mockAuthService.EXPECT().DeleteRefreshToken(gomock.Any(), userId, deviceId).Return(nil)

		req, _ := http.NewRequest(http.MethodGet, "/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message": "Logout"}`, w.Body.String())

		// Check if cookies are cleared
		cookies := w.Result().Cookies()
		for _, cookie := range cookies {
			assert.Equal(t, -1, cookie.MaxAge)
			assert.Empty(t, cookie.Value)
		}
	})
}

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mock_services.NewMockIAuthService(ctrl)
	mockUserService := mock_services.NewMockIUserService(ctrl)
	authHandler := handlers.NewAuthHandler(mockAuthService, mockUserService)

	t.Run("should return 400 when validatedBody is missing", func(t *testing.T) {
		router := gin.Default()
		router.POST("/register", authHandler.Register)

		req, _ := http.NewRequest(http.MethodPost, "/register", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error": "Error"}`, w.Body.String())
	})

	t.Run("should return 500 when validatedBody has invalid type", func(t *testing.T) {
		router := gin.Default()
		router.POST("/register", func(c *gin.Context) {
			c.Set("validatedBody", "invalid-type") // Simulate incorrect type
			authHandler.Register(c)
		})

		req, _ := http.NewRequest(http.MethodPost, "/register", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "invalid type for validatedBody"}`, w.Body.String())
	})

	t.Run("should return 409 if CreateUser fails", func(t *testing.T) {
		router := gin.Default()
		router.POST("/register", func(c *gin.Context) {
			c.Set("validatedBody", dto.CreateUser{
				Name:     "John Doe",
				Username: "johndoe",
				Email:    "john@example.com",
				Password: "securepassword",
			})
			authHandler.Register(c)
		})
		mockAuthService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to create user"))
		req, _ := http.NewRequest(http.MethodPost, "/register", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.JSONEq(t, `{"errors": "failed to create user"}`, w.Body.String())
	})

	t.Run("should return 500 if GenerateToken fails", func(t *testing.T) {
		user := &models.User{ID: uuid.New(), Email: "john@example.com"}
		router := gin.Default()
		router.POST("/register", func(c *gin.Context) {
			c.Set("validatedBody", dto.CreateUser{
				Name:     "John Doe",
				Username: "johndoe",
				Email:    "john@example.com",
				Password: "securepassword",
			})
			authHandler.Register(c)
		})
		mockAuthService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(user, nil)
		mockAuthService.EXPECT().GenerateToken(user.ID).Return("", errors.New("token generation failed"))
		req, _ := http.NewRequest(http.MethodPost, "/register", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"errors": "Something went wrong"}`, w.Body.String())
	})

	t.Run("should return 500 if SendVerificationEmail fails", func(t *testing.T) {
		user := &models.User{ID: uuid.New(), Email: "john@example.com"}
		token := "some-token"
		router := gin.Default()
		router.POST("/register", func(c *gin.Context) {
			c.Set("validatedBody", dto.CreateUser{
				Name:     "John Doe",
				Username: "johndoe",
				Email:    "john@example.com",
				Password: "securepassword",
			})
			authHandler.Register(c)
		})
		mockAuthService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(user, nil)
		mockAuthService.EXPECT().GenerateToken(user.ID).Return(token, nil)
		mockAuthService.EXPECT().SendVerificationEmail(user.Name, user.Email, token).Return(errors.New("email send failed"))
		req, _ := http.NewRequest(http.MethodPost, "/register", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"errors": "Something went wrong"}`, w.Body.String())
	})

	t.Run("should return 201 if registration is successful", func(t *testing.T) {
		user := &models.User{ID: uuid.New(), Email: "john@example.com"}
		token := "some-token"
		router := gin.Default()
		router.POST("/register", func(c *gin.Context) {
			c.Set("validatedBody", dto.CreateUser{
				Name:     "John Doe",
				Username: "johndoe",
				Email:    "john@example.com",
				Password: "securepassword",
			})
			authHandler.Register(c)
		})
		mockAuthService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(user, nil)
		mockAuthService.EXPECT().GenerateToken(user.ID).Return(token, nil)
		mockAuthService.EXPECT().SendVerificationEmail(user.Name, user.Email, token).Return(nil)
		reqBody, _ := json.Marshal(dto.CreateUser{
			Name:     "John Doe",
			Username: "johndoe",
			Email:    "john@example.com",
			Password: "securepassword",
		})
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		expectedMsg := `{"message": "An email has been sent to john@example.com. Please follow the instruction to verify your account."}`
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, expectedMsg, w.Body.String())
	})
}

func TestRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mock_services.NewMockIAuthService(ctrl)
	mockUserService := mock_services.NewMockIUserService(ctrl)
	authHandler := handlers.NewAuthHandler(mockAuthService, mockUserService)

	t.Run("should return 401 if cookies are missing", func(t *testing.T) {
		router := gin.Default()
		router.GET("/refresh-token", authHandler.RefreshToken)

		req, _ := http.NewRequest(http.MethodGet, "/refresh-token", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 401 if refresh token verification fails", func(t *testing.T) {
		router := gin.Default()
		router.GET("/refresh-token", func(c *gin.Context) {
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_REFRESH_TOKEN, Value: "invalid-token"})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_DEVICE_ID, Value: uuid.New().String()})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_USER_ID, Value: uuid.New().String()})

			authHandler.RefreshToken(c)
		})

		mockAuthService.EXPECT().VerifyRefreshToken(gomock.Any(), gomock.Any(), gomock.Any(), "invalid-token").Return(errors.New("invalid token"))

		req, _ := http.NewRequest(http.MethodGet, "/refresh-token", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error": "Unauthorized"}`, w.Body.String())
	})

	t.Run("should return 500 if generating new token fails", func(t *testing.T) {
		userId := uuid.New()
		deviceId := uuid.New()

		router := gin.Default()
		router.GET("/refresh-token", func(c *gin.Context) {
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_REFRESH_TOKEN, Value: "valid-token"})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_DEVICE_ID, Value: deviceId.String()})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_USER_ID, Value: userId.String()})

			authHandler.RefreshToken(c)
		})

		mockAuthService.EXPECT().VerifyRefreshToken(gomock.Any(), userId, deviceId, "valid-token").Return(nil)
		mockAuthService.EXPECT().GenerateToken(userId).Return("", errors.New("failed to generate token"))

		req, _ := http.NewRequest(http.MethodGet, "/refresh-token", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Something went wrong"}`, w.Body.String())
	})

	t.Run("should return 500 if deleting old refresh token fails", func(t *testing.T) {
		userId := uuid.New()
		deviceId := uuid.New()

		router := gin.Default()
		router.GET("/refresh-token", func(c *gin.Context) {
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_REFRESH_TOKEN, Value: "valid-token"})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_DEVICE_ID, Value: deviceId.String()})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_USER_ID, Value: userId.String()})

			authHandler.RefreshToken(c)
		})

		mockAuthService.EXPECT().VerifyRefreshToken(gomock.Any(), userId, deviceId, "valid-token").Return(nil)
		mockAuthService.EXPECT().GenerateToken(userId).Return("new-access-token", nil)
		mockAuthService.EXPECT().DeleteRefreshToken(gomock.Any(), userId, deviceId).Return(errors.New("failed to delete old token"))

		req, _ := http.NewRequest(http.MethodGet, "/refresh-token", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Something went wrong"}`, w.Body.String())
	})

	t.Run("should return 500 if new refresh token generation fails", func(t *testing.T) {
		userId := uuid.New()
		deviceId := uuid.New()

		router := gin.Default()
		router.GET("/refresh-token", func(c *gin.Context) {
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_REFRESH_TOKEN, Value: "valid-token"})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_DEVICE_ID, Value: deviceId.String()})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_USER_ID, Value: userId.String()})

			authHandler.RefreshToken(c)
		})

		mockAuthService.EXPECT().VerifyRefreshToken(gomock.Any(), userId, deviceId, "valid-token").Return(nil)
		mockAuthService.EXPECT().GenerateToken(userId).Return("new-access-token", nil)
		mockAuthService.EXPECT().DeleteRefreshToken(gomock.Any(), userId, deviceId).Return(nil)
		mockAuthService.EXPECT().GenerateRefreshToken().Return("", "", errors.New("failed to generate refresh token"))

		req, _ := http.NewRequest(http.MethodGet, "/refresh-token", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Something went wrong"}`, w.Body.String())
	})

	t.Run("should return 200 with new tokens on success", func(t *testing.T) {
		userId := uuid.New()
		deviceId := uuid.New()

		router := gin.Default()
		router.GET("/refresh-token", func(c *gin.Context) {
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_REFRESH_TOKEN, Value: "valid-token"})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_DEVICE_ID, Value: deviceId.String()})
			c.Request.AddCookie(&http.Cookie{Name: constants.COOKIE_USER_ID, Value: userId.String()})

			authHandler.RefreshToken(c)
		})

		mockAuthService.EXPECT().VerifyRefreshToken(gomock.Any(), userId, deviceId, "valid-token").Return(nil)
		mockAuthService.EXPECT().GenerateToken(userId).Return("new-access-token", nil)
		mockAuthService.EXPECT().DeleteRefreshToken(gomock.Any(), userId, deviceId).Return(nil)
		mockAuthService.EXPECT().GenerateRefreshToken().Return("new-refresh-token", "hashed-token", nil)
		mockAuthService.EXPECT().StoreRefreshToken(gomock.Any(), userId, deviceId, "hashed-token").Return(nil)

		req, _ := http.NewRequest(http.MethodGet, "/refresh-token", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"token": "Bearer new-access-token"}`, w.Body.String())

		// Check if new refresh token is set in cookies
		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "new-refresh-token", cookies[0].Value)
	})
}

func TestGetAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mock_services.NewMockIAuthService(ctrl)
	mockUserService := mock_services.NewMockIUserService(ctrl)
	authHandler := handlers.NewAuthHandler(mockAuthService, mockUserService)

	t.Run("should return 400 if authenticatedUserId is missing", func(t *testing.T) {
		router := gin.Default()
		router.GET("/get-auth", authHandler.GetAuth)

		req, _ := http.NewRequest(http.MethodGet, "/get-auth", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error": "validated body not exists"}`, w.Body.String())
	})

	t.Run("should return 400 if authenticatedUserId is not a UUID", func(t *testing.T) {
		router := gin.Default()
		router.GET("/get-auth", func(c *gin.Context) {
			c.Set("authenticatedUserId", "not-a-uuid") // Invalid type
			authHandler.GetAuth(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/get-auth", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error": "invalid user id"}`, w.Body.String())
	})

	t.Run("should return 400 if user is not found", func(t *testing.T) {
		userId := uuid.New()

		router := gin.Default()
		router.GET("/get-auth", func(c *gin.Context) {
			c.Set("authenticatedUserId", userId)
			authHandler.GetAuth(c)
		})

		mockUserService.EXPECT().GetUserById(gomock.Any(), userId).Return(nil, errors.New("user not found"))

		req, _ := http.NewRequest(http.MethodGet, "/get-auth", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error": "user not found"}`, w.Body.String())
	})

	t.Run("should return 200 with user data", func(t *testing.T) {
		userId := uuid.New()
		mockUser := &models.User{ // Ensure it's a pointer to models.User
			ID:        userId,
			Name:      "John Doe",
			Email:     "john.doe@example.com",
			Username:  "",
			Role:      "",
			Provider:  "",
			CreatedAt: "",
			UpdatedAt: "",
		}

		router := gin.Default()
		router.GET("/get-auth", func(c *gin.Context) {
			c.Set("authenticatedUserId", userId)
			authHandler.GetAuth(c)
		})

		mockUserService.EXPECT().GetUserById(gomock.Any(), userId).Return(mockUser, nil) // Correct return type

		req, _ := http.NewRequest(http.MethodGet, "/get-auth", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		expectedJSON := `{
			"user": {
				"id": "` + userId.String() + `",
				"name": "John Doe",
				"email": "john.doe@example.com",
				"username": "",
				"role": "",
				"provider": "",
				"created_at": "",
				"updated_at": ""
			}
		}`
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, expectedJSON, w.Body.String())
	})

}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mock_services.NewMockIAuthService(ctrl)
	mockUserService := mock_services.NewMockIUserService(ctrl)

	authHandler := handlers.NewAuthHandler(mockAuthService, mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", func(c *gin.Context) {
		c.Set("validatedBody", dto.Login{
			Identity: "test@example.com",
			Password: "password123",
		})
		authHandler.Login(c)
	})

	t.Run("should return 200 with token on successful login", func(t *testing.T) {
		userID := uuid.New()
		existingUser := &models.User{
			ID:       userID,
			Email:    "test@example.com",
			Password: "hashed_password",
		}

		mockAuthService.EXPECT().GetUserByIdentity(gomock.Any(), "test@example.com").Return(existingUser, nil)
		mockAuthService.EXPECT().VerifyPassword("hashed_password", "password123").Return(true)
		mockAuthService.EXPECT().GenerateToken(userID).Return("test_token", nil)
		mockAuthService.EXPECT().GenerateRefreshToken().Return("refresh_token", "hashed_refresh_token", nil)
		mockAuthService.EXPECT().StoreRefreshToken(gomock.Any(), userID, gomock.Any(), "hashed_refresh_token").Return(nil)

		req, _ := http.NewRequest(http.MethodPost, "/login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Bearer test_token")
	})

	t.Run("should return 401 if password is incorrect", func(t *testing.T) {
		existingUser := &models.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			Password: "hashed_password",
		}

		mockAuthService.EXPECT().GetUserByIdentity(gomock.Any(), "test@example.com").Return(existingUser, nil)
		mockAuthService.EXPECT().VerifyPassword("hashed_password", "password123").Return(false)

		req, _ := http.NewRequest(http.MethodPost, "/login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "wrong password")
	})

	t.Run("should return 404 if user is not found", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserByIdentity(gomock.Any(), "test@example.com").Return(nil, errors.New("user not found"))

		req, _ := http.NewRequest(http.MethodPost, "/login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "user not found")
	})
}
