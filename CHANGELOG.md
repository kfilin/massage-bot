# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
