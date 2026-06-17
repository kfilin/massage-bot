package googlecalendar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// withEnv sets env vars for the duration of the test and restores on cleanup.
func withEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	old := make(map[string]string, len(kv))
	for k, v := range kv {
		old[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
	t.Cleanup(func() {
		for k, v := range old {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	})
}

// TestGetToken_FromEnvValid covers the happy path: env var holds valid
// JSON with both AccessToken and RefreshToken. Expiry stays as parsed.
func TestGetToken_FromEnvValid(t *testing.T) {
	tok := oauth2.Token{
		AccessToken:  "ya29.access",
		RefreshToken: "1//refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}
	b, _ := json.Marshal(tok)
	withEnv(t, map[string]string{"GOOGLE_TOKEN_JSON": string(b)})

	got, err := getToken(nil) // config is unused on the env path
	if err != nil {
		t.Fatalf("getToken: %v", err)
	}
	if got.AccessToken != tok.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", got.AccessToken, tok.AccessToken)
	}
	if got.RefreshToken != tok.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", got.RefreshToken, tok.RefreshToken)
	}
}

// TestGetToken_FromEnvZeroExpiryForcesRefresh covers the branch where
// Expiry is zero but RefreshToken is present: Expiry is forced into
// the past so the TokenSource refreshes the token.
func TestGetToken_FromEnvZeroExpiryForcesRefresh(t *testing.T) {
	tok := oauth2.Token{
		AccessToken:  "ya29.access",
		RefreshToken: "1//refresh",
		TokenType:    "Bearer",
		// Expiry omitted
	}
	b, _ := json.Marshal(tok)
	withEnv(t, map[string]string{"GOOGLE_TOKEN_JSON": string(b)})

	got, err := getToken(nil)
	if err != nil {
		t.Fatalf("getToken: %v", err)
	}
	if !got.Expiry.Before(time.Now()) {
		t.Errorf("Expiry should be in the past, got %v", got.Expiry)
	}
}

// TestSaveToken_RoundTrip writes a token and reads it back, verifying
// the file is created with the expected JSON.
func TestSaveToken_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "token.json") // saveToken doesn't create parent dirs

	want := &oauth2.Token{
		AccessToken:  "ya29.test",
		RefreshToken: "1//test",
		TokenType:    "Bearer",
		Expiry:       time.Now().Truncate(time.Second),
	}

	saveToken(path, want)

	// Read it back via tokenFromFile semantics.
	got, err := tokenFromFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.AccessToken != want.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", got.AccessToken, want.AccessToken)
	}
	if got.RefreshToken != want.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", got.RefreshToken, want.RefreshToken)
	}

	// Verify file mode is 0600.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Errorf("file mode: got %o, want 0600", mode)
	}
}
