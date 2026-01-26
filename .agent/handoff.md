# Handoff: Next Steps & Priorities (v4.3.0)

Current state: **v4.3.0 Smart Communication**. The bot now proactively manages reminders and archives all interactions (including admin replies) to the medical records.

## 🔴 HIGH PRIORITY (Phase 2: TWA Evolution)

1. **Conditional Cancellation buttons**:
    - Update TWA logic to ONLY show the "Cancel" button if `(session_time - now) > 72h`.
    - Provide a "Contact Vera" link as a fallback.

2. **Health Dashboard Improvements**:
    - Polish the responsive layout for the new medical record groupings (Photos, Scans, etc.).
    - Implement a "Next Appointment" countdown on the home screen.

## 🟡 MEDIUM PRIORITY (Phase 3: Administrative Suite)

1. **Internal Admin Dashboard**:
   - Create a Telegram-native view for Vera to see "Today's Agenda" and confirmation statuses in a single message/keyboard.

2. **Billing & Logs**:
    - Enhance the `LogEvent` mechanism to allow Vera to manually append clinical notes or billable amounts directly via the bot.

## 🟢 LOW PRIORITY (Technical Debt)

1. **Robust Free/Busy Query**:
   - Integrate actual Google Calendar Free/Busy API for bulletproof schedule management.

2. **Automated Backups**:
    - Integrate the `CreateBackup` repository method with the bot's `/backup` command and S3/Telegram storage.

---
*Current Gold Standard: `ddbe67c` (v4.3.0)*
