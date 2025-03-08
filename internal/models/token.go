package models

import "github.com/google/uuid"

type Token struct {
	ID        int       `json:"id"`
	Hash      string    `json:"hash"`
	IsRevoked bool      `json:"is_revoked"`
	DeviceId  uuid.UUID `json:"device_id"`
	UserId    uuid.UUID `json:"user_id"`
	ExpiredAt string    `json:"expired_at"`
}
