# Project Backlog

## 📋 Observation & Improvement Ideas

### 1. "Last Visit" Logic Refinement

- **Issue**: Currently, the bot might consider the most recently *created* appointment as the "Last Visit", or it might just be the last one in a list. When a patient books multiple slots (e.g., Jan 24, then Jan 30, then Jan 26), the "Last Visit" might show Jan 26 if it was the last one processed, which is chronologically incorrect for a "Last" (most future) visit indicator.
- **Goal**: Ensure the "Last Visit" display in the bot and Medical Card reflects the chronologically latest scheduled appointment in the future (or the most recent past one if no future ones exist).
- **Context**: Patients find it confusing if the sequence doesn't follow the calendar.

---
*Created: 2026-01-21 16:55*
