package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
)

// Helper to generate valid hash for testing
func signInitData(data map[string]string, botToken string) string {
	// 1. Remove hash if present (should not be, but just in case)
	delete(data, "hash")

	// 2. Sort keys
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 3. Build data check string
	var dataCheckArr []string
	for _, k := range keys {
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, data[k]))
	}
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// 4. Calculate HMAC
	h1 := hmac.New(sha256.New, []byte("WebAppData"))
	h1.Write([]byte(botToken))
	secretKey := h1.Sum(nil)

	h2 := hmac.New(sha256.New, secretKey)
	h2.Write([]byte(dataCheckString))
	return hex.EncodeToString(h2.Sum(nil))
}

func TestValidateInitData(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	tests := []struct {
		name         string
		data         map[string]string
		wantID       string
		wantName     string
		wantErr      bool
		overrideHash string // If set, use this hash instead of calculating valid one
		removeUser   bool
	}{
		{
			name: "Valid Data",
			data: map[string]string{
				"query_id":  "AAGL...",
				"user":      `{"id":12345,"first_name":"John","last_name":"Doe","username":"jdoe"}`,
				"auth_date": "1672531200",
			},
			wantID:   "12345",
			wantName: "John Doe",
			wantErr:  false,
		},
		{
			name: "Valid Data - No Last Name",
			data: map[string]string{
				"query_id":  "AAGL...",
				"user":      `{"id":54321,"first_name":"Jane","last_name":""}`,
				"auth_date": "1672531200",
			},
			wantID:   "54321",
			wantName: "Jane",
			wantErr:  false,
		},
		{
			name: "Invalid Hash",
			data: map[string]string{
				"query_id": "AAGL...",
				"user":     `{"id":12345,"first_name":"John"}`,
			},
			overrideHash: "invalid_hash_value",
			wantErr:      true,
		},
		{
			name: "Missing User Data",
			data: map[string]string{
				"query_id": "AAGL...",
				// user key missing
			},
			wantID:     "",
			wantErr:    true,
			removeUser: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare query params
			params := url.Values{}
			for k, v := range tt.data {
				params.Set(k, v)
			}

			// Calculate valid hash if not constructing invalid test case
			hash := tt.overrideHash
			if hash == "" {
				hash = signInitData(tt.data, botToken)
			}
			params.Set("hash", hash)

			initData := params.Encode()
			// If we want to simulate missing hash parameter entirely, we would manipulate initData string
			// but here we test validation logic mostly.

			// Execute
			gotID, gotName, err := validateInitData(initData, botToken)

			// Verify
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInitData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotID != tt.wantID {
					t.Errorf("validateInitData() gotID = %v, want %v", gotID, tt.wantID)
				}
				if gotName != tt.wantName {
					t.Errorf("validateInitData() gotName = %v, want %v", gotName, tt.wantName)
				}
			}
		})
	}
}

func TestValidateHMAC(t *testing.T) {
	secret := "my_secret_key"
	id := "123456789"

	validToken := generateHMAC(id, secret)

	if !validateHMAC(id, validToken, secret) {
		t.Errorf("validateHMAC failed for valid token")
	}

	if validateHMAC(id, "invalid_token", secret) {
		t.Errorf("validateHMAC succeeded for invalid token")
	}

	if validateHMAC("wrong_id", validToken, secret) {
		t.Errorf("validateHMAC succeeded for wrong ID")
	}
}
