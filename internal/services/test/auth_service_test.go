package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"my-go-api/internal/dto"
	"my-go-api/internal/mocks"
	"my-go-api/internal/models"
	"my-go-api/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockIUserRepository(ctrl)
	mockUtils := mocks.NewMockIUtils(ctrl)

	authService := services.NewAuthService(mockUserRepo, mockUtils, nil, "uri")

	ctx := context.Background()
	req := dto.CreateUser{
		Name:     "John Doe",
		Username: "johndoe",
		Email:    "john@example.com",
		Password: "securepassword",
	}

	t.Run("it should fail if username already exists", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByUsername(ctx, req.Username).Return(&models.User{}, nil)

		user, err := authService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "username is registered", err.Error())
	})

	t.Run("it should fail if email already exists", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByUsername(ctx, req.Username).Return(nil, sql.ErrNoRows)
		mockUserRepo.EXPECT().GetByEmail(ctx, req.Email).Return(&models.User{}, nil)

		user, err := authService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "email is registered", err.Error())
	})

	t.Run("it should fail if password hashing fails", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByUsername(ctx, req.Username).Return(nil, sql.ErrNoRows)
		mockUserRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, sql.ErrNoRows)
		mockUtils.EXPECT().HashPassword(req.Password).Return("", errors.New("hashing failed"))

		user, err := authService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "hashing failed", err.Error())
	})

	t.Run("it should fail if user creation fails", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByUsername(ctx, req.Username).Return(nil, sql.ErrNoRows)
		mockUserRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, sql.ErrNoRows)
		mockUtils.EXPECT().HashPassword(req.Password).Return("hashedpassword", nil)
		mockUserRepo.EXPECT().Create(ctx, req.Name, req.Username, req.Email, "hashedpassword").
			Return(nil, errors.New("db error"))

		user, err := authService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "create user failed", err.Error())
	})

	t.Run("it should succeed when all validations pass", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByUsername(ctx, req.Username).Return(nil, sql.ErrNoRows)
		mockUserRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, sql.ErrNoRows)
		mockUtils.EXPECT().HashPassword(req.Password).Return("hashedpassword", nil)

		expectedUser := &models.User{
			ID:       uuid.New(),
			Name:     req.Name,
			Username: req.Username,
			Email:    req.Email,
			Password: "hashedpassword",
		}

		mockUserRepo.EXPECT().Create(ctx, req.Name, req.Username, req.Email, "hashedpassword").
			Return(expectedUser, nil)

		user, err := authService.CreateUser(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser, user)
	})
}

func TestValidateToken(t *testing.T) {
	t.Run("it should fail because utility.ValidateToken return err", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUtils := mocks.NewMockIUtils(ctrl)
		mockUtils.EXPECT().ValidateToken(gomock.Any()).Return(nil, errors.New("some errors"))
		authService := services.NewAuthService(nil, mockUtils, nil, "uri")
		payload, err := authService.ValidateToken("token")
		assert.Error(t, err)
		assert.Nil(t, payload)
		assert.Equal(t, err.Error(), "invalid token")
	})

	t.Run("it should fail because token expired", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUtils := mocks.NewMockIUtils(ctrl)
		mockClaims := &jwt.MapClaims{
			"exp":    float64(time.Now().Add(-1 * time.Hour).UnixMilli()),
			"userId": uuid.New().String(),
		}
		t.Log(time.Now())
		mockUtils.EXPECT().ValidateToken(gomock.Any()).Return(mockClaims, nil)
		authService := services.NewAuthService(nil, mockUtils, nil, "uri")
		payload, err := authService.ValidateToken("token")
		assert.Error(t, err)
		assert.Nil(t, payload)
		assert.Equal(t, err.Error(), "token expired")
	})

	t.Run("it should works because mockClaims.UserId equals to payload.UserId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUtils := mocks.NewMockIUtils(ctrl)
		userId := uuid.New()
		mockClaims := &jwt.MapClaims{
			"exp":    float64(time.Now().Add(1 * time.Hour).UnixMilli()),
			"userId": userId.String(),
		}
		mockUtils.EXPECT().ValidateToken(gomock.Any()).Return(mockClaims, nil)
		authService := services.NewAuthService(nil, mockUtils, nil, "uri")
		payload, err := authService.ValidateToken("token")
		assert.NoError(t, err)
		assert.Equal(t, payload.UserId, userId)
	})
}

