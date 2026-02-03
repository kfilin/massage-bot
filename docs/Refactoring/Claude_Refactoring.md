# Refactoring & Test Coverage Implementation Plan

## Current State Analysis

| Package | Current Coverage | Target | LOC |
|---------|-----------------|--------|-----|
| `config` | 91.7% ✅ | 90%+ | ~80 |
| `services/appointment` | 80.7% ✅ | 85%+ | ~420 |
| `adapters/googlecalendar` | 26.2% 🟡 | 80%+ | ~380 |
| `logging` | 9.1% 🔴 | 80%+ | ~120 |
| `storage` | 0% 🔴 | 80%+ | ~780 |
| `domain` | 0% 🔴 | 80%+ | ~110 |
| `delivery/telegram` | 0% 🔴 | 60%+ | ~400 |
| `monitoring` | 0% 🔴 | 80%+ | ~50 |
| `cmd/bot` | 0% 🔴 | 40%+ | ~660 |

**Overall Target**: 80%+ project-wide coverage

---

## Phase 1: Critical Fixes (Same Day)

### 1.1 Fix CI/CD Go Version & Add Static Analysis

#### [MODIFY] [ci.yml](file:///home/kirillfilin/Documents/massage-bot/.github/workflows/ci.yml)

- Change line 24 from `go-version: 1.24` to `go-version: 1.23`
- Add `golangci-lint` step after vet
- Add `go test -race` for race condition detection

```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v3
  with:
    version: latest

- name: Run Race Detector
  run: go test -race -short ./...
```

---

### 1.2 Consolidate Version String

#### [NEW] [version.go](file:///home/kirillfilin/Documents/massage-bot/internal/version/version.go)

```go
package version

const (
    Version   = "v5.3.6"
    Edition   = "Clinical Edition"
    FullName  = Version + " " + Edition
)
```

#### [MODIFY] [main.go](file:///home/kirillfilin/Documents/massage-bot/cmd/bot/main.go)

- Remove hardcoded `"v5.3.6 Clinical Edition"` and `"v5.1.0"`
- Import `internal/version` and use `version.Version`
- Remove unreachable `time.Sleep` and second `Fatalf`

---

### 1.3 Remove Duplicate Handler

#### [MODIFY] [bot.go](file:///home/kirillfilin/Documents/massage-bot/internal/delivery/telegram/bot.go)

Remove duplicate line 184: `b.Handle("/ban", bookingHandler.HandleBan)`

---

## Phase 2: Logging Consolidation

### Files to Update (Replace `log.Printf` → `logging.*`)

| File | Occurrences | Effort |
|------|-------------|--------|
| `postgres_repository.go` | ~15 | Medium |
| `adapter.go` | ~5 | Low |
| `bot.go` | ~3 | Low |
| `webapp.go` | ~20 | Medium |

#### [MODIFY] [logger.go](file:///home/kirillfilin/Documents/massage-bot/internal/logging/logger.go)

Fix race condition in `Get()`:

```go
func Get() *zap.SugaredLogger {
    once.Do(func() {
        Init(os.Getenv("LOG_LEVEL") == "DEBUG")
    })
    return logger
}
```

---

## Phase 3: Test Coverage Implementation

### 3.0 Test Infrastructure Setup

#### [NEW] [test/helpers.go](file:///home/kirillfilin/Documents/massage-bot/test/helpers.go)

Shared test utilities:

```go
package test

func MustParseTime(s string) time.Time { ... }
func NewMockRepo() *MockRepo { ... }
```

#### [NEW] test/fixtures/

Directory for test data:

- `sample_calendar_event.json`
- `sample_patient.json`

---

### 3.1 Domain Tests

#### [NEW] [models_test.go](file:///home/kirillfilin/Documents/massage-bot/internal/domain/models_test.go)

```go
package domain

import "testing"

func TestSplitSummary(t *testing.T) {
    tests := []struct {
        name     string
        summary  string
        wantLen  int
        wantSvc  string
        wantName string
    }{
        {"Standard", "Massage - John Doe", 2, "Massage", "John Doe"},
        {"ServiceOnly", "Massage", 1, "Massage", ""},
        {"MultiDash", "Full Body Massage - Jane - Doe", 2, "Full Body Massage", "Jane - Doe"},
        {"Empty", "", 1, "", ""},
    }
    // ...table-driven test implementation
}

func TestTimeConstants(t *testing.T) {
    if WorkDayStartHour < 0 || WorkDayStartHour > 23 { t.Error("invalid start hour") }
    if WorkDayEndHour < 0 || WorkDayEndHour > 23 { t.Error("invalid end hour") }
    if ApptTimeZone == nil { t.Error("timezone not initialized") }
}
```

**Expected Coverage**: 80%+

---

### 3.2 Storage Tests (Largest Gap)

#### [NEW] [postgres_repository_test.go](file:///home/kirillfilin/Documents/massage-bot/internal/storage/postgres_repository_test.go)

