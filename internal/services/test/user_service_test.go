package services_test

import (
	"context"
	"errors"
	"my-go-api/internal/mocks"
	"my-go-api/internal/models"
	"my-go-api/internal/services"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIUserRepository(ctrl)
	userService := services.NewUserService(mockRepo)

	ctx := context.Background()
	userID := uuid.New()
	mockUser := &models.User{ID: userID, Name: "Test User"}

	t.Run("UpdateUser - success", func(t *testing.T) {
		mockRepo.EXPECT().Update(ctx, mockUser).Return(mockUser, nil)
		updatedUser, err := userService.UpdateUser(ctx, mockUser)
		assert.NoError(t, err)
		assert.Equal(t, mockUser, updatedUser)
	})

	t.Run("GetUserById - success", func(t *testing.T) {
		mockRepo.EXPECT().GetById(ctx, userID).Return(mockUser, nil)
		user, err := userService.GetUserById(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, mockUser, user)
	})

	t.Run("GetAllUsers - success", func(t *testing.T) {
		mockRepo.EXPECT().GetAll(ctx).Return([]models.User{*mockUser}, nil)
		users, err := userService.GetAllUsers(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, mockUser.ID, users[0].ID)
	})

	t.Run("GetUserById - not found", func(t *testing.T) {
		mockRepo.EXPECT().GetById(ctx, userID).Return(nil, errors.New("user not found"))
		user, err := userService.GetUserById(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}
