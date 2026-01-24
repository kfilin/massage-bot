# Checkpoint- **Commit**: `0aa07d6`

- **Date**: 2026-01-24
- **Status**: **v4.2.1 Clinical Edition Stable**. Implemented visit history UI, fixed 24h sync bug, added appointment status filtering, and performed a direct, permanent scrub of legacy boilerplate from all patient records.
- **Rollback Command**: `git reset --hard 0aa07d6`
- **UI/UX**: "История посещений" section added to TWA; visit stats fixed (now using full history).
- **Quality**: Cancelled events and Admin Blocks filtered from statistics.

## ✅ Accomplishments

1. **Visit History & Sync Refinement**:
    - **Full History Sync**: Fixed a bug where TWA stats were limited to the last 24 hours of data.
    - **Status Tracking**: Added `Status` field to `domain.Appointment` and GCal adapter.
    - **Smart Filtering**: Excluded `cancelled` events and manual "Admin Blocks" from clinical counts and history.
    - **History UI**: Added a clean "History of Visits" list to the Patient Card TWA (last 5 visits).
2. **Template Cleanup**:
    - **Redundancy Removal**: Removed the empty "Ссылки на документы" section from Markdown cards.

## 💡 Learnings

- **GCal Status Nuance**: Google Calendar events remain in the `List` results even after deletion if `ShowDeleted` isn't managed carefully, but checking the `Status == "cancelled"` property is the most robust way to ensure clinical accuracy.
- **Sync Depth**: Relying on a short time window (e.g., 24h) for patient statistics is fragile in a medical context where patients may have long gaps between visits. Always default to full-history scans for "Gold Standard" stats unless performance becomes a primary constraint.
- **Remote Names**: Stick to the explicitly defined remote names (`github`, `gitlab`) to ensure alignment with existing automated scripts (like `deploy_home_server.sh`).

## 🚧 Current Blockers & Risks

- **Free/Busy Logic**: Still using basic overlap checks; full Google Calendar Free/Busy integration is the next priority.

## 🔜 Next Steps

1. **Free/Busy Query**: Implement genuine free/busy logic for robust schedule management.
2. **Backlog Prioritization**: Review and prune `backlog.md` for the next technical sprint.

---
*Created by Antigravity AI.*
