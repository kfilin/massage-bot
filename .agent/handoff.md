# Handoff: Next Steps & Priorities

Current state: **v4.2.0 Booking Overhaul**. Stable and production-ready.

## 🔴 HIGH PRIORITY (Technical Debt)

1. **Implement Robust Free/Busy Query**:
   - Transition `GetAvailableTimeSlots` in `internal/services/appointment/service.go` to use the actual Google Calendar Free/Busy API.
   - **Goal**: Prevent overlaps with physical events not created by the bot.

2. **Refine Token Expiry Metric**:
   - Update `client.go` to fix the "0.0 days" refresh token reporting issue in production.

## 🟡 MEDIUM PRIORITY (Analytics)

1. **Grafana Dashboard**:
   - Build a dashboard using the metrics on `:8083`.

## 🟢 LOW PRIORITY (Maintenance)

1. **Automated Backups**:
   - Implement ZIP-based patient data backups.

---
*Current Gold Standard: `4d64549`*
