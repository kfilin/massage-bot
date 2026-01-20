# 💆 Massage Bot: Project Hub

## 📍 Project Vision

A Telegram Bot for booking massage services, integrated with Google Calendar, PostgreSQL, and a TWA Medical Card dashboard.

---

## 🏗️ Technical Foundation

- **Language**: Go 1.23
- **Database**: PostgreSQL (Session Storage & Patient Records)
- **External Integrations**: Google Calendar API, Groq (Transcription), PDF Generation.
- **Infrastructure**: Docker Compose on Home Server, GitHub -> GitLab Mirroring.

## 🌍 Environment & Networking (Source of Truth)

- **Home Server IP**: `192.168.1.102`
- **SSH Access**: Port `2222`, User `kirill`.
- **Bot Endpoint**: `vera-bot-local.kfilin.icu`
- **Internal Ports**:
  - `8080`: Main Bot/TWA
  - `8081`: Health Check
  - `8082`: Web App Backend
- **Caddy**: Connected via `caddy-test-net` bridge.

## 🔄 Git Workflow

- **Master Branch**: Primary source of truth.
- **Remotes**: `github` (dev), `gitlab` (CI/CD / Registry).
- **Rule**: Always push to both remotes to keep the pipeline alive.

## 🚧 Current Development Status

- **Last Stable Build**: `6d126dd` (v3.1.3).
- **Major Blocker**: SSL/Port conflict on local server (Home Server Caddy vs Docker Caddy).
- **Next Feature**: Re-integration of polished Medical Card UI and returning patient logic.

---

## 📂 Organizational Structure

- `.agent/Collaboration-Blueprint.md`: The "Operating System" for how we work.
- `.agent/last_session.md`: The continuity bridge for AI agents.
- `scripts/deploy_home_server.sh`: The source of truth for deployments.
- `docs/session_3a.md`: Detailed history of the Medical Card refactor.
