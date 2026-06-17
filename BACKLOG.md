# Project Backlog

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
### 35. [TODO] WebApp Handler Refactoring
- **Status**: Backlog
- **Priority**: Medium
- **Goal**: Move WebApp handlers from `cmd/bot/webapp_handler.go` to `internal/delivery/web/`.
- **Rationale**: Currently, handlers are mixed with application entry logic. Moving them to `internal/` ensures better architecture alignment and follows the "package by feature/layer" pattern used in the rest of the project.

### 36. [IN PROGRESS] Test Coverage Hardening (80%+)
- **Status**: In Progress — resumed 2026-06-17
- **Priority**: High
- **Goal**: Increase repository test coverage from ~42% to 80%+.
- **Progress (after Sprint 5 + 2026-06-17)**: Overall ~42% → ~70%+. Key gains: storage 32→68.7%, handlers 40→78.1%, cmd/bot 27→66.1%, googlecalendar 53→69.7%, transcription 23→88.2%, services/appointment 86→92.5%.
- **Per-function wins**: NewTranscribeHandler 73→100%, NewUpdatePatientHandler 73.8→100%, tokenFromFile 33.3→100%, backoff 83.3→100%, NewWebAppHandler 57.3→87.9%.
- **Next targets**: `reminder` (81.5%), `domain` (91.7%), `logging` (91.2%), `services/appointment` (92.5%). Structural ceilings: `delivery/telegram` (4.2%), googlecalendar OAuth (getToken/saveToken 0%).
- **Tests added across sessions**: ~121 new test functions.

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

### 41. [DONE] Production Bot Offline — Fix Container Collision + Missing .env
- **Status**: Completed (2026-06-17) — prod bot back online, port 8082 bound, LOG_LEVEL=INFO, auth_date check deployed.
- **Resolution**: Root cause was a zombie `massage-bot` container stuck in `Created` state from a prior failed `docker compose up`, holding the port 8082 Docker-internal allocation even though the port appeared free in `ss` (the allocation is in Docker's network, not the host's listen list). `docker rm -f massage-bot` + `docker compose up -d app` brought it back. The original P0 path-collison narrative (gateway-2 holding 8082) was outdated — gateway-2 had moved to 8088. The .env / TG_BOT_TOKEN issue from the earlier `vera-bot-massage-bot-1` failure (Apr 2026 gitlab image) was resolved by the time the new `massage-bot` image ran (env_file directive honoured). `LOG_LEVEL=DEBUG → INFO` applied via `docker compose up -d` (restart doesn't re-read `env_file`). Removed `HEALTH_PORT=8081` dead override line from `/opt/vera-bot/.env`. Removed the old `vera-bot-massage-bot-1` Exited container. **Remaining from this issue's secondary items**: `chmod 600 /opt/vera-bot/.env` (still 0644) and the `docker-compose.override.yml` on the server (see #44/#45).
- **Commits**: `08a50a9` (auth_date), `fe16a0e` (webapp move), `7975878` (booking split), `e66ef8a` (memory limit), `119bc85`/`d7ed7ba`/`b6a9fdb`/`8c99feb` (P2s).
- **Status**: Backlog (P0 — production down)
- **Priority**: P0
- **Goal**: Restore the vera-bot service. Two stacked issues confirmed by server inspection:
  1. `vera-bot-massage-bot-1` exited with code 1 at `2026-06-16T20:45:36Z`. **Actual root cause** (not port collision): `No .env file found or error loading .env file` → `Environment variable TG_BOT_TOKEN is not set` (fatal). `/opt/vera-bot/.env` exists with the token — `docker compose up` either ran from wrong CWD or `env_file` directive wasn't honored. A second `massage-bot` container exists in `Created` status with zero logs (from `deploy.sh`'s `docker run` path).
  2. Port 8082 is held by `gateway-2` (29h uptime, `0.0.0.0:8082->8080/tcp`). Compose's default `HOST_WEBAPP_PORT:-8082` will continue to collide even after #1 is fixed.
- **Secondary issues to fix in the same change window**:
  - `docker-compose.override.yml` redefines the service as `massage-bot`, conflicting with the `app` service in the base compose — caused the dual-container mess.
  - `deploy.sh` has no `cd $APP_DIR` and uses `-p 8080:8080 -p 8081:8081` (wrong ports). Replace with `deploy_home_server.sh` semantics or remove.
  - `chmod 600 /opt/vera-bot/.env` — currently `0644` (group-readable credentials in production, also confirmed in `what_to_fix.md` #4.1).
- **Proposed fix sequence** (NOT to execute without explicit user approval):
  ```bash
  ssh server "cd /opt/vera-bot && \
    docker compose down && \
    docker rm -f massage-bot 2>/dev/null || true && \
    chmod 600 .env && \
    rm docker-compose.override.yml && \
    sed -i 's/HOST_WEBAPP_PORT:-8082/HOST_WEBAPP_PORT:-8086/' .env && \
    docker compose pull && \
    docker compose up -d && \
    sleep 15 && \
    curl -s http://localhost:8086/health"
  ```
- **Verification**: `docker ps | grep massage-bot` shows Up, `/health` returns 200, smoke-test booking flow on staging first, then repeat on prod.
- **Source**: User report (2026-06-16) + `ssh server` inspection + `what_to_fix.md` cross-reference.

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

### 44. [TODO] CI/CD Pipeline Audit
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
- ❌ `docker-compose.override.yml` on server (167 bytes) — diff vs repo not yet inspected; presumed server-side drift.
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
- **Known drift** (from server inspection):
  - Server `AGENTS.md` is 1550 bytes, references skills that don't exist in current `.agent/skills/` (`database-expert`, `devops-harness`, `ai-integration-expert`, `twa-aesthetics`). Stale — was probably from an older bootstrap.
  - `docker-compose.override.yml` exists on server (167 bytes) but NOT in local repo — implies uncommitted server-side drift, or it was deleted locally.
  - `deploy.sh` on server is dated `дек 26 19:48` (Dec 26, 2025) — much older than other files dated May 2026.
  - `.env` on server is dated `апр 24 08:40` (Apr 24, 2026) — credentials haven't been rotated since April, despite being `0644`.
  - `.env.example` on server is dated `мая 13 14:36` (May 13, 2026) — newer than `.env`, suggesting `.env.example` was updated after `.env` was last touched.
- **Tasks**:
  1. `ssh server "cd /opt/vera-bot && git status && git log --oneline -5"` to see server's actual state.
  2. Diff server files vs local files for non-git-tracked drift (`docker-compose.override.yml`, `AGENTS.md`, `deploy.sh`).
  3. Decide per file: commit server-specific to repo, OR remove as "experimental".
  4. Establish convention: server is read-only except for `data/` and `.env`; everything else comes from git.
  5. Update `startup.md` reference paths once #43 is done.
- **Verification**: `git status` clean on both sides; deploy succeeds without manual intervention; no `docker-compose.override.yml`-style files reappear.

---

#### Last updated: 2026-06-17 (post-P2s; #41 + #44 mostly complete (image pinning, backup verify, manual gate audit done); #34 routing extracted + tested; reminder 81.5→91.4%)

