package models

import "github.com/google/uuid"

type Token struct {
	ID        int       `json:"id"`
	Value     string    `json:"value"`
	TokenId   uuid.UUID `json:"token_id"`
	IsRevoked bool      `json:"is_revoked"`
	UserId    string    `json:"user_id"`
	CreatedAt string    `json:"created_at"`
	ExpiredAt string    `json:"expired_at"`
}
