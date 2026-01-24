# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [4.2.0] - 2026-01-24

### Added

- "Back" button navigation for booking flow (Date → Service, Time → Date).
- 72h cancellation warning in bot confirmation and Patient's Card.
- Summarized document grouping in Patient's Card (Scans, Photos, Videos, Voice Messages, Texts, Others).
- Professional "Conventional Commits" standard and squashed history.

### Changed

- Simplified booking slots to hourly intervals (09:00 - 18:00) to ensure therapist breaks.
- Localized TWA badge to "КАРТА ПАЦИЕНТА".
- Optimized TWA layout for mobile devices (responsive stacking of stat boxes).
- Refined Markdown rendering for clinical notes (fixed headers/bold text).
- Verified and enforced 50MB file upload limits across all interfaces.

## [4.1.0] - 2026-01-23

### Added

- **Clinical Storage 2.0**: Permanent switch back to Markdown-mirrored filesystem for Obsidian/WebDAV sync.
- Bi-directional sync between Postgres and Markdown files.
- ID-suffix folder tracking allowing therapist folder renaming in Obsidian.
- Infrastructure: Moved metrics to port 8083; established Prometheus/Grafana baseline.
- Concurrency locking and slot caching for booking stability.

## [4.0.0] - 2026-01-20

### Changed

- **The PDF Pivot**: Decommissioned StirlingPDF and removed PDF generation buttons. Reverted to browser-native `window.print()` for stability and lower overhead.
- **The Postgres Return**: Re-implemented PostgreSQL as the primary metadata store. The scale of the project (patient records, analytics, audit logs) made a relational DB essential for long-term reliability.

## [3.1.15] - 2026-01-18

### Added

- Robust name extraction from Google Calendar events.
- Quiet self-healing logic for session management and medical card auto-auth.

## [2.0.0] - 2025-11-20

### Changed

- **Experiment Phase**: Attempted transition to standalone file-based storage and PDF-only exports.
- Temporary removal of PostgreSQL in favor of pure FS-based state (later reverted in v4.0.0).

## [1.0.0] - 2025-10-15

### Added

- Initial Bot Foundation: Persistent menus and Google Meet integration.
- Initial PostgreSQL integration for analytics and sessions.
- Clinical records stored as semi-structured `.md` files.
- 24h smart cancellation policy implementation.
