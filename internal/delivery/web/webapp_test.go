package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/presentation"
)

func TestGenerateHMAC(t *testing.T) {
	secret := "test-secret"
	id := "12345"

	token := generateHMAC(id, secret)

	if len(token) == 0 {
		t.Fatal("Expected non-empty HMAC token")
	}

	token2 := generateHMAC(id, secret)
	if token != token2 {
		t.Error("HMAC generation is not deterministic")
	}

	token3 := generateHMAC("99999", secret)
	if token == token3 {
		t.Error("Different IDs should produce different tokens")
	}
}

func TestValidateHMAC_Valid(t *testing.T) {
	secret := "test-secret"
	id := "12345"

	token := generateHMAC(id, secret)
	if !validateHMAC(id, token, secret) {
		t.Error("Expected valid HMAC")
	}
}

func TestValidateHMAC_Invalid(t *testing.T) {
	secret := "test-secret"
	id := "12345"

	if validateHMAC(id, "wrong-token", secret) {
		t.Error("Expected invalid HMAC")
	}
}

func TestValidateInitData_MissingHash(t *testing.T) {
	initData := "user=12345"
	_, _, err := validateInitData(initData, "bot-token")
	if err == nil {
		t.Error("Expected error for missing hash")
	}
}

func TestValidateInitData_InvalidHash(t *testing.T) {
	initData := "hash=invalidhash&user=12345"
	_, _, err := validateInitData(initData, "bot-token")
	if err == nil {
		t.Error("Expected error for invalid hash")
	}
}

func TestValidateInitData_MissingUser(t *testing.T) {
	token := "test-bot-token"
	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", time.Now().Unix()),
	}
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var arr []string
	for _, k := range keys {
		arr = append(arr, k+"="+data[k])
	}
	checkString := strings.Join(arr, "\n")

	h1 := hmac.New(sha256.New, []byte("WebAppData"))
	h1.Write([]byte(token))
	secret := h1.Sum(nil)

	h2 := hmac.New(sha256.New, secret)
	h2.Write([]byte(checkString))
	hash := hex.EncodeToString(h2.Sum(nil))

	data["hash"] = hash
	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	_, _, err := validateInitData(initData, token)
	if err == nil {
		t.Error("Expected error for missing user data")
	}
}

func TestValidateInitData_MissingUserJSON(t *testing.T) {
	token := "test-bot-token"
	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", time.Now().Unix()),
		"user":      "not-valid-json",
	}
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var arr []string
	for _, k := range keys {
		arr = append(arr, k+"="+data[k])
	}
	checkString := strings.Join(arr, "\n")

	h1 := hmac.New(sha256.New, []byte("WebAppData"))
	h1.Write([]byte(token))
	secret := h1.Sum(nil)

	h2 := hmac.New(sha256.New, secret)
	h2.Write([]byte(checkString))
	hash := hex.EncodeToString(h2.Sum(nil))

	data["hash"] = hash
	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	_, _, err := validateInitData(initData, token)
	if err == nil {
		t.Error("Expected error for invalid user JSON")
	}
}

func TestValidateInitData_Success(t *testing.T) {
	token := "test-bot-token"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe"}`
	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", time.Now().Unix()),
		"user":      userJSON,
	}
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var arr []string
	for _, k := range keys {
		arr = append(arr, k+"="+data[k])
	}
	checkString := strings.Join(arr, "\n")

	h1 := hmac.New(sha256.New, []byte("WebAppData"))
	h1.Write([]byte(token))
	secret := h1.Sum(nil)

	h2 := hmac.New(sha256.New, secret)
	h2.Write([]byte(checkString))
	hash := hex.EncodeToString(h2.Sum(nil))

	data["hash"] = hash
	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	userID, fullName, err := validateInitData(initData, token)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if userID != "12345" {
		t.Errorf("Expected user ID '12345', got '%s'", userID)
	}
	if fullName != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", fullName)
	}
}

func TestValidateInitData_EmptyFirstNameLastName(t *testing.T) {
	token := "test-bot-token"
	userJSON := `{"id":99999,"first_name":"","last_name":""}`
	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", time.Now().Unix()),
		"user":      userJSON,
	}
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var arr []string
	for _, k := range keys {
		arr = append(arr, k+"="+data[k])
	}
	checkString := strings.Join(arr, "\n")

	h1 := hmac.New(sha256.New, []byte("WebAppData"))
	h1.Write([]byte(token))
	secret := h1.Sum(nil)

	h2 := hmac.New(sha256.New, secret)
	h2.Write([]byte(checkString))
	hash := hex.EncodeToString(h2.Sum(nil))

	data["hash"] = hash
	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	_, fullName, err := validateInitData(initData, token)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if fullName != "Пациент" {
		t.Errorf("Expected fallback name 'Пациент', got '%s'", fullName)
	}
}

func TestSendTelegramMessage_APIError(t *testing.T) {
	sendTelegramMessage("invalid-token", "123", "test message")
}

func TestValidateInitData_MissingAuthDate(t *testing.T) {
	token := "test-bot-token"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe"}`
	data := map[string]string{
		"user": userJSON,
	}
	hash := signTestInitData(data, token)
	data["hash"] = hash

	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	_, _, err := validateInitData(initData, token)
	if err == nil {
		t.Fatal("Expected error for missing auth_date, got nil")
	}
	if !strings.Contains(err.Error(), "auth_date") {
		t.Errorf("Expected error mentioning auth_date, got: %v", err)
	}
}

