# 💆 Massage Bot: Project Hub

## 📍 Project Vision

A professional clinical ecosystem for massage therapists. Features interactive booking, automated medical records, and cross-device synchronization via Obsidian/WebDAV.

---

## 🏗️ Technical Foundation

- **Version**: v5.7.0 (Native Agentic OS)
- **Language**: **Go 1.24** (Alpine-based)
- **Database**: PostgreSQL 15 (Metadata & Sync Status)
- **Clinical Storage**: **Markdown-mirrored Filesystem** (Clinical Storage 2.0)
- **Infrastructure**: Docker Compose on Home Server (Prod: 8082, Test: 9082).
- **Networks**: Shared `caddy-test-net` + Isolated `massage-bot-internal`.

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

- **Commit**: `[PENDING]` (v5.7.0 Native OS)
- **Date**: 2026-02-10
- **Status**: **STABLE**. Native Antigravity OS implemented.
- **Rollback**: `661a641` (v5.6.5)

---

## 🧠 Collaboration Rules (The "Operating System")

1. **Logic Over Compliance**: I am a partner, not a script. I push back on debt.
2. **Hypothesis First**: No code changes without Observation -> Hypothesis -> Verification.
3. **Native Automation**: Use `/checkpoint` and `/changelog` for session management.
4. **Knowledge Powered**: Use KIs and Skills to load specific domain expertise.
