package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var refreshToken string

var config *oauth2.Config

func SetGoogleRefreshToken(token string) {
	refreshToken = token
}

func SetGetGoogleOAuthConfig(cId, pId, cSe, ru string) {
	credentials := fmt.Sprintf(`{
		"installed": {
			"client_id": "%s",
			"project_id": "%s",
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://oauth2.googleapis.com/token",
			"client_secret": "%s",
			"redirect_uris": ["%s"]
		}
	}`, cId, pId, cSe, ru)
	cfg, err := google.ConfigFromJSON([]byte(credentials), gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Error parsing OAuth config: %v", err)
	}
	config = cfg
}

func GetTokenFromRefreshToken(config *oauth2.Config) *oauth2.Token {
	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("Unable to refresh token: %v", err)
	}
	return newToken
}

func SendEmail() error {
	token := GetTokenFromRefreshToken(config)
	client := config.Client(context.Background(), token)
	service, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create Gmail client: %v", err)
	}

	subject := "Subject: Test Email with OAuth2 in Go\n"
	body := "Hello! This is a test email sent using Gmail API in Go."
	rawMessage := base64.URLEncoding.EncodeToString([]byte("To: arridhaamrad@gmail.com\r\n" + subject + "\r\n" + body))

	message := &gmail.Message{Raw: rawMessage}

	_, err = service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return err
	}
	fmt.Println("Email sent successfully!")
	return nil
}
