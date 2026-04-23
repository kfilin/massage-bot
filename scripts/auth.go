package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func main() {
	godotenv.Load(".env") // Try to load .env from parent dir
	
	credsFromEnv := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	var credsBytes []byte
	if credsFromEnv != "" {
		credsBytes = []byte(credsFromEnv)
	} else {
		var err error
		credsBytes, err = os.ReadFile("credentials.json")
		if err != nil {
			log.Fatalf("No credentials found: %v", err)
		}
	}

	config, err := google.ConfigFromJSON(credsBytes, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v", err)
	}

	config.RedirectURL = "http://localhost:8080"

	authCodeChan := make(chan string)
	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "Authentication successful! You can close this tab.")
			authCodeChan <- code
		} else {
			fmt.Fprintf(w, "Authentication failed or code not found in redirect.")
		}
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe error: %v", err)
		}
	}()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("\n=== ACTION REQUIRED ===\nGo to the following link in your browser:\n\n%v\n\nWaiting for authorization...\n", authURL)

	var authCode string
	select {
	case authCode = <-authCodeChan:
	case <-time.After(5 * time.Minute):
		log.Fatalf("authorization timed out")
	}

	go server.Shutdown(context.Background())

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("unable to retrieve token from web: %v", err)
	}

	tokJson, _ := json.Marshal(tok)
	fmt.Printf("\n=== SUCCESS! YOUR NEW TOKEN ===\n\nGOOGLE_TOKEN_JSON='%s'\n\n", string(tokJson))
}
