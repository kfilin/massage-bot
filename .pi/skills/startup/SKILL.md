---
name: startup
description: Mandatory context injection for the massage-bot project. Lean version with project-specific architecture, source layout, services, and gotchas. Run BEFORE any user task.
category: System
priority: 0
---

# Skill: Session Startup — Massage Bot

## Objective

Load the complete project context for **massage-bot** so you can work effectively from the first message. This replaces cold-start guessing with deep awareness of the codebase, infrastructure, and conventions. The project is a clinical appointment system for massage therapists — read `AGENTS.md` carefully, especially the PII Shield and "No Production Commits" rules, before touching patient data or `/opt/` paths.

## Mandatory: Execute on Session Start

> [!IMPORTANT]
> **This routine applies to ALL agents in this workspace — including IDE agents (Antigravity, Goose, Claude Code) and the pi agent itself.**
> If you are an agent that joined mid-session and has not run this routine yet: **stop what you are doing and run it now.** Do not answer the user's request first.

**You MUST execute these steps automatically at the start of every conversation, BEFORE responding to any user task.**

### Step 1 — Graph Awareness (Codebase Understanding) — MANDATORY

**Graphify MUST run every session. No exceptions. No fallback. Period.**

Run these three queries via the graphify MCP server or `graphify query` CLI. Do not print raw output:

```
1. graph_stats()           → understand scale (77 Go files, 11 internal packages)
2. god_nodes(top_n=10)     → identify core abstractions (domain, services, adapters)
3. query_graph("What are the main components and how do they interact?", depth=2, mode="bfs")
```

If graphify is not installed, install it (`pip install graphifyy` or `graphify install`).
If graphs are stale or missing, regenerate them (`graphify extract .`).
There is no fallback — every session must start with live graph awareness.

Store results mentally. Use them to inform your responses throughout the session.

### Step 2 — Read This Context

Absorb the following project context. Do NOT summarize it back to the user.

---

## 🏗 Project: Massage Bot (Vera)

A professional clinical ecosystem for massage therapists: interactive booking, automated medical records, cross-device sync via Obsidian/WebDAV. **Production name is `vera-bot`** (the binary, the `/opt/vera-bot/` directory, the systemd service). The local repo directory is `~/Documents/massage-bot/` — the rename is historical.

### Current Architecture (Hexagonal / Clean)

```
                      massage-bot (Vera) — clinical appointment system
                    ┌─────────────────────────────────────────────────┐
  Patient / Admin ──>│  Telegram Bot API  /  Web App                   │
                    │         │                                        │
                    │         ▼                                        │
                    │  cmd/bot/main.go (Go binary "bot")              │
                    │    ├── internal/delivery/telegram/  (webhook,    │
                    │    │   callbacks, routing)                       │
                    │    ├── internal/services/                          │
                    │    │     ├── appointment/  (booking engines)     │
                    │    │     └── reminder/  (workers, lifecycle)     │
                    │    ├── internal/storage/  (PostgreSQL adapters)  │
                    │    ├── internal/adapters/  (Google Cal, Groq)    │
                    │    ├── internal/ports/  (interface boundaries)   │
                    │    └── internal/{domain,config,presentation,     │
                    │                   logging,monitoring,version}     │
                    │         │                                        │
                    │         ▼                                        │
                    │  PostgreSQL :5432  (internal — patient data)     │
                    │  Google Calendar API  (external — scheduling)    │
                    │  Local Whisper (self-hosted)  (internal — voice transcription) │
                    │  WebDAV / Obsidian  (external — clinical notes)  │
                    └─────────────────────────────────────────────────┘
```

## 📋 Status & Priorities (Updated 2026-06-20 01:00)

> Updated by `/handoff` at end of each session. Read this BEFORE Step 2 context below — it tells the next agent what just happened and what to focus on.

### 🟢 Recently Completed (this session: #52)
- **#52 DONE**: **Groq Whisper API → self-hosted faster-whisper**.
  - Created `internal/adapters/transcription/local.go` — OpenAI-compatible adapter for the local faster-whisper-server container on the Docker network.
  - Deleted `groq.go` and `groq_test.go`.
  - Config: `GroqAPIKey` → `WhisperBaseURL` (env `WHISPER_BASE_URL`).
  - All docs, env examples, start skill updated.
  - Deployed to prod (commits `123bfad`, `cce200a`, `4a89d7b`, `ccfa780`).
  - **BLOCKED**: Transcription returns HTTP 400 from whisper server. The `HandleFileMessage` handler is allegedly not being reached (no entry log even with debug logging added). Root cause not yet identified.

