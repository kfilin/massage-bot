# 📂 Project File Structure Explained

This document provides a detailed overview of every file and directory in the **Vera Massage Bot** project.

---

## 🤖 Agent Context & AI Collaboration (`.agent/`)

| File | Purpose |
| :--- | :--- |
| `.agent/Project-Hub.md` | **The Source of Truth**. High-level project vision, tech stack, and milestone tracking. |
| `.agent/Collaboration-Blueprint.md` | **The Operating System**. Defines how AI and Humans work together (The "Gold Standard"). |
| `.agent/last_session.md` | **Continuity Bridge**. Technical summary of the previous development session. |
| `.agent/handoff.md` | **Instruction Set**. Specific goals and high-priority tasks for the current session. |
| `.agent/backlog.md` | **Future Roadmap**. Ideas, refinements, and technical debt to be addressed. |
| `.agent/Scripts-Inventory.md` | **Toolbox**. Detailed documentation of all scripts in `scripts/`. |
| `.agent/workflows/` | **Standard Ops**. Reusable workflows (e.g., `/checkpoint`). |

---

## 🏗️ Core Infrastructure & Docker

| File | Purpose |
| :--- | :--- |
| `docker-compose.yml` | **Master Controller**. The primary orchestration file for production. |
| `Dockerfile` | **Blueprint**. Instructions to build the bot's container image. |
| `deploy/docker-compose.prod.yml` | Production-specific Docker overrides. |
| `deploy/docker-compose.test-override.yml` | **Test Env**. Overrides for isolated test environment (Ports/Net). |
| `deploy/docker-compose.example.yml` | Template for environment-specific orchestration. |

---

## 🛡️ Reverse Proxy & Deployments

| Folder / File | Purpose |
| :--- | :--- |
| `deploy/Caddyfile` | **Proxy Config**. Handles HTTPS and routes traffic to the bot. |
| `deploy/k8s/` | **Kubernetes**. Production manifests (Deployment, Secrets, ConfigMap). |
| `deploy/Caddyfile.example` | Template for host-level Caddy configuration. |

---

## 💻 Source Code (Go)

| Folder / File | Purpose |
| :--- | :--- |
| `cmd/bot/main.go` | **Entry Point**. Initializes the bot and health servers. |
| `internal/` | **The Brain**. Domain logic, database repository, and service layer. |
| `go.mod` / `go.sum` | **Dependencies**. External Go libraries. |
| `Makefile` | **Shortcuts**. Automation for build/test/metrics tasks. |

---

## 📜 Documentation

| Folder / File | Purpose |
| :--- | :--- |
| `README.md` | **Home Page**. Quick-start and general overview. |
| `CHANGELOG.md` | **Version History**. Chronological log of all notable changes. |
| `docs/DEVELOPER.md` | **Technical Guide**. Security and architecture details. |
| `docs/USER_GUIDE.md` | **Patient Guide (EN)**. Instructions for bot users in English. |
| `docs/USER_GUIDE_RU.md` | **Patient Guide (RU)**. Detailed Russian manual (Master version). |
| `docs/VERA_GUIDE_RU.md` | **Therapist Guide**. Record management instructions. |
| `docs/metrics.md` | **Monitoring Reference**. List of instrumented Prometheus metrics. |
| `docs/ProdArchitecture.md` | **Gap Analysis**. Comparison of current Home Lab setup vs. Enterprise Best Practices. |

---

## 📦 Data & Storage

| Folder | Purpose |
| :--- | :--- |
| `data/patients/` | **Clinical Records**. Secure Markdown storage for patient cards. |
| `data/backups/` | **ZIP Archives**. Staging area for automated Telegram backups. |
| `logs/access.log` | **Audit Trail**. Traffic and application logs. |

---

### Last updated: 2026-01-29 (v5.2.0 Dual Environment)
