package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
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

func TestUpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_services.NewMockIUserService(ctrl)
	handler := handlers.NewUserHandler(mockService)

	router := gin.Default()
	router.PUT("/user/:id", func(c *gin.Context) {
		// Simulating middleware setting "validatedBody" before handler is called
		c.Set("validatedBody", map[string]interface{}{
			"name":  "Updated Name",
			"email": "updated@example.com",
		})
		handler.Update(c)
	})

	validUserID := uuid.New()

	t.Run("should return 500 for invalid user ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/user/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Invalid user id"}`, w.Body.String())
	})

	t.Run("should return 400 when request body is missing", func(t *testing.T) {
		routerNoBody := gin.Default()
		routerNoBody.PUT("/user/:id", handler.Update)

		req, _ := http.NewRequest(http.MethodPut, "/user/"+validUserID.String(), nil)
		w := httptest.NewRecorder()
		routerNoBody.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error": "Error"}`, w.Body.String())
	})

	t.Run("should return 404 when user not found", func(t *testing.T) {
		mockService.EXPECT().GetUserById(gomock.Any(), validUserID).Return(nil, sql.ErrNoRows)

		req, _ := http.NewRequest(http.MethodPut, "/user/"+validUserID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error": "User not found"}`, w.Body.String())
	})

	t.Run("should return 500 when service fails", func(t *testing.T) {
		mockService.EXPECT().GetUserById(gomock.Any(), validUserID).Return(nil, errors.New("some error"))

		req, _ := http.NewRequest(http.MethodPut, "/user/"+validUserID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Something went wrong"}`, w.Body.String())
	})

	t.Run("should return 200 on successful update", func(t *testing.T) {
		existingUser := &models.User{
			ID:        validUserID,
			Name:      "Old Name",
			Email:     "old@example.com",
			Username:  "",
			Password:  "",
			Provider:  "",
			Role:      "",
			CreatedAt: "",
			UpdatedAt: "",
		}

		mockService.EXPECT().GetUserById(gomock.Any(), validUserID).Return(existingUser, nil)
		mockService.EXPECT().UpdateUser(gomock.Any(), gomock.AssignableToTypeOf(&models.User{}))

		reqBody, _ := json.Marshal(map[string]interface{}{
			"name":  "Updated Name",
			"email": "updated@example.com",
		})

		req, _ := http.NewRequest(http.MethodPut, "/user/"+validUserID.String(), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		expectedResponse := `{
			"user": {
				"id": "` + validUserID.String() + `",
				"name": "Updated Name",
				"email": "updated@example.com",
				"username":  "",
				"provider":  "",
				"role":      "",
				"created_at": "",
				"updated_at": ""
			}
		}`

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, expectedResponse, w.Body.String())
	})
}

func TestGetUserById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_services.NewMockIUserService(ctrl)
	handler := handlers.NewUserHandler(mockService)

	router := gin.Default()
	router.GET("/user/:id", handler.GetUserById)

	userId := uuid.New()
	mockUser := &models.User{ID: userId, Name: "John Doe", Email: "john@example.com"}

	t.Run("should return 200 with user data", func(t *testing.T) {
		mockService.EXPECT().GetUserById(gomock.Any(), userId).Return(mockUser, nil)

		req, _ := http.NewRequest(http.MethodGet, "/user/"+userId.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 when user not found", func(t *testing.T) {
		mockService.EXPECT().GetUserById(gomock.Any(), userId).Return(nil, sql.ErrNoRows)

		req, _ := http.NewRequest(http.MethodGet, "/user/"+userId.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("should return 500 when user not found", func(t *testing.T) {
		mockService.EXPECT().GetUserById(gomock.Any(), userId).Return(nil, errors.New("some errors"))

		req, _ := http.NewRequest(http.MethodGet, "/user/"+userId.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGetAllUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_services.NewMockIUserService(ctrl)
	handler := handlers.NewUserHandler(mockService)

	router := gin.Default()
	router.GET("/users", handler.GetAll)

	mockUsers := []models.User{
		{ID: uuid.New(), Name: "John Doe", Email: "john@example.com"},
		{ID: uuid.New(), Name: "Jane Doe", Email: "jane@example.com"},
	}

	t.Run("should return 200 with user list", func(t *testing.T) {
		mockService.EXPECT().GetAllUsers(gomock.Any()).Return(mockUsers, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
