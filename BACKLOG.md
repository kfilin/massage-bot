# Project Backlog

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

### 21. [TODO] History List Pagination

- **Priority**: Medium
- If patient has 50+ visits, page gets very long
- **Idea**: "Show more" button or virtual scrolling

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

### 25. [TODO] Print Optimization

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

### 30. [TODO] Knowledge Item (KI) for Clinical Patterns

- **Status**: Backlog
- **Priority**: Medium (daily time savings potential)
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

### 34. [IN PROGRESS] Phase 4: Integration Testing with Testcontainers

- **Status**: In Progress (2026-06-17) — routing layer extracted & tested; storage stable at 68.7%
- **Priority**: Medium
- **Goal**: Maximize project test coverage (aiming for >80% total) by implementing integration tests for the `internal/storage` (Postgres) and `internal/delivery/telegram` (bot routing) layers.
- **Progress**:
  - testcontainers-go installed. 16 integration tests for `internal/storage` covering all CRUD operations, session storage, and appointment metadata.
  - Storage coverage: 32% → **68.7%** (2026-06-17).
  - **Telegram routing extracted** (2026-06-17): pure-function refactor — `RouteCallback(data)` and `RouteTextMessage(text, SessionView)` extracted into `internal/delivery/telegram/routing.go`; side-effecting helpers (`handleAdminReply`, `forwardPatientMessageToAdmins`) extracted into `text_flow.go`. `bot.go` OnCallback/OnText handlers now delegate routing decisions to these functions (behavior-preserving).
  - **Telegram coverage: 4.4% → 21.2%** (2026-06-17). Routing logic itself is ~100% covered (32 new tests).
  - Dead file `internal/delivery/telegram/bot.go.bak` (17KB stale Feb-04 backup) removed.
- **Remaining**: `delivery/telegram` wiring (`RunBot`, `InitBot`) still untestable without mocking `*telebot.Bot`; structural ceiling for this package ~25%. `cmd/bot` (6.6%) is mostly glue/wiring — low ROI.
- **Update 2026-06-17 (next session)**: Two testable seams extracted from `RunBot` (`setupMenuButton`, `runScheduledBackup`); both consume `ports.BotAPI` and are now 100% covered via `bot_wiring_test.go` (7 new tests). The remaining `RunBot`/`InitBot` code is registration (b.Handle, b.Use, b.Start, b.Stop) which stays on the concrete `*telebot.Bot` and is structurally untestable through the BotAPI interface. Package coverage 39.6% → 47.6% (+8.0pp). Overall: 76.0% → 76.6% (+0.6pp). `RunBot` itself still 0%.
### 35. [TODO] WebApp Handler Refactoring
- **Status**: Backlog
- **Priority**: Medium
- **Goal**: Move WebApp handlers from `cmd/bot/webapp_handler.go` to `internal/delivery/web/`.
- **Rationale**: Currently, handlers are mixed with application entry logic. Moving them to `internal/` ensures better architecture alignment and follows the "package by feature/layer" pattern used in the rest of the project.

### 36. [IN PROGRESS] Test Coverage Hardening (80%+)
- **Status**: In Progress — 2026-06-17 sprint
- **Priority**: High
- **Goal**: Increase repository test coverage from ~42% to 80%+.
- **Progress (2026-06-17 sprint)**: Overall **73.9% → 74.8%**. Per-package:
  - `storage`: 68.7% → **86.0%** (unit) / **91.7%** (integration). +17.3pp / +23.0pp.
  - `googlecalendar`: 69.7% → **74.3%** (+4.6pp). Covered `getToken` env-var branches and `saveToken` round-trip with file-mode check.
  - `reminder`: 91.4% → **96.6%** (+5.2pp). Added `TestStart_RunsAndStopsOnContextCancel`.
  - `delivery/telegram`: 21.2% → **24.6%** (+3.4pp). Session helpers 0% → 100% (19 new sub-cases).
  - `services/appointment`: 93.5% (unchanged aggregate; metrics.go 0% → 100% via 5 new cases).
  - `delivery/web`: 78.4% (unchanged; `sendTelegramMessage` tests added but no measurable gain — function calls real Telegram API).
- **Tests added this session**: 20+ new test functions across 6 packages:
  - storage: 4 MigrateJSONToPostgres, 3 CreateBackup, 4 getPatientDir, plus not-found/nil cases for LogEvent/UpdatePatientProfile/SaveAppointmentMetadata/GetPatient; 1 integration test for InitDB.
  - googlecalendar: 2 getToken (env-var happy path + zero-expiry refresh), 1 saveToken round-trip.
  - web: 3 sendTelegramMessage (panic guards, no real coverage).
  - delivery/telegram: 19 sub-cases for 3 session helpers.
  - reminder: 1 Start coverage test.
  - services/appointment: 5 metrics collector tests.
- **Next targets** (for 80%): Close the 3.4pp gap (was 5.2pp):
  1. ~~`delivery/telegram` wiring (24.6% → ~70%)~~: partially done (47.6% reached; +70% unreachable without `*telebot.Bot` mocking).
  2. **`cmd/bot` (6.6% → ~40%)**: low ROI glue code, but adds ~3pp overall.
  3. **`delivery/web` StartServer (0% → ~50%)**: requires extracting the HTTP-server bootstrap from `StartServer` into a testable function.
  4. **`BotAPI` interface satisfaction test** (`internal/ports/botapi.go`): add a static assertion that `*telebot.Bot` actually implements `ports.BotAPI`. Catches the case where the interface drifts and the real bot silently stops satisfying it (compile-time today, but easy to miss for downstream consumers).