Strategy: Use interface mocking for `*sqlx.DB` or use `sqlmock`:

```go
package storage

import (
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"
)

func TestSavePatient(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()
    sqlxDB := sqlx.NewDb(db, "sqlmock")
    
    repo := NewPostgresRepository(sqlxDB, t.TempDir())
    
    patient := domain.Patient{TelegramID: "123", Name: "Test"}
    
    mock.ExpectExec("INSERT INTO patients").WillReturnResult(sqlmock.NewResult(1, 1))
    
    err := repo.SavePatient(patient)
    if err != nil { t.Errorf("unexpected error: %v", err) }
}

func TestGetPatient_NotFound(t *testing.T) { ... }
func TestIsUserBanned(t *testing.T) { ... }
func TestSavePatientDocument(t *testing.T) { ... }
func TestGenerateHTMLRecord(t *testing.T) { ... }
func TestMdToHTML(t *testing.T) { ... }
```

**Key Functions to Cover**:

- `SavePatient`, `GetPatient` (DB operations)
- `mdToHTML` (markdown conversion)
- `GenerateHTMLRecord` (template generation)
- `SavePatientDocumentReader` (file operations)
- `CreateBackup` (requires mock `exec.Command`)

**Expected Coverage**: 80%+

---

### 3.3 Logging Tests

#### [MODIFY] [logger_test.go](file:///home/kirillfilin/Documents/massage-bot/internal/logging/logger_test.go)

Add tests for:

```go
func TestRedactPII(t *testing.T) {
    tests := []struct{input, expected string}{
        {"User 123456789 logged in", "User [REDACTED_ID] logged in"},
        {"Email: test@example.com", "Email: [REDACTED_EMAIL]"},
        {"Short 12345", "Short 12345"}, // Should NOT redact short numbers
    }
}

func TestInit_DebugMode(t *testing.T) { ... }
func TestInit_ProductionMode(t *testing.T) { ... }
func TestLoggerWrappers(t *testing.T) { ... }
```

**Expected Coverage**: 80%+

---

### 3.4 Google Calendar Adapter Tests

#### [MODIFY] [adapter_test.go](file:///home/kirillfilin/Documents/massage-bot/internal/adapters/googlecalendar/adapter_test.go)

Add tests for:

- `eventToAppointment` (currently not covered)
- `isNotFound` helper
- TGID extraction from description

**Expected Coverage**: 80%+

---

### 3.5 WebApp Tests

#### [NEW] [webapp_test.go](file:///home/kirillfilin/Documents/massage-bot/cmd/bot/webapp_test.go)

```go
package main

import "testing"

func TestGenerateHMAC(t *testing.T) {
    result := generateHMAC("12345", "secret")
    if result == "" { t.Error("empty HMAC") }
}

func TestValidateHMAC(t *testing.T) {
    secret := "test-secret"
    id := "user123"
    validToken := generateHMAC(id, secret)
    
    if !validateHMAC(id, validToken, secret) { t.Error("valid token rejected") }
    if validateHMAC(id, "invalid", secret) { t.Error("invalid token accepted") }
}

func TestValidateInitData(t *testing.T) { ... }
```

**Expected Coverage**: 60%+

---

### 3.6 Monitoring Tests

#### [NEW] [metrics_test.go](file:///home/kirillfilin/Documents/massage-bot/internal/monitoring/metrics_test.go)

```go
package monitoring

import "testing"

func TestMetricsRegistration(t *testing.T) {
    // Verify all metrics are properly registered
    if ApiRequestsTotal == nil { t.Error("ApiRequestsTotal not registered") }
    if BookingLeadTimeDays == nil { t.Error("BookingLeadTimeDays not registered") }
    // ...
}
```

**Expected Coverage**: 80%+

---

## Verification Plan

### Automated Tests

Run all tests with coverage:

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

**Success Criteria**: `total: (statements) 80.0%` or higher

### Per-Package Coverage Check

```bash
go test -cover ./internal/domain
go test -cover ./internal/storage
go test -cover ./internal/logging
go test -cover ./internal/adapters/googlecalendar
go test -cover ./internal/monitoring
go test -cover ./cmd/bot
```

### CI Verification

After Phase 1 completion, push to trigger CI:

```bash
git push origin feature/refactoring
```

**Success Criteria**: GitHub Actions passes on Go 1.23

---

## Estimated Effort

| Phase | Effort | Priority |
|-------|--------|----------|
| Phase 1: Critical Fixes | 1 hour | P0 |
| Phase 2: Logging Consolidation | 2 hours | P1 |
| Phase 3: Test Coverage | 6-8 hours | P2 |
| **Total** | **~10 hours** | |

---

## Dependencies

- **sqlmock**: `go get github.com/DATA-DOG/go-sqlmock` (for storage tests)
- No other new dependencies required

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Markdown sync breaks | Use `t.TempDir()` for isolated file tests |
| Google API mocking complex | Focus on `eventToAppointment` unit tests |
| Telegram handler testing | Mock `telebot.Context` interface |
