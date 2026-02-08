# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v5.6.4] - 2026-02-08

### Fixed (v5.6.4)

- **Metrics Intelligence**: Fixed `vera_active_sessions` to accurately track real-time concurrent users (was previously 0).
- **Token Monitoring**: Resolved `vera_token_expiry_days` bug where the metric would initialize to negative values or become stale. It now auto-refreshes on every token usage and initializes correctly on startup.

### Changed (v5.6.4)

- **Dashboard**: Updated Grafana configuration to focus on actionable signals:
  - Removed "Clinical Note Length" (vanity metric).
  - Added "API Latency" and "Request Rate" for external dependency health.
  - Added "Avg Booking Lead Time" to track patient planning behavior.

## [v5.6.3] - 2026-02-08

### Changed (v5.6.3)

- **TWA UI**: Added scrollable note container to "Medical Card" to prevent long histories from overflowing the screen.
- **Dialogue View**: Implemented automatic date separators (`📅 DD.MM.YYYY`) in patient notes history for better readability.
- **Transcription**: Added "продолжение следует" to the hallucination filter to reduce noise in silent/background-noise voice messages.
- **Documentation**: Comprehensive update of `README.md`, `VERA_GUIDE_RU.md`, `USER_GUIDE.md`, and technical docs to align with v5.6.3 code state.

### Fixed (v5.6.3)

- **TWA Button**: Fixed a JavaScript variable scope error (`ReferenceError: p is not defined`) causing the "New Appointment" button to be unresponsive.
- **Admin Voice Reply**: Fixed an issue where admin voice replies were not being routed to the patient or logged in the dialogue history.

## [v5.6.2] - 2026-02-07

### Added (v5.6.2)

- **Walkthrough Guide**: Created a new `walkthrough.md` in the root directory for convenient testing and feature verification (as requested by the user).

### Fixed (v5.6.2)

- **Admin TWA Access**: Resolved a critical logic error in `webapp.go` where an admin's administrative status was lost correctly when viewing patient records.
- **TWA Cancellation**: Fixed critical redirect bug by replacing page reloading with instant DOM updates. Cancellation is now seamless and does not trigger navigation.
- **Manual Booking**: Fixed logic where "Manual Booking" assigned the appointment to the Admin instead of the patient (by using correct deep-link ID logic).
- **Documentation**: Renamed `.agent/sop/feature-release.md` to `docs/CI_CD_Pipeline.md` for better discoverability.

## [v5.6.1] - 2026-02-06

### Added (v5.6.1)

- **Multi-Therapist Support**: Updated configuration and handlers to support a list of therapist IDs. Both Vera and the lead admin now receive real-time booking notifications.
- **Dynamic Bot Links**: Refactored startup to automatically detect bot username (`@vera_massage_bot` or `@VeraDevBot`). All TWA links (deep links, record buttons) are now dynamically generated using the active bot's identity.
- **Patient Notifications**: Implemented automated confirmation messages for patients when an admin books an appointment on their behalf via the "Manual Booking" flow.
- **WebDAV Identity**: WebDAV server now identifies itself with the correct deployment context.

### Changed (v5.6.1)

- **Admin Logic**: Expanded admin permissions to include all therapists, enabling them to use administrative deep links in the TWA.
- **Deployment**: Consolidated network configuration to prioritize stability and outbound API connectivity.

### Fixed (v5.6.1)

- **Link Redirection**: Fixed a bug where TWA "Записать" buttons occasionally pointed to the wrong bot instance across dev/prod environments.
- **TWA Auth Regression**: Resolved an issue where therapists were treated as regular patients in the TWA, preventing them from accessing admin search/booking.

## [v5.6.0] - 2026-02-06

### Added (v5.6.0)

- **TWA Actions**: Implemented "Add Appointment" (➕) and "View Card" (📄) quick-action buttons in Admin Search.
- **Deep Linking**: Added support for `manual_ID` start parameter to pre-fill patient data when admins trigger booking from TWA.
- **Patient CTA**: Added "🗓 Записаться" (Schedule) button for patients in TWA with direct bot redirection.
- **UI/UX**: Implemented high-density stat cards and conditional "Empty States" (icons + help text) for all TWA sections.
- **Mobile responsiveness**: Redesigned stat grid for mobile (2-column layout + spanning "Next Appointment").

