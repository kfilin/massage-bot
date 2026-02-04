# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v5.4.0] - 2026-02-04

### Added

- **Menu Button**: Added "Открыть карту" (Open Card) button for quick one-click access to the Telegram Web App directly from the bot's chat interface.

### Fixed

- **TWA Cancellation**: Fixed critical bug where cancelled appointments would reappear on patient cards after page refresh. Root cause was that cancellations only deleted from Google Calendar but not from the local database cache.
- **Repository Interface**: Added `DeleteAppointment` method to properly remove cancelled appointments from both Google Calendar AND the local database.
- **Menu Button Panic**: Fixed nil pointer dereference in SetMenuButton by using raw API call instead of telebot library method.

### Changed

- **Service Architecture**: Updated `appointment.Service` to include database repository for proper appointment lifecycle management.
- **Test Coverage**: Expanded to **37.8%** overall (up from 19.5%, +18.3pp), exceeding the 30% target.

## [v5.3.7] - 2026-02-01

### Fixed

- **App Stability**: Permanently resolved "Invalid Token" and crash loops by decoupling the WebApp server lifespan from the Telegram Bot connection within `main.go`. The WebApp now remains online even if the Bot cannot reach the Telegram API.
- **DNS Collision**: Resolved a critical Docker DNS conflict where Caddy round-robined requests between Production (`massage-bot`) and Test (`massage-bot-test`) containers on the shared network.
- **Build System**: Forced cache validation in `bot.go` to prevent stale binaries from persisting in deployments.

### Added

- **Request Tracing**: Added debug logging to `webapp.go` middleware to trace incoming requests (Method, URL, RemoteAddr) for easier routing diagnosis.

## [v5.3.4] - 2026-01-31

### Added

- **Admin Arsenal**: Restored `/edit_name <telegram_id> <new_name>` command. This allows admins to manually correct patient names in both the database and Markdown synchronization files, bypassing fragile auto-sync logic.

## [v5.3.3] - 2026-01-31

### Fixed

- **TWA Authentication**: Implemented permanent `AUTH ERROR` diagnostics to resolve "Invalid Token" issues.
- **Network Topology**: Isolated database traffic to a private `bot-db-net` bridge. This prevents DNS collisions between Prod and Test environments on the same host and ensures `db` hostname consistently resolves correctly.
- **Environment Parity**: Synchronized all Git remotes (GitHub, GitLab) and updated project standards to ensure parity across Local, Test, and Production instances.

## [v5.2.3] - 2026-01-30

### Fixed

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

### Added

- **Dual Folder Strategy**: Deployed a fully isolated Test Environment (`/opt/vera-bot-test`) running on ports 9082 (App) and 9083 (Metrics), separate from Production.
- **Dedicated Config**: Added `docker-compose.test-override.yml` and parameterized port usage for flexible deployments.

## [5.1.1] - 2026-01-29

### Fixed

- **TWA Performance**: Restored "Lightning Fast" loading speeds by switching from synchronous Google Calendar sync to a **Local Database Cache** strategy.
- **TWA Cancellation**: Removed blocking confirmation dialog that caused freezing on iOS devices; cancellation is now instant.
- **Data Integrity**: Fixed a critical schema bug where the `appointments` table was missing, causing cache failures.
- **Network Reliability**: Added `ngrok-skip-browser-warning` header to prevent silent failures during local testing (harmless in production).
- **Media Uploads**: Adjusted file size limit validation to 20MB to align with Telegram Bot API constraints.

## [5.1.0] - 2026-01-27

### Added

- **Enhanced Observability**: Implemented application-wide `DEBUG` logs for database initialization, repository operations, Google Calendar API requests, and Telegram bot interactions.
- **Duplicati Integration**: Verified and documented the setup of a local Duplicati instance for incremental, encrypted backups of clinical data and metadata.

### Changed

- **Database Logging**: Refined PostgreSQL logging in `docker-compose.yml` to captue only data-modifying queries (`log_statement=mod`) and disabled connection/disconnection logs to eliminate health-check noise.
- **Documentation Cleanup**: consolidated and updated the `.agent` and `docs/` directories, removing 1000+ lines of redundant or outdated documentation.

## [5.0.0] - 2026-01-26

### Added

- **Phase 4: Technical Excellence** (Series Finale)
- **Robust Scheduling**: Successfully migrated to the official Google Calendar **Free/Busy API**. This ensures 100% accurate collision detection for available slots, automatically respecting "Out of Office", manual blocks, and overlapping events created outside the bot.
- **Automated Backups 2.0**: Implemented a comprehensive backup system that archives both the PostgreSQL database (`pg_dump`) and the clinical Markdown directory (`data/patients/`).
- **Backup Worker**: Added a background ticker that performs an automated backup every 24 hours and sends it directly to the therapist via Telegram.
- **Manual Backups**: Updated the `/backup` command for admins to trigger an immediate full ZIP archive delivery.

### Changed

- **Infrastructure**: Updated Dockerfile to include `postgresql-client` and `zip` for integrated backup capabilities.
- **Availability Logic**: Refactored the internal scheduler to perform a just-in-time Free/Busy check during the final confirmation step, eliminating the risk of race-condition double bookings.