func TestValidateInitData_InvalidAuthDate(t *testing.T) {
	token := "test-bot-token"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe"}`
	data := map[string]string{
		"auth_date": "not-a-number",
		"user":      userJSON,
	}
	hash := signTestInitData(data, token)
	data["hash"] = hash

	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	_, _, err := validateInitData(initData, token)
	if err == nil {
		t.Fatal("Expected error for invalid auth_date, got nil")
	}
	if !strings.Contains(err.Error(), "invalid auth_date") {
		t.Errorf("Expected error mentioning invalid auth_date, got: %v", err)
	}
}

func TestValidateInitData_ExpiredAuthDate(t *testing.T) {
	token := "test-bot-token"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe"}`
	// 2 hours ago — exceeds the 1-hour initDataMaxAge window.
	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", time.Now().Add(-2*time.Hour).Unix()),
		"user":      userJSON,
	}
	hash := signTestInitData(data, token)
	data["hash"] = hash

	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	_, _, err := validateInitData(initData, token)
	if err == nil {
		t.Fatal("Expected error for expired initData, got nil")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected error mentioning expiry, got: %v", err)
	}
}

func TestValidateInitData_FutureAuthDate(t *testing.T) {
	token := "test-bot-token"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe"}`
	// 1 hour in the future — exceeds the 5-minute clock-skew tolerance.
	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", time.Now().Add(1*time.Hour).Unix()),
		"user":      userJSON,
	}
	hash := signTestInitData(data, token)
	data["hash"] = hash

	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	_, _, err := validateInitData(initData, token)
	if err == nil {
		t.Fatal("Expected error for future auth_date, got nil")
	}
	if !strings.Contains(err.Error(), "future") {
		t.Errorf("Expected error mentioning future timestamp, got: %v", err)
	}
}

func TestValidateInitData_FreshAuthDate(t *testing.T) {
	token := "test-bot-token"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe"}`
	// 30 seconds ago — well within the 1-hour window.
	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", time.Now().Add(-30*time.Second).Unix()),
		"user":      userJSON,
	}
	hash := signTestInitData(data, token)
	data["hash"] = hash

	var parts []string
	for k, v := range data {
		parts = append(parts, k+"="+v)
	}
	initData := strings.Join(parts, "&")

	userID, fullName, err := validateInitData(initData, token)
	if err != nil {
		t.Fatalf("Expected success for fresh initData, got: %v", err)
	}
	if userID != "12345" || fullName != "John Doe" {
		t.Errorf("Unexpected result: id=%s name=%s", userID, fullName)
	}
}

func TestWebAppHandler_InvalidHMAC(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	secret := "test-secret"

	repo := &mockRepo{}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()

	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{}, secret)

	req, _ := http.NewRequest("GET", "/?id=12345&token=invalid-token", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 (loading page), got %d", rr.Code)
	}
}

func TestWebAppHandler_AdminNoTarget(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{
		patient: domain.Patient{TelegramID: adminID, Name: "Admin"},
		getAllPatientsFunc: func() ([]domain.Patient, error) {
			return []domain.Patient{
				{TelegramID: "200", Name: "Patient A"},
			}, nil
		},
	}
	service := &mockApptService{}
	presenter, _ := presentation.NewWebPresenter()

	handler := NewWebAppHandler(repo, service, presenter, botToken, []string{adminID}, "secret")

	initData := makeInitData(adminID, "Admin", botToken)
	req, _ := http.NewRequest("GET", "/?initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}
}

func TestNewSearchHandler_WithQuery(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{}
	handler := NewSearchHandler(repo, botToken, []string{adminID})
	initData := makeInitData(adminID, "Admin", botToken)

	req, _ := http.NewRequest("GET", "/api/search?q=test_patient&initData="+url.QueryEscape(initData), nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	var result []domain.Patient
	_ = fmt.Sprintf("result: %v", result)
}

func TestNewCancelHandler_NotFound(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	service := &mockApptService{appointments: map[string]domain.Appointment{}}
	presenter := presentation.NewBotPresenter()
	handler := NewCancelHandler(service, botToken, []string{adminID}, presenter)

	initData := makeInitData(adminID, "Admin", botToken)
	body := fmt.Sprintf(`{"id":"nonexistent","initData":"%s"}`, url.QueryEscape(initData))
	req, _ := http.NewRequest("POST", "/cancel", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Handler returns 400 for not found appointments
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusNotFound {
		t.Errorf("Expected 400 or 404, got %d", rr.Code)
	}
}

func TestDraftHandler_DiscardFlow(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{
		media: domain.PatientMedia{ID: "media-1", PatientID: "200", Transcript: "Test transcript"},
	}

	handler := NewDraftHandler(repo, botToken, []string{adminID}, "secret")
	initData := makeInitData(adminID, "Admin", botToken)
	body := map[string]string{
		"id":       "media-1",
		"initData": initData,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/draft/discard", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestNewUpdatePatientHandler_NotesTooLong(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	adminID := "100"

	repo := &mockRepo{}
	handler := NewUpdatePatientHandler(repo, botToken, []string{adminID})

	initData := makeInitData(adminID, "Admin", botToken)
	longNotes := strings.Repeat("x", 10001)
	body := fmt.Sprintf(`{"telegramID":"200","name":"Test","notes":"%s","initData":"%s"}`, longNotes, url.QueryEscape(initData))
	req, _ := http.NewRequest("POST", "/api/patient/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for too-long notes, got %d", rr.Code)
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
