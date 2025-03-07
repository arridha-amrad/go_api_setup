package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DB        DbConfig
	Port      string
	SecretKey string
}

type DbConfig struct {
	DbUrl        string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

var SECRET_KEY string

func LoadEnv() (*Config, error) {
	env := os.Getenv("GO_ENV")

	envFile := ".env.prod"

	if env == "development" {
		envFile = ".env.dev"
	}

	if err := godotenv.Load(envFile); err != nil {
		return nil, err
	}

	vMaxOpenConns, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	if err != nil {
		return nil, err
	}
	vMaxIdleConns, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DB: DbConfig{
			DbUrl:        os.Getenv("DB_URL"),
			MaxOpenConns: vMaxOpenConns,
			MaxIdleConns: vMaxIdleConns,
			MaxIdleTime:  os.Getenv("DB_MAX_IDLE_TIME"),
		},
		Port:      os.Getenv("PORT"),
		SecretKey: os.Getenv("SECRET_KEY"),
	}

	return cfg, nil
}