### Changed (v5.6.0)

- **Record Template**: Cleaned up and unified CSS/HTML for better performance and maintainability.
- **Backlog**: Marked items 18 (Stat Cards) and 20 (Empty States) as DONE.

### Fixed (v5.6.0)

- **Data Formatting**: Standardized date display (DD.MM.YYYY) across all TWA cards, removing redundant time strings where unnecessary.

## [v5.5.2] - 2026-02-06

### Added (v5.5.2)

- **Organization**: Created root `ARCHIVE/` directory for historical session logs and documentation.
- **Agent SOPs**: Renamed `.agent/workflows/` to `.agent/sop/` to distinguish from CI/CD machine workflows.
- **Convention**: Implemented ISO 8601 (`YYYY-MM-DD`) naming for `handoff` and `last_session` files for chronological sorting.
- **Automation**: Updated Checkpoint SOP with automated rotation and archiving logic.

### Fixed (v5.5.2)

- **Linting**: Fixed over 100 Markdown linting errors across all documentation files (headings, tables, code blocks).
- **Documentation**: Corrected broken links and updated `docs/files.md` to reflect the new structure.

## [v5.5.1] - 2026-02-06

### Added (v5.5.1)

- **Accessibility**: Implemented full keyboard navigation (`:focus-visible` outlines) and Screen Reader support (`aria-expanded`, `role="button"`) for the TWA Patient Card.
- **Clean Code**: Integrated accessibility logic directly into `record_template.go` (no external deps).

## [v5.5.0] - 2026-02-05

### Added (v5.5.0)

- **GitHub→GitLab Mirroring**: Automated repository sync using HTTPS + Personal Access Token. Pushes to GitHub automatically trigger GitLab CI/CD pipeline.
- **TWA InitData Auth**: Cancel appointments now use Telegram's native `initData` authentication instead of URL tokens. Cryptographically signed by Telegram, never expires during session.
- **Admin TWA Features**: Admins can now view a full patient list (autoload) and override the 72h cancellation restriction in the Web App.
- **Bot Command**: Added `/patients` command to list recent patients with direct TWA links.

### Changed (v5.5.0)

- **CI/CD Architecture**: GitHub handles tests/builds only; GitLab handles all deployments. Eliminates duplicate deploy attempts.
- **TWA Cancel UX**: Better error messages for patients instead of cryptic "Invalid token" errors.
- **Documentation**: Updated `Collaboration-Blueprint.md` to "Gold Standard" and `files.md` to reflect current project structure.

### Fixed (v5.5.0)

- **Invalid Token Error**: Fixed HMAC mismatch caused by whitespace handling inconsistencies between Go backend and JS frontend (`ids` are now trimmed).
- **Stale Token Bug**: Resolved issue where TWA cancel would fail with "Недействительный токен" after deployments due to stale URL tokens. InitData auth is session-based and immune to this.
- **Deploy Scripts in Git**: Added `scripts/` directory to Git tracking (was previously in `.gitignore`).

## [v5.4.0] - 2026-02-04

### Added (v5.4.0)

- **Menu Button**: Added "Открыть карту" (Open Card) button for quick one-click access to the Telegram Web App directly from the bot's chat interface.

### Fixed (v5.4.0)

- **TWA Cancellation**: Fixed critical bug where cancelled appointments would reappear on patient cards after page refresh. Root cause was that cancellations only deleted from Google Calendar but not from the local database cache.
- **Repository Interface**: Added `DeleteAppointment` method to properly remove cancelled appointments from both Google Calendar AND the local database.
- **Menu Button Panic**: Fixed nil pointer dereference in SetMenuButton by using raw API call instead of telebot library method.

### Changed (v5.4.0)

- **Service Architecture**: Updated `appointment.Service` to include database repository for proper appointment lifecycle management.
- **Test Coverage**: Expanded to **37.8%** overall (up from 19.5%, +18.3pp), exceeding the 30% target.

## [v5.3.7] - 2026-02-01

### Fixed (v5.3.7)

