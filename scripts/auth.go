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

	fmt.Printf("\n=== ACTION REQUIRED ===\nGo to the following link in your browser:\n\n%v\n\n", authURL)
	fmt.Printf("After authorizing, you will be redirected to localhost:8080/?code=...\n")
	fmt.Printf("Copy the value of the 'code' parameter from your browser's URL bar.\n\n")
	
	fmt.Print("Paste the authorization code here: ")
	
	// Read from stdin for headless servers
	reader := bufio.NewReader(os.Stdin)
	authCode, err = reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read code: %v", err)
	}
	authCode = strings.TrimSpace(authCode)

	if authCode == "" {
		log.Fatalf("No code provided")
	}

	go server.Shutdown(context.Background())

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("unable to retrieve token from web: %v", err)
	}

	tokJson, _ := json.Marshal(tok)
	fmt.Printf("\n=== SUCCESS! YOUR NEW TOKEN ===\n\nGOOGLE_TOKEN_JSON='%s'\n\n", string(tokJson))
}
