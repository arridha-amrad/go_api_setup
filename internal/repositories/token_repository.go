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

func (s *TokenRepository) Insert(ctx context.Context, userId, deviceId uuid.UUID, hash string) (*models.Token, error) {
	token := &models.Token{}
	query := `
		INSERT INTO tokens (device_id, user_id, hash)
		VALUES ($1, $2, $3)
		RETURNING id, hash, is_revoked, device_id, user_id, expired_at
	`
	if err := s.db.QueryRowContext(ctx, query, deviceId, userId, hash).Scan(
		&token.ID, &token.Hash, &token.IsRevoked, &token.DeviceId, &token.UserId, &token.ExpiredAt,
	); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *TokenRepository) GetToken(ctx context.Context, userId, deviceId uuid.UUID) (*models.Token, error) {
	token := &models.Token{}
	query := `
		SELECT id, hash, is_revoked, device_id, user_id, expired_at
		FROM tokens
		WHERE user_id=$1 AND device_id=$2
	`
	if err := s.db.QueryRowContext(ctx, query, userId, deviceId).Scan(
		&token.ID, &token.Hash, &token.IsRevoked, &token.DeviceId, &token.UserId, &token.ExpiredAt,
	); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *TokenRepository) Remove(ctx context.Context, userId, deviceId uuid.UUID) error {
	query := `DELETE FROM tokens WHERE user_id=$1 AND device_id=$2`
	result, err := s.db.ExecContext(ctx, query, userId, deviceId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return nil
	}
	return nil
}
