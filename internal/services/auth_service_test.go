package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"my-go-api/internal/models"
	"my-go-api/internal/repositories/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUserByIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock the user repository
	mockUserRepo := mocks.NewMockIUserRepository(ctrl)

	// Create the AuthService with the mocked repository
	authService := &AuthService{
		userRepo: mockUserRepo,
	}

	// Define test cases
	tests := []struct {
		name        string
		identity    string
		mockUser    *models.User
		mockError   error
		expected    *models.User
		expectedErr string
	}{
		{
			name:     "identity is email - user found",
			identity: "test@example.com",
			mockUser: &models.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
			},
			mockError:   nil,
			expected:    &models.User{ID: uuid.New(), Email: "test@example.com", Username: "testuser"},
			expectedErr: "",
		},
		{
			name:        "identity is email - user not found",
			identity:    "notfound@example.com",
			mockUser:    nil,
			mockError:   sql.ErrNoRows,
			expected:    nil,
			expectedErr: "user not found",
		},
		{
			name:     "identity is username - user found",
			identity: "testuser",
			mockUser: &models.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
			},
			mockError:   nil,
			expected:    &models.User{ID: uuid.New(), Email: "test@example.com", Username: "testuser"},
			expectedErr: "",
		},
		{
			name:        "identity is username - user not found",
			identity:    "notfounduser",
			mockUser:    nil,
			mockError:   sql.ErrNoRows,
			expected:    nil,
			expectedErr: "user not found",
		},
		{
			name:        "identity is email - repository error",
			identity:    "error@example.com",
			mockUser:    nil,
			mockError:   errors.New("repository error"),
			expected:    nil,
			expectedErr: "repository error",
		},
		{
			name:        "identity is username - repository error",
			identity:    "erroruser",
			mockUser:    nil,
			mockError:   errors.New("repository error"),
			expected:    nil,
			expectedErr: "repository error",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations
			if strings.Contains(tt.identity, "@") {
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), tt.identity).
					Return(tt.mockUser, tt.mockError).
					Times(1)
			} else {
				mockUserRepo.EXPECT().
					GetByUsername(gomock.Any(), tt.identity).
					Return(tt.mockUser, tt.mockError).
					Times(1)
			}

			// Call the function
			user, err := authService.GetUserByIdentity(context.Background(), tt.identity)

			// Assertions
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Username, user.Username)
			}
		})
	}
}
