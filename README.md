# 💆 Vera Massage Bot (Technical Excellence v5.0.0)

![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)
![Docker](https://img.shields.io/badge/Docker-Enabled-blue?logo=docker)
![License](https://img.shields.io/badge/License-Private-red)

A professional Telegram-based ecosystem for clinical massage practice. Developed for the Vera studio in Fethiye, this bot combines interactive scheduling with a robust medical recording system and seamless cross-device synchronization via **WebDAV**.

---

## 🌟 High-Value Features

### 📅 Zero-Collision Scheduling (v5.0)

The definitive scheduling engine powered by the official **Google Calendar Free/Busy API**:

- **100% Accuracy**: Respects "Out of Office", manual blocks, and external calendar overlays.
- **Just-in-Time Verification**: Eliminates race conditions by re-verifying availability at the exact moment of confirmation.

### 💾 Automated Backups 2.0 (v5.0)

Disaster recovery that runs itself:

- **Comprehensive Archival**: Daily ZIP backups containing the full PostgreSQL database (`pg_dump`) and patient Markdown files (`data/patients/`).
- **Telegram Delivery**: Encrypted archives are delivered directly to the Admin every 24 hours.
- **Self-Healing Storage**: Local temporary archives are purged after delivery to prevent disk bloat.

### 🩺 Clinical Storage 2.0

A structured **Markdown-mirrored** architecture for patient records:

- **Bi-directional Sync**: Edits made in the database reflect in `.md` files in real-time.
- **Obsidian Mobility**: Connect **Obsidian Mobile** (iOS/Android) via WebDAV to manage medical cards directly on your phone.
- **Suffix-based Lookup**: Therapist-friendly folder naming (e.g., `Ivan Ivanov (123456)`) while maintaining strict ID-based tracking.

### 🔔 Interactive Reminders

- **72h/24h Interactive Flow**: Ticker-based worker requests patient confirmation.
- **Loop-Closed Messaging**: Admins can reply to patient inquiries directly via the bot using the `✍️ Ответить` interface.
- **72h Cancellation Rule**: Enforced notice period for self-service cancellations to reduce administrative burden.

### 🔒 Enterprise Logic

- **Transcription**: Automated voice-to-text conversion for consultation notes.
- **Shadow Banning**: Polite rejection of unwanted users through the "No slots available" middleware.
- **DB Resilience**: Built-in 5-attempt retry loop for PostgreSQL connectivity.

---

## 🏗 System Architecture

The project follows a clean architecture pattern, prioritizing stability and dependency isolation.

- **Backend**: Go 1.24 (Fiber / Telebot v3)
- **Database**: PostgreSQL 15+ (Transactional integrity)
- **Sync**: WebDAV Server (CORS/OPTIONS enabled for Mobile clients)
- **Monitoring**: Prometheus/Grafana stack on port 8083.
- **Deployment**: Docker Compose with resource guards.

---

## 🚀 Quick Start (Production)

### 1. Requirements

- Docker & Docker Compose
- Google Cloud credentials (`credentials.json`)
- Telegram Bot Token (@BotFather)

### 2. Environment Setup

Create a `.env` file from `.env.example`.

### 3. Deploy

```bash
docker-compose up -d --build
```

---
*Created by Kirill Filin with Gemini Assistance. Gold Standard Checkpoint: v5.0.0 Stable (2026-01-26).*
