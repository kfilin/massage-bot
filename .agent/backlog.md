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

### 18. [DONE] Stat Cards - Information Density

- **Status**: Completed (2026-02-06).
- **Resolution**: Combined First Visit + Total Visits in one card, improved mobile responsiveness with a 2-column grid, and made "Next Appointment" more prominent with a dedicated highlight style.

### 19. [DONE] Accessibility Improvements

- **Status**: Completed in v5.5.1 (2026-02-06).
- **Resolution**:
  - Added `:focus-visible` outline styles for keyboard navigation.
  - Added `aria-expanded` attributes to collapsible sections via JS.
  - Added `role="button"` and `tabindex="0"` to headers.
  - Added `aria-label` to iconic links.

### 20. [DONE] Empty States Enhancement

- **Status**: Completed (2026-02-06).
- **Resolution**: Added icons and descriptive text for empty sections (Notes, History, Docs) to improve UX for new patients.

### 21. [TODO] History List Pagination

- **Priority**: Medium
- If patient has 50+ visits, page gets very long
- **Idea**: "Show more" button or virtual scrolling

### 22. [TODO] Visual Hierarchy Enhancement

- **Priority**: Medium
- Voice Transcripts section could have distinct styling (e.g., chat bubble style)
- **Note**: Generic file icons rejected by user. Only custom/premium visuals to be used.

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

#### Last updated: 2026-02-09

## 🚀 Future Integrations (Added 2026-02-09)

### 27. [TODO] Apple Health Integration

- **Status**: Planned (Backlog)
- **Goal**: Deep sync of patient data to Apple Health.
- **Note**: Requires native iOS app or Shortcuts. TWA can mostly just *mimic the look*.

### 28. [TODO] Apple Wallet Pass

- **Status**: Planned (Backlog)
- **Goal**: Generate `.pkpass` for appointments so patients can add them to Apple Wallet.
