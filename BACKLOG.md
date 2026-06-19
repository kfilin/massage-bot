# Project Backlog

## 🟢 Session 2026-06-19 14:45 — #51 verified on server, back button works

- [x] **#51 DONE**: TWA back button fix verified on server — `setupBackButton()` now uses `sessionStorage` instead of `window.history.length`, works in both directions (search → patient → search).
- [x] **Second round (14:45)**: Also added:
  - `app.js`: `tg.BackButton.hide()` **before** `window.location.href` to prevent TWA from retaining button state across page loads
  - `search.html`: inline script force-hides back button on search page before `DOMContentLoaded`
- [x] **Deployed to prod (commit `c7a97da`)**: `ssh server "/opt/vera-bot" → docker compose build + up`. Health 200, back button works.

## 🟢 Session 2026-06-19 13:25 — golangci-lint install + lint fixes + prod deploy

- [x] **Installed golangci-lint v1.64.8** (was missing, `make lint` failed).
- [x] **Created `.golangci.yml`** with linter config.
- [x] **Fixed `Makefile` targets**: lint/vet/test/cover now use `./cmd/... ./internal/...` to avoid `postgres_data/` permission issue.
- [x] **Fixed 27 lint issues** across 14 files:
  - Unchecked errors (Encode, Write, Send, Delete) in production + test code
  - Removed unused mock types/functions
  - Simplified nil+len checks (gosimple S1009)
  - Inlined unused `expiryDays` variable (ineffassign)
  - Fixed nil dereference potential (staticcheck SA5011)
  - Removed empty branches (staticcheck SA9003)
  - Added missing `logging` import
- [x] **Prod deploy** (commit `128e7f8`): `SKIP_PORT_CHECK=1 ./scripts/deploy.sh prod` — image rebuilt, containers recreated, health 200.

## 🟢 Session 2026-06-19 01:08 — Google Calendar migration: second batch + patient linking tool

- [x] **Migration pagination**: Added pagination + `--since` flag to `doMigrate` so it reads ALL vfilinav events beyond 500.
- [x] **Second batch migrated**: 885 events from 2025-10-03 → 2026-06-19 copied to project calendar (1385 total now).
- [x] **`link-patients` command built**: `scripts/data_migration.go` — scans all events, groups by name, lets you assign TGIDs per patient, batch-updates descriptions with `TGID:XXX\n`, imports to DB.

## 🟢 Session 2026-06-18 23:57 — Pre-release cleanup, documentation refresh, data migration

- [x] **Documentation pass**: Updated README (coverage 80%, deploy scripts, structure), USER_GUIDE/EN/RU, DEVELOPER (Go version, test strategy, admin commands), CHANGELOG (Phases 44-47), AGENT_USER_MANUAL (rewritten for real project), docs/files.md (full rewrite), docs/CI_CD_Pipeline.md, docs/API.md (TWA endpoints), docs/VERA_GUIDE_RU.md, docs/metrics_setup.md, docs/ProdArchitecture.md, data/README.md, metrics.md, docs/backlog_design.md
- [x] **Archived stale files**: walkthrough.md → ARCHIVE/WALKTHROUGH/, what_to_fix.md → ARCHIVE/REVIEWS/
- [x] **DB wiped clean**: TRUNCATE all tables (patients, appointments, metadata, media, analytics, sessions, blacklist). Clean slate.
- [x] **Calendar cleaned**: 16 mock events deleted from veramassagist calendar.
- [x] **Migration script built**: `scripts/data_migration.go` with 5 subcommands (clean-db, list-events, clean-calendar, auth, migrate).
- [x] **Data migrated**: 500 events from vfilinav@gmail.com → project calendar via OAuth. Token generated and saved locally.
- [x] **#34 marked DONE** (called at 80.0% — structural ceiling accepted).
- [x] **#30 marked DEFERRED** (user said "significantly later").
- **Pending**: Prod deploy blocked by port 8082 collision. Run `SKIP_PORT_CHECK=1 ./scripts/deploy.sh prod`.
- **Pending**: Calendar events need verification after deploy.

## 🟢 Session 2026-06-17 23:25 — P0: Production Crash Loop Resolved (Orphaned Docker Network)

- [x] Diagnosed prod crash loop via `nc -w 3` + `ping` from app container → 100% packet loss on `massage-bot-internal` bridge in **both directions** (app↔db), even with correct ARP, correct routes, fresh network, and post-`systemctl restart docker`. Bridge had `fdb_n_learned 2` while caddy-test-net had 44, indicating veth endpoints from a prior `vera-bot` project were still attached to a phantom `vera-bot_bot-db-net` network.
- [x] Fixed with `docker compose down && docker network prune -f && docker compose up -d`. **4 stale networks** pruned (`vera-bot_bot-db-net`, `openclaw_default`, `diun_default`, `massage-bot-test_bot-db-net`). Prod back online — `/health` returns 200, bot authenticated as `@vera_massage_bot`, Reminder Service started.
- [x] Closed out deferred items from prior #41: `chmod 600 /opt/vera-bot/.env` (was 0644) and removed stale `docker-compose.override.yml` (167 bytes from 2026-01-20, referenced dead `registry.gitlab.com/kfilin/massage-bot:latest` image, service name `massage-bot` didn't match main `app` service — was a no-op override).
- [x] Updated #41, #44, #45 with new root cause + removed resolved bullets. Forensic snapshots saved at `/tmp/vera-prod-pre-fix.{log,json}` and `/tmp/vera-net-pre-fix.json` for the record.
- **Note**: `No .env file found` log line is a false positive — godotenv looks for a literal `.env` file in the binary's CWD, but env vars ARE loaded via compose's `env_file:` directive. The line appears even when env is correctly loaded. Do not chase this; it is benign noise.

## 🟢 Session 2026-06-17 13:55 — Startup: replace project-agnostic with project-specific

