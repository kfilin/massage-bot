# 💆 Massage Bot: Project Hub

## 📍 Project Vision

A professional clinical ecosystem for massage therapists. Features interactive booking, automated medical records, and cross-device synchronization via Obsidian/WebDAV.

---

## 🏗️ Technical Foundation

- **Version**: v5.3.5-stable (Clinical & Hardened)
- **Language**: **Go 1.24** (Alpine-based)
- **Database**: PostgreSQL 15 (Metadata & Sync Status)
- **Clinical Storage**: **Markdown-mirrored Filesystem** (Clinical Storage 2.0)
- **Infrastructure**: Docker Compose on Home Server (Prod: 8082, Test: 9082).
- **Networks**: Shared `caddy-test-net` + Isolated `bot-db-net`. **Hardened**: `DB_HOST=massage-bot-db` to prevent DNS ghosting.

---

## 🚀 Deployment & Workflows

All deployment scripts are in the `scripts/` directory.

### 1. Production Deployment (Manual)

Triggered by pushing to `gitlab` (triggers pipeline) or running manually on server:

```bash
./scripts/deploy_home_server.sh
```

### 2. Test Environment (Twin Strategy)

A fully isolated environment running on `vera-bot-test.kfilin.icu`:

```bash
./scripts/deploy_test_server.sh
```

### 3. Local Development (`local-dev`)

- **Mode**: `WEBAPP_DEV_MODE=true` in `.env` allows login without Telegram HMAC.
- **Run**: `docker compose up -d` (requires `docker network create caddy-test-net`).

### 4. Backups & Metrics

- **Metrics**: `./scripts/report_metrics.sh` (CLI Dashboard).
- **Backups**: `./scripts/backup_data.sh` (Zips `data/` directory).

---

## 💎 Gold Standard Checkpoint

- **Commit**: `59c4f69` (Manual Appointment Visibility & Master View)
- **Date**: 2026-01-31
- **Status**: **STABLE**. Manual appointments tracked uniquely; Admin master view enabled.
- **Rollback**: `7a00cd9` (Initial Manual Flow)

---

## 🧠 Collaboration Rules (The "Operating System")

1. **Human-Led, AI-Assisted**: You define the goal, I suggest 3 options with trade-offs.
2. **Docs as Fuel**: Every feature = Code + Tests + Rationale.
3. **Smart Logs**: Use `git log` and this Hub to track decisions (ADRs).
4. **Checkpoint**: Use `/checkpoint` to flush context and update this Hub.
