# Refactoring Proposals

## 1. Goal: Centralise All Runtime‑Configurable Values

Centralize all configuration (environment variables, defaults) into a single `config` package to remove hard-coded globals and improve testability.

### 1.1 Implementation Plan

#### 1.1.1 Create a `config` package

* **File**: `internal/config/config.go`
* **Struct**: `Config` struct holding:
  * `WorkDayStartHour` (int)
  * `WorkDayEndHour` (int)
  * `ApptTimeZone` (*time.Location)
  * `SlotDuration` (time.Duration)
  * `CacheTTL` (time.Duration)
* **Mechanism**: Reads values from environment variables (`WORKDAY_START_HOUR`, `WORKDAY_END_HOUR`, `APPT_TIMEZONE`, `APPT_SLOT_DURATION`, `APPT_CACHE_TTL`).
* **Logging**: Logs effective configuration at startup.
* **Testing Helper**: Provides a `With(cfg *Config) func()` helper for tests.

#### 1.1.2 Remove duplicated globals

* Delete `SlotDuration`, `ApptTimeZone`, `WorkDayStartHour`, `WorkDayEndHour` **init** block from `internal/domain/models.go` and `internal/services/appointment/service.go`.
* Keep thin **proxy functions** in `domain/models.go` for backward compatibility.
* Update imports to include `github.com/kfilin/massage-bot/internal/config`.

#### 1.1.3 Update business logic

* In `services/appointment/service.go`, replace globals with `config.Default`.
* Ensure `NowFunc` checks use `config.Default.ApptTimeZone`.

---

## 2. Goal: Eliminate Hard-Coded Secrets

Remove all hard-coded secrets (API keys, tokens) from the source code and move them to environment variables.

### 2.1 Implementation Plan

#### 2.1.1 Actions

* **Search**: Identify secret-like strings (AWS, Google, Telegram tokens) using `rg -i "(AKIA|AIza|[0-9A-Za-z]{35,})"`.
* **Externalize**: Replace literals with `os.Getenv("VAR_NAME")` (e.g., `TELEGRAM_BOT_TOKEN`).
* **GitIgnore**: Ensure files like `credentials.json` are ignored.
* **Document**: Add "Configuration" section to README.md.
* **Validate**: Add startup checks for mandatory secrets in `config/init()`.

---

## 3. Goal: Raise Test Coverage to ~80%

Increase unit test coverage to ~80% to ensure stability during refactoring.

### 3.1 Implementation Plan

#### 3.1.1 Target Tests

* **Google Calendar Adapter** (`adapter.go`): Mock `calendar.Service`, verify `Create` and error handling.
* **Appointment Service** (`service.go`): Table-driven tests for `GetAvailableTimeSlots`.
* **Telegram Delivery** (`booking.go`): Stub `BotAPI` to verify user-visible error messages.
* **Metrics** (`monitoring/metrics.go`): Verify counter increments.
* **Config Loading**: Verify env loading.

#### 3.1.2 Implementation Steps

1. Add `internal/mocks` package.
2. Write table-driven test files.
3. Run `go test -cover ./...`.

---

## 4. Goal: Introduce Structured, PII‑Safe Logging

Replace standard `log` package with a structured logger (`zap`) that automatically redacts Personal Identifiable Information (PII).

### 4.1 Implementation Plan

#### 4.1.1 Implementation Steps

* **Choose Logger**: Use `go.uber.org/zap`.
* **Wrapper**: Create `internal/logging/logger.go` exposing `Infof`, `Warnf`, `Errorf`, `Debugf`.
* **Redaction**: Implement `redactPII(s string)` using regex `\d{9,}` to hide Telegram IDs.
* **Migration**: Replace all `log.Printf` calls with `logging.Get().Info/Error`.
* **Env Control**: control verbosity via `LOG_LEVEL` (INFO in prod, DEBUG in dev).

---

## 5. Goal: Cache Free‑Busy Results (short‑term)

Implement a short-term cache for free-busy results to reduce the number of calls to the Google Calendar API, improving performance and avoiding rate limits.

### 5.1 Implementation Plan

1. **Analyze & Critique**: Evaluate the proposed caching strategy in `internal/services/appointment/service.go`.
2. **Setup Metrics**: Add `FreeBusyCacheHits` and `FreeBusyCacheMisses` to the monitoring package.
3. **Implement Caching**: Add the cache structure and logic to the `appointment.Service`.
4. **Testing**: Verify caching logic with unit tests and mocks.

---

## 6. Goal: Add Project Documentation & CI

Establish a robust foundation for the project with comprehensive documentation, automated CI/CD pipelines, and health monitoring.

### 6.1 Implementation Plan

1. **Analyze & Critique**: Review existing documentation and potential CI integration points.
2. **README.md**: Create a comprehensive README with architecture, setup, and contribution guides.
3. **CI Workflow**: Implement `ci.yml` for automated linting, testing, and Docker builds.
4. **Health Check**: Add `/healthz` endpoint for liveness probes.

---

## 7. Goal: Refactor Service Into Smaller Components

Refactor the monolithic `appointment.Service` into smaller, single-responsibility components to improve testability and maintainability.

### 7.1 Implementation Plan

1. **Analyze & Critique**: Evaluate the current service structure and propose a decoupling strategy.
2. **SlotEngine**: Extract pure business logic (availability calculation) into a separate component.
3. **MetricsCollector**: Decouple Prometheus instrumentation via an interface.
4. **Service Coordination**: Update the main Service to coordinate these components.

---

## 8. Goal: Redact Sensitive Information From Logs

Implement comprehensive PII (Personally Identifiable Information) redaction in logs to protect user privacy (Telegram IDs, Emails, Phone Numbers).

### 8.1 Implementation Plan

1. **Analyze & Critique**: Identify gaps in the existing redaction (which only covered Telegram IDs).
2. **Advanced Redaction**: Support Emails and International Phone Numbers.
3. **Full Wrapper**: Ensure all logging methods (Info, Debug, Error, etc.) go through redaction.
4. **Verification**: Add tests for the new redaction logic.

---

## 9. Goal: Add a Graceful‑Shutdown Path

Implement a graceful shutdown mechanism to ensure the application cleans up resources (DB connections, background workers) and flushes logs/metrics before exiting.

### 9.1 Implementation Plan

1. **Analyze & Critique**: The app currently uses `select{}` blocking and `ListenAndServe`, which precludes graceful exit.
2. **Signal Handling**: Listen for SIGINT/SIGTERM.
3. **Orchestration**: Use `sync.WaitGroup` to wait for components.
4. **Component Refactor**: Update HTTP servers and Bot to accept Context cancellation.

---

## 10. Goal: Final Verification & Deliverable Audit

Perform a final audit of the entire refactoring project to ensure all goals are met and artifacts are delivered.

### 10.1 Implementation Plan

1. **Goal Coverage Audit**: Verify implementation of Goals 1-9.
2. **Final Integrity Check**: Verify tests, coverage, and infrastructure.
3. **Documentation Deliverables**: Ensure README and Refactoring history are complete.