- [x] Replaced `.pi/skills/startup/SKILL.md` project-agnostic skeleton (hydrated on 2026-06-17 morning) with project-specific lean version (241 lines). Now bakes in: architecture (Hexagonal/Clean, Go binary, telebot, PostgreSQL, Google Calendar, Groq Whisper, WebDAV), source layout (77 Go files, 11 internal packages), deploy targets (local docker, test :8086, prod :8082), and top 6 gotchas (PII Shield, No Production Commits, pre-commit audit, credentials.json, coverage.out, `.agent_legacy*` cleanup). **Pending**: remove `.agent_legacy*` dirs in a dedicated cleanup session.

## 🟢 Session 2026-06-14 08:20 — Vault Project Scoping Alignment

- [x] Aligned handoff path to `Bridge/massage-bot/Checkpoints/`
- [x] Migrated handoff notes from flat `Bridge/Checkpoints/` to project-scoped directory
- [x] Updated `handoff.md`, `obsidian_management.md`, `startup.md` skills
- [x] No Go code changes required (uses own data directory, not shared vault)

## 📋 Observation & Improvement Ideas

### 1. [DONE] "History of Visits" Data Accuracy & Utility

- **Status**: Completed in v4.2.1.
- **Resolution**: Implemented full-history sync, appointment status filtering, and a TWA history UI.
- **Commit**: `ba80b18`

### 2. [OBSOLETE] Robust TWA Authentication (InitData)

- **Status**: Obsolete (per user direction).
- **Note**: Logic for initData validation is no longer a priority or has been superseded by other auth flows.

### 3. [DONE] Smart Forwarding & Loop Closure

- **Status**: Completed in v4.3.0.
- **Resolution**: Implemented auto-logging of patient inquiries and a reciprocal "Reply" interface for admins that archives whole conversations to the Med-Card.

### 4. [DONE] Professional Reminder Service

- **Status**: Completed in v4.3.0.
- **Resolution**: Built a ticker-based service for 72h and 24h interactive notifications with confirmation tracking.

### 5. [DONE] Robust Scheduling (Free/Busy API)

- **Status**: Completed in v5.0.0.
- **Resolution**: Migrated to official Google Calendar Free/Busy API for 100% accurate slot detection.

### 7. [DONE] Automated Backups 2.0

- **Status**: Completed in v5.0.0.
- **Resolution**: Implemented ZIP archival of DB + Files with daily Telegram delivery to Admin.

### 12. [DONE] Local Duplicati Backup Setup

- **Status**: Completed (2026-01-27).
- **Resolution**: User set up and verified Duplicati instance on the home server for incremental, encrypted backups of the `./data` directory.

### 14. [DONE] Admin Patient Name Edit

- **Status**: Completed (2026-01-31).
- **Resolution**: Implemented `/edit_name <id> <new_name>` command for admins.
- **Note**: User noted difficulty in self-testing due to overlapping Admin/Patient roles, but logic is verified for admin IDs.

### 15. [DONE] Manual Appointment Creation

