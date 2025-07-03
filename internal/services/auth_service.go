package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"my-go-api/internal/dto"
	"my-go-api/internal/models"
	"my-go-api/internal/repositories"
	"my-go-api/internal/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type IAuthService interface {
	SendVerificationEmail(name, email, token string) error
	StoreRefreshToken(ctx context.Context, jti, userId, deviceId uuid.UUID, hash string) error
	DeleteRefreshToken(ctx context.Context, userId, deviceId uuid.UUID) error
	VerifyRefreshToken(ctx context.Context, userId, deviceId uuid.UUID, token string) error
	GenerateRefreshToken() (string, string, error)
	GenerateToken(userId, jti uuid.UUID) (string, error)
	VerifyPassword(hashedPassword string, plainPassword string) bool
	GetUserByIdentity(ctx context.Context, identity string) (*models.User, error)
	ValidateToken(tokenString string) (*TokenPayload, error)
	CreateUser(ctx context.Context, req dto.CreateUser) (*models.User, error)
}

type authService struct {
	appUri    string
	userRepo  repositories.IUserRepository
	tokenRepo repositories.ITokenRepository
	redisRepo repositories.IRedisRepository
	utility   utils.IUtils
}

func NewAuthService(
	userRepo repositories.IUserRepository,
	utility utils.IUtils,
	tokenRepo repositories.ITokenRepository,
	redisRepo repositories.IRedisRepository,
	appUri string,
) IAuthService {

	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		appUri:    appUri,
		utility:   utility,
		redisRepo: redisRepo,
	}

}

func (s *authService) SendVerificationEmail(name, email, token string) error {
	var link = s.appUri + fmt.Sprintf("/email-verification?token=%s", token)
	var subject = "Email verification"
	var emailBody = fmt.Sprintf("Hello %s.\n\n Please follow this link to verify your new account\n\n%s", name, link)
	err := s.utility.SendEmailWithGmail(subject, emailBody, email)
	if err != nil {
		return err
	}
	return nil
}

func (s *authService) StoreRefreshToken(
	ctx context.Context,
	jti, userId, deviceId uuid.UUID,
	hash string) error {
	key := fmt.Sprintf("refresh-token:%s", hash)
	if err := s.redisRepo.HSet(key, map[string]any{
		"userId":   userId,
		"deviceId": deviceId,
		"jti":      jti,
	}, time.Hour*24*7); err != nil {
		return err
	}
	return nil
}

// func (s *authService) StoreRefreshToken(ctx context.Context, userId, deviceId uuid.UUID, hash string) error {
// 	_, err := s.tokenRepo.Insert(ctx, userId, deviceId, hash)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (s *authService) DeleteRefreshToken(ctx context.Context, userId, deviceId uuid.UUID) error {
	err := s.tokenRepo.Remove(ctx, userId, deviceId)
	if err != nil {
		return err
	}
	return nil
}

func (s *authService) VerifyRefreshToken(ctx context.Context, userId, deviceId uuid.UUID, token string) error {
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
	if existingToken.Hash != s.utility.HashWithSHA256(token) {
		return errors.New("unrecognized token")
	}
	return nil
}

func (s *authService) GenerateRefreshToken() (string, string, error) {
	raw, err := s.utility.GenerateRandomBytes(32)
	if err != nil {
		return "", "", errors.New("failed to generate refresh token")
	}
	hash := s.utility.HashWithSHA256(raw)
	return raw, hash, nil
}

func (s *authService) GenerateToken(userId, jti uuid.UUID) (string, error) {
	token, err := s.utility.GenerateToken(userId, jti)
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	return token, nil
}

func (s *authService) VerifyPassword(hashedPassword string, plainPassword string) bool {
	if err := s.utility.VerifyPassword(hashedPassword, plainPassword); err != nil {
		return false
	}
	return true
}

func (s *authService) GetUserByIdentity(ctx context.Context, identity string) (*models.User, error) {
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

func (s *authService) ValidateToken(tokenString string) (*TokenPayload, error) {
	claims, err := s.utility.ValidateToken(tokenString)
	log.Println(claims)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	expireTime, ok := (*claims)["exp"].(float64)
	if !ok {
		return nil, errors.New("failed to covert to int64")
	}
	expirationTime := time.UnixMilli(int64(expireTime))
	if expirationTime.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	userId, ok := (*claims)["userId"].(string)
	if !ok {
		return nil, errors.New("failed to covert to uuid")
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

func (u *authService) CreateUser(ctx context.Context, req dto.CreateUser) (*models.User, error) {
	existingUser, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username has been taken")
	}

	existingUser, err = u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email has been taken")
	}

	hashedPassword, err := u.utility.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.Create(ctx, req.Name, req.Username, req.Email, hashedPassword)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("create user failed")
	}

	return user, nil
}
