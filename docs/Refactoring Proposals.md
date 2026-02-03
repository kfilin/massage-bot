# Detailed Summary of the “massage‑bot” Improvement Plan

Everything is broken down into **high‑level goals** → **concrete tasks** → **implementation notes** (code locations, example snippets, commands to run).

---

## 1. Goal: Centralise All Runtime‑Configurable Values

### 1.1 Create a `config` package

| File | Path | Content (short description) |
| :--- | :--- | :--- |
| `internal/config/config.go` | **new** | • Defines a `Config` struct that holds: • `WorkDayStartHour` (int); • `WorkDayEndHour` (int); • `ApptTimeZone` (*time.Location); • `SlotDuration` (time.Duration); • `CacheTTL` (time.Duration); • Reads values from environment variables (`WORKDAY_START_HOUR`, `WORKDAY_END_HOUR`, `APPT_TIMEZONE`, `APPT_SLOT_DURATION`, `APPT_CACHE_TTL`). • Logs the effective configuration at startup. • Provides a `With(cfg *Config) func()` helper for tests (temporary override). |
| `internal/config/config_test.go` | **new** | Minimal test that sets env vars, calls `init()`, and checks that the fields contain the expected values. |

### Key design points

* All values are loaded **once** in the package `init()`; they are immutable after that (except via the test helper).
* Default fall‑backs replicate the current hard‑coded defaults (9 AM‑6 PM, “Europe/Istanbul”, 60 min slot, 2 min cache TTL).
* Errors parsing the environment are logged as warnings, not fatal (except for a completely missing timezone – we fall back to UTC).

### 1.2 Remove duplicated globals

* Delete the `SlotDuration`, `ApptTimeZone`, `WorkDayStartHour`, `WorkDayEndHour` **init** block from `internal/domain/models.go` and from `internal/services/appointment/service.go`.
* Keep thin **proxy functions** in `domain/models.go` for backward compatibility (e.g., `func WorkDayStartHour() int { return config.Default.WorkDayStartHour }`).
* Update **all imports** to include `github.com/kfilin/massage-bot/internal/config` where needed (service, adapters, tests).

### 1.3 Update the business logic to use the new config

* In `services/appointment/service.go` replace every reference to the old globals with the corresponding `config.Default` value (or the proxy functions).
* Where the code previously used the globals directly (e.g., `loc := ApptTimeZone`), change to `loc := config.Default.ApptTimeZone`.
* Ensure the `NowFunc`‑based time checks also use `config.Default.ApptTimeZone`.

### 1.4 Run the test suite to verify compilation

```bash
go test ./... -run TestConfig
```

If any compile‑time errors arise (missing imports, mismatched names), fix them and re‑run until green.

---

## 2. Goal: Eliminate Hard‑Coded Secrets

| Action | Details |
| :--- | :--- |
| **Search for secret‑like strings** | `rg -i "(AKIA\|AIza\|[0-9A-Za-z]{35,})" .` — list any tokens that look like AWS keys, Google API keys, or Telegram bot tokens. |
| **Move tokens to environment variables** | Replace literals such as `telegram_api_data/8037374978:AAFPG1Gc…` with `os.Getenv("TELEGRAM_BOT_TOKEN")`. Add the env‑var read in `internal/delivery/telegram/bot.go` (or wherever the token is currently loaded). |
| **Add to `.gitignore`** | Ensure any file that still contains a secret (e.g., `credentials.json`) is listed in `.gitignore` to prevent accidental commits. |
| **Document required env variables** | Extend the new `README.md` (see §7) with a “Configuration” section that lists all env vars (`WORKDAY_START_HOUR`, `APPT_TIMEZONE`, `TELEGRAM_BOT_TOKEN`, `GOOGLE_CALENDAR_ID`, etc.). |
| **Validate at start‑up** | In `config/init()` (or a small `security` helper) check that mandatory secrets are not empty; if they are, `log.Fatalf` with a clear message. |

---

## 3. Goal: Raise Test Coverage to ~80 %

| Area | Current state | Target tests | Example test ideas |
| :--- | :--- | :--- | :--- |
| **Google Calendar adapter** (`adapter.go`) | 0 % | 90 % | – Mock `calendar.Service` (use an interface wrapper); – Verify `Create` builds the expected `Event` (summary, description, conference data); – Verify error handling (invalid StartTime, empty CalendarID). |
| **Appointment service – slot calculation** (`service.go`) | 0 % | 85 % | – Table‑driven tests for `GetAvailableTimeSlots` with a fake repository that returns a pre‑defined set of busy intervals; – Edge cases: slot crossing day‑boundary, slot exactly at end‑hour, slot overlapping a busy interval. |
| **Telegram delivery handlers** (`booking.go`, `debug_test.go`) | 0 % | 75 % | – Use the `telegram-bot-api` mock or a stubbed `BotAPI` to simulate `SendMessage`, `EditMessageReplyMarkup`, etc.; – Verify that invalid input yields the expected user‑visible error message. |
| **Metrics** (`monitoring/metrics.go`) | 0 % | 70 % | – Simple unit test that calls a metric (e.g., `ServiceBookingsTotal.WithLabelValues("test").Inc()`) and checks that the Prometheus collector has the correct value using `prometheus/testutil`. |
| **Config loading** (`config/config_test.go`) | – | 100 % | – Already covered in §1.2. |

