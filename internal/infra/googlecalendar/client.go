package googlecalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func NewClient(ctx context.Context, credentialsFile string) (*calendar.Service, error) {
	config, err := getConfig(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	tok, err := getToken(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return calendar.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, tok)))
}

func getConfig(credentialsFile string) (*oauth2.Config, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	return google.ConfigFromJSON(b, calendar.CalendarScope)
}

func getToken(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	tokFile := "token.json"
	if tok, err := tokenFromFile(tokFile); err == nil {
		return tok, nil
	}

	tok, err := getTokenFromWeb(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := saveToken(tokFile, tok); err != nil {
		log.Printf("Warning: failed to cache token: %v", err)
	}
	return tok, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	return tok, json.NewDecoder(f).Decode(tok)
}

func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to this URL and authorize:\n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("failed to read auth code: %w", err)
	}

	return config.Exchange(ctx, code)
}

func saveToken(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}
