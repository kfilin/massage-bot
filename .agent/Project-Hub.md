# 💆 Massage Bot: Project Hub

## 📍 Project Vision

A Telegram Bot for booking massage services, integrated with Google Calendar, PostgreSQL, and a TWA Medical Card dashboard.

---

## 🏗️ Technical Foundation

- **Language**: Go 1.23
- **Database**: PostgreSQL (Session Storage & Patient Records)
- **External Integrations**: Google Calendar API, Groq (Transcription).
- **Infrastructure**: Docker Compose on Home Server, GitHub -> GitLab Mirroring.

- **Caddy**: Connected via `caddy-test-net` bridge.

## 🏗️ Development Strategy: Sandbox vs. Home Server

- **Status**: The USER prefers to minimize Sandbox usage.
- **Home Server (Primary)**: The target for all verified changes. Deployment is triggered via Git push (GitLab CI).
- **Local Sandbox (Fallback)**: Used ONLY for:
  - Rapid CSS/UI iterations on the Medical Card.
  - Testing destructive logic (database migrations, mass record deletes).
  - Bypassing Home Server networking/SSL conflicts during development.
- **Rule**: If a feature can be tested safely on the Home Server, do so. Do not default to the sandbox unless necessary.

## 🔄 Git Workflow

- **Master Branch**: Primary source of truth.
- **Remotes**: `github` (dev), `gitlab` (CI/CD / Registry).
- **Rule**: Always push to both remotes to keep the pipeline alive.

## 🚧 Current Development Status

- **Last Stable Build**: `6d126dd` (v3.1.3).
- **Major Blocker**: SSL/Port conflict on local server (Home Server Caddy vs Docker Caddy) - *Resolved by using Home Server's internal host ports.*
- **Next Feature**: Verification of Visit Syncing and UI reordering.

---

## 📂 Organizational Structure

- `.agent/Collaboration-Blueprint.md`: The "Operating System" for how we work.
- `.agent/last_session.md`: The continuity bridge for AI agents.
- `scripts/deploy_home_server.sh`: The source of truth for deployments.
- `docs/session_3a.md`: Detailed history of the Medical Card refactor.

---

## 🔧 Maintenance: Google OAuth Token Renewal

- **Next Renewal Due**: ~2026-07-09 (Check logs for `invalid_grant`)
- **Procedure**:
  1. Generate URL: `https://accounts.google.com/o/oauth2/auth?client_id=451987724111-3p6vs3dvk96sh42gaeuplcp26t9t3998.apps.googleusercontent.com&redirect_uri=http://localhost&response_type=code&scope=https://www.googleapis.com/auth/calendar&access_type=offline&prompt=consent`
  2. Get `code` from redirect URL.
  3. Exchange: `curl -d "client_id=451987724111-3p6vs3dvk96sh42gaeuplcp26t9t3998.apps.googleusercontent.com&client_secret=${CLIENT_SECRET}&code=${CODE}&redirect_uri=http://localhost&grant_type=authorization_code" https://oauth2.googleapis.com/token`
  4. Update `GOOGLE_TOKEN_JSON` in server `.env` and restart.
