# Part 4: Introduce Structured, PII‑Safe Logging

## Goal

Replace standard `log` package with a structured logger (`zap`) that automatically redacts Personal Identifiable Information (PII).

## 4.1 Implementation Plan (Source: Refactoring Proposals.md)

### 4.1.1 Implementation Steps

* **Choose Logger**: Use `go.uber.org/zap`.
* **Wrapper**: Create `internal/logging/logger.go` exposing `Infof`, `Warnf`, `Errorf`, `Debugf`.
* **Redaction**: Implement `redactPII(s string)` using regex `\d{9,}` to hide Telegram IDs.
* **Migration**: Replace all `log.Printf` calls with `logging.Get().Info/Error`.
* **Env Control**: control verbosity via `LOG_LEVEL` (INFO in prod, DEBUG in dev).

## 4.2 Solution Status

### **Completed**

> [!NOTE]
> **Status**: Restored and Verified.
> **Verification**:
>
> * `internal/logging` created with Zap and Redaction.
> * `main.go`, `config`, `adapter`, `service`, `bot.go` migrated to structured logging.
> * `go test internal/logging` PASSED.
> * `go build ./cmd/bot` PASSED.

* **Implementation**:
  * Created `internal/logging` package wrapping `uber-go/zap`.
  * Implemented automatic PII redaction for strings containing 9+ digits.
* **Migration**:
  * Replaced all `log.Printf`, `log.Println`, `log.Fatal` calls in core, config, storage, delivery, services, and adapters.
  * Updated `main.go`, `config.go`, `webapp.go`, `health.go` and all service/adapter files.
* **Verification**:
  * Verified logs output in JSON format.
  * Verified `LOG_LEVEL=DEBUG` works.
  * Verified PII redaction logic with unit tests.
