package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Import the pgx driver
)

func Connect(url, maxIdleTime string, maxOpenConns, maxIdleConns int) (*sql.DB, error) {

	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	duration, err := time.ParseDuration(maxIdleTime)
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
