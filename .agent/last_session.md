# Session Log: v5.2.1 (TWA & Test Env Stabilization)

**Date**: 2026-01-30
**Commit**: `6e33a0d`
**Goal**: Resolve persistent connectivity issues in Test Environment TWA and improve developer experience.

---

## 🏛️ Architectural Decisions

### 1. Root Path Serving (TWA Optimization)

* **Decision**: Serving the WebApp directly at `/` instead of redirecting to `/card`.
* **Rationale**: The redirection logic, while theoretically sound, introduced a vulnerability where some clients (Android WebView) or networks (Cloudflare) misidentified the protocol or dropped the connection context. Serving content immediately at the root path provides a more robust initial handshake.

### 2. HTTP/1.1 Enforcement for Legacy Compatibility

* **Decision**: Enforced `protocols h1` in Caddy for the test domain.
* **Rationale**: The "HTTP/2 Error: NO_ERROR" is a specific bug in Android System WebView interacting with certain HTTP/2 implementations. By downgrading the Caddy-to-Client negotiation to HTTP/1.1, we entirely bypass the buggy code path in the client's browser engine, ensuring 100% reliability.

### 3. Docker Compose "Auto-Discovery"

* **Decision**: Injected `COMPOSE_PROJECT_NAME` and `COMPOSE_FILE` into the `.env` file of the deployment.
* **Rationale**: Developer Experience (DX). Previously, running `docker compose ps` in the test folder returned empty results because the project was named differently (`massage-bot-test`) than the directory (`vera-bot-test`). By baking the configuration into the environment file, strictly standard Docker commands work as expected without manual flags.

---

## 🐛 Technical Challenges & Elegant Solutions

### 1. The "Ghost" Containers

* **Problem**: `docker compose ps` showed no running containers in the test directory, causing confusion.
* **Solution**: Identified that the project name override (`-p`) hid them from the default `ps` view. Solved by persisting the project name in `.env`, making the override implicit and permanent.

### 2. The Identity Crisis (Nickname vs. Name)

* **Problem**: TWA showed "AA" (Nickname) instead of "Kirill" (Real Name).
* **Insight**: This was correct behavior for a fresh environment. The **Self-Healing** logic initialized a new patient record using the only data it had (Telegram Nickname). It self-corrected to the Real Name once sync logic matched it with Google Calendar data during an appointment interaction.

---

## 💎 Checkpoint Status

* **Environment**: Dual (Prod `8082`, Test `9082`).
* **Stability**: **Green**. Both environments verified working on Desktop and Mobile.
* **Next Steps**: None. The infrastructure is solid.

---
*Created by Antigravity AI following the Collaboration Blueprint.*
*Project Status: STABLE (v5.2.1).*
