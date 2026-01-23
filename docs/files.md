# 📂 Project File Structure Explained

This document provides a detailed overview of every file and directory in the **Vera Massage Bot** project.

---

## 🏗️ Core Infrastructure & Docker

| File | Purpose |
| :--- | :--- |
| `docker-compose.yml` | **Master Controller**. The primary orchestration file for production. |
| `Dockerfile` | **Blueprint**. Instructions to build the bot's container image. |
| `deploy/docker-compose.prod.yml` | Production-specific Docker overrides. |
| `deploy/docker-compose.dev.yml` | Development-specific Docker overrides. |
| `deploy/docker-compose.example.yml` | Template for environment-specific orchestration. |

---

## 🛡️ Reverse Proxy (SSL)

| Folder / File | Purpose |
| :--- | :--- |
| `deploy/Caddyfile` | **Proxy Config**. Handles HTTPS and routes traffic to the bot. |
| `deploy/Caddyfile.example` | Template for the server's host-level Caddy configuration. |
| `deploy/Caddyfile.dev` | Simplified proxy configuration for local development. |

---

## 💻 Source Code (Go)

| Folder / File | Purpose |
| :--- | :--- |
| `cmd/` | **Entry Points**. Contains `bot/main.go` and the WebApp server logic. |
| `internal/` | **The Brain**. Contains business logic, database handlers, and services. |
| `go.mod` / `go.sum` | **Dependencies**. Lists all external libraries used by the bot. |
| `Makefile` | **Shortcuts**. Commands like `make build` or `make test`. |

---

## 📜 Documentation

| Folder / File | Purpose |
| :--- | :--- |
| `README.md` | **Home Page**. General overview and quick-start guide. |
| `docs/DEVELOPER.md` | **Technical Guide**. In-depth info for maintainers (Architecture/Security). |
| `docs/USER_GUIDE_RU.md` | **Patient Guide**. Instructions for clients using the bot. |
| `docs/VERA_GUIDE_RU.md` | **Therapist Guide**. Instructions for medical record management. |
| `docs/files.md` | **This File**. The project's navigational map. |

---

## 📦 Data & Storage

| Folder | Purpose |
| :--- | :--- |
| `data/` | **Clinical Records**. Stores patient medical cards as `.md` files. |
| `postgres_data/` | **Database Files**. Persistent storage for the PostgreSQL database. |
| `logs/` | **Audit Trail**. Application logs for debugging and monitoring. |

---

## 🚀 Automation & CI/CD

| File / Folder | Purpose |
| :--- | :--- |
| `scripts/` | Shell scripts for deployment, backups, and token renewal. |
| `deploy/k8s/` | **Kubernetes**. Manifests for K8s-based deployments. |
| `.gitlab-ci.yml` | GitLab pipeline for building images and auto-deploying. |
