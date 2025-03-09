package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"my-go-api/internal/models"
	"my-go-api/internal/repositories"
	"my-go-api/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	appUri    string
	userRepo  repositories.IUserRepository
	tokenRepo *repositories.TokenRepository
}

func NewAuthService(userRepo repositories.IUserRepository, tokenRepo *repositories.TokenRepository, appUri string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		appUri:    appUri,
	}
}

func (s *AuthService) SendVerificationEmail(name, email, token string) error {
	var link = s.appUri + fmt.Sprintf("/email-verification?token=%s", token)
	var subject = "Email verification"
	var emailBody = fmt.Sprintf("Hello %s.\n\n Please follow this link to verify your new account\n\n%s", name, link)
	err := utils.SendEmail(subject, emailBody, email)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) StoreRefreshToken(ctx context.Context, userId, deviceId uuid.UUID, hash string) error {
	_, err := s.tokenRepo.Insert(ctx, userId, deviceId, hash)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) DeleteRefreshToken(ctx context.Context, userId, deviceId uuid.UUID) error {
	err := s.tokenRepo.Remove(ctx, userId, deviceId)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) VerifyRefreshToken(ctx context.Context, userId, deviceId uuid.UUID, token string) error {
	existingToken, err := s.tokenRepo.GetToken(ctx, userId, deviceId)
	if err != nil {
		return errors.New("stored token not found")
	}
	if existingToken.IsRevoked {
		return errors.New("reuse token detected")
	}
	t, err := time.Parse(time.RFC3339, existingToken.ExpiredAt)
	if err != nil {
		return errors.New("failed to parsed string expiredAt to time")
	}
	if t.UnixMilli() < time.Now().UnixMilli() {
		return errors.New("refresh token is expired")
	}
	if existingToken.Hash != utils.HashWithSHA256(token) {
		return errors.New("unrecognized token")
	}
	return nil
}

func (s *AuthService) GenerateRefreshToken() (string, string, error) {
	raw, err := utils.GenerateRandomBytes(32)
	if err != nil {
		return "", "", errors.New("failed to generate refresh token")
	}
	hash := utils.HashWithSHA256(raw)
	return raw, hash, nil
}

func (s *AuthService) GenerateToken(userId uuid.UUID) (string, error) {
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

type TokenPayload struct {
	UserId uuid.UUID
}

func (s *AuthService) ValidateToken(tokenString string, tokenType string) (*TokenPayload, error) {
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	expireTime, ok := (*claims)["exp"].(float64)
	if !ok {
		return nil, errors.New("failed to covert to float64")
	}
	expirationTime := time.Unix(int64(expireTime), 0)
	if expirationTime.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	userId, ok := (*claims)["userId"].(string)
	if !ok {
		return nil, errors.New("failed to covert to string")
	}
	userIdUUD, err := uuid.Parse(userId)
	if err != nil {
		return nil, err
	}
	payload := &TokenPayload{
		UserId: userIdUUD,
	}
	return payload, nil
}