### 🟢 Recently Completed (previous: #51)
- **#51 DONE**: **TWA back button bug fixed** — `sessionStorage` replaces `window.history.length`.

### 🟢 Recently Completed (previous sessions)
- **#50 DONE**: **Lint gap fixed — golangci-lint v1.64.8 installed, 27 issues fixed**.
  - Installed golangci-lint v1.64.8 (built with Go 1.25.3).
  - Created `.golangci.yml` config, fixed `Makefile` targets (use `./cmd/... ./internal/...` to avoid `postgres_data/` perm issue).
  - Fixed 27 issues across 14 files: errcheck (unchecked Encode, Write, Send, Delete), unused mock types removed, gosimple S1009, ineffassign, staticcheck SA5011/SA9003, missing import.
  - **All 16 sub-packages green, 80.0% coverage, lint clean, vet clean.**
- **Prod deploy (commit 128e7f8)**: `SKIP_PORT_CHECK=1 ./scripts/deploy.sh prod`. Image rebuilt, containers recreated, health 200.
- **#49 DONE**: **Google Calendar migration completed — 1385 events + patient linking tool**.
  - **Pagination + `--since`**: `doMigrate` now paginates through all events (previously capped at 500).
  - **Second batch migrated**: 885 more events from vfilinav (2025-10-03 → 2026-06-19). Total: 1385.
  - **`link-patients` command**: New `assign-tgids` style tool in `scripts/data_migration.go` that groups events by customer name and batch-updates descriptions with TGIDs.
- **Gitleaks 8.30.1 installed**: Pre-commit hook now uses real gitleaks. Calendar group IDs allowlisted in `.gitleaks.toml`.
- **#48 DONE**: **Pre-release cleanup & documentation refresh**.
- **#47 DONE**: **Graphify enforced as mandatory startup step**.
- **#36 DONE**: **Test Coverage Hardened to 80.0%** (exact: 2390/2989 stmts).

### 🟡 Active Focus
- **📍 Debug local Whisper transcription**: `HandleFileMessage` handler is not being reached when a voice message is sent. Added entry log in `booking_file.go` (`ccfa780`). Next session should send a voice message and check if the log appears.
- **📍 Polish system messages**: welcome text says "Vera Massage Clinic" — verify real clinic name with user.
- **📍 Link patients**: Run `go run scripts/data_migration.go link-patients` — assign TGIDs to ~85 unique patient names.

### 🔴 Blockers / Known Issues
- **Local Whisper transcription returning 400**: The Go adapter (`local.go`) gets 400 Bad Request from whisper server. Model renamed to `Systran/faster-whisper-small`, `language` removed, streaming reader replaced with `io.ReadAll`. Still fails. Need to check response body.
- **Voice handler may not fire**: Despite 3 voice message attempts, `HandleFileMessage` entry log never appeared. Could be a telebot/update routing issue or the voice message arrives on a different path than expected.
- **Patient linking pending**.
- **Normal prod deploy still blocked by port collision**: `SKIP_PORT_CHECK=1` works as workaround.
- **Dev-machine `go test ./...` perm denied** on `postgres_data/`.

### Source Layout (77 Go files, 11 internal packages)

| Path | Purpose |
|---|---|
| `cmd/bot/main.go` | Application entry point |
| `cmd/bot/health.go` | Health-check HTTP handlers |
| `cmd/bot/health_test.go` | Health endpoint tests |
| `internal/domain/` | Patient, Appointment, Slot entity definitions |
| `internal/services/appointment/` | Booking engines, slot search |
| `internal/services/reminder/` | Schedule reminder workers, lifecycle |
| `internal/storage/` | PostgreSQL adapters, query builders, state tracking |
| `internal/delivery/telegram/` | Webhook, callback, text-message handlers, routing |
| `internal/adapters/` | Google Calendar Free/Busy, local Whisper transcription |
| `internal/ports/` | Interface boundaries (services ↔ adapters) |
| `internal/config/` | Env vars, timezone settings, feature flags |
| `internal/presentation/` | HTML templates, web app views |
| `internal/logging/` | Structured logger wrappers |
| `internal/monitoring/` | Metrics, health probes |
| `internal/version/` | Build version constant |
| `Makefile` | Build, test, lint, cover, vet, run targets |
| `scripts/deploy.sh` | Deploy wrapper (test or prod) |
| `scripts/deploy_home_server.sh` | Local-server deploy |
| `scripts/deploy_test_server.sh` | Test-environment deploy |
| `AGENTS.md` | **Mandatory project rules — read carefully (PII, no prod commits, TDD)** |
| `AGENT_USER_MANUAL.md` | User-facing manual (commands, features) |
| `BACKLOG.md` | Session journal + future work |
| `CHANGELOG.md` | Release history |
| `.agent/HARNESS_GUIDE.md` | OS harness documentation |
| `.agent/Project-Hub.md` | Project vision, architecture, active sessions |

