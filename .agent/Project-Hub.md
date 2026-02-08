# 💆 Massage Bot: Project Hub

## 📍 Project Vision

A professional clinical ecosystem for massage therapists. Features interactive booking, automated medical records, and cross-device synchronization via Obsidian/WebDAV.

---

## 🏗️ Technical Foundation

- **Version**: v5.6.3 (TWA UI & Voice Fixes)
- **Language**: **Go 1.24** (Alpine-based)
- **Database**: PostgreSQL 15 (Metadata & Sync Status)
- **Clinical Storage**: **Markdown-mirrored Filesystem** (Clinical Storage 2.0)
- **Infrastructure**: Docker Compose on Home Server (Prod: 8082, Test: 9082).
- **Networks**: Shared `caddy-test-net` + Isolated `massage-bot-internal`. **Hardened**: Named bridge for stability.

---

## 🚀 Deployment & SOPs

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

- **Metrics**:
  - CLI: `./scripts/report_metrics.sh`
  - **Visual**: Grafana Dashboard + Prometheus Scrape Config in `deploy/monitoring/`.
- **Backups**: `./scripts/backup_data.sh` (Zips `data/` directory).

---

## 💎 Gold Standard Checkpoint

- **Commit**: `f6605dc6d18dcc69676d2326715ae20f42b08637` (v5.6.3 Stable)
- **Date**: 2026-02-08
- **Status**: **STABLE**. Documentation fully synchronized. TWA UI Refinements & Voice Reply Fixes.
- **Rollback**: `280fc14` (v5.6.2)

---

## 🧠 Collaboration Rules (The "Operating System")

1. **Human-Led, AI-Assisted**: You define the goal, I suggest 3 options with trade-offs.
2. **Docs as Fuel**: Every feature = Code + Tests + Rationale.
3. **Smart Logs**: Use `git log` and this Hub to track decisions (ADRs).
4. **Checkpoint**: Follow `.agent/sop/checkpoint.md` to rotate handoffs into `ARCHIVE/` and update this Hub.
