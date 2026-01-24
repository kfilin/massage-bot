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

## [4.1.1] - 2026-01-24

### Added

- Comprehensive Metrics Infrastructure (Prometheus on port 8083).
- Business Intelligence reporter: `scripts/report_metrics.sh`.
- Analytics for Visit Loyalty (New vs Returning) and Booking Lead Time.
- Metrics documentation and team onboarding workflows.

## [4.1.0] - 2026-01-23

### Added

- Bi-directional sync between Postgres and Markdown files (Clinical Storage 2.0).
- ID-suffix folder tracking allowing therapist folder renaming in Obsidian.
- WebDAV protocol support for Obsidian Mobile synchronization.
- Concurrency locking and slot caching for booking stability.
- HTML sanitization for patient record generation.

## [3.1.15] - 2026-01-18

### Added

- Robust name extraction from Google Calendar events.
- Quiet self-healing logic for session management.
- Medical card UI overhaul.

## [3.1.12] - 2026-01-16

### Added

- Automated 2h visit reminders for patients.
- Dynamic medical card sync with Google Calendar.
- Aggressive regex scrubbing for clinical notes to improve readability.

## [3.1.8] - 2025-11-27

### Fixed

- Initialization error handling (replaced panics with educational messages for invalid tokens).
- Google Mirroring stability fixes for CI/CD pipelines.

## [3.15.0] - 2026-01-22

### Fixed

- Stabilized project backbone after experimental PDF generation phase.
- Reverted to `window.print()` for reliable PDF exports across devices.
- Recovered core clinical logic and appointment handling.