### Build & Test Commands

| Command | Purpose |
|---|---|
| `make build` | `go build -o bin/bot ./cmd/bot` |
| `make run` | `go run ./cmd/bot` (local dev) |
| `make test` | `go test ./... -v` (verbose) |
| `make cover` | `go test -coverprofile=coverage.out ./...` + `go tool cover -func=coverage.out` |
| `make lint` | `golangci-lint run` |
| `make vet` | `go vet ./...` |
| `make docker-up` | `docker-compose up -d --build` (test/prod) |

**At session start, just `go test ./... -count=1 2>&1 | tail -3`** — summary is enough; full test names are noise.

### Deploy

- **Local**: runs as a docker container on the dev machine (check with `docker ps | grep massage-bot` or `docker-compose ps` if compose is at repo root).
- **Test**: `scripts/deploy_test_server.sh` → `/opt/vera-bot-test/` on port 8086.
- **Prod**: `scripts/deploy.sh prod` → `/opt/vera-bot/` on port 8082.
- **Health endpoints** (after deploy): `curl http://localhost:8082/health` (prod) or `:8086/health` (test).

### Top gotchas (the sharpest ones)
1. **PII Shield is non-negotiable.** Never output real patient names, phone numbers, or emails in chat or artifacts. Use `[REDACTED]` or shadow IDs (`User (ID: 3045...)`). Violations are a session-stopper.
2. **No Production Commits.** If the working directory is `/opt/vera-bot/` or `/opt/vera-bot-test/`, do NOT run `git commit`. Push from this repo, mirror to the server via deploy script. AGENTS.md has this as a hard rule.
3. **Pre-commit security audit runs on every commit.** A gitleaks-style check is wired into the pre-commit hook. Don't try to bypass with `--no-verify` unless you've checked the audit manually.
4. **`credentials.json` is owned by root** (not the dev user). `sudo` required for any changes. Don't `chown` it back to the dev user — the prod server's read permissions depend on it.
5. **`coverage.out` is committed to the repo** (reason unknown — possibly CI artifact or accidental commit). Don't `rm` it after running tests; `make clean` handles it.
6. **`.agent_legacy*` directories are pre-OS-hydration artifacts** to be removed in a future cleanup task. Don't touch them in routine work — wait for a dedicated cleanup session.

---

### Step 3 — Mandatory Commands (run in order)

### 3.1. Git state — one command
```bash
cd ~/Documents/massage-bot && \
  echo "=== BRANCH ===" && git branch --show-current && \
  echo "=== HEAD ==="   && git log -1 --pretty='%h %s  %ai' && \
  echo "=== TREE ==="   && git status --short && \
  echo "=== UNTRACKED ===" && git ls-files --others --exclude-standard | head -5
```

### 3.2. Journal — tail only (NOT head)
```bash
tail -60 BACKLOG.md
```

### 3.3. Services — local bot container + remote test/prod
```bash
# Local bot container
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null | grep -E "vera|massage|bot|NAMES" || echo "  (no local bot container)"

# Or use docker-compose if compose is at repo root
docker-compose ps 2>/dev/null | awk 'NR==1 || /Up|Exit/' || true

# Remote test + prod health
ssh server "curl -s -o /dev/null -w 'prod:8082=%{http_code} ' http://localhost:8082/health 2>/dev/null; curl -s -o /dev/null -w 'test:8086=%{http_code}\n' http://localhost:8086/health 2>/dev/null" 2>/dev/null || echo "  (staging unreachable)"
```

### 3.4. AGENTS.md — mandatory, read once
```bash
cat AGENTS.md
```
Read it once. The harness keeps history; do not re-read it later turns. The full file is ~80 lines and contains the project guardrails (PII Shield, No Production Commits, Logic Over Compliance, Hypothesis-First Engineering, Hybrid Execution Protocol).