- **App Stability**: Permanently resolved "Invalid Token" and crash loops by decoupling the WebApp server lifespan from the Telegram Bot connection within `main.go`. The WebApp now remains online even if the Bot cannot reach the Telegram API.
- **DNS Collision**: Resolved a critical Docker DNS conflict where Caddy round-robined requests between Production (`massage-bot`) and Test (`massage-bot-test`) containers on the shared network.
- **Build System**: Forced cache validation in `bot.go` to prevent stale binaries from persisting in deployments.

### Added (v5.3.7)

- **Request Tracing**: Added debug logging to `webapp.go` middleware to trace incoming requests (Method, URL, RemoteAddr) for easier routing diagnosis.

## [v5.3.4] - 2026-01-31

### Added (v5.3.4)

- **Admin Arsenal**: Restored `/edit_name <telegram_id> <new_name>` command. This allows admins to manually correct patient names in both the database and Markdown synchronization files, bypassing fragile auto-sync logic.

## [v5.3.3] - 2026-01-31

### Fixed (v5.3.3)

- **TWA Authentication**: Implemented permanent `AUTH ERROR` diagnostics to resolve "Invalid Token" issues.
- **Network Topology**: Isolated database traffic to a private `bot-db-net` bridge. This prevents DNS collisions between Prod and Test environments on the same host and ensures `db` hostname consistently resolves correctly.
- **Environment Parity**: Synchronized all Git remotes (GitHub, GitLab) and updated project standards to ensure parity across Local, Test, and Production instances.

## [v5.2.3] - 2026-01-30

### Fixed (v5.2.3)

- **TWA Connectivity**: Resolved connectivity issues by switching test network to `external` mode.
- **Documentation**: Streamlined project documentation structure.

## [v5.2.2] - 2026-01-30

### 🚀 Infrastructure & DevOps

- **Twin Environment Unification**:
  - **Standardized Networking**: Both Production and Test now use `caddy-test-net` as the single source of truth defined in the base `docker-compose.yml`.
  - **Cleanup**: Removed legacy `caddy-proxy` sidecar and `bot-network` from Production override.
  - **Result**: Production and Test containers are now structurally identical.
- **CI/CD Pipeline Refactor**:
  - **Target Shift**: Pipeline now automatically deploys to **Test Environment** (`deploy_test_environment`) on push to master.
  - **Manual Production**: Added `deploy_production` job (manual trigger) for explicit promotion.
  - **Cleanup**: Removed legacy Kubernetes job (`deploy_staging`).

## [v5.2.1] - 2026-01-29

### Added (v5.2.1)

- **Dual Folder Strategy**: Deployed a fully isolated Test Environment (`/opt/vera-bot-test`) running on ports 9082 (App) and 9083 (Metrics), separate from Production.
- **Dedicated Config**: Added `docker-compose.test-override.yml` and parameterized port usage for flexible deployments.

## [5.1.1] - 2026-01-29

### Fixed (5.1.1)

- **TWA Performance**: Restored "Lightning Fast" loading speeds by switching from synchronous Google Calendar sync to a **Local Database Cache** strategy.
- **TWA Cancellation**: Removed blocking confirmation dialog that caused freezing on iOS devices; cancellation is now instant.
- **Data Integrity**: Fixed a critical schema bug where the `appointments` table was missing, causing cache failures.
- **Network Reliability**: Added `ngrok-skip-browser-warning` header to prevent silent failures during local testing (harmless in production).
- **Media Uploads**: Adjusted file size limit validation to 20MB to align with Telegram Bot API constraints.

## [5.1.0] - 2026-01-27

### Added (5.1.0)

- **Enhanced Observability**: Implemented application-wide `DEBUG` logs for database initialization, repository operations, Google Calendar API requests, and Telegram bot interactions.
- **Duplicati Integration**: Verified and documented the setup of a local Duplicati instance for incremental, encrypted backups of clinical data and metadata.

### Changed (5.1.0)

- **Database Logging**: Refined PostgreSQL logging in `docker-compose.yml` to captue only data-modifying queries (`log_statement=mod`) and disabled connection/disconnection logs to eliminate health-check noise.
- **Documentation Cleanup**: consolidated and updated the `.agent` and `docs/` directories, removing 1000+ lines of redundant or outdated documentation.

