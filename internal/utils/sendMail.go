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

func (u *utility) CreateGoogleOauth2Config() *oauth2.Config {
	credentials := fmt.Sprintf(`{
		"installed": {
			"client_id": "%s",
			"project_id": "%s",
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://oauth2.googleapis.com/token",
			"client_secret": "%s",
			"redirect_uris": ["%s"]
		}
	}`, u.google.ClientId, u.google.ProjectId, u.google.ClientSecret, u.appUri)
	cfg, err := google.ConfigFromJSON([]byte(credentials), gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Error parsing OAuth config: %v", err)
	}
	return cfg
}

func (u *utility) GetTokenFromRefreshToken(config *oauth2.Config) *oauth2.Token {

	token := &oauth2.Token{RefreshToken: u.google.RefreshToken}
	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("Unable to refresh token: %v", err)
	}
	return newToken
}

func (u *utility) SendEmailWithGmail(subject, body, address string) error {
	config := u.CreateGoogleOauth2Config()
	token := u.GetTokenFromRefreshToken(config)
	client := config.Client(context.Background(), token)
	service, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create Gmail client: %v", err)
	}

	s := "Subject: " + subject + "\n"

	rawMessage := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("To: %s\r\n", address) + s + "\r\n" + body))

	message := &gmail.Message{Raw: rawMessage}

	_, err = service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return err
	}
	fmt.Println("Email sent successfully!")
	return nil
}
