package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type IRedisRepository interface {
	HSet(key string, data map[string]any, expiry time.Duration) error
	HGet(key string, field string) (string, error)
}

type redisRepository struct {
	rdb *redis.Client
}

func NewRedisRepository(rdb *redis.Client) IRedisRepository {
	return &redisRepository{
		rdb: rdb,
	}
}

func (s *redisRepository) HSet(key string, data map[string]any, expiration time.Duration) error {
	ctx := context.Background()
	err := s.rdb.HSet(ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash in Redis: %w", err)
	}
	// Set expiration if provided
	if expiration > 0 {
		err = s.rdb.Expire(ctx, key, expiration).Err()
		if err != nil {
			return fmt.Errorf("failed to set expiration: %w", err)
		}
	}
	return nil
}

func (s *redisRepository) HGet(key string, field string) (string, error) {
	ctx := context.Background()
	result, err := s.rdb.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("field %s not found: %w", field, err)
		}
		return "", fmt.Errorf("redis HGet failed: %w", err)
	}
	return result, nil
}