## [4.4.3] - 2026-01-26

### Fixed

- **TWA Cancellation**: Restored Admin Alert for patients who are also admins (now only receive one alert due to deduplicated recipient list).

## [4.4.2] - 2026-01-26

### Fixed

- **TWA Cancellation**: Deduplicated Telegram notifications for admins.
- **TWA Cancellation**: Prevented sending "Admin Alert" to a patient who is also an admin.

## [4.4.1] - 2026-01-26

### Fixed

- **TWA Cancellation**: Added bot notifications for both patient and admins when a record is cancelled via the Web App.

## [4.4.0] - 2026-01-26

### Added

- **Phase 2: TWA Evolution**
- Conditional "Cancel" buttons in TWA (enforced 72h-notice rule).
- Real-time "Next Appointment" countdown on the TWA home screen.
- Support for multiple future appointments list in TWA.
- Server-side `/cancel` endpoint with HMAC security and 72h enforcement.
- "Contact Vera" fallback link for late cancellations.
- Improved CSS responsive layout for clinical groupings.

## [4.3.0] - 2026-01-26

### Added

- **Reminder Service**: New background worker (10-min ticker) for automated patient notifications.
- **Interactive Reminders**: Interactive `[✅ Подтвердить]` and `[❌ Отменить]` buttons for 72h and 24h appointment windows.
- **Smart Admin Reply**: Forwarded patient messages now include a `[✍️ Ответить]` button, allowing admins to reply directly through the bot.
- **Automatic Archiving**: All patient inquiries (text/voice) and admin responses are automatically logged to the patient's medical card (Postgres & Markdown).
- **Confirmation Tracking**: New database metadata layer to track appointment confirmation status.

### Changed

- **Messaging Loop**: Refined auto-reply logic for unknown patient inputs ("Ваше сообщение получено и передано Вере.").
- **Bot Persona**: Professionalized communication persona for better patient guidance.

### Fixed

- **Name Input Flow**: Fixed a regression in the booking flow where name input was bypassed by the forwarding middleware.

## [4.2.2] - 2026-01-24

### Fixed

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

### Changed

- **Scheduler Logic**: Simplified booking slots to hourly intervals (09:00 - 18:00) to ensure therapist breaks.
- **Localization**: Localized TWA badge to "КАРТА ПАЦИЕНТА".
- **Responsive Design**: Optimized TWA layout for mobile devices (responsive stacking of stat boxes).
- **Markdown Purity**: Refined Markdown rendering for clinical notes (fixed headers/bold text).
- **Safety**: Verified and enforced 50MB file upload limits across all interfaces.

## [4.1.0] - 2026-01-23

### Added

- **Clinical Storage 2.0**: Permanent switch back to Markdown-mirrored filesystem for Obsidian/WebDAV sync.
- **Suffix Tracking**: Implemented `(TelegramID)` folder suffix tracking, allowing therapist-led folder renames in Obsidian.
- **Metrics Stack**: Established Prometheus/Grafana baseline on port 8083.
- **Resilience**: Added a 5-attempt retry loop for Postgres database connections.

## [4.0.0] - 2026-01-20

### Changed

- **Architecture Pivot**: Decommissioned StirlingPDF in favor of browser-native `window.print()`.
- **The Postgres Return**: Re-implemented PostgreSQL as the primary metadata store for long-term scalability.

## [3.1.15] - 2026-01-18

### Added

- **Smart Registration**: Robust name extraction from Google Calendar and "Quiet Self-Healing" session management.
- **TWA Auth Expansion**: implemented `initData` self-healing for seamless web-app authentication.

## [3.1.8] - 2025-11-27

### Added

- **Voice Intelligence**: Integrated **Groq (Whisper)** for voice note transcription.
- **Policy Shift**: Extended the cancellation window from 24h to **72h**.
- **Admin Alerts**: Cross-admin notifications for time blocks and new bookings.

## [2.5.0] - 2025-11-15

### Changed

- **Menu Evolution**: Switched from one-time keyboards to a persistent **Main Menu** pattern for better UX.
- **Scheduling**: Implemented the "No Weekend" rule, filtering out Saturdays and Sundays from the calendar.

## [2.1.0] - 2025-11-10

### Added

- **Admin Arsenal**: Introduced the `/block` command for manual schedule blocking.
- **Security**: Implementation of a **Blacklist** to prevent problematic user registrations.
- **Google Meet**: Automated generation of video call links for all consultations.

## [2.0.0] - 2025-11-01

### Changed

- **Experiment Phase**: Temporary removal of PostgreSQL in favor of pure FS-based state.
- **The OAuth Port Dance**: Successfully resolved host conflicts by moving from port 8080 to **18080** and establishing a dedicated `HEALTH_PORT=8081`.

## [1.0.0] - 2025-10-15

### Added

- **Initial Core**: Bot structure with Google Calendar integration.
- **Persistence**: Initial Postgres setup for sessions and `token.json` migration to the `data/` volume.
- **Standard**: established the "Magic Question" architectural review process.