func TestGetUserByIdentity(t *testing.T) {
	t.Run("It should work. Execute userRepo.GetByEmail because input identity is an email", func(t *testing.T) {
		input := "test@mail.com"
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockIUserRepository(ctrl)
		mockUserRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(&models.User{ID: uuid.New(), Email: "test@mail.com"}, nil)
		authService := services.NewAuthService(mockUserRepo, nil, nil, "")
		user, err := authService.GetUserByIdentity(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, input, user.Email)
	})
	t.Run("It should work. Execute userRepo.GetByUsername because input identity is not an email", func(t *testing.T) {
		input := "test_username"
		ctrl := gomock.NewController(t)
		mockUserRepo := mocks.NewMockIUserRepository(ctrl)
		mockUserRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(&models.User{ID: uuid.New(), Username: "test_username"}, nil)
		authService := services.NewAuthService(mockUserRepo, nil, nil, "")
		user, err := authService.GetUserByIdentity(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, input, user.Username)
	})
	t.Run("It should failed. Execute userRepo.GetByEmail but corresponding user is not found", func(t *testing.T) {
		input := "test@mail.com"
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockIUserRepository(ctrl)
		mockUserRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(nil, sql.ErrNoRows)
		authService := services.NewAuthService(mockUserRepo, nil, nil, "")
		user, err := authService.GetUserByIdentity(context.Background(), input)
		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "user not found")
	})
	t.Run("It should failed. Execute userRepo.GetByUsername but corresponding user is not found", func(t *testing.T) {
		input := "test_username"
		ctrl := gomock.NewController(t)
		mockUserRepo := mocks.NewMockIUserRepository(ctrl)
		mockUserRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(nil, sql.ErrNoRows)
		authService := services.NewAuthService(mockUserRepo, nil, nil, "")
		user, err := authService.GetUserByIdentity(context.Background(), input)
		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "user not found")
	})
}

func TestSendVerificationEmail(t *testing.T) {
	t.Run("it should work", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUtils := mocks.NewMockIUtils(ctrl)
		mockUtils.EXPECT().SendEmailWithGmail("Email verification", gomock.Any(), "test@example.com").Return(nil)
		authService := services.NewAuthService(nil, mockUtils, nil, "")
		err := authService.SendVerificationEmail("John", "test@example.com", "some-token")
		assert.NoError(t, err)
	})
	t.Run("it should fail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUtils := mocks.NewMockIUtils(ctrl)
		mockUtils.EXPECT().SendEmailWithGmail("Email verification", gomock.Any(), "test@example.com").Return(errors.New("some errors"))
		authService := services.NewAuthService(nil, mockUtils, nil, "")
		err := authService.SendVerificationEmail("John", "test@example.com", "some-token")
		assert.Error(t, err)
	})
}

func TestStoreRefreshToken(t *testing.T) {
	t.Run("it should work", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)
		mockTokenRepo.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.Token{}, nil)
		authService := services.NewAuthService(nil, nil, mockTokenRepo, "")
		err := authService.StoreRefreshToken(context.Background(), uuid.New(), uuid.New(), "some-hash")
		assert.NoError(t, err)
	})
	t.Run("it should fail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)
		mockTokenRepo.EXPECT().Insert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some errors"))
		authService := services.NewAuthService(nil, nil, mockTokenRepo, "")
		err := authService.StoreRefreshToken(context.Background(), uuid.New(), uuid.New(), "some-hash")
		assert.Error(t, err)
	})
}

