# Checkpoint Summary: 2026-01-20 (Post-Cleanup)

## 🎯 Current Technical State

- **Bot Version**: v3.1.11.
- **Environment**: Local Dev Sandbox established (`docker-compose.dev.yml`, `Caddyfile.dev`).
- **Sandbox Status**: Fully functional with local DB (`db-dev`) and internal SSL (8443).

## ✅ Accomplishments

1. **Aggressive Note Scrubbing (v3.1.11)**: Implemented regex-based scrubbing in TWA to reliably remove legacy automated booking entries (even those with symbols/emojis) from clinical notes.
2. **Dynamic Medical Card Sync**: TWA now fetches live GCal data. If visits are deleted in GCal, they are removed from the card instantly.
3. **Timezone Consistency**: Added `TZ=Europe/Istanbul` to all Docker environments to resolve the 3-hour offset issue.
4. **Project Backlog**: Initialized `backlog.md` to track sorting improvements and other UX feedback.

## 🚧 Current Blockers & Risks

- **GitLab CI Status**: Monitor the pipeline run for v3.1.11. If the server card still shows old logic, verify the Docker image rebuild on the home server.

## 🔜 Next Steps

1. **GCal Sorting**: Finalize chronological sorting of appointments in the "Last Visit" calculation.
2. **Verification**: Confirm TWA "Save PDF" fidelity on various mobile devices.

## 📂 Active logic

- Project OS: `.agent/Collaboration-Blueprint.md`
- Project Variables: `.agent/Project-Hub.md`
- Codebase Map: `.agent/File-Map.md`
