# 💆 Vera Massage Bot (Clinical Edition v4.1.0)

![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)
![Docker](https://img.shields.io/badge/Docker-Enabled-blue?logo=docker)
![License](https://img.shields.io/badge/License-Private-red)

A professional Telegram-based ecosystem for clinical massage practice. Developed for the Vera studio in Fethiye, this bot combines interactive scheduling with a robust medical recording system and seamless cross-device synchronization via **WebDAV**.

---

## 🌟 High-Value Features

### 🩺 Clinical Storage 2.0

The centerpiece of the v4 reconstruction. It moves away from scattered JSONs to a structured **Markdown-mirrored** architecture:

- **Bi-directional Sync**: Edits made in the database reflect in `.md` files in real-time.
- **Categorized Folder Logic**: Automated organization of patient data into `scans/`, `images/`, and `messages/`.
- **Obsidian Mobility**: Connect **Obsidian Mobile** (iOS/Android) via WebDAV to manage medical cards directly on your phone.
- **Suffix-based Lookup**: Thermapist-friendly folder naming (e.g., `Ivan Ivanov (123456)`) while maintaining strict ID-based tracking.

### 📱 Premium TWA Interface (v4)

A high-end Telegram Web App experience replacing legacy PDF generation:

- **Clinical White Theme**: Modern, minimalist UI optimized for medical professionals.
- **Auth Self-Healing**: Automated JS logic to handle TWA token expiration without user friction.
- **Live GCal Stats**: Visit counts, first/last dates, and upcoming appointments synced directly from Google Calendar APIs.

### 🔔 Automated Reminders & Scheduling

- **2h Notification Worker**: Automatic Telegram reminders sent to patients exactly 2 hours before their visit.
- **72h Cancellation Rule**: Enforced notice period for self-service cancellations to reduce administrative burden.
- **Smart GCal Filtering**: Admin blocks (`⛔ Gym`, `⛔ Lunch`) are automatically filtered out from clinical history.

### 🔒 Enterprise Logic

- **Transcription**: Automated voice-to-text conversion for consultation notes.
- **Shadow Banning**: Polite rejection of unwanted users through the "No slots available" middleware.
- **DB Resilience**: Built-in 5-attempt retry loop for PostgreSQL connectivity with exponential backoff.

---

## 🏗 System Architecture

The project follows a clean architecture pattern, prioritizing stability and dependency isolation.

- **Backend**: Go 1.24 (Fiber / Net-HTTP)
- **Database**: PostgreSQL 15+ (Transactional integrity)
- **Sync**: WebDAV Server (CORS/OPTIONS enabled for Mobile clients)
- **Deployment**: Docker Compose with CPU/Memory reservation guards.

---

## 🚀 Quick Start (Production)

### 1. Requirements

- Docker & Docker Compose
- Google Cloud credentials (`credentials.json`)
- Telegram Bot Token (@BotFather)

### 2. Environment Setup

Create a `.env` file based on the following:

```bash
TG_BOT_TOKEN="your_token"
TG_ADMIN_ID="your_id"
TG_THERAPIST_ID="vera_id"
WEBAPP_SECRET="secure_token" # Used for WebDAV Auth
DATA_DIR="/app/data"
DB_URL="postgres://user:pass@db:5432/massage_bot"
TZ="Europe/Istanbul"
```

### 3. Deploy

```bash
# Clone and build
git clone https://github.com/kfilin/massage-bot.git
cd massage-bot
docker-compose up -d --build
```

---

## 📚 Documentation & Guides

- **[📖 Инструкция для Веры (RU)](VERA_GUIDE_RU.md)** — Подключение Obsidian на iPhone.
- **[📖 User Guide (RU)](USER_GUIDE_RU.md)** — Основные функции для пациентов.
- **[🛠 Developer Guide](DEVELOPER.md)** — Архитектура и обслуживание.

---
*Maintained by Kirill Filin & AntiGravity AI. Based on the v3.15 Stable Backbone.*