### Coverage Implementation Steps

1. Add a `internal/mocks` package with interfaces for the Google Calendar client and the Telegram Bot (auto‑generated with `mockgen` or handwritten).
2. Write the table‑driven test files in the same directories (`*_test.go`).
3. Run `go test -cover ./...` and inspect coverage.
4. If any package stays under 70 % after the above, add a minimal test (even a simple “does not panic” call) to lift it.

### Running the full coverage suite

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out   # open in browser to verify %
```

---

## 4. Goal: Introduce Structured, PII‑Safe Logging

| Step | Details |
| :--- | :--- |
| **Choose a logger** | Add `go.uber.org/zap` (or `github.com/rs/zerolog`). Prefer Zap for its `SugaredLogger` for minimal code changes. |
| **Create a logger wrapper** (`internal/logging/logger.go`) | that: • Exposes `Infof`, `Warnf`, `Errorf`, `Debugf`; • Redacts fields known to be PII (`CustomerTgID`, `Notes`, `VoiceTranscripts`); • Adds a request‑ID field when a `context.Context` contains one (`ctx.Value("request_id")`). |
| **Replace all `log.Printf` calls** | with the wrapper (batch replace via `rg -l "log.Printf"` then edit). Example: `log.Printf("DEBUG: ...")` → `logger.Debugf(ctx, "...", ...)` |
| **Guard against overly‑verbose production logs** | Initialise the logger at **INFO** level in prod, **DEBUG** in dev (controlled by env var `LOG_LEVEL`). |
| **Add unit tests for the redaction** | Write a test that logs a struct containing a TG ID and asserts that the output string contains `"REDACTED"` instead of the real ID. |

---

## 5. Goal: Cache Free‑Busy Results (short‑term)

| Component | Change |
| :--- | :--- |
| **storage layer** (`internal/services/appointment/service.go`) | Add a struct field `freeBusyCache map[string]struct{slots []domain.TimeSlot; expires time.Time}` protected by `cacheMu sync.RWMutex`. |
| **GetFreeBusy wrapper** | In `GetAvailableTimeSlots` (and anywhere else that calls `repo.GetFreeBusy`), first look up the cache key `calendarID\|date`. If a valid entry exists (`expires.After(time.Now())`), return it; otherwise call the repo, store the result with`expires = time.Now().Add(config.Default.CacheTTL)`. |
| **Cache TTL** | Already present in the `config` struct (`CacheTTL`). Default 2 min, configurable via env var. |
| **Testing** | Add a unit test that injects a fake repo returning a fixed slice of busy slots, calls `GetAvailableTimeSlots` twice, and asserts that the repo’s method was invoked **once** (use a mock repository that counts calls). |
| **Metrics** | Increment a counter `FreeBusyCacheHits` / `FreeBusyCacheMisses` in `monitoring`. |

---

## 6. Goal: Add Project Documentation & CI

### 6.1 README.md

* **Introduction** – what the bot does (Telegram‑based appointment booking, Google Calendar integration).
* **Architecture diagram** – a simple PlantUML block diagram (delivery → service → ports → adapters).
* **Setup** – required env vars, how to obtain a Telegram bot token, how to create a Google service‑account JSON, how to configure the calendar ID.
* **Running locally** – `go run ./cmd/massage-bot` (or the binary).
* **Testing** – `go test ./...`, `go test -cover`.
* **Docker** – sample `docker build` and `docker run` commands.
* **Contributing** – linting (`golangci-lint run`), commit style, PR checklist.

### 6.2 GitHub Actions Workflow (`.github/workflows/ci.yml`)

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '^1.22'   # or latest stable
      - name: Install dependencies
        run: go mod tidy
      - name: Lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: v1.60
      - name: Run tests & coverage
        run: |
          go test ./... -coverprofile=coverage.out
          go tool cover -func=coverage.out
      - name: Build Docker image
        run: |
          docker build -t massage-bot:${{ github.sha }} .
```

*The workflow validates lint, unit tests, coverage, and builds a Docker image – all in under 5 minutes.*

### 6.3 Health‑Check Endpoint

* Add a small HTTP server (e.g., `net/http`) that serves `/healthz` with `200 OK` and JSON `{ "status":"ok" }`.
* Register it in `main.go` (or a new `internal/health` package).
* Use the same server for Prometheus metrics (`/metrics`).
* Hook into `context.Cancel` on SIGTERM so the HTTP server shuts down gracefully.

