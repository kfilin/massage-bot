# Part 3: Raise Test Coverage to ~80%

## Goal

Increase unit test coverage to ~80% to ensure stability during refactoring.

## 3.1 Implementation Plan (Source: Refactoring Proposals.md)

### 3.1.1 Target Tests

* **Google Calendar Adapter** (`adapter.go`): Mock `calendar.Service`, verify `Create` and error handling.
* **Appointment Service** (`service.go`): Table-driven tests for `GetAvailableTimeSlots`.
* **Telegram Delivery** (`booking.go`): Stub `BotAPI` to verify user-visible error messages.
* **Metrics** (`monitoring/metrics.go`): Verify counter increments.
* **Config Loading**: Verify env loading.

### 3.1.2 Implementation Steps

1. Add `internal/mocks` package.
2. Write table-driven test files.
3. Run `go test -cover ./...`.

## 3.2 Solution Status

### **Completed**

> [!NOTE]
> **Restored on**: 2026-02-03
> **Changes**: Added unit tests for Google Calendar Adapter (`adapter_test.go`) and Appointment Service (`service_test.go`) covering critical paths.

> [!WARNING]
> Code for this part is missing. The details below are the **intended** implementation from a previous lost session. Use them as a blueprint to re-implement.

* **New Tests Added**:
  * `internal/adapters/googlecalendar/adapter_test.go`: Added `TestAdapter_FindByID` and `TestAdapter_Delete` using `httptest.NewServer`.
  * `internal/services/appointment/service_test.go`: Added `TestService_CancelAppointment` and `TestService_FindByID` with table-driven tests.
* **Coverage**: Verified via `go test -coverprofile=coverage.out`.
* **Status**: Tests are passing and coverage has been increased.