- **Status**: Completed (2026-01-31).
- **Resolution**: Added `/create_appointment` command. Implemented unique ID tracking (`manual_<name>`) and an Admin Master View in "My Appointments" to ensure manual bookings are visible and linked to dedicated patient Med-Cards.
- **Implementation**: [booking.go](file:///home/kirillfilin/Documents/massage-bot/internal/delivery/telegram/handlers/booking.go), [bot.go](file:///home/kirillfilin/Documents/massage-bot/internal/delivery/telegram/bot.go)

### 16. [DONE] DB & Stability Hardening (v5.3.6)

- **Status**: Completed (2026-02-01).
- **Resolution**: Implemented `connect_timeout` for DB, added startup crash-loop delays, and documented external API visibility issues in a stability report.
- **Note**: System currently healthy internally but failing to reach Telegram API.

---

## 🎨 TWA UI/UX Improvements (Added 2026-02-04)

### 17. [DONE] Quick Wins - Dark Mode, Animations, Loading States

- **Status**: Completed in v5.4.0.
- **Resolution**: Added dark mode support, section fade-in animations, button loading states, and visual differentiation for upcoming vs history sections.

### 18. [DONE] Stat Cards - Information Density

- **Status**: Completed (2026-02-06).
- **Resolution**: Combined First Visit + Total Visits in one card, improved mobile responsiveness with a 2-column grid, and made "Next Appointment" more prominent with a dedicated highlight style.

### 19. [DONE] Accessibility Improvements

- **Status**: Completed in v5.5.1 (2026-02-06).
- **Resolution**:
  - Added `:focus-visible` outline styles for keyboard navigation.
  - Added `aria-expanded` attributes to collapsible sections via JS.
  - Added `role="button"` and `tabindex="0"` to headers.
  - Added `aria-label` to iconic links.

### 20. [DONE] Empty States Enhancement

- **Status**: Completed (2026-02-06).
- **Resolution**: Added icons and descriptive text for empty sections (Notes, History, Docs) to improve UX for new patients.

### 21. [DONE] History List Pagination

- **Status**: ✅ DONE 2026-06-18
- **Resolution** (commit `f33ebb9`): server-side pagination (default limit 30, max 100) plus an AJAX "Show more" flow.
- **Approach**: `?limit=N&offset=M` query params on the card endpoint. Handler reads them, slices the in-memory `appts` list (the full list is still loaded for `TotalVisits` / `FirstVisit` / `LastVisit` stats). `?partial=history` returns just the visit cards + a possible "Show more" button (rendered by the new `RenderHistoryFragment` presenter + `{{define "history_fragment"}}` block in `card.html`). The `loadMoreHistory` function in `app.js` fetches the partial URL (preserving auth params), parses the response, and appends to the history list — no scroll-jump, no full page reload.
- **DB primitive**: new `ports.Repository.GetAppointmentHistoryPaginated(id, limit, offset) ([]Appointment, hasMore, error)`. The hasMore signal is computed via the limit+1 trick (no separate `COUNT(*)` query). Tested with 4 sqlmock tests (ExactPage, HasMore, Offset, DBError). Not called from the handler in this commit (the handler slices the already-loaded list), but available as a primitive for any future caller.
- **Tests added**: 4 storage + 5 web handler + 1 presenter = 10 new tests. All 15 packages green. JS syntax validated with `node -c`.
- **Verification**: `bash scripts/deploy.sh prod` would redeploy; the change is doc-only safe to skip — but the next routine deploy will pick it up.
- **Not addressed**: virtual scrolling was an alternative in the original BACKLOG; chose the simpler "Show more" approach because the therapist's real upper bound is ~100 visits (1-2 per week over 1-2 years), so even a "Show more" once or twice covers the full history. Virtual scrolling would add JS complexity for no practical gain.

### 22. [TODO] Visual Hierarchy Enhancement

- **Priority**: Medium
- Voice Transcripts section could have distinct styling (e.g., chat bubble style)
- **Note**: Generic file icons rejected by user. Only custom/premium visuals to be used.

### 23. [TODO] Success Feedback

- **Priority**: Medium
- After successful cancellation, show success toast/banner before reload
- Add subtle success animation

### 24. [TODO] Offline Support

- **Priority**: Low
- Add service worker for offline viewing of cached data
- Show "offline" indicator when disconnected

### 25. [DONE] Print Optimization (@media print CSS) — Implemented, needs verification

- **Priority**: Low
- Add `@media print` styles for proper printing
- Hide interactive elements in print view

### 26. [TODO] Performance Optimization

- **Priority**: Low
- Lazy load history items below the fold
- Consider skeleton loading states for slow connections

---

#### Last updated: 2026-06-14

### 29. [TODO] Native MCP for PostgreSQL

- **Status**: Backlog
- **Priority**: P3 (ship when a concrete DB-heavy task demands it)
- **Goal**: Connect a native MCP server to the database for direct structured querying.
- **Rationale**: Currently, the agent relies on pasted schema snippets or `.md` skill docs for DB context. A live MCP connection enables real-time schema exploration, diagnostic queries, and migration verification without human intermediation.
- **Dependencies**: Must run in dev/test environments only — never production. Requires Docker network access to the PostgreSQL container.
- **Implementation Sketch**:
  1. Evaluate existing PostgreSQL MCP servers (e.g., `@modelcontextprotocol/server-postgres`).
  2. Add MCP config to `docker-compose.yml` (dev only, gated by env var).
  3. Update `database-expert` SKILL.md to reference the MCP capability.
  4. Test: agent can query `information_schema`, describe tables, and verify migration results.
- **Risk**: MCP injects schemas into context on every step — monitor token consumption. If overhead is too high, consider a CLI-based alternative.
- **Source**: Article comparison (Level 5), backlog since v5.7.0.

### 30. [DEFERRED] Knowledge Item (KI) for Clinical Patterns

- **Status**: Deferred — implementation pushed significantly later
- **Priority**: Low (no earlier than next major feature cycle)
- **Goal**: Build a "Shared Clinical Brain" — structured massage therapy protocols usable by both the Agent (for context) and the TWA (for quick-entry templates).
- **Detail**: See [clinical-patterns-ki.md](file:///home/kirillfilin/Documents/massage-bot/.agent/backlog/clinical-patterns-ki.md) for the original specification.

#### The Problem

The therapist types similar notes multiple times per day. Common recurring patterns include:
- **Pre-session assessments**: "Жалобы: напряжение в шейном отделе, ограничение подвижности" — the structure is identical, only the body zone and symptoms change.
- **Post-session notes**: "Проведён массаж [зона]. Обнаружены триггерные точки в [мышца]. Рекомендовано: [список]" — same template, different fill-ins.
- **Care instructions**: Standard post-massage recommendations (питьевой режим, ограничение нагрузок, тепло/холод) are given to nearly every patient but re-typed or re-dictated each time.
- **Contraindication flags**: Checking and documenting contraindications follows the same structure per body zone.

This is manual, repetitive work that a template + AI assist system can reduce from minutes to seconds per patient.

#### Phased Implementation

**Phase 1: Template Snippets (TWA-side, no AI needed)**
1. Define 10-15 most common clinical templates in a structured format (JSON or Markdown).
2. Add a "📋 Шаблоны" button to the TWA Edit Modal (next to the 🎙️ voice button).
3. Clicking a template inserts a pre-filled structure into the `editNotes` textarea with `[placeholder]` markers the therapist can quickly fill in.
4. Templates are stored in a `data/clinical-templates/` directory, editable by the therapist without code changes.

**Example template** ("Зажим шейного отдела"):
```markdown
## Осмотр [DATE]
**Жалобы:** напряжение в шейном отделе, [доп. жалобы]
**Пальпация:** триггерные точки в [мышцы], тонус [повышен/норма]
**Проведено:** массаж ШВЗ, [техники]
**Рекомендации:**
- Питьевой режим 1.5-2л
- Ограничение нагрузок 24ч
- [доп. рекомендации]
```

**Phase 2: AI-Assisted Note Generation (Agent-side)**
1. Create clinical protocol KIs in Russian (sourced from user's materials).
2. When the therapist records a voice memo, the AI transcription pipeline can:
   - Detect the body zone mentioned in the recording.
   - Auto-suggest the matching template, pre-filled with detected values.
   - Offer to append standard recommendations for that zone.
3. The therapist reviews and edits before saving — human remains in the loop.

**Phase 3: Pattern Learning (future)**
1. Analyze existing `therapist_notes` across patients to identify the most frequent note structures.
2. Propose new templates based on actual usage patterns.
3. Track which templates are used most and refine them.

#### Dependencies
- **Phase 1**: TWA changes to `record_template.go` (add template picker UI). New `data/clinical-templates/` directory with `.md` template files.
- **Phase 2**: Requires `ai-integration-expert` skill enhancement. May benefit from PostgreSQL MCP (backlog #29) for pattern analysis.
- **Phase 3**: Requires sufficient data volume (50+ notes per template category).

#### Automation Candidates (recurring daily cases)

| Pattern | Frequency | Automation |
|---|---|---|
| Neck/shoulder tension assessment | 3-5x/day | Template + voice detect |
| Lower back pain protocol | 2-3x/day | Template with zone-specific recs |
| Post-massage care instructions | Every patient | One-tap insert of standard recs |
| Contraindication checklist | Every new patient | Structured checklist template |
| Follow-up scheduling notes | Every patient | Auto-generate from appointment data |

#### Success Criteria
- **Phase 1**: Therapist can insert a structured template in ≤2 taps. Note entry time per patient reduced by 50%+.
- **Phase 2**: Voice memo → structured note with ≤1 manual edit.
- **Source**: Original specification + user feedback (March 2026): "There are cases that happen more than once a day. These at least can be somewhat automated."

### 31. [TODO] Fix metrics.md

- **Status**: Backlog
- **Goal**: Correct the metrics access paths (local vs home server) and update the documentation.
- **Note**: The metrics server resides on the home server, not locally.

### 32. [TODO] Background Agent Research

- **Status**: Backlog (Research only — no implementation without findings review)
- **Priority**: Low (future-facing)
- **Goal**: Deeply investigate background agent patterns (Ralph loops, Dispatch, cloud sandboxes) and produce an ADR with conclusions on if/how/when to adopt them for this project.
- **Rationale**: Level 7 of the agentic engineering ladder promises autonomous work, but carries high risk for clinical production data. Unsupervised agents can cause silent regressions, data corruption, or security exposure. A thorough research-first approach is mandatory.
- **Dependencies**: Requires solid Level 6 (harness engineering) to be fully in place — no background agents without automated backpressure.
- **Research Scope**:
  1. Survey existing tools: Dispatch, Inspect (Ramp), Claude Code Background Tasks, GitHub Codex.
  2. Identify lowest-risk entry points: read-only agents (docs freshness checker, `/report` as cron).
  3. Evaluate trust boundaries: what can a background agent touch? What must remain human-gated?
  4. Cost analysis: token consumption, infra complexity, maintenance overhead.
  5. Produce a formal ADR: "Should we adopt background agents?" with YES/NO/CONDITIONAL recommendation.
- **Success Criteria**: ADR reviewed and approved by human before any implementation begins.
- **Source**: Article comparison (Level 7), March 2026.

### 33. [TODO] Universal Agentic OS Template

- **Status**: Backlog
- **Priority**: Medium (high strategic value for future projects)
- **Goal**: Extract the project-agnostic patterns from our Agentic OS into a reusable starter kit that can bootstrap any new project.
- **Rationale**: Our system (rules, skills, workflows, KIs, session management) has matured through 6+ months of real-world use on production software. The patterns are battle-tested but currently coupled to the Massage Bot's domain. Decoupling them creates a competitive advantage for any future project.
- **Dependencies**: All OS upgrades from this session must be stable and verified first (turbo gates, `/review`, `AGENTS.md`, etc.).
- **Implementation Scope**:
  1. **Identify universal vs. project-specific**: Separate domain-agnostic rules (`hypothesis-first`, `constraints-not-checklists`, `logic-over-compliance`) from project-specific ones (`no-server-commits`, `pii-shield`).
  2. **Template structure**: Create a `.agent-template/` with placeholder `Project-Hub.md`, generic `AGENTS.md`, workflow templates (`/checkpoint`, `/changelog`, `/review`), and empty `skills/` and `rules/` directories.
  3. **Bootstrap script**: A one-command setup that copies the template into a new project and prompts for project-specific values (name, stack, deployment target).
  4. **Documentation**: Write a "Getting Started with Agentic OS" guide explaining each component and when to customize it.
  5. **Version and publish**: Track the template as a standalone repo or Git subtree.
- **Success Criteria**: A new project can be bootstrapped to Level 4-5 of the agentic engineering ladder within 30 minutes using the template.
- **Source**: Article comparison + cross-model review, March 2026.

### 34. [DONE] Phase 4: Integration Testing with Testcontainers

- **Status**: ✅ DONE 2026-06-18
- **Priority**: Medium
- **Goal**: Maximize project test coverage (aiming for >80% total) by implementing integration tests for the `internal/storage` (Postgres) and `internal/delivery/telegram` (bot routing) layers.
- **Resolution**: Called done at **80.0% overall coverage** (target met). The remaining uncovered code (entry point `main()` at 0%, `RunBot`/`InitBot` wiring at 0%) operates on concrete `*telebot.Bot` — these are thin registration/glue layers that only change during major refactors and provide diminishing returns for testing effort.
- **Progress**:
  - testcontainers-go installed. 16 integration tests for `internal/storage` covering all CRUD operations, session storage, and appointment metadata.
  - Storage coverage: 32% → **68.7%** (2026-06-17), later to 86% unit / 92% integration via #36 hardening.
  - **Telegram routing extracted** (2026-06-17): pure-function refactor — `RouteCallback(data)` and `RouteTextMessage(text, SessionView)` extracted into `internal/delivery/telegram/routing.go`; side-effecting helpers extracted into `text_flow.go`.
  - **Telegram coverage: 4.4% → 47.6%** (2026-06-17). Routing logic ~100% covered (32 tests).
  - Two testable seams extracted from `RunBot` (`setupMenuButton`, `runScheduledBackup`) — both 100% covered via `bot_wiring_test.go`.
  - Dead file `internal/delivery/telegram/bot.go.bak` (17KB stale Feb-04 backup) removed.
- **Structural ceiling accepted**: `main()` 0%, `RunBot` 0%, `InitBot` 0% — all are thin wiring on concrete `*telebot.Bot`, inherently untestable without an interface refactor that doesn't justify its effort.
### 35. [DONE] WebApp Handler Refactoring
- **Status**: ✅ DONE
- **Resolution**: Handlers already reside in `internal/delivery/web/webapp_handler.go` and `server.go`. `cmd/bot/main.go` calls `web.StartServer()`; no handlers exist in `cmd/bot/`. Architecture is clean.

### 36. [DONE] Test Coverage Hardening (80%+)
- **Status**: ✅ DONE 2026-06-18
- **Priority**: High
- **Goal**: Increase repository test coverage from ~42% to 80%+.
- **Final push (2026-06-18)**: **76.8% → 80.0%** (+3.2pp).
  - `cmd/bot`: 6.6% → **16.1%** — extracted `createHealthMux()` from `startHealthServer`, tested routes (`createHealthMux` 100%, `testStartServer_Lifecycle` 80%).
  - `delivery/web`: 79.3% → **88.5%** — extracted `createWebAppMux()` from `StartServer` and tested:
    - All routes registered (TestCreateWebAppMux_RoutesRegistered)
    - Static assets served (TestCreateWebAppMux_StaticAssets)
    - WebDAV disabled by default (TestCreateWebAppMux_NoWebDAV)
    - WebDAV enabled via env vars (status page, redirect, auth, CORS, wrong password, nonexistent dir, file path, Obsidian client)
    - Server lifecycle (TestStartServer_Lifecycle)
  - `internal/ports`: added BotAPI compile-time interface assertion (TestBotAPIImplemented)
  - `NewWebAppHandler`: added tests for DB error during history load, self-heal without name fallback
  - `NewCancelHandler`: now at **100%** — added test for cancel service error path
- **Key refactors**: extracted `createWebAppMux()` and `createHealthMux()` so route setup is testable via httptest without real server startup.
- **Next targets** (beyond 80%): `cmd/bot` `main()` (0%), `StartServer` (78.6% — server lifecycle only), `cmd/bot` `startHealthServer` (80% — shutdown error branches). All remaining uncovered blocks are hard-to-trigger error-only branches (ListenAndServe failure, Shutdown error, template parse fail).

### 37. [DONE] Grafana Dashboard Sync
- **Status**: ✅ DONE 2026-06-18
- **Resolution**: Added 4 new panels to `deploy/monitoring/grafana_dashboard.json`:
  - Free/Busy Cache Hits (stat)
  - Free/Busy Cache Misses (stat)
  - Bot Commands (stat)
  - Clinical Note Length (stat)
- Updated panel layout to accommodate new panels (y positions shifted). Dashboard now covers all metrics from `internal/monitoring/metrics.go` and `docs/API.md`.
### 38. [TODO] Refine DEVELOPER.MD
- **Status**: Backlog
- **Priority**: Low
- **Goal**: Add more technical details, edge cases, and "deep dive" sections to `DEVELOPER.MD`.
- **Rationale**: Ensure the onboarding process is as smooth as possible for new developers.

### 39. [DONE] Fix Stale Vet-Failing Test Files
- **Status**: Completed (2026-06-14)
- **Resolution**: Both test files were already fixed by prior refactors. `go vet ./...` passes clean across all packages. Test suites pass with no errors.
- **Commit**: TBD (no code change needed — backlog cleanup only)

### 40. [DONE] Universal Collaboration Harness Migration
- **Status**: Completed (2026-05-14)
- **Goal**: Port the high-density Antigravity collaboration harness to the Massage Bot ecosystem.
- **Resolution**: Migrated rules, skills, and Hub structure. Verified identical collaboration DNA with the Agentic Lab project.
- **Commit**: `3637715`

---

## 🟠 Infrastructure Hygiene Sprint — 2026-06-16

> **Trigger**: User session review of `what_to_fix.md` + server inspection revealed a stack of long-standing hygiene issues that all need to be resolved together. Not blocking P0 fix in #41 but should be sequenced after it.

### 41. [DONE] Production Bot Offline — Two Stacked Incidents
- **Status**: Completed (2026-06-17, two separate incidents)
- **Priority**: P0
- **Incident 1 (morning)**: Zombie `massage-bot` container in `Created` state from a prior failed `docker compose up` was holding the port 8082 Docker-internal allocation. Resolved with `docker rm -f massage-bot` + `docker compose up -d app`. The earlier "env_file not honored" narrative from the Apr 2026 `vera-bot-massage-bot-1` failure was already resolved by the time the new `massage-bot` image ran. Also: `LOG_LEVEL=DEBUG → INFO`, removed dead `HEALTH_PORT=8081` line from `.env`, removed old Exited container.
- **Incident 2 (evening, 19:35–23:25)**: Bot in crash loop with `dial tcp 172.19.0.2:5432: i/o timeout` from app to db on the `massage-bot-internal` bridge. **Root cause**: orphaned `vera-bot_bot-db-net` Docker network from a prior `vera-bot` project (test environment) was holding stale veth endpoints, conflicting with the new `massage-bot-internal` bridge. Bridge had `fdb_n_learned 2` (the two endpoints registered but not learnable as reachable) and 0 packets traversed `DOCKER-FORWARD` for `br-302b26045b3e`. `docker compose down` alone did not help (the DB container had "Up 3 hours" because it survived a `systemctl restart docker` while running). `systemctl restart docker` + `docker compose up -d` did not help either (containers just reattached to existing namespace). **Fix that worked**: `docker compose down && docker network prune -f && docker compose up -d`. Pruned 4 stale networks (`vera-bot_bot-db-net`, `openclaw_default`, `diun_default`, `massage-bot-test_bot-db-net`). Prod `/health` 200, bot authenticated, no more crash loop.
- **Deferred cleanups also closed in incident 2**:
  - `chmod 600 /opt/vera-bot/.env` (was 0644 — group-readable credentials) ✓
  - Removed stale `docker-compose.override.yml` (167 bytes from 2026-01-20, referenced dead `registry.gitlab.com/kfilin/massage-bot:latest`, service name `massage-bot` didn't match main `app` service — was a no-op) ✓
- **Investigation notes for future reference**:
  - `No .env file found or error loading .env file: open .env` is a **benign false positive** in logs — godotenv looks for a literal `.env` file in the binary's CWD, but env vars ARE loaded via compose's `env_file:` directive (confirmed by `docker exec printenv DB_HOST` returning `db`). Do not chase.
  - `nc -zv` from busybox is misleading — returns 0 even when actual TCP connection times out. Use `nc -w 5` (with timeout) for real connectivity test.
  - Bridge drop with 0 packets in DOCKER-FORWARD = veth endpoint problem. The fix is `docker network prune -f` after compose down, not just `docker compose down`.
- **Commits**: `08a50a9` (auth_date), `fe16a0e` (webapp move), `7975878` (booking split), `e66ef8a` (memory limit), `119bc85`/`d7ed7ba`/`b6a9fdb`/`8c99feb` (P2s). Incident 2 fix was a server-side operational change; no code commit needed (docker-compose.yml was already correct).
- **Forensic snapshots**: `/tmp/vera-prod-pre-fix.log` (1345 lines), `/tmp/vera-prod-pre-fix.json`, `/tmp/vera-net-pre-fix.json` on the server, retained for the record.

### 42. [DONE] Audit and Slim `.agent/` Folder — Migrate to Agentic OS v2 Template
- **Status**: Completed (2026-06-17)
- **Priority**: High (blocks future agent sessions; current harness has stale + duplicated content)
- **Resolution**: Identified `agentic-os` (`/home/kirillfilin/Projects/agentic-os/`) as canonical template source via WORKSPACE.md. Added new **Operational Rules** section to `agentic-os/AGENTS.md` (commit `0919985`). Re-hydrated `massage-bot/AGENTS.md` from template (commit `0c039b6`) with extended Child DOX Index. Created `.pi/skills/` mirror (Agent Skills standard) for all 7 canonical skills. Refreshed `.agent/Project-Hub.md`. Created handoff document for agentic-lab-2.0 cleanup at `Cleanup/cleanup.md`. Single AGENTS.md pattern (with merged Operational Rules) works in pi's auto-load model.
- **Commits**: `0919985` (agentic-os), `0c039b6` (massage-bot hydration)
- **Goal**: Analyze the current `.agent/` folder, retain only project-specific value, and adopt the **agentic-lab-2.0 OS template** as the canonical harness.
- **Current `.agent/` inventory** (11 files):
  | File | Decision | Reason |
  |---|---|---|
  | `HARNESS_GUIDE.md` (3.8 KB) | REMOVE | Universal meta-doc, superseded by agentic-lab-2.0 docs |
  | `handoff.md` (root) | REMOVE | Outdated — references "Phase 13 Redis" from an old session |
  | `Project-Hub.md` | KEEP & REFRESH | Project-specific, needs current state |
  | `project-config.env` | KEEP | `HYDRATED=true`, `PROJECT_NAME=massage-bot`, `GIT_MAIN_BRANCH=master` |
  | `skills/startup.md` | KEEP & REFRESH | Mandatory startup, references graphify which is mostly offline |
  | `skills/handoff.md` | KEEP & REFRESH | End-of-session routine |
  | `skills/obsidian_management.md` | KEEP | Vault path config |
  | `skills/hydrate-harness.md` | REMOVE | Superseded by agentic-lab-2.0 model |
  | `skills/fleet-orchestration.md` | KEEP & SYNC | agentic-lab-2.0 source skill |
  | `skills/fleet-roles.md` | KEEP & SYNC | agentic-lab-2.0 source skill |
  | `skills/multi-modal-intelligence.md` | KEEP & SYNC | agentic-lab-2.0 source skill |
  | `global-skills/` (15 files) | KEEP | Anthropic-style methodology library (TDD, debugging, etc.) |
  | `global-skills_legacy_20260613/` (25 files) | REMOVE | Legacy duplicate workspace; user confirmed "anything experimental can be removed" |
- **Template files to import from agentic-lab-2.0** (after project-agnostic cleanup — see constraint below):
  - [AgenticLab 2.0/Checkpoints/2026-06-15-handoff.md](file:///home/kirillfilin/Projects/agentic-lab-2.0/docs/AgenticLab%202.0/Checkpoints/2026-06-15-handoff.md) — handoff format reference
  - [AgenticLab 2.0/Checkpoints/2026-06-15-handoff-phase3.md](file:///home/kirillfilin/Projects/agentic-lab-2.0/docs/AgenticLab%202.0/Checkpoints/2026-06-15-handoff-phase3.md)
  - [AGENTIC_OS_INTEGRATION.md](file:///home/kirillfilin/Projects/agentic-lab-2.0/docs/AGENTIC_OS_INTEGRATION.md) — explains how a project integrates with Agentic OS
  - [Agent-Prompt.md](file:///home/kirillfilin/Projects/agentic-lab-2.0/docs/Agent-Prompt.md) — session entry / cost control rules
  - [AGENTS.md](file:///home/kirillfilin/Projects/agentic-lab-2.0/docs/AGENTS.md) — **docs-folder** AGENTS.md (note: differs from root AGENTS.md; this one governs the `docs/` subfolder)
  - [Bot-Commands.md](file:///home/kirillfilin/Projects/agentic-lab-2.0/docs/Bot-Commands.md) — command surface reference
- **⚠️ Constraint**: User explicitly stated the agentic-lab-2.0 docs **must be cleaned up first to be project-agnostic before adoption**, and "we'll take care of them together". So:
  - Do NOT copy these files verbatim into massage-bot.
  - First, in a joint session with the user, scrub domain-specific references (e.g., "Agentic Lab", "agentic-lab", "bridge-2", "orchestrator-2", "Pilot/Concierge/Architect" role names if not relevant) and parameterize them.
  - Then import the cleaned versions into massage-bot's `.agent/` and `docs/`.
- **Success criteria**:
  - `.agent/` has only project-relevant files; no duplicates with `global-skills/`.
  - `Project-Hub.md` reflects current state (Sprint #5 coverage done, Phase 4 testcontainers in progress, Clinical Patterns backlog #30 active).
  - `startup.md` and `handoff.md` follow the agentic-lab-2.0 pattern.
  - `legacy_20260613` folder removed.
  - One end-to-end session passes startup → handoff without warnings.

### 43. [TODO] Consolidate Local Projects to `/home/Projects/`
- **Status**: Backlog
- **Priority**: Medium (housekeeping; enables consistent scripts and IDE paths)
- **Goal**: Move all active code projects into a single root: `/home/kirillfilin/Projects/`.
- **Current scattered locations** (from `ls /home/kirillfilin/Documents/`):
  - `massage-bot/` → move to `/home/kirillfilin/Projects/massage-bot/`
  - `watchtower-masterbot/` → move to `/home/kirillfilin/Projects/watchtower-masterbot/`
  - `mcp-server/` → move to `/home/kirillfilin/Projects/mcp-server/`
  - Others under `Documents/` need audit (anything with `.git`, `go.mod`, `package.json`).
- **Pre-existing `/home/kirillfilin/Projects/` content** (keep in place):
  - `agentic-lab` (legacy, see #42 for cleanup)
  - `agentic-lab-2.0` (template source, do NOT move)
  - `agentic-os` (template source, do NOT move)
  - `Antigravity_On_Steroids` (template source, do NOT move)
  - `TIL` (knowledge base, do NOT move)
- **Tasks**:
  1. Audit `/home/kirillfilin/Documents/` for project directories.
  2. Create new paths under `/home/kirillfilin/Projects/`.
  3. `git remote -v` to confirm each project before move.
  4. `mv` (preserves git metadata), then `git status` to verify clean.
  5. Update IDE workspace files (e.g., `massage-bot.code-workspace`) and shell aliases if they hardcode `Documents/` paths.
  6. Update `startup.md`, `Project-Hub.md`, and any `.agent/skills/*.md` that reference `~/Documents/massage-bot/`.
- **Risk**: Hardcoded paths in scripts, CI, and IDE configs. Need `grep -r "Documents/massage-bot"` and `grep -r "Documents/watchtower-masterbot"` audit.
- **Verification**: All projects still build/test; IDE workspaces open; CI references resolve.

### 44. [DONE] CI/CD Pipeline Audit
- **Status**: Backlog
- **Priority**: Medium
- **Goal**: Audit `.gitlab-ci.yml` and deploy scripts, fix known flaws.
- **Known issues** (from user + server inspection + cross-validation with `what_to_fix.md`):
  - **Go version mismatch**: `.gitlab-ci.yml` L7 declares `GO_VERSION: "1.25.3"` but L14 uses `image: golang:1.23` for tests. `Dockerfile` uses `golang:1.25-alpine`. Tests and prod builds run on different Go versions — subtle incompatibilities possible. `what_to_fix.md` #2.4 also flagged this.
  - **`deploy.sh` is broken** (see #41): wrong ports, no `cd $APP_DIR`, uses `docker run` while production uses `docker compose`.
  - **`deploy_home_server.sh` is the working path** — uses `docker compose -f docker-compose.yml -f deploy/docker-compose.prod.yml build --no-cache --pull`. Audit but don't break.
  - **Other CI images unpinned**: `docker:latest`, `alpine:latest` — can break silently on upstream changes.
  - **No port-collision pre-flight**: the current P0 incident could have been caught at deploy time.

**Audit progress** (2026-06-17):
- ✅ Go version mismatch: `.gitlab-ci.yml` test stage now `image: golang:1.25.3-alpine` (commit `ef5e173`).
- ✅ `deploy.sh`: New in-repo wrapper at `scripts/deploy.sh` (commit `ea1377a`) — `test|prod` arg, port-collision pre-flight, docker compose under the hood. Old per-env scripts (`deploy_home_server.sh`, `deploy_test_server.sh`) kept for now to avoid breaking GitLab CI which still calls them.
- ✅ `docker-compose.yml` resource limits: `memory: 512M` added (commit `e66ef8a`).
- ✅ **Image pinning** (2026-06-17): `alpine:latest` → `alpine:3.21` (CI runtime + Dockerfile); `docker:latest` → `docker:27-cli`; `docker:dind` → `docker:27-dind`; `caddy:latest` → `caddy:2.8-alpine` (dev compose). All base images now pinned to specific versions.
- ✅ **Backup-restore verification** (2026-06-17): New `scripts/verify_backup.sh` with explicit exit codes (0/1/2/3/4) for happy path, missing arg, corrupt ZIP, missing entries, and invalid JSON. Tested against synthetic good/bad/corrupt backups.
- ✅ **GitLab CI manual prod gate audit** (2026-06-17): Gate is functionally correct (`when: manual` + `needs: ["run-tests"]` + branch restriction). Migrated deprecated `only:` to modern `rules:` syntax. Added `environment:` blocks for both deploy jobs (enables GitLab deployment board + rollback UI).
- ✅ **`go test ./...` permission fix** (2026-06-17): `.gitlab-ci.yml` test job changed from `go test ./...` to `go test ./cmd/... ./internal/...`. The old pattern failed on this dev machine with `open postgres_data: permission denied` (UID 70 postgres owns the volume); would have failed in CI too on any runner with similar permission layout. Also applies to `go vet` calls.
- ✅ **Testcontainers as direct dep** (2026-06-17): `go.mod` now has `github.com/testcontainers/testcontainers-go` and `modules/postgres` as direct requires (were `// indirect` because the import is gated by `//go:build integration`).
- ✅ **`docker-compose.override.yml` removed** (2026-06-17): 167-byte stale override from 2026-01-20 was deleted from `/opt/vera-bot/` during the #41 incident-2 fix. Was a no-op (service name `massage-bot` didn't match main `app` service) and referenced the dead `registry.gitlab.com/kfilin/massage-bot:latest` image. No replacement needed — the base `docker-compose.yml` is the source of truth.
- **Tasks**:
  1. Pin all images to specific versions (e.g., `golang:1.25-alpine`, `docker:27-dind`, `alpine:3.20`).
  2. Add port-collision pre-flight check to deploy scripts.
  3. Add `memory: 512m` to `docker-compose.yml` resource limits (currently only CPU).
  4. Add backup verification step (daily ZIP backup exists per startup.md, but restore is never tested).
  5. Review GitLab CI manual gate for prod — verify it actually gates correctly.
  6. Decide: keep `deploy.sh` as a thin wrapper around `docker compose`, OR delete it.
- **Source**: User (2026-06-16) + `what_to_fix.md` review (cross-validated with server inspection).

### 45. [DONE] Git Sync Hygiene (PC ↔ Server)
- **Status**: ✅ DONE 2026-06-18
- **Server-state inspection findings** (2026-06-18 01:15):
  - **AGENTS.md drift was already resolved** — the BACKLOG's claim of 1550B and stale skill references was stale itself; the server's `AGENTS.md` is 10793B and matches local byte-for-byte. Probably resolved by the recent deploy of `8e150f4` (which did `git reset --hard origin/master`).
  - **HEAD was 2 commits behind** (`9577fcc` vs `5c920a4` on origin) — the server had my deploy.sh fix in its working tree (from a `scp` during #46) but the git tree still had the old version. Resolved by running the same `git fetch && git reset --hard origin/master` pattern that `scripts/deploy.sh` uses.
  - **Untracked drift removed**:
    - 🗑️ `/opt/vera-bot/deploy.sh` (root, 1245B, Dec 26 2025) — legacy script, used `docker pull registry.gitlab.com/...` against a dead image. Superseded by `scripts/deploy.sh`.
    - 🗑️ `/opt/vera-bot/docker-compose.yml.backup` (1387B, Jan 9 2026) — stale backup, not in active use.
  - **Intentionally left alone**:
    - `.env.backup` (gitignored, 0600 perms, Jan 9 2026) — pre-credential-rotation snapshot. Could be useful as a fallback; deletion is a separate decision.
    - `.env`, `.env.test`, `data/`, `data_test/`, `telegram_api_data/`, `credentials.json` — all gitignored, all expected to be server-local.
- **New convention** (added to `AGENTS.md` Guardrails): **Server Read-Only Convention** — `/opt/vera-bot/` is read-only except for `data/`, `credentials.json`, `.env`, `.env.test`. All other changes flow through `scripts/deploy.sh prod` (which does `git reset --hard origin/master`). No `scp`, no `ssh ... vi`, no `git commit` on the server. Root cause of the drift I just fixed was a `scp` during #46 that bypassed the deploy script.
- **Verification**: `git status` clean on both sides; deploy script's pre-flight (now fixed in #46) works on the live server; prod health 200 throughout.
- **Tasks** (all DONE):
  1. ✅ `ssh server` state inspection.
  2. ✅ Diffed server files vs local; identified 3 drift items.
  3. ✅ Deleted 2 stale files; left 1 (`.env.backup`) as optional follow-up.
  4. ✅ Server synced with origin/master (`5c920a4`).
  5. ✅ Convention documented in `AGENTS.md` Guardrails.
  6. ✅ `startup.md` path updates deferred — depends on #43 (project dir rename), still open.

### 47. [DONE] Enforce Graphify as Mandatory Startup Step
- **Status**: ✅ DONE 2026-06-18
- **Resolution**: Removed "skip silently" fallback from `.pi/skills/startup/SKILL.md` Step 1. Added **Graphify Mandatory (No Skip)** guardrail to `AGENTS.md`. Hardened AGENTS.md startup procedure language to "MANDATORY, no skip — install if missing, rebuild if stale."
- **Files changed**:
  - `AGENTS.md` — added guardrail + hardened Step 1 language
  - `.pi/skills/startup/SKILL.md` — Step 1 rewritten (no fallback), status section updated
- **Tasks** (all DONE):
  1. ✅ Updated startup SKILL.md Step 1 — graphify MUST run, no fallback.
  2. ✅ Added "Graphify Mandatory" guardrail to AGENTS.md.
  3. ✅ Ran graphify queries: 905 nodes, 2,938 edges, 26 communities.
  4. ✅ Pushed to GitHub + GitLab.

### 46. [DONE] Fix `scripts/deploy.sh` port-collision pre-flight (broken for normal deploys)
- **Status**: ✅ DONE 2026-06-18
- **Resolution** (commit `2e007da`): replaced the naive `ss` check with a **smart pre-flight** that only aborts when the port is bound by something OUTSIDE the `vera-bot` compose project. Verified live on the server against three states: our container bound to 8082 (proceeds), free port (proceeds), simulated rogue binding on 9999 (aborts with diagnostic info).
- **Approach chosen**: option 2 (smarter check) rather than option 1 (`--force` flag). The `ss -tlnp` PID-parsing approach in the original BACKLOG text is not viable on the server (cross-namespace, no `CAP_NET_ADMIN`), so we identify "our" binding via the `com.docker.compose.project=vera-bot` docker label — reliable on both local and server.
- **Bonus fix**: guarded the `.env` `HOST_WEBAPP_PORT=` read with `|| true` — under `set -euo pipefail`, grep returning 1 on a missing key would have aborted the script silently.
- **Mirrored** to `/opt/vera-bot/scripts/deploy.sh`. Prod health 200 after push.
- **Tasks** (all DONE):
  1. ~~Add a `--force` / `--skip-port-check` flag~~ — superseded by smarter check.
  2. ✅ Make the pre-flight smarter — only abort if the bound process is NOT in our compose project.
  3. ✅ Tested live (3 scenarios, all pass).
  4. ✅ Pushed to GitHub + GitLab, mirrored to server, prod health 200.
- **Source**: Discovered during 2026-06-18 prod deploy of commit `8e150f4` (see `~/Documents/my_obsidian_vault/Bridge/massage-bot-project/Checkpoints/Handoff-2026-06-18-0005.md`, "Decisions" section, last bullet).

---

## 📋 Session 2026-06-20 00:30 — #52 Groq→local Whisper switch (DONE)

### 52. [DONE] Switch Groq Whisper API → self-hosted faster-whisper

- [x] **Created `internal/adapters/transcription/local.go`** — self-hosted Whisper adapter
  - OpenAI-compatible multipart POST to `http://whisper:8000/v1/audio/transcriptions`
  - No API key, ru language forced initially (later removed to match agentic-lab)
  - 120s timeout, full response body read + json.Unmarshal
  - Matches agentic-lab-2.0 connect_handler.go pattern exactly
- [x] **Deleted `groq.go` and `groq_test.go`**
- [x] **Updated config**: `GroqAPIKey` → `WhisperBaseURL` (env `WHISPER_BASE_URL`)
- [x] **Updated all docs**: README, DEVELOPER.md, files.md, .env.example
- [x] **Updated `.env` on server**: `WHISPER_BASE_URL=http://whisper:8000/v1/audio/transcriptions`
- [x] **Deployed to prod** (commits: `123bfad`, `cce200a`, `4a89d7b`, `ccfa780`)
- [x] **Resolved Blocker**: Stopped the local dev bot container (`massage-bot-app-1`) which was running in the background on the developer machine and stealing Long Polling updates.
- [x] **Verified Transcription**: Once the local bot container was stopped, the production bot successfully received updates, invoked the local Whisper instance (`Systran/faster-whisper-small`), transcribed the audio, and generated the review buttons in Telegram.

#### Last updated: 2026-06-20 01:25

---

#### Previous: Last updated: 2026-06-18 20:45 (#47 done)


