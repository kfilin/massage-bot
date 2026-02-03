# 💆 Vera Massage Bot (Technical Excellence v5.3.6)

![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)
![Test Coverage](https://img.shields.io/badge/coverage-33.4%25-green)
![Docker](https://img.shields.io/badge/Docker-Enabled-blue?logo=docker)
![License](https://img.shields.io/badge/License-Private-red)

## 🛠️ Refactoring Status (2026-02-03)

**Current Phase**: 3 - Code Quality
**Progress**: 33.4% Code Coverage (Target: 30%+)
**Focus**: Cleanup, Linting, & Documentation
**Documentation**: [docs/Refactoring/README.md](docs/Refactoring/README.md)

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
- **Monitoring**: Prometheus/Grafana stack on port 8083 (with decoupled `MetricsCollector`).
- **Deployment**: Docker Compose with resource guards.

---

## � Project Structure

- `cmd/bot`: Main application entry point.
- `internal/domain`: Core business logic and shared models.
- `internal/services`: Business logic implementation (appointments, etc.).
- `internal/storage`: Database persistence (PostgreSQL).
- `internal/delivery`: Transport layer (Telegram bot handlers).
- `internal/adapters`: External integrations (Google Calendar).
- `internal/monitoring`: Prometheus metrics.
- `internal/logging`: Structured logging.

---

## �🚀 Quick Start (Production)

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

### Development & Testing

Run the full test suite locally:

```bash
# Run all tests
make test
# OR
go test ./...
```

To check for linting errors and code quality:

```bash
golangci-lint run
```

To build the bot binary:

```bash
make build
# OR
go build -o bin/bot ./cmd/bot
```

---

## ⚙️ Configuration

The bot is configured entirely via environment variables.

| Variable | Description | Required | Reference |
| :--- | :--- | :--- | :--- |
| `TG_BOT_TOKEN` | Telegram Bot API Token | Yes | [BotFather](https://t.me/BotFather) |
| `TG_ADMIN_ID` | Telegram ID of the primary admin | Yes | [userinfobot](https://t.me/userinfobot) |
| `ALLOWED_TELEGRAM_IDS` | Comma-separated list of allowed user IDs | Yes | Users allowed to book |
| `GOOGLE_CREDENTIALS_JSON` | Content of Google Service Account JSON | Yes* | *Either this or PATH required |
| `GOOGLE_CREDENTIALS_PATH` | Path to Google Service Account JSON | Yes* | *Either this or JSON required |
| `GOOGLE_CALENDAR_ID` | Calendar ID to manage (usually email) | No | Defaults to `primary` |
| `GROQ_API_KEY` | API Key for Voice Transcription | No | [Groq Console](https://console.groq.com) |
| `TG_THERAPIST_ID` | Telegram ID for "Ask Therapist" feature | No | Defaults to Admin ID |
| `WEBAPP_URL` | Public URL for the Mini App | No | e.g. `https://vera.massage/app` |
| `WEBAPP_SECRET` | Secret key for Web App JWT signature | No | Required if Web App used |
| `WEBAPP_PORT` | Port for the Web App server | No | Defaults to `8080` |
| `WORKDAY_START_HOUR` | Start of working day (0-23) | No | Defaults to 8 |
| `WORKDAY_END_HOUR` | End of working day (0-23) | No | Defaults to 20 |
| `APPT_TIMEZONE` | Timezone for scheduling | No | Defaults to `Europe/Istanbul` |
| `APPT_SLOT_DURATION` | Duration of one slot | No | Defaults to `1h` |
| `APPT_CACHE_TTL` | Cache TTL for free/busy | No | Defaults to `5m` |

---
*Created by Kirill Filin with Gemini Assistance. Gold Standard Checkpoint: v5.3.6 Stable (2026-02-03).*