## [5.0.0] - 2026-01-26

### Added (5.0.0)

- **Phase 4: Technical Excellence** (Series Finale)
- **Robust Scheduling**: Successfully migrated to the official Google Calendar **Free/Busy API**. This ensures 100% accurate collision detection for available slots, automatically respecting "Out of Office", manual blocks, and overlapping events created outside the bot.
- **Automated Backups 2.0**: Implemented a comprehensive backup system that archives both the PostgreSQL database (`pg_dump`) and the clinical Markdown directory (`data/patients/`).
- **Backup Worker**: Added a background ticker that performs an automated backup every 24 hours and sends it directly to the therapist via Telegram.
- **Manual Backups**: Updated the `/backup` command for admins to trigger an immediate full ZIP archive delivery.

### Changed (5.0.0)

- **Infrastructure**: Updated Dockerfile to include `postgresql-client` and `zip` for integrated backup capabilities.
- **Availability Logic**: Refactored the internal scheduler to perform a just-in-time Free/Busy check during the final confirmation step, eliminating the risk of race-condition double bookings.

## [4.4.3] - 2026-01-26

### Fixed (4.4.3)

- **TWA Cancellation**: Restored Admin Alert for patients who are also admins (now only receive one alert due to deduplicated recipient list).

## [4.4.2] - 2026-01-26

### Fixed (4.4.2)

- **TWA Cancellation**: Deduplicated Telegram notifications for admins.
- **TWA Cancellation**: Prevented sending "Admin Alert" to a patient who is also an admin.

## [4.4.1] - 2026-01-26

### Fixed (4.4.1)

- **TWA Cancellation**: Added bot notifications for both patient and admins when a record is cancelled via the Web App.

## [4.4.0] - 2026-01-26

### Added (4.4.0)

- **Phase 2: TWA Evolution**
- Conditional "Cancel" buttons in TWA (enforced 72h-notice rule).
- Real-time "Next Appointment" countdown on the TWA home screen.
- Support for multiple future appointments list in TWA.
- Server-side `/cancel` endpoint with HMAC security and 72h enforcement.
- "Contact Vera" fallback link for late cancellations.
- Improved CSS responsive layout for clinical groupings.

## [4.3.0] - 2026-01-26

### Added (4.3.0)

- **Reminder Service**: New background worker (10-min ticker) for automated patient notifications.
- **Interactive Reminders**: Interactive `[✅ Подтвердить]` and `[❌ Отменить]` buttons for 72h and 24h appointment windows.
- **Smart Admin Reply**: Forwarded patient messages now include a `[✍️ Ответить]` button, allowing admins to reply directly through the bot.
- **Automatic Archiving**: All patient inquiries (text/voice) and admin responses are automatically logged to the patient's medical card (Postgres & Markdown).
- **Confirmation Tracking**: New database metadata layer to track appointment confirmation status.

### Changed (4.3.0)

- **Messaging Loop**: Refined auto-reply logic for unknown patient inputs ("Ваше сообщение получено и передано Вере.").
- **Bot Persona**: Professionalized communication persona for better patient guidance.

### Fixed (4.3.0)

- **Name Input Flow**: Fixed a regression in the booking flow where name input was bypassed by the forwarding middleware.

## [4.2.2] - 2026-01-24

### Fixed (4.2.2)

- **Navigation**: Fixed routing issues for "Back to Service" and "Back to Date" navigation buttons in the booking flow.

## [4.2.1] - 2026-01-24

### Added (v4.2.1)

- **Visit History UI**: New "История посещений" section in TWA showing the 5 most recent confirmed visits.
- **Status Tracking**: Appointment status (confirmed/cancelled) is now synchronized from Google Calendar.

### Changed (v4.2.1)

- **Direct Scrubbing**: Instead of a complex migration, implemented a direct, permanent scrub of the legacy "Ссылки на документы" boilerplate within the `SyncAll` startup flow.

### Fixed (v4.2.1)

- **Sync Logic**: Fixed a bug where TWA visit statistics were limited to a 24-hour window; now uses full history.
- **Data Accuracy**: Cancelled events and "Admin Blocks" are now correctly excluded from clinical visit counts and history.

