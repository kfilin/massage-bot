# Changelog

All notable changes to the **Massage Bot** project will be documented in this file.

## [Phase 47: Graphify Mandatory Enforcement] - 2026-06-18

### Changed
- **`AGENTS.md`**: Added "Graphify Mandatory (No Skip)" guardrail. Startup procedure Step 1 hardened: "MANDATORY, no skip — install if missing, rebuild if stale."
- **`.pi/skills/startup/SKILL.md`**: Removed "skip silently" fallback from Step 1 — graphify MUST run every session.

## [Phase 46: Fix deploy.sh Port-Collision Pre-Flight] - 2026-06-18

### Changed
- **`scripts/deploy.sh`**: Replaced naive `ss` port check with smart pre-flight that only aborts when the port is bound by a process OUTSIDE the vera-bot compose project. Uses `com.docker.compose.project=vera-bot` label for reliable identification on both local and server.
- **`scripts/deploy.sh`**: Guarded `.env` `HOST_WEBAPP_PORT=` read with `|| true` to prevent silent abort under `set -euo pipefail` when key is missing.

### Fixed
- Production deploy was aborting because `ss` detected the port from our own container as a "collision." Smart pre-flight now distinguishes our bindings from rogue ones.

## [Phase 45: Git Sync Hygiene (PC ↔ Server)] - 2026-06-18

### Changed
- **`AGENTS.md`**: Added **Server Read-Only Convention** guardrail — `/opt/vera-bot/` is read-only except for `data/`, `credentials.json`, `.env`, `.env.test`. Code/config flows exclusively through `scripts/deploy.sh prod`.

### Housekeeping
- Cleaned up 2 stale untracked files on server: legacy `deploy.sh` (used `docker run` on dead image) and `docker-compose.yml.backup`.
- Server HEAD was 2 commits behind origin — resolved via `git fetch && git reset --hard origin/master`.

## [Phase 44: CI/CD Image Pinning & Backup Verification] - 2026-06-17

### Changed
- **`.gitlab-ci.yml`**: Pinned all base images to specific versions: `docker:latest` → `docker:27-cli`, `docker:dind` → `docker:27-dind`, `alpine:latest` → `alpine:3.21` (×2). Migrated deprecated `only:` syntax to modern `rules:` for both deploy jobs. Added `environment:` blocks (test, production) for GitLab deployment board UI. Kept `when: manual` + `needs: ["run-tests"]` for prod gate.
- **`Dockerfile`**: Runtime stage base image pinned from `alpine:latest` → `alpine:3.21`. Builder stage already pinned via `golang:1.25-alpine`.
- **`deploy/docker-compose.dev.yml`**: Pinned `caddy:latest` → `caddy:2.8-alpine`.

### Added
- **`scripts/verify_backup.sh`**: Backup restoration verification script with explicit exit codes (0=pass, 4=invalid JSON). Validates ZIP integrity, required entries, JSON parsing.

### Fixed
- **`.gitlab-ci.yml`**: Go version mismatch resolved — test stage now uses `golang:1.25.3-alpine` (was `1.23`). Permission fix: `go test ./cmd/... ./internal/...` instead of `./...` to avoid `postgres_data` permission errors in CI.

## [Phase 42: Telegram Routing Extraction & Reminder Lifecycle] - 2026-06-17

### Added
- **`internal/delivery/telegram/routing.go`**: Pure-function extraction of OnCallback and OnText routing logic.
- **`internal/delivery/telegram/routing_test.go`**: 32 table-driven tests covering all 13 callback prefixes.
- **`internal/delivery/telegram/text_flow.go`**: Side-effecting helpers `handleAdminReply` and `forwardPatientMessageToAdmins`.
- **`reminder.RunLoopForTest`**: Inner goroutine of `Start()` extracted for testable lifecycle.

### Changed
- **`internal/delivery/telegram/bot.go`**: OnCallback/OnText handlers refactored to delegate to `RouteCallback` / `RouteTextMessage`.
- **`internal/services/reminder/service.go`**: `Start(ctx)` delegates ticker loop to `RunLoopForTest`.

### Removed
- **`internal/delivery/telegram/bot.go.bak`** (17KB stale Feb-04 backup) deleted.

## [Phase 40: Universal Collaboration Harness Migration] - 2026-05-14

### Changed
- Migrated project to Antigravity standard harness structure (`.agent/` + `global-skills/`).
- Moved backlog from `.agent/backlog.md` to root `BACKLOG.md`.
- Un-ignored `.agent/` in `.gitignore`.

## [Phase 36: Test Coverage Hardening (80%+)] - 2026-06-18

### Changed
- Coverage: **76.8% → 80.0%** (+3.2pp).
- Extracted `createHealthMux()` from `startHealthServer` (tested routes: 100%).
- Extracted `createWebAppMux()` from `StartServer` (all routes, static assets, WebDAV, lifecycle tested: ~89%).
- `NewCancelHandler`: now at **100%** (added cancel service error path).

## [Phase 34: Integration Testing with Testcontainers] - 2026-06-17

### Added
- **`testcontainers-go`** as dependency (PostgreSQL module).
- 16 integration tests for `internal/storage` covering all CRUD, session storage, appointment metadata.
- Integration build tag: `//go:build integration`.

### Changed
- Storage coverage: 32% → **68.7%**.
- `internal/storage/init.go`: `InitDB` supports testcontainers connection string override.