func TestDeleteRefreshToken(t *testing.T) {
	t.Run("it should work because tokenRepo.Remove return nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)
		mockTokenRepo.EXPECT().Remove(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		authService := services.NewAuthService(nil, nil, mockTokenRepo, "")
		err := authService.DeleteRefreshToken(context.Background(), uuid.New(), uuid.New())
		assert.NoError(t, err)
	})
	t.Run("it should fail because tokenRepo.Remove return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)
		mockTokenRepo.EXPECT().Remove(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some errors"))
		authService := services.NewAuthService(nil, nil, mockTokenRepo, "")
		err := authService.DeleteRefreshToken(context.Background(), uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

func TestVerifyRefreshToken(t *testing.T) {
	t.Run("it should fail because tokenRepo.GetToken return err", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)
		mockTokenRepo.EXPECT().GetToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some errors"))
		authService := services.NewAuthService(nil, nil, mockTokenRepo, "test.com")
		err := authService.VerifyRefreshToken(context.Background(), uuid.New(), uuid.New(), "token")
		assert.Error(t, err)
	})

	t.Run("it should fail because token isRevoked true", func(t *testing.T) {
		token := models.Token{
			ID:        1,
			Hash:      "some bytes",
			IsRevoked: true,
			DeviceId:  uuid.New(),
			UserId:    uuid.New(),
			ExpiredAt: "",
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)
		mockTokenRepo.EXPECT().GetToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(&token, nil)
		authService := services.NewAuthService(nil, nil, mockTokenRepo, "test.com")
		err := authService.VerifyRefreshToken(context.Background(), uuid.New(), uuid.New(), "token")
		assert.Error(t, err)
	})

	t.Run("it should fail because current time exceeds token expiredAt", func(t *testing.T) {
		exp := time.Now().Add(-5 * time.Hour).Format(time.RFC3339) // Corrected formatting
		t.Log(exp)
		token := models.Token{
			ID:        1,
			Hash:      "some bytes",
			IsRevoked: false, // This should trigger "reuse token detected" before expiration check
			DeviceId:  uuid.New(),
			UserId:    uuid.New(),
			ExpiredAt: exp,
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)
		mockTokenRepo.EXPECT().GetToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(&token, nil)
		authService := services.NewAuthService(nil, nil, mockTokenRepo, "test.com")
		err := authService.VerifyRefreshToken(context.Background(), uuid.New(), uuid.New(), "token")
		assert.Error(t, err)
	})

	t.Run("it should fail because hash not matched", func(t *testing.T) {
		exp := time.Now().Add(5 * time.Hour).Format(time.RFC3339)
		token := models.Token{
			ID:        1,
			Hash:      "some bytes",
			IsRevoked: false,
			DeviceId:  uuid.New(),
			UserId:    uuid.New(),
			ExpiredAt: exp,
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUtils := mocks.NewMockIUtils(ctrl)
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)

		mockTokenRepo.EXPECT().GetToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(&token, nil)
		mockUtils.EXPECT().HashWithSHA256(gomock.Any()).Return("hash")

		authService := services.NewAuthService(nil, mockUtils, mockTokenRepo, "test.com")
		err := authService.VerifyRefreshToken(context.Background(), uuid.New(), uuid.New(), "token")
		assert.Error(t, err)
	})

	t.Run("it should work because token exists, isRevoked false, current time < token.expiredAt and has the same hash", func(t *testing.T) {
		exp := time.Now().Add(5 * time.Hour).Format(time.RFC3339)
		token := models.Token{
			ID:        1,
			Hash:      "same",
			IsRevoked: false,
			DeviceId:  uuid.New(),
			UserId:    uuid.New(),
			ExpiredAt: exp,
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUtils := mocks.NewMockIUtils(ctrl)
		mockTokenRepo := mocks.NewMockITokenRepository(ctrl)

		mockTokenRepo.EXPECT().GetToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(&token, nil)

		mockUtils.EXPECT().HashWithSHA256(gomock.Any()).Return("same")

		authService := services.NewAuthService(nil, mockUtils, mockTokenRepo, "test.com")
		err := authService.VerifyRefreshToken(context.Background(), uuid.New(), uuid.New(), "token")
		assert.NoError(t, err)
	})
}
