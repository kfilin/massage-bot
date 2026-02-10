# Part 1: Centralise All Runtime‑Configurable Values

## Goal

Centralize all configuration (environment variables, defaults) into a single `config` package to remove hard-coded globals and improve testability.

## 1.1 Implementation Plan (Source: Refactoring Proposals.md)

### 1.1.1 Create a `config` package

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

### 1.1.2 Remove duplicated globals

* Delete `SlotDuration`, `ApptTimeZone`, `WorkDayStartHour`, `WorkDayEndHour` **init** block from `internal/domain/models.go` and `internal/services/appointment/service.go`.
* Keep thin **proxy functions** in `domain/models.go` for backward compatibility.
* Update imports to include `github.com/kfilin/massage-bot/internal/config`.

### 1.1.3 Update business logic

* In `services/appointment/service.go`, replace globals with `config.Default`.
* Ensure `NowFunc` checks use `config.Default.ApptTimeZone`.

## 1.2 Solution Status

### **Completed**

* **Architecture**: Moved from global configuration variables to Dependency Injection (DI).
* **New Package**: All configuration logic is now in `internal/config`, defined in `Config` struct.
* **Services**: `AppointmentService`, `ReminderService`, `PostgresRepository`, and Telegram `BookingHandler` now accept `*config.Config` in their constructors.
* **Cleanup**: Legacy global variables (`ApptTimeZone`, `SlotDuration`, `WorkDayStartHour`, `WorkDayEndHour`) have been removed from `internal/domain`.
* **Entry Point**: `cmd/bot/main.go` loads the config via `config.LoadConfig()` and injects it into all dependencies.
* **Verification**: The entire test suite (`go test ./...`) passes.
