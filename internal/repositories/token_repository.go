package repositories

import (
	"context"
	"database/sql"
	"my-go-api/internal/models"

	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (s *TokenRepository) Insert(ctx context.Context, tokenId, userId, value string) (*models.Token, error) {
	token := &models.Token{}
	query := `
		INSERT INTO tokens (token_id, user_id, value)
		VALUES ($1, $2, $3)
		RETURNING id, token_id, user_id, value, is_revoked, created_at, expired_at
	`
	if err := s.db.QueryRowContext(ctx, query, tokenId, userId, value).Scan(
		&token.ID, &token.TokenId, &token.UserId, &token.Value, &token.IsRevoked, &token.CreatedAt, &token.ExpiredAt,
	); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *TokenRepository) GetByTokenId(ctx context.Context, tokenId string) (*models.Token, error) {
	token := &models.Token{}
	query := `
		SELECT id, token_id, user_id, value, is_revoked, created_at, expired_at
		FROM tokens
		WHERE token_id=$1
	`
	if err := s.db.QueryRowContext(ctx, query, tokenId).Scan(
		&token.ID, &token.TokenId, &token.UserId, &token.Value, &token.IsRevoked, &token.CreatedAt, &token.ExpiredAt,
	); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *TokenRepository) Remove(ctx context.Context, tokenId uuid.UUID) (bool, error) {
	query := `DELETE FROM token WHERE token_id=$1`

	result, err := s.db.ExecContext(ctx, query, tokenId)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return false, nil
	}

	return true, nil
}
