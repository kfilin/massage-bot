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

### 5. Reliable Available Slots Calculation

- **Issue**: The current logic for calculating available slots in `internal/adapters/googlecalendar/adapter.go` is "basic" and uses placeholders for genuinely checking free/busy times.
- **Goal**: Implement a robust Free/Busy query across Google Calendar to account for all types of events, working hours, and multi-service durations.
- **Context**: Confirmed placeholder in `adapter.go:L134`: `// This is a placeholder and needs actual implementation`.

### 6. Patient Discovery & Metadata Extraction

- **Issue**: `customerName` in `eventToAppointment` is often a placeholder or derived from a simple heuristic.
- **Goal**: Refine summary parsing and implement a background sync or more robust caching for patient metadata.
- **Context**: Confirmed heuristic in `adapter.go:L381`. While v4.2.1 improved specific sync logic, general inbound metadata extraction remains basic.

### 7. Automated Clinical Backups

- **Issue**: `CreateBackup` in `PostgresRepository` is empty. A shell script `scripts/backup_data.sh` exists for local ZIPs but isn't integrated into the bot or cloud flow.
- **Goal**: Integrate backup logic into the bot (or cron) with off-site redundancy (e.g., S3 or Telegram Admin upload).
- **Context**: `postgres_repository.go:L564` is empty. `scripts/backup_data.sh` handles local `data/` zipping but lacks cloud integration.

### 8. [IN PROGRESS] Phase 2: Personal Cabinet (TWA) Evolution

- **Goal**: Conditional button rendering (Cancel only >72h), improved layout, and live health metrics.
- **Status**: Planning stage.

### 9. [PLANNED] Phase 3: Administrative Suite

- **Goal**: In-bot dashboard for Vera to manage sessions and billing.

### 10. Automated Cancellation for Unconfirmed Bookings

- **Goal**: Automatically cancel an appointment if it remains "Unconfirmed" after a certain period (e.g., at T-48h).
- **Status**: Under evaluation (User debating on timeframes). Added to backlog (2026-01-26).
- **Context**: Helps keep the schedule clean but requires careful timing logic.

---
*Last updated: 2026-01-26 11:25*