---

## 7. Goal: Refactor `Service` Into Smaller Components

| New sub‑components | Responsibility |
| :--- | :--- |
| `repoAdapter` (implements `ports.AppointmentRepository`) | All calls to Google Calendar / DB. |
| `slotEngine` (struct with `GetAvailableTimeSlots`, `IsSlotFree`) | Pure business logic, no external I/O. |
| `metricsCollector` (wraps `monitoring` calls) | Separate concern for metric instrumentation. |
| `Service` (public) | Holds the three components and coordinates them (`CreateAppointment`, `CancelAppointment`, etc.). |

### Service Refactoring Implementation Steps

1. Create three new files in `internal/services/appointment/` (`repo_adapter.go`, `slot_engine.go`, `metrics_collector.go`).
2. Define small interfaces for each (e.g., `type SlotEngine interface { GetAvailableTimeSlots(... ) ([]domain.TimeSlot,error) }`).
3. In `service.go`, replace the monolithic struct with:

```go
type Service struct {
    repo    ports.AppointmentRepository
    engine  SlotEngine
    metrics MetricsCollector
    NowFunc func() time.Time
    mu      sync.Mutex
}
```

1. Update the constructor:

```go
func NewService(repo ports.AppointmentRepository) *Service {
    return &Service{
        repo:    repo,
        engine:  NewSlotEngine(repo),
        metrics: NewMetricsCollector(),
        NowFunc: time.Now,
    }
}
```

1. Adjust all method bodies to delegate to the appropriate sub‑component.
2. Add unit tests for each sub‑component individually (they are easier to mock).

**Result** – the service is now **single‑responsibility**, testable in isolation, and future extensions (e.g., swapping Google Calendar for Outlook) only require a new `repoAdapter`.

---

## 8. Goal: Redact Sensitive Information From Logs

| Action | Code change |
| :--- | :--- |
| **Introduce a helper** `func redactPII(s string) string` | Replace any occurrence of a TG ID, email, phone number with `"[REDACTED]"`. Simple regex: `re := regexp.MustCompile(\d{9,})`. |
| **Use the helper** in the logger wrapper (see §4) | Before writing the final message. |
| **Add a test** | That feeds a string containing a TG ID and asserts the output contains `[REDACTED]`. |
| **Scan the codebase** | For any direct `fmt.Printf("%s", clientID)` statements that bypass the logger and replace them. |

---

## 9. Goal: Add a Graceful‑Shutdown Path

* In `main.go` (or wherever the bot is started) do:

```go
ctx, cancel := signal.NotifyContext(context.Background(),
    os.Interrupt, syscall.SIGTERM)
defer cancel()

// start telegram bot, http server, background workers with `ctx`
go startTelegramBot(ctx)
go startHTTPServer(ctx)

// block until ctx is cancelled
<-ctx.Done()
log.Println("Shutting down …")
```

* Ensure each long‑running goroutine listens to `ctx.Done()` and returns.
* For the HTTP server, use `srv.Shutdown(ctx)`.

---

## 10. Goal: Verify All Changes End‑to‑End

| Command | Expected result |
| :--- | :--- |
| `go test ./...` | **All tests pass**, coverage ≥ 80 % (≈ 85 %). |
| `golangci-lint run` | **No lint errors** (or only intentionally ignored ones). |
| `go build ./cmd/massage-bot` | **Successful build**; binary size unchanged (≈ 10 MB). |
| `docker build -t massage-bot .` | **Docker image builds** without errors. |
| `docker run -e TELEGRAM_BOT_TOKEN=… -e GOOGLE_CALENDAR_ID=primary -p 8080:8080 massage-bot` | Bot starts, logs “healthz ready”, and `/healthz` returns `200`. |
| `curl http://localhost:8080/healthz` | JSON `{ "status":"ok" }`. |
| `curl http://localhost:8080/metrics` | Prometheus metrics appear (including new `FreeBusyCacheHits`). |
| Manual test: send a Telegram `/book` command → verify a Google Calendar event is created. | Appointment appears in Google Calendar, bot replies with confirmation containing a **redacted** TG ID. |

If any step fails, edit the offending file, run the failing command again, and repeat until the whole suite passes.

---

## 11. Deliverables (what you will hand‑over)

1. **`internal/config`** package (code + tests).
2. Updated **domain** and **service** files that use the config.
3. **Secrets‑free** source (no token literals).
4. **Unit‑test suite** with ≥ 80 % coverage.
5. **Logger wrapper** with PII redaction.
6. **Free‑busy cache** implementation & metrics.
7. **README.md**, **CI workflow**, **Dockerfile**, **health‑check** endpoint.
8. **Refactored Service** split into `repoAdapter`, `slotEngine`, `metricsCollector`.
9. **Graceful‑shutdown** logic in `main.go`.
10. **Documentation** (README, CI notes, deployment steps).

---
