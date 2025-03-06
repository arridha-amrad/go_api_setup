package database

import (
	"context"
	"database/sql"
	"my-go-api/internal/config"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Import the pgx driver
)

func Connect(cfg config.DbConfig) (*sql.DB, error) {

	db, err := sql.Open("pgx", cfg.DbUrl)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	duration, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
