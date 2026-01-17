package googlecalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/kfilin/massage-bot/internal/monitoring"
)

// NewGoogleCalendarClient creates and authenticates a Google Calendar service client.
func NewGoogleCalendarClient() (*calendar.Service, error) {
	ctx := context.Background()

	var credsBytes []byte
	credsFromEnv := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credsFromEnv != "" {
		credsBytes = []byte(credsFromEnv)
		log.Println("Loaded Google credentials from GOOGLE_CREDENTIALS_JSON environment variable.")
	} else {
		var err error
		credsBytes, err = ioutil.ReadFile("credentials.json")
		if err != nil {
			return nil, fmt.Errorf("unable to read client secret file (credentials.json) and GOOGLE_CREDENTIALS_JSON not set: %v", err)
		}
		log.Println("Loaded Google credentials from credentials.json file.")
	}

	config, err := google.ConfigFromJSON(credsBytes, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file/env to config: %v", err)
	}

	// Explicitly set the RedirectURL to match our manual listener
	config.RedirectURL = "http://localhost:8080" // <--- IMPORTANT: Still setting this

	token, err := getToken(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get Google API token: %w", err)
	}

	// Update expiry metric
	// We try to catch the long-term refresh token expiry if available,
	// otherwise fallback to access token expiry.
	expiryDays := 180.0 // Default 6 months as per user expectation

	tokenJSON := os.Getenv("GOOGLE_TOKEN_JSON")
	var rawData struct {
		RefreshTokenExpiresIn float64 `json:"refresh_token_expires_in"`
	}
	if err := json.Unmarshal([]byte(tokenJSON), &rawData); err == nil && rawData.RefreshTokenExpiresIn > 0 {
		expiryDays = rawData.RefreshTokenExpiresIn / 86400
		log.Printf("DEBUG: Detected Refresh Token expiry in %.1f days", expiryDays)
	} else if !token.Expiry.IsZero() {
		expiryDays = time.Until(token.Expiry).Hours() / 24
		log.Printf("DEBUG: Falling back to Access Token expiry: %.2f hours", expiryDays*24)
	}

	client := config.Client(ctx, token)
	monitoring.UpdateTokenExpiry(expiryDays)

	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

// getToken retrieves a token from the environment variable or file, or performs initial authentication.
func getToken(config *oauth2.Config) (*oauth2.Token, error) {
	// Try to get token from environment variable first
	tokenFromEnv := os.Getenv("GOOGLE_TOKEN_JSON")
	if tokenFromEnv != "" {
		var tok oauth2.Token
		err := json.Unmarshal([]byte(tokenFromEnv), &tok)
		if err == nil {
			log.Println("Loaded Google token from GOOGLE_TOKEN_JSON environment variable.")
			// If Expiry is zero, it means it was likely pasted from a raw Google response
			// that uses "expires_in" instead of "expiry". We force it to the past
			// so that TokenSource will automatically refresh it using the RefreshToken.
			if tok.Expiry.IsZero() && tok.RefreshToken != "" {
				log.Println("DEBUG: Token expiry is missing, forcing refresh check.")
				tok.Expiry = time.Now().Add(-1 * time.Hour)
			}
			return &tok, nil
		}
		log.Printf("CRITICAL: Failed to unmarshal GOOGLE_TOKEN_JSON from env: %v. Raw length: %d", err, len(tokenFromEnv))
	} else {
		log.Println("DEBUG: GOOGLE_TOKEN_JSON environment variable is empty.")
	}

	// Fallback to data/token.json file for local development
	tok, err := tokenFromFile("data/token.json")
	if err == nil {
		log.Println("Loaded Google token from data/token.json file.")
		return tok, nil
	}
	log.Printf("Warning: Failed to load token from data/token.json: %v. Initiating new authentication.", err)

	// --- START MANUAL LISTENER FOR OAUTH CALLBACK ---
	authCodeChan := make(chan string) // Channel to receive the authorization code
	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux} // Server listening on port 8080

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "Authentication successful! You can close this tab.")
			authCodeChan <- code
		} else {
			fmt.Fprintf(w, "Authentication failed or code not found in redirect.")
		}
	})

	// Start the server in a goroutine so it doesn't block
	go func() {
		log.Printf("Listening for OAuth callback on %s", config.RedirectURL) // Removed / for cleaner logging
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			// This error means the server failed to start, e.g., port in use.
			// It's the key to debugging your "localhost refused to connect" problem.
			log.Fatalf("HTTP server ListenAndServe error: %v", err) // Use Fatalf to stop if server can't start
		}
	}()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)
	fmt.Println("Waiting for authorization code via http://localhost:8080...")

	var authCode string
	select {
	case authCode = <-authCodeChan:
		// Received the code from the web server
		if authCode == "" {
			return nil, fmt.Errorf("failed to receive authorization code via web server. Check logs for server errors.")
		}
	case <-time.After(5 * time.Minute): // Timeout after 5 minutes
		// --- FIX 1: Check error return value of server.Shutdown ---
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down HTTP server during timeout: %v", err)
		}
		// --- END FIX 1 ---
		return nil, fmt.Errorf("authorization timed out. No code received within 5 minutes.")
	}

	// Shut down the server gracefully after receiving the code
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// --- FIX 1 (applied again): Check error return value of server.Shutdown ---
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down HTTP server: %v", err)
		}
		// --- END FIX 1 ---
	}()

	tok, err = config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	// Save the new token to file (for local development)
	saveToken("data/token.json", tok)
	log.Println("New Google token generated and saved locally. Remember to update GOOGLE_TOKEN_JSON on Heroku!")

	return tok, nil
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	if err == nil && tok.Expiry.IsZero() && tok.RefreshToken != "" {
		tok.Expiry = time.Now().Add(-1 * time.Hour)
	}
	return tok, err
}

// saveToken saves a token to a file.
func saveToken(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	// --- FIX 2: Check error return value of Encode ---
	if err := json.NewEncoder(f).Encode(token); err != nil {
		log.Printf("Error encoding token to file: %v", err)
	}
	defer f.Close() // Ensure defer is after any potential error that would close f
}
