# 💆 Massage Bot: Project Hub

## 📍 Vision
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
- **Metrics**: CLI: `./scripts/report_metrics.sh`, Grafana Dashboard + Prometheus.
- **Backups**: `./scripts/backup_data.sh` (Zips `data/` directory).

---

## ❄️ Session Initialization (Cold Start)

Any agent entering this workspace **MUST** follow **[AGENTS.md](../AGENTS.md)** — which mandates executing `global-skills/startup.md` before responding to any user request.

---

## 🧠 The Antigravity Constitution

### Engineering Rigor
1. **Hypothesis First**: No code changes without Observation -> Hypothesis -> Verification. See `global-skills/quality-gates.md`.
2. **Interactive Verification**: UI/UX changes require a `walkthrough.md` with screenshots/recordings.
3. **Logic Over Compliance**: I am a partner, not a script. I push back on debt. [Details](./rules/logic-over-compliance.md).

### Platform Constraints
4. **Omnichannel Awareness**: Format responses appropriately — Telegram uses HTML, Discord uses Markdown.
5. **Budget Consciousness**: Route cheap tasks to lightweight models. Don't invoke heavy models for trivial parsing. [Details](./rules/budget-consciousness.md).
6. **Approval Gates**: All state-mutating tools (`write_file`, `delete_file`, `run_command`) require explicit user approval.
7. **Native Automation**: Use `/handoff` and `/changelog` for session management.
8. **Anti-Overengineering**: Search for optimal, lean solutions before committing to complexity. [Details](./rules/anti-overengineering.md).
9. **Constraints, Not Checklists**: Meta-rules for system health. [Details](./rules/constraints-not-checklists.md).

---

## 🗃️ Key Files

| File | Purpose |
| :--- | :--- |
| `.agent/HARNESS_GUIDE.md` | **Master manual** for the Collaboration Harness (rules, workflows, hydration) |
| `global-skills/startup.md` | **Mandatory session startup** — architecture + Graphify queries |
| `global-skills/handoff.md` | End-of-session routine — updates startup.md, BACKLOG.md, writes checkpoint |
| `BACKLOG.md` | Pending tasks and priorities |
| `CHANGELOG.md` | Full phase history |
| `DEVELOPER.md` | Technical architecture deep-dive |
| `docker-compose.yml` | Single source of truth for all services |
