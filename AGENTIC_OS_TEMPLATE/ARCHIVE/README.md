# 🗄️ Project Archive

This directory stores historical records of the project's evolution to keep the active `.agent/` directory clean and performant.

## 📂 Structure

- **`CHANGELOGS/`**: Historical versioned changelogs (e.g., `CHANGELOG_2026-02-10.md`).
- **`SESSIONS/`**: Archived `last_session.md` files for full audit trails.
- **`HANDOFFS/`**: Past mission statements and task lists.

## 📝 Retention Policy

- Files are rotated and archived automatically via the `/checkpoint` and `/changelog` workflows.
- Only the **current** `handoff.md`, `last_session.md`, and `CHANGELOG.md` (or the latest date-stamped version) should remain in the root or `.agent/` folder.
