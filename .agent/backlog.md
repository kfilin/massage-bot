# Project Backlog

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

#### Last updated: 2026-03-20

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

### 34. [TODO] Phase 4: Integration Testing with Testcontainers

- **Status**: Backlog
- **Priority**: Medium
- **Goal**: Maximize project test coverage (aiming for >80% total) by implementing integration tests for the `internal/storage` (Postgres) and `internal/delivery/telegram` (bot routing) layers.
- **Rationale**: Pure unit tests hit a ceiling at ~42% repository coverage due to the heavy reliance on raw DB queries (`postgres_repository.go` is over 1000 lines) and external API interactions (`telebot.Bot` methods). To properly verify the database layer without complex and brittle SQL mocks, we need a real, ephemeral database.
- **Implementation Scope**:
  1. Add `github.com/testcontainers/testcontainers-go` and `github.com/testcontainers/testcontainers-go/modules/postgres` to the project.
  2. Create a test suite for `internal/storage` that spins up an isolated Postgres container before tests run, applies the schemas, and tests the CRUD operations against it.
  3. Explore using `httptest` servers to mock Telegram API responses so `bot.go` logic can be verified.
- **Success Criteria**: `internal/storage` achieves >80% test coverage, raising the total project coverage significantly above the current ~42% baseline.
### 35. [TODO] WebApp Handler Refactoring
- **Status**: Backlog
- **Priority**: Medium
- **Goal**: Move WebApp handlers from `cmd/bot/webapp_handler.go` to `internal/delivery/web/`.
- **Rationale**: Currently, handlers are mixed with application entry logic. Moving them to `internal/` ensures better architecture alignment and follows the "package by feature/layer" pattern used in the rest of the project.

### 36. [TODO] Test Coverage Hardening (80%+)
- **Status**: Backlog
- **Priority**: High
- **Goal**: Increase repository test coverage from ~42% to 80%+.
- **Rationale**: Critical logic in `internal/storage` and `internal/delivery/telegram` needs better verification. Complements item #34 (Testcontainers).

### 37. [TODO] Grafana Dashboard Sync
- **Status**: Backlog
- **Priority**: Medium
- **Goal**: Update Grafana dashboards to include the new metrics documented in `docs/API.md`.
- **Rationale**: Ensures the therapist and admin have visual parity with the underlying telemetry.