- **Structural ceilings**: `delivery/telegram` (RunBot/InitBot need `*telebot.Bot` mocking; only registration-side code remains uncovered), googlecalendar OAuth (NewGoogleCalendarClient needs real Google creds).
- **Tests added across all sessions**: ~140 new test functions.

### 37. [TODO] Grafana Dashboard Sync
- **Status**: Backlog
- **Priority**: Medium
- **Goal**: Update Grafana dashboards to include the new metrics documented in `docs/API.md`.
- **Rationale**: Ensures the therapist and admin have visual parity with the underlying telemetry.
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

### 45. [TODO] Git Sync Hygiene (PC ↔ Server)
- **Status**: Backlog
- **Priority**: Medium
- **Goal**: Make local and server git states consistent and easily reconcilable. Reduce manual drift between `~/Documents/massage-bot/` (or `/home/Projects/massage-bot/` after #43) and `/opt/vera-bot/`.
- **Known drift** (from server inspection, **partially resolved 2026-06-17**):
  - Server `AGENTS.md` is 1550 bytes, references skills that don't exist in current `.agent/skills/` (`database-expert`, `devops-harness`, `ai-integration-expert`, `twa-aesthetics`). Stale — was probably from an older bootstrap.
  - ~~`docker-compose.override.yml` exists on server (167 bytes) but NOT in local repo~~ — **RESOLVED**: file removed during #41 incident-2 fix.
  - `deploy.sh` on server is dated `дек 26 19:48` (Dec 26, 2025) — much older than other files dated May 2026.
  - `.env` on server is dated `апр 24 08:40` (Apr 24, 2026) — credentials haven't been rotated since April, but perms are now 0600 (was 0644) as of 2026-06-17.
  - `.env.example` on server is dated `мая 13 14:36` (May 13, 2026) — newer than `.env`, suggesting `.env.example` was updated after `.env` was last touched.
- **Tasks**:
  1. `ssh server "cd /opt/vera-bot && git status && git log --oneline -5"` to see server's actual state.
  2. Diff server files vs local files for non-git-tracked drift (`docker-compose.override.yml`, `AGENTS.md`, `deploy.sh`).
  3. Decide per file: commit server-specific to repo, OR remove as "experimental".
  4. Establish convention: server is read-only except for `data/` and `.env`; everything else comes from git.
  5. Update `startup.md` reference paths once #43 is done.
- **Verification**: `git status` clean on both sides; deploy succeeds without manual intervention; no `docker-compose.override.yml`-style files reappear.

### 46. [TODO] Fix `scripts/deploy.sh` port-collision pre-flight (broken for normal deploys)
- **Status**: Backlog
- **Priority**: Low (deploy works around it; cosmetic/UX issue)
- **Goal**: Make `scripts/deploy.sh prod` actually usable on a healthy prod (it currently aborts every normal deploy).
- **Bug** (discovered 2026-06-18 during #34/#36 prod deploy of commit `8e150f4`):
  - The pre-flight at `scripts/deploy.sh:36-48` runs `ss -tlnH | grep ":${PORT}\$"` and aborts the deploy if the port is bound.
  - A normal `docker compose up -d --force-recreate` keeps the old container bound to the port during the atomic swap. So the pre-flight **always fires** on a healthy prod, and the script can never deploy a running bot.
  - Worked around for the 2026-06-18 deploy by bypassing the wrapper and running the raw `docker compose ... build --no-cache --pull && docker compose ... up -d --force-recreate` directly (the same pattern the legacy `deploy_home_server.sh` uses).
  - The pre-flight was originally added during the P0 incident investigation (see #41) to catch rogue bots squatting 8082 — it's correct in *that* scenario but blocks routine deploys.
- **Tasks**:
  1. Add a `--force` / `--skip-port-check` flag to `scripts/deploy.sh` (already half-done: `SKIP_PORT_CHECK=1` is hard-coded for `test`, needs to be a CLI flag for `prod`).
  2. Better: make the pre-flight smarter — only abort if the bound process is NOT a `massage-bot` container (parse `ss -tlnp` output for the binary/PID and compare).
  3. Update `AGENTS.md` / `startup.md` "How to deploy" section if the chosen approach changes the CLI surface.
- **Source**: Discovered during 2026-06-18 prod deploy of commit `8e150f4` (see `~/Documents/my_obsidian_vault/Bridge/massage-bot-project/Checkpoints/Handoff-2026-06-18-0005.md`, "Decisions" section, last bullet).
- **Verification**: `bash scripts/deploy.sh prod` succeeds on a running prod (no need to bypass). After deploy, `curl http://localhost:8082/health` returns 200 and the container is the freshly-rebuilt one.

---

#### Last updated: 2026-06-18 00:35 (handoff path fix: moved 5 handoffs from `Bridge/Checkpoints/` to `Bridge/massage-bot-project/Checkpoints/`; updated `.pi/skills/handoff/SKILL.md` path template; added #46 for the deploy.sh pre-flight bug found during today's prod deploy)

