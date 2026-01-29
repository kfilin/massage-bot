# 💆 Massage Bot: Project Hub

## 📍 Project Vision

A professional clinical ecosystem for massage therapists. Features interactive booking, automated medical records, and cross-device synchronization via Obsidian/WebDAV.

---

## 🏗️ Technical Foundation

- **Version**: v5.2.2-stable (Unified Twin Environments)
- **Language**: **Go 1.24** (Alpine-based)
- **Database**: PostgreSQL 15 (Metadata & Sync Status)
- **Clinical Storage**: **Markdown-mirrored Filesystem** (Clinical Storage 2.0)
- **Integrations**: Google Calendar API (Free/Busy v3), Groq (Whisper Transcription).
- **Protocols**: **WebDAV** (for Obsidian Mobile sync).
- **Infrastructure**: Docker Compose on Home Server with CPU/RAM guards.
- **Environments**:
  - **Prod**: `/opt/vera-bot` (Port 8082).
  - **Test**: `/opt/vera-bot-test` (Port 9082) - Fully Isolated.

---

## 🏗️ Development Strategy

- **Home Server (Primary)**: The target for all verified changes. Deployment is triggered by pushing to **GitHub**, which mirrors to GitLab. Access via `ssh server`.
- **Clinical Storage 2.0**: Bi-directional sync between DB and `.md` files in `data/patients/`.
- **Sync Rule**: ID suffix tracking `(TelegramID)` allows therapist to rename patient folders in Obsidian without breaking the bot.

---

## 🔄 Git Workflow

- **Master Branch**: Primary source of truth.
- **Rule**: All restoration work from the "PDF experiment" has been consolidated into the `master` branch on the stable `v5.0.0` backbone.

---

## 🚀 Deployment & CI/CD Workflow

This project uses a dual-remote setup with automated mirroring to maintain sync and trigger builds.

1. **Push to GitHub (`origin`)**: **Manual step**. Primary source for code and metrics.
2. **Push to GitLab (`gitlab`)**: **Manual step**. Pushing to GitLab triggers the CI/CD pipeline (`.gitlab-ci.yml`) for production deployment.
3. **GitLab Pipeline**: Builds the Docker image, pushes to registry, and deploys to the home server.
    - Runs tests (Go 1.24).
    - Builds the Docker image and pushes it to the GitLab Registry.
    - Triggers the `.gitlab-ci.yml` pipeline.
    - **Auto-Deploys to Test** (`vera-bot-test`).
    - **Waits for Manual Click** to deploy to Production.

> [!NOTE]
> Both remotes must be updated manually. Ensure changes are pushed to both **GitHub** (for tracking) and **GitLab** (to trigger the production pipeline).

---

---

## 🚧 Current Development Status

- **Status**: **Stable & Production Ready**.
- **Core Features**:
  - [x] WebDAV / Obsidian Mobile Sync.
  - [x] Premium TWA Dashboard (**Lightning Fast DB Cache**).
  - [x] Automated Interactive Reminders (72h/24h).
  - [x] Smart Forwarding with Loop Closure (Admin Reply).
  - [x] Hierarchical Storage (scans/images/messages).
  - [x] DB Resilience (Extended timeout strategy).
  - [x] **Zero-Collision Scheduling** (Free/Busy API).
  - [x] **Backups 2.0** (DB + FS ZIP + 24h Tele-Delivery).
  - [x] **Duplicati Integration** (Incremental encrypted local backups).
  - [x] **Test Environment** (Running on `vera-bot-test.kfilin.icu`).

---

## 📂 Organizational Structure

- `.agent/Collaboration-Blueprint.md`: The "Operating System" for how we work.
- `.agent/last_session.md`: The continuity bridge for AI agents.
- `.agent/handoff.md`: Direct instructions for the next session.
- `scripts/deploy_home_server.sh`: The source of truth for deployments.

---

## 💎 Gold Standard Checkpoint

- **Commit**: `cf8f017` (v5.2.2 Unified)
- **Date**: 2026-01-30
- **Status**: **v5.2.2 UNIFIED**. Architectures are identical.
  - **Single Source of Truth**: Base `docker-compose.yml` defines the network (`caddy-test-net`).
  - **Pipeline Sync**: GitLab Auto-Deploys to Staging; Manual Deploy to Prod.
- **Rollback Command**: `git reset --hard cf8f017`

---

## 🔧 Maintenance: Google OAuth Token Renewal

- **Next Renewal Due**: ~2026-07-09 (Check logs for `invalid_grant`)
- **Procedure**: See detailed logs in `scripts/renew_token.sh`.