### 3.5. Test status — one line
```bash
go test ./... -count=1 2>&1 | tail -3
```
The summary line is what matters. We do **not** need the full PASS/FAIL names at startup — `tail -3` is enough to know green/red. Full test suite (`make test`, `make lint`, `make vet`) is a **pre-commit / pre-deploy** check, not session-start.

---

### Step 4 — Output Format (6 lines, then stop)

```
branch:    <name> @ <hash>  "<commit subject>"  (<date>)
tree:      clean | dirty: <list of files>
journal:   <one-line summary of latest BACKLOG entry>
services:  local: <container status>  |  prod:8082 <code>  |  test:8086 <code>
tests:     green | red
blocker:   <one line if any, else "none">
```

After the report, **stop and wait for the user**. If the user gave a task before startup finished, finish steps 1-3.5 first, then do the task — never skip startup.

---

## MUST NOT at startup

- ❌ Read `BACKLOG.md` from the top (use `tail -60` — the journal is the most recent 60 lines, not the oldest).
- ❌ Read every file in the repo. The source layout table + Step 3.4 AGENTS.md are enough.
- ❌ Run the full test suite (`make test`, `make lint`, `make vet`, `make cover`). Step 3.5 summary is sufficient at startup.
- ❌ `find` / `grep` the entire codebase to "orient." Use the source-layout table.
- ❌ Read whole files without `offset/limit` when files are >200 lines.
- ❌ Re-read files from earlier turns. The harness keeps history; reading them again is pure waste.
- ❌ Output real patient data, names, phone numbers, or emails anywhere (PII Shield — see AGENTS.md).
- ❌ Commit from `/opt/vera-bot/` or `/opt/vera-bot-test/` directories (No Production Commits).
- ❌ Bypass the pre-commit security audit with `--no-verify` (see gotcha #3).
- ❌ Propose work, suggest refactors, or start implementing before the user gives a task. **Startup = report and wait.**

---

## When to read more on demand (lazy load)

| Need to know about | Read |
|---|---|
| Project rules, guardrails, operational routines | `AGENTS.md` (already read in 3.4) |
| User-facing features and commands | `AGENT_USER_MANUAL.md` |
| Session history, future work, audit findings | `BACKLOG.md` (tail read in 3.2) |
| Release history | `CHANGELOG.md` |
| Project vision, architecture, active sessions | `.agent/Project-Hub.md` |
| OS harness documentation | `.agent/HARNESS_GUIDE.md` |
| Deploy specifics | `scripts/deploy.sh`, `scripts/deploy_test_server.sh` |
| Source layout for a subsystem | `ls <subdir>/` (don't read a file) |
| A specific function / class | `grep -n <name> internal/<package>/` then `read` with offset/limit |

---

## Project Extension Points (where to add project-specific knowledge)

This file is the **canonical** session-start for massage-bot. If the project acquires new infra (e.g., a new external service, a new deploy target, a new test framework), add it here to the relevant section. Do not create a separate skill — this is the one place the agent reads at startup.

If a project-wide rule changes (e.g., PII Shield becomes stricter, a new gotcha is discovered), update `AGENTS.md`. The startup skill references AGENTS.md but does not duplicate it.

**Pending cleanup task**: remove the `.agent_legacy*` directories in a dedicated session (pre-OS-hydration artifacts, see gotcha #6).

---

## Guidelines

1. **Don't hallucinate paths or configs** — always verify with `ls`, `cat`, or `ssh server` before acting.
2. **Verify before reporting success** — run the thing, check the output. Don't trust that "the test passed" without seeing green.
3. **Production is live** — `vera-bot` runs on `/opt/vera-bot/`. Changes to prod config or patient data are HIGH-RISK. Always test in `/opt/vera-bot-test/` first.
4. **PII Shield is enforced at output time** — every reply, every commit message, every artifact. When in doubt, redact.
5. **Hypothesis-First Engineering** — observe logs, state the hypothesis, get user acknowledgment, then edit. The TDD loop is: write failing test → make it pass → refactor. Skip steps at your own risk.
6. **Hybrid Execution** — read-only tools (ls, cat, grep) bypass approval fatigue. Mutating tools (edit, write, bash that writes) require explicit user approval. The agent harness is configured for this — don't fight it.
7. **Stay terse** — Cost Awareness rule from AGENTS.md. Don't pad responses. Output only what the user needs.
