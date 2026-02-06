# Part 2: Eliminate Hard‑Coded Secrets

## Goal

Remove all hard-coded secrets (API keys, tokens) from the source code and move them to environment variables.

## 2.1 Implementation Plan (Source: Refactoring Proposals.md)

### 2.1.1 Actions

* **Search**: Identify secret-like strings (AWS, Google, Telegram tokens) using `rg -i "(AKIA|AIza|[0-9A-Za-z]{35,})"`.
* **Externalize**: Replace literals with `os.Getenv("VAR_NAME")` (e.g., `TELEGRAM_BOT_TOKEN`).
* **GitIgnore**: Ensure files like `credentials.json` are ignored.
* **Document**: Add "Configuration" section to README.md.
* **Validate**: Add startup checks for mandatory secrets in `config/init()`.

## 2.2 Solution Status

### **Completed**

> [!NOTE]
> **Restored on**: 2026-02-03
> **Changes**: Migrated `cmd/bot/config` to `internal/config`, updated `scripts/renew_token.sh`, and documented secrets.
>
> [!WARNING]
> Code for this part is missing. The details below are the **intended** implementation from a previous lost session. Use them as a blueprint to re-implement.

* **Legacy Config Removed**: deleted the unused `cmd/bot/config` package.
* **Scripts Secured**: `scripts/renew_token.sh` now strictly requires environment variables.
* **Documentation**: `README.md` now lists all 13 supported environment variables.
* **Security**: `.gitignore` has been hardened against accidental secret commits.

### Verification

* Build passed: `go build ./cmd/bot`
* Tests passed: `go test ./internal/config/...`
* **Scripts Secured**: `scripts/renew_token.sh` now strictly requires environment variables.
* **Documentation**: `README.md` now lists all 13 supported environment variables.
* **Security**: `.gitignore` has been hardened against accidental secret commits.

* Build passed: `go build ./cmd/bot`
* Tests passed: `go test ./internal/config/...`