### Removed (v4.2.1)

- **Redundancy**: Removed the empty "Ссылки на документы" placeholder from Markdown cards.

### Decision Rationale (v4.2.1)

- **Full History Sync**: Pivoted from a 24-hour sliding window to a full history scan for visit statistics. This ensures that a patient's "First Visit" and "Total Visits" remain accurate even if they haven't visited in months.
- **Explicit Status Filtering**: Introduced a `Status` field to appointments to distinguish between "Confirmed" and "Cancelled" events. This prevents administrative noise (cancellations/blocks) from inflating clinical metrics.

## [4.2.0] - 2026-01-24

### Added (v4.2.0)

- **Navigation 2.0**: "Back" button navigation for booking flow (Date → Service, Time → Date).
- **Policy Enforcement**: 72h cancellation warning in bot confirmation and Patient's Card.
- **Categorized Clinical Data**: Summarized document grouping in Patient's Card (Scans, Photos, Videos, Voice Messages, Texts, Others).
- **Professionalism**: Professional "Conventional Commits" standard and squashed history.

### Changed (4.2.0)

- **Scheduler Logic**: Simplified booking slots to hourly intervals (09:00 - 18:00) to ensure therapist breaks.
- **Localization**: Localized TWA badge to "КАРТА ПАЦИЕНТА".
- **Responsive Design**: Optimized TWA layout for mobile devices (responsive stacking of stat boxes).
- **Markdown Purity**: Refined Markdown rendering for clinical notes (fixed headers/bold text).
- **Safety**: Verified and enforced 50MB file upload limits across all interfaces.

## [4.1.0] - 2026-01-23

### Added (4.1.0)

- **Clinical Storage 2.0**: Permanent switch back to Markdown-mirrored filesystem for Obsidian/WebDAV sync.
- **Suffix Tracking**: Implemented `(TelegramID)` folder suffix tracking, allowing therapist-led folder renames in Obsidian.
- **Metrics Stack**: Established Prometheus/Grafana baseline on port 8083.
- **Resilience**: Added a 5-attempt retry loop for Postgres database connections.

## [4.0.0] - 2026-01-20

### Changed (4.0.0)

- **Architecture Pivot**: Decommissioned StirlingPDF in favor of browser-native `window.print()`.
- **The Postgres Return**: Re-implemented PostgreSQL as the primary metadata store for long-term scalability.

## [3.1.15] - 2026-01-18

### Added (3.1.15)

- **Smart Registration**: Robust name extraction from Google Calendar and "Quiet Self-Healing" session management.
- **TWA Auth Expansion**: implemented `initData` self-healing for seamless web-app authentication.

## [3.1.8] - 2025-11-27

### Added (3.1.8)

- **Voice Intelligence**: Integrated **Groq (Whisper)** for voice note transcription.
- **Policy Shift**: Extended the cancellation window from 24h to **72h**.
- **Admin Alerts**: Cross-admin notifications for time blocks and new bookings.

## [2.5.0] - 2025-11-15

### Changed (2.5.0)

- **Menu Evolution**: Switched from one-time keyboards to a persistent **Main Menu** pattern for better UX.
- **Scheduling**: Implemented the "No Weekend" rule, filtering out Saturdays and Sundays from the calendar.

## [2.1.0] - 2025-11-10

### Added (2.1.0)

- **Admin Arsenal**: Introduced the `/block` command for manual schedule blocking.
- **Security**: Implementation of a **Blacklist** to prevent problematic user registrations.
- **Google Meet**: Automated generation of video call links for all consultations.

## [2.0.0] - 2025-11-01

### Changed (2.0.0)

- **Experiment Phase**: Temporary removal of PostgreSQL in favor of pure FS-based state.
- **The OAuth Port Dance**: Successfully resolved host conflicts by moving from port 8080 to **18080** and establishing a dedicated `HEALTH_PORT=8081`.

## [1.0.0] - 2025-10-15

### Added (1.0.0)

- **Initial Core**: Bot structure with Google Calendar integration.
- **Persistence**: Initial Postgres setup for sessions and `token.json` migration to the `data/` volume.
- **Standard**: established the "Magic Question" architectural review process.
