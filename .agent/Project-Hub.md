# 💆 Massage Bot: Project Hub

## 📍 Project Vision

A professional clinical ecosystem for massage therapists. Features interactive booking, automated medical records, and cross-device synchronization via Obsidian/WebDAV.

---

## 🏗️ Technical Foundation

- **Version**: v4.3.0 (Smart Reminders & Loop Closure)
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

1. **Push to GitHub (`origin`)**: **Manual step**. Primary source for code and metrics.
2. **Push to GitLab (`gitlab`)**: **Manual step**. Pushing to GitLab triggers the CI/CD pipeline (`.gitlab-ci.yml`) for production deployment.
3. **GitLab Pipeline**: Builds the Docker image, pushes to registry, and deploys to the home server.
    - Runs tests (Go 1.24).
    - Builds the Docker image and pushes it to the GitLab Registry.
    - Triggers the `deploy_home_server.sh` script on the target server.

> [!NOTE]
> Both remotes must be updated manually. Ensure changes are pushed to both **GitHub** (for tracking) and **GitLab** (to trigger the production pipeline).

---

---

## 🚧 Current Development Status

- **Status**: **Stable & Production Ready**.
- **Core Features**:
  - [x] WebDAV / Obsidian Mobile Sync.
  - [x] Premium TWA Dashboard.
  - [x] Automated Interactive Reminders (72h/24h).
  - [x] Smart Forwarding with Loop Closure (Admin Reply).
  - [x] Hierarchical Storage (scans/images/messages).
  - [x] DB Resilience (Extended timeout strategy).

---

## 📂 Organizational Structure

- `.agent/Collaboration-Blueprint.md`: The "Operating System" for how we work.
- `.agent/last_session.md`: The continuity bridge for AI agents.
- `.agent/handoff.md`: Direct instructions for the next session.
- `scripts/deploy_home_server.sh`: The source of truth for deployments.

---

## 💎 Gold Standard Checkpoint

- **Commit**: `ddbe67c`
- **Date**: 2026-01-26
- **Status**: **v4.3.0 Stable**. Implemented Smart Reminders and Closed-Loop communication.
- **Rollback Command**: `git reset --hard ddbe67c`

---

## 🔧 Maintenance: Google OAuth Token Renewal

- **Next Renewal Due**: ~2026-07-09 (Check logs for `invalid_grant`)
- **Procedure**: See detailed logs in `scripts/renew_token.sh`.
