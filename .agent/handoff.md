# Handoff: Next Steps & Priorities

Current state: **v4.1.1 Clinical Intelligence**. Stable and instrumented.

## 🔴 HIGH PRIORITY (Technical Debt)

1. **Implement Robust Free/Busy Query**:
   - The comment in `internal/adapters/googlecalendar/adapter.go:L134` points to a basic placeholder.
   - **Goal**: Implement a genuine Free/Busy check using the Google Calendar API to prevent even physical overlaps not tracked by the bot.

2. **Refine Token Expiry Metric**:
   - Update `client.go` to better detect `refresh_token_expires_in` in the production environment to fix the "0.0 days" report issue.

## 🟡 MEDIUM PRIORITY (Analytics)

1. **Grafana Dashboard Implementation**:
   - Use the raw metrics exposed on `:8083/metrics` to build a professional dashboard matching the CLI report data.

## 🟢 LOW PRIORITY (Maintenance)

1. **Automated Backups Implementation**:
   - Replace the `CreateBackup` placeholder in `postgres_repository.go` with actual ZIP logic.

---
*Current Gold Standard: `77e9e8d`*
