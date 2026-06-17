# Changelog

All notable changes to the **Massage Bot** project will be documented in this file.

## [Phase 42: CI/CD Image Pinning & Backup Verification] - 2026-06-17

### Changed
- **`.gitlab-ci.yml`**: Pinned all base images to specific versions: `docker:latest` → `docker:27-cli`, `docker:dind` → `docker:27-dind`, `alpine:latest` → `alpine:3.21` (×2). Migrated deprecated `only:` syntax to modern `rules:` for both deploy jobs. Added `environment:` blocks (test, production) for GitLab deployment board UI. Kept `when: manual` + `needs: ["run-tests"]` for prod gate (functionally correct).
- **`Dockerfile`**: Runtime stage base image pinned from `alpine:latest` to `alpine:3.21`. Builder stage already pinned via `golang:1.25-alpine`.
- **`deploy/docker-compose.dev.yml`**: Pinned `caddy:latest` to `caddy:2.8-alpine`.

### Added
- **`scripts/verify_backup.sh`**: Backup restoration verification script. Validates ZIP integrity (`unzip -t`), checks for required entries (`blacklist.txt`, `patients/`, `token.json`), extracts to temp dir and spot-checks a `patient.json` parses. Explicit exit codes: 0=pass, 1=arg/missing, 2=corrupt, 3=missing entries, 4=invalid JSON. Tested against synthetic good/bad/corrupt backups.

### Verified
- All base images now have specific version tags (no `:latest` for upstream images); the only remaining `massage-bot:latest` is the local build output tag.
- `go test -short ./...` clean across all packages.
- YAML syntax valid for `.gitlab-ci.yml` and all compose files.

Refs BACKLOG #44 (CI/CD Pipeline Audit — 5 of 6 tasks done; remaining: `docker-compose.override.yml` server drift inspection).

## [Phase 41: Telegram Routing Extraction & Reminder Lifecycle] - 2026-06-17

### Added
- **`internal/delivery/telegram/routing.go`**: Pure-function extraction of OnCallback and OnText routing logic. `RouteCallback(data)` returns the matched prefix or `(_, false)`; `RouteTextMessage(text, SessionView)` returns a `TextAction` enum representing the routing decision across all four priority levels (command fallback, main-menu buttons, admin-reply, awaiting-confirmation, safety fallbacks, default flow).
- **`internal/delivery/telegram/routing_test.go`**: 32 table-driven tests covering all 13 callback prefixes, exact-match callbacks, edge cases (empty data, whitespace, unknown prefix), and the full text-priority ladder.
- **`internal/delivery/telegram/text_flow.go`**: Imperative helpers `handleAdminReply` and `forwardPatientMessageToAdmins` extracted from the OnText handler closure. They own the side effects (Telegram sends, Med-Card writes) that depend on `*telebot.Bot`.
- **`reminder.RunLoopForTest`**: Inner goroutine of `Start()` extracted so tests can drive ticks on a manual channel and assert lifecycle behavior (cancellation, multiple ticks, stop callback).

### Changed
- **`internal/delivery/telegram/bot.go`**: `RunBot` OnCallback and OnText handlers refactored to delegate routing decisions to `RouteCallback` / `RouteTextMessage`. Inline side-effecting code moved to helpers in `text_flow.go`. Behavior is preserved (same handler dispatches, same priority ladder).
- **`internal/services/reminder/service.go`**: `Start(ctx)` now delegates the ticker loop to `RunLoopForTest(ctx, ticks, stop)`; production behavior unchanged.

### Removed
- **`internal/delivery/telegram/bot.go.bak`** (17KB stale Feb-04 backup file) deleted.

### Test Coverage
- `internal/delivery/telegram`: **4.4% → 21.2%** (routing logic now ~100% covered).
- `internal/services/reminder`: **81.5% → 91.4%** (Start() ticker lifecycle tested).
- All other packages: unchanged (no regressions).
- `go vet ./...` clean across all packages.

## [Phase 40: Universal Collaboration Harness Migration] - 2026-05-14

### Added
- **Universal Project Harness**: Migrated the project to the new Antigravity standard, including the `.agent/` and `global-skills/` unified structure.
- **The Antigravity Constitution**: Established a 12-point master list of rules (Philosophy, Engineering Rigor, and Platform Guardrails) as the unified standard.
- **Dedicated Rule Files**: Deployed high-context rule files for `logic-over-compliance`, `anti-overengineering`, `constraints-not-checklists`, `context-compaction`, `no-server-commits`, `budget-consciousness`, and `pii-shield`.
- **Informative Project Hub**: Rebuilt the Hub with full information density, preserving all clinical technical foundation details while adding the new Ecosystem rules.
- **Hydration System**: Initialized `.agent/project-config.env` and hydrated the harness for the `massage-bot` context.

### Changed
- **Backlog Promotion**: Moved the project backlog from `.agent/backlog.md` to the root `BACKLOG.md` for better visibility.
- **Version Control**: Un-ignored the `.agent/` directory in `.gitignore` to track the collaboration harness.
