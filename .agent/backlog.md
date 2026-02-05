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

### 5. [DONE] Robust Scheduling (Free/Busy API)

- **Status**: Completed in v5.0.0.
- **Resolution**: Migrated to official Google Calendar Free/Busy API for 100% accurate slot detection.

### 7. [DONE] Automated Backups 2.0

- **Status**: Completed in v5.0.0.
- **Resolution**: Implemented ZIP archival of DB + Files with daily Telegram delivery to Admin.

### 12. [DONE] Local Duplicati Backup Setup

- **Status**: Completed (2026-01-27).
- **Resolution**: User set up and verified Duplicati instance on the home server for incremental, encrypted backups of the `./data` directory.

### 14. [DONE] Admin Patient Name Edit

- **Status**: Completed (2026-01-31).
- **Resolution**: Implemented `/edit_name <id> <new_name>` command for admins.
- **Note**: User noted difficulty in self-testing due to overlapping Admin/Patient roles, but logic is verified for admin IDs.

### 15. [DONE] Manual Appointment Creation

- **Status**: Completed (2026-01-31).
- **Resolution**: Added `/create_appointment` command. Implemented unique ID tracking (`manual_<name>`) and an Admin Master View in "My Appointments" to ensure manual bookings are visible and linked to dedicated patient Med-Cards.
- **Implementation**: [booking.go](file:///home/kirillfilin/Documents/massage-bot/internal/delivery/telegram/handlers/booking.go), [bot.go](file:///home/kirillfilin/Documents/massage-bot/internal/delivery/telegram/bot.go)

### 16. [DONE] DB & Stability Hardening (v5.3.6)

- **Status**: Completed (2026-02-01).
- **Resolution**: Implemented `connect_timeout` for DB, added startup crash-loop delays, and documented external API visibility issues in a stability report.
- **Note**: System currently healthy internally but failing to reach Telegram API.

---

## 🎨 TWA UI/UX Improvements (Added 2026-02-04)

### 17. [DONE] Quick Wins - Dark Mode, Animations, Loading States

- **Status**: Completed in v5.4.0.
- **Resolution**: Added dark mode support, section fade-in animations, button loading states, and visual differentiation for upcoming vs history sections.

### 18. [TODO] Stat Cards - Information Density

- **Priority**: High
- On mobile, stat cards stack vertically taking too much space
- "Обновлено" is less useful than showing "Next Appointment"  
- **Idea**: Combine first visit + total visits in one card

### 19. [TODO] Accessibility Improvements

- **Priority**: High
- Add `aria-labels` on interactive elements
- Add `:focus-visible` styles for keyboard navigation
- Improve contrast ratios for muted text

### 20. [TODO] Empty States Enhancement

- **Priority**: High
- Current "Документов пока нет" is plain text
- **Idea**: Add icon/illustration and call-to-action

### 21. [TODO] History List Pagination

- **Priority**: Medium
- If patient has 50+ visits, page gets very long
- **Idea**: "Show more" button or virtual scrolling

### 22. [TODO] Visual Hierarchy Enhancement

- **Priority**: Medium
- Voice Transcripts section could have distinct styling
- Documents section could show file type icons

### 23. [TODO] Success Feedback

- **Priority**: Medium
- After successful cancellation, show success toast/banner before reload
- Add subtle success animation

### 24. [TODO] Offline Support

- **Priority**: Low
- Add service worker for offline viewing of cached data
- Show "offline" indicator when disconnected

### 25. [TODO] Print Optimization

- **Priority**: Low
- Add `@media print` styles for proper printing
- Hide interactive elements in print view

### 26. [TODO] Performance Optimization

- **Priority**: Low
- Lazy load history items below the fold
- Consider skeleton loading states for slow connections

---
*Last updated: 2026-02-04*
