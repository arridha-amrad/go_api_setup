package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type Config struct {
	DB           DbConfig
	Port         string
	SecretKey    string
	GoogleOAuth2 GoogleOAuth2Config
	AppUri       string
}

type DbConfig struct {
	DbUrl        string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type GoogleOAuth2Config struct {
	ProjectId    string
	ClientId     string
	ClientSecret string
	RefreshToken string
}

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
		AppUri:    os.Getenv("APP_URI"),
		Port:      os.Getenv("PORT"),
		SecretKey: os.Getenv("SECRET_KEY"),
		GoogleOAuth2: GoogleOAuth2Config{
			ProjectId:    os.Getenv("GOOGLE_PROJECT_ID"),
			ClientId:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RefreshToken: os.Getenv("GOOGLE_REFRESH_TOKEN"),
		},
	}
	return cfg, nil
}

func SetGetGoogleOAuthConfig(cId, pId, cSe, ru string) *oauth2.Config {
	credentials := fmt.Sprintf(`{
		"installed": {
			"client_id": %s
			"project_id": %s
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://oauth2.googleapis.com/token",
			"client_secret": %s
			"redirect_uris": %s
		}
	}`, cId, pId, cSe, ru)
	config, err := google.ConfigFromJSON([]byte(credentials), gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Error parsing OAuth config: %v", err)
	}
	return config
}
