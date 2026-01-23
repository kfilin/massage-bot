# Project Backlog

## 📋 Observation & Improvement Ideas

### 1. "History of Visits" Data Accuracy & Utility

- **Issue**: Cancelled appointments in Google Calendar might still influence the "First/Last Visit" dates or appear in the history table even if the total count decreases correctly.
- **Goal**: Decide if the "History of Visits" block is needed. If it is kept, ensure cancelled/deleted events are filtered out completely from the chronological calculations.
- **Context**: The user noticed that while the visit count badge updates correctly, the history table sometimes retains cancelled future entries.

### 2. Robust TWA Authentication (InitData)

- **Issue**: Opening the Web App from the Telegram "Menu Button" or static links results in a "Missing ID or Token" error because the current logic requires query parameters.
- **Goal**: Implement validation using `window.Telegram.WebApp.initData`. This allows identifying the user securely regardless of how the app was opened.
- **Context**: The user sent a screenshot showing the "Missing id or token" error when presumably opening the app from a non-dynamic source.

### 3. WebDAV & VPN Clinical Privacy

- **Issue**: WebDAV is currently accessible via the public internet (Basic Auth). While functional, it exposes medical records to brute-force attempts.
- **Goal**: Restrict WebDAV access to the **SoftEther SSTP VPN** local network. Update Caddy/Docker configuration to only allow connections from VPN IP ranges.
- **Context**: The user runs a home-based SSTP VPN, which is the "Gold Standard" for securing clinical data at rest.

### 4. Reliable Available Slots Calculation

- **Issue**: The current logic for calculating available slots in `internal/adapters/googlecalendar/adapter.go` is "basic" and uses placeholders for genuinely checking free/busy times.
- **Goal**: Implement a robust Free/Busy query across Google Calendar to account for all types of events, working hours, and multi-service durations.
- **Context**: Comment in `adapter.go:L134` and `L165` highlights this as a "placeholder" that "needs to be refined".

### 5. Patient Discovery & Metadata Extraction

- **Issue**: `customerName` in `eventToAppointment` is often a placeholder or derived from a simple split. `TotalVisits` and `First/Last Visit` sync is complex.
- **Goal**: Refine summary parsing and implement a background sync or more robust caching for patient metadata to ensure data consistency between GCal and Postgres.
- **Context**: `adapter.go:L378` and recent sync logic refinements.

### 6. Automated Clinical Backups

- **Issue**: `CreateBackup` in `PostgresRepository` is currently a simplified placeholder returning an empty string.
- **Goal**: Implement automated daily ZIP backups of the Postgres DB and the `data/patients` Markdown directory, with optional cloud upload (e.g., S3 or Telegram Admin send).
- **Context**: `postgres_repository.go:L430`.

---
*Last updated: 2026-01-24 02:05*
