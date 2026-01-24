# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [4.2.0] - 2026-01-24

### Added

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
- **Suffix Tracking**: Implemented `(TelegramID)` folder suffix tracking, allowing the therapist to rename folders in Obsidian without breaking bot links.
- **Metrics Stack**: Moved metrics to port 8083; established Prometheus/Grafana baseline.
- **Resilience**: Added a 5-attempt retry loop for Postgres database connections.

## [4.0.0] - 2026-01-20

### Changed

- **Architecture Pivot**: Decommissioned StirlingPDF and removed PDF generation buttons. Reverted to browser-native `window.print()` for stability.
- **Storage Strategy**: Re-implemented PostgreSQL as the primary metadata store, moving away from purely file-based state for enterprise-level reliability.

## [3.1.15] - 2026-01-18

### Added

- **Smart Registration**: Robust name extraction from Google Calendar events and "Quiet Self-Healing" logic for session management.
- **TWA Auth Expansion**: Implemented `initData` self-healing to automatically authenticate TWA users.

## [3.1.8] - 2025-11-27

### Added

- **Voice Intelligence**: Integrated **Groq (Whisper)** for high-speed voice note transcription in clinical records.
- **Policy Shift**: Extended the therapist's cancellation grace period from 24h to **72h (3 days)** based on clinical requirements.
- **Russian Localization**: Initial translation of core bot interfaces and documentation into Russian.

## [1.0.0] - 2025-10-15

### Added

- **Foundations**: Persistent main menus, Google Meet integration, and initial PostgreSQL/Analytics setup.
- **Interview Filter**: Established the "Magic Question" professional standard for architectural choices.
- **Initial Sync**: Basic Markdown clinical storage and 24h cancellation policy.
