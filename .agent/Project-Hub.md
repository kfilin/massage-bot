# 💆 Massage Bot: Project Hub

## 📍 Project Vision

A professional clinical ecosystem for massage therapists. Features interactive booking, automated medical records, and cross-device synchronization via Obsidian/WebDAV.

---

## 🏗️ Technical Foundation

- **Version**: v4.1.0 (Clinical Edition)
- **Language**: **Go 1.24** (Alpine-based)
- **Database**: PostgreSQL 15 (Metadata & Sync Status)
- **Clinical Storage**: **Markdown-mirrored Filesystem** (Clinical Storage 2.0)
- **Integrations**: Google Calendar API, Groq (Whisper Transcription).
- **Protocols**: **WebDAV** (for Obsidian Mobile sync).
- **Infrastructure**: Docker Compose on Home Server with CPU/RAM guards.

---

## 🏗️ Development Strategy

- **Home Server (Primary)**: The target for all verified changes. Deployment is triggered by pushing to **GitHub**, which mirrors to GitLab. Access via `ssh server`.
- **Clinical Storage 2.0**: Bi-directional sync between DB and `.md` files in `data/patients/`.
- **Sync Rule**: ID suffix tracking `(TelegramID)` allows therapist to rename patient folders in Obsidian without breaking the bot.

---

## 🔄 Git Workflow

- **Master Branch**: Primary source of truth.
- **Rule**: All restoration work from the "PDF experiment" has been consolidated into the `master` branch on the stable `v3.15` backbone.

---

## 🚀 Deployment & CI/CD Workflow

This project uses a dual-remote setup with automated mirroring to maintain sync and trigger builds.

1. **Push to GitHub (`origin`)**: This is the primary entry point for all changes.
2. **Automated Mirror**: GitHub Actions (see `.github/workflows/mirror.yml`) automatically force-pushes the `master` branch to GitLab.
3. **GitLab Pipeline**: GitLab receives the push and triggers the CI/CD pipeline (`.gitlab-ci.yml`), which:
    - Runs tests (Go 1.24).
    - Builds the Docker image and pushes it to the GitLab Registry.
    - Triggers the `deploy_home_server.sh` script on the target server.

> [!IMPORTANT]
> Always push to **GitHub** first to ensure both repositories are in sync. Pushing directly to GitLab should only be done for debugging the pipeline itself, as it leaves GitHub outdated.

---

---

## 🚧 Current Development Status

- **Status**: **Stable & Production Ready**.
- **Core Features**:
  - [x] WebDAV / Obsidian Mobile Sync.
  - [x] Premium "PDF-free" TWA Dashboard.
  - [x] Automated 2h Visit Reminders.
  - [x] Hierarchical Storage (scans/images/messages).
  - [x] DB Resilience (5x retry loop).

---

## 📂 Organizational Structure

- `.agent/Collaboration-Blueprint.md`: The "Operating System" for how we work.
- `.agent/last_session.md`: The continuity bridge for AI agents.
- `.agent/handoff.md`: Direct instructions for the next session.
- `scripts/deploy_home_server.sh`: The source of truth for deployments.

---

## 💎 Gold Standard Checkpoint

- **Commit**: `f7b055643f4ffda7fe1e7b5add64be28ab9dfa82`
- **Date**: 2026-01-23
- **Status**: **The Gold Standard**. WebDAV fixed, documentation updated, mirroring refined, and CI/CD verified.
- **Rollback Command**: `git reset --hard f7b0556`

---

## 🔧 Maintenance: Google OAuth Token Renewal

- **Next Renewal Due**: ~2026-07-09 (Check logs for `invalid_grant`)
- **Procedure**: See detailed logs in `scripts/renew_token.sh`.
