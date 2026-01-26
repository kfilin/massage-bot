# 💆 Vera Massage Bot (Clinical Edition v4.3.0)

![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)
![Docker](https://img.shields.io/badge/Docker-Enabled-blue?logo=docker)
![License](https://img.shields.io/badge/License-Private-red)

A professional Telegram-based ecosystem for clinical massage practice. Developed for the Vera studio in Fethiye, this bot combines interactive scheduling with a robust medical recording system and seamless cross-device synchronization via **WebDAV**.

---

## 🌟 High-Value Features

### 🩺 Clinical Storage 2.0

The centerpiece of the v4 reconstruction. It moves away from scattered JSONs to a structured **Markdown-mirrored** architecture:

- **Bi-directional Sync**: Edits made in the database reflect in `.md` files in real-time.
- **Obsidian Mobility**: Connect **Obsidian Mobile** (iOS/Android) via WebDAV to manage medical cards directly on your phone.
- **Suffix-based Lookup**: Therapist-friendly folder naming (e.g., `Ivan Ivanov (123456)`) while maintaining strict ID-based tracking.

### 📱 Premium TWA Interface (v4)

A high-end Telegram Web App experience for clinical management:

- **Clinical White Theme**: Modern, minimalist UI optimized for medical professionals.
- **Live GCal Stats**: Visit counts, first/last dates, and upcoming appointments synced directly from Google Calendar APIs.
- **PDF Export**: Generate professional medical reports directly from the TWA dashboard.

### 🔔 Automated Reminders & Smart Reply

- **72h/24h Interactive Reminders**: Ticker-based worker scanning for upcoming visits to request patient confirmation.
- **Loop-Closed Messaging**: Admins can reply to patient inquiries directly via the bot using the `✍️ Ответить` interface.
- **Zero-Manual Logging**: Every patient-therapist interaction is automatically archived in the clinical medical record.
- **72h Cancellation Rule**: Enforced notice period for self-service cancellations to reduce administrative burden.

### 🔒 Enterprise Logic

- **Transcription**: Automated voice-to-text conversion for consultation notes.
- **Shadow Banning**: Polite rejection of unwanted users through the "No slots available" middleware.
- **DB Resilience**: Built-in 5-attempt retry loop for PostgreSQL connectivity.

---

## 🏗 System Architecture

The project follows a clean architecture pattern, prioritizing stability and dependency isolation.

- **Backend**: Go 1.24 (Fiber / Net-HTTP)
- **Database**: PostgreSQL 15+ (Transactional integrity)
- **Sync**: WebDAV Server (CORS/OPTIONS enabled for Mobile clients)
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

*Created by Kirill Filin with Gemini Assistance. Checkpoint: v4.3.0-clinical (2026-01-26).*
