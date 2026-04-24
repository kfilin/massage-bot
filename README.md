# 💆 Vera Massage Bot (Technical Excellence v5.7.0)

![Go Version](https://img.shields.io/badge/Go-1.25.3-00ADD8?style=flat&logo=go)
![Test Coverage](https://img.shields.io/badge/coverage-42.0%25-green)
![Docker](https://img.shields.io/badge/Docker-Enabled-blue?logo=docker)
![License](https://img.shields.io/badge/License-Private-red)

## 🚀 Project Status

**Version**: v5.7.0 (Stable)
**Status**: Active Production
**Latest Feature**: Telegram Web App (TWA) & Voice Intelligence

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

### 📱 Telegram Web App (TWA)

A full-featured Mini App integrated directly into Telegram:

- **For Patients**:
  - **Booking Wizard**: Visual calendar, service selection, and slot picker.
  - **Medical Card**: View visit history, upcoming appointments, and personal data.
  - **Fast Action**: "Book Now" buttons and "Next Appointment" countdowns.
- **For Admins**:
  - **Patient Search**: Live search across the entire database.
  - **Manual Booking**: "Create Appointment" flow to book on behalf of patients.
  - **Full History**: Access to all patient notes, files, and visit logs.

### 🎙️ Voice Intelligence (Groq/Whisper)

- **Transcription**: All voice messages from patients are automatically transcribed using **Groq's Whisper API**.
- **Context**: Transcriptions are saved to the patient's medical card (Postgres + Markdown) and forwarded to the therapist.
- **Filtering**: Intelligent filtering removes "hallucinations" (e.g., "Silence", "Thank you") from empty voice notes.

### 🛡️ Security & Hardening (v5.7.0)

- **Gitleaks Protection**: Integrated pre-commit hooks to prevent accidental leakage of API keys or bot tokens.
- **Environment Isolation**: Strict separation between `.env`, `.env.test`, and production secrets.
- **PII Shielding**: Medical records are stored in a dedicated `data/` volume with restricted filesystem access.

### 🛠️ System Resilience

- **Health Monitoring**: Dedicated `/health` endpoint and Prometheus metrics for real-time stability tracking.
- **Auto-Recovery**: Built-in 5-attempt retry loop for PostgreSQL connectivity and self-healing TWA authentication.
- **Graceful Shutdown**: Orchestrated termination of goroutines to ensure data integrity during updates.

---

## 🏗 System Architecture

The project follows a **Hexagonal / Clean Architecture** pattern, prioritizing stability and dependency isolation.

```mermaid
graph TD
    User((Patient/Admin)) <--> TG[Telegram Bot API]
    TG <--> App[Massage Bot Core]
    
    subgraph "Internal Services"
        App <--> Booking[Booking Engine]
        App <--> Trans[Whisper Transcription]
        App <--> Remind[Reminder Service]
    end
    
    subgraph "External Adapters"
        Booking <--> Google[(Google Calendar)]
        Trans <--> Groq[Groq AI]
    end
    
    subgraph "Persistence & Sync"
        App <--> DB[(PostgreSQL 15)]
        App <--> MD[Markdown Records]
        MD <--> WebDAV[WebDAV Server]
        WebDAV <--> Obsidian[[Obsidian Mobile]]
    end
    
    subgraph "Observability"
        App --> Prom[Prometheus]
        Prom --> Graf[Grafana]
    end
```

- **Backend**: Go 1.25.3 (Standard Library HTTP + Telebot v3)
- **Database**: PostgreSQL 15+ (Transactional integrity)
- **Frontend**: Telegram Web App (Vanilla JS + CSS, zero-dependency)
- **Sync**: WebDAV Server for Obsidian mobility.
- **Monitoring**: Prometheus/Grafana stack on port 8083. [View API Docs](docs/API.md).
- **Deployment**: "Twin Strategy" (Staging & Production) using Docker Compose.

---

## 📂 Project Structure

- `cmd/bot`: Application entry point, Health server, and **Web App routing**.
- `internal/domain`: Core entities (Patient, Appointment, Slot).
- `internal/services`: Domain logic (Booking engine, Reminders, Transcription).
- `internal/storage`: Persistence layer (PostgreSQL, Sessions, and **HTML Templates**).
- `internal/delivery/telegram`: Telegram bot handlers and middleware.
- `internal/adapters`: Third-party integrations (Google Calendar, Groq).
- `internal/ports`: Interface definitions for architectural boundaries.
- `internal/config`: Environment-based configuration management.
- `internal/monitoring`: Prometheus metrics and performance collectors.

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
# Production Deployment (Manual)
./scripts/deploy_home_server.sh

# Test Environment Deployment
./scripts/deploy_test_server.sh
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
| `TG_THERAPIST_ID` | Comma-separated list for Therapist notifications | No | Defaults to Admin ID |
| `WEBAPP_URL` | Public URL for the Mini App | No | e.g. `https://vera.massage/app` |
| `WEBAPP_SECRET` | Secret key for Web App JWT signature | No | Required if Web App used |
| `WEBAPP_PORT` | Port for the Web App server | No | Defaults to `8080` |
| `WORKDAY_START_HOUR` | Start of working day (0-23) | No | Defaults to 8 |
| `WORKDAY_END_HOUR` | End of working day (0-23) | No | Defaults to 20 |
| `APPT_TIMEZONE` | Timezone for scheduling | No | Defaults to `Europe/Istanbul` |
| `APPT_SLOT_DURATION` | Duration of one slot | No | Defaults to `1h` |
| `APPT_CACHE_TTL` | Cache TTL for free/busy | No | Defaults to `5m` |

---
*Created by Kirill Filin with Gemini Assistance. Gold Standard Checkpoint: v5.7.0 Stable (2026-04-20).*
