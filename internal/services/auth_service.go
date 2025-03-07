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

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	tokenRepo *repositories.TokenRepository
}

func NewAuthService(userRepo *repositories.UserRepository, tokenRepo *repositories.TokenRepository) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

func (s *AuthService) StoreRefreshToken(ctx context.Context, tokenId, userId, value string) error {
	_, err := s.tokenRepo.Insert(ctx, tokenId, userId, value)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) DeleteRefreshToken(ctx context.Context, tokenId uuid.UUID) (bool, error) {
	ok, err := s.tokenRepo.Remove(ctx, tokenId)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *AuthService) VerifyRefreshToken(ctx context.Context, tokenId, userId string) error {
	existingToken, err := s.tokenRepo.GetByTokenId(ctx, tokenId)
	if err != nil {
		return err
	}

	if existingToken.UserId != userId || !existingToken.IsRevoked {
		return errors.New("token invalid")
	}

	payload, err := s.ValidateToken(existingToken.Value, "refresh")

	if err != nil {
		return err
	}

	if payload.UserId == userId && payload.TokenId == tokenId {
		return nil
	}
	return errors.New("token data is invalid")

}

func (s *AuthService) GenerateToken(userId, tokenId, tokenType string) (string, error) {

	var myTokenType utils.TokenType
	switch tokenType {
	case "refresh":
		myTokenType = utils.RefreshToken
	case "access":
		myTokenType = utils.AccessToken
	case "link":
		myTokenType = utils.LinkToken
	default:
		return "", errors.New("undefined token")
	}

	token, err := utils.GenerateToken(userId, tokenId, myTokenType)
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
	UserId  string
	TokenId string
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

	currTokenType, ok := (*claims)["type"].(string)

	if !ok || currTokenType != tokenType {
		return nil, errors.New("failed to recognize the token")
	}

	userId, ok := (*claims)["user_id"].(string)
	if !ok {
		return nil, errors.New("failed to covert to string")
	}

	tokenId, ok := (*claims)["token_id"].(string)
	if !ok {
		return nil, errors.New("failed to covert to string")
	}

	payload := &TokenPayload{
		UserId:  userId,
		TokenId: tokenId,
	}
	return payload, nil
}
