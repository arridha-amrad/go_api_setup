package services

import (
	"context"
	"database/sql"
	"errors"
	"my-go-api/internal/models"
	"my-go-api/internal/repositories"
	"my-go-api/pkg/utils"
	"strings"
	"time"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) GenerateToken(userId string) (string, error) {
	token, err := utils.GenerateToken(userId)
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	return token, nil
}

func (s *AuthService) VerifyPassword(hashedPassword string, plainPassword string) bool {
	if err := utils.VerifyPassword(hashedPassword, plainPassword); err != nil {
		return false
	}
	return true
}

func (s *AuthService) GetUserByIdentity(ctx context.Context, identity string) (*models.User, error) {
	var user *models.User
	if strings.Contains(identity, "@") {
		existingUser, err := s.userRepo.GetByEmail(ctx, identity)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.New("user not found")
			}
			return nil, err
		}
		user = existingUser
	} else {
		existingUser, err := s.userRepo.GetByUsername(ctx, identity)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.New("user not found")
			}
			return nil, err
		}
		user = existingUser
	}
	return user, nil
}

func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		return "", errors.New("invalid token")
	}

	expireTime, ok := (*claims)["exp"].(float64)
	if !ok {
		return "", errors.New("failed to covert to float64")
	}
	expirationTime := time.Unix(int64(expireTime), 0)
	if expirationTime.Before(time.Now()) {
		return "", errors.New("token expired")
	}

	userId, ok := (*claims)["user_id"].(string)
	if !ok {
		return "", errors.New("failed to covert to string")
	}

	return userId, nil
}
