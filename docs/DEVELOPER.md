# 🛠 Vera Massage Bot - Developer Guide (v5.0.0)

Technical documentation for maintainers and developers. The project is currently in its **v5.0.0 (Technical Excellence)** state, featuring zero-collision scheduling and automated disaster recovery.

## 🏗 Architecture Overview

- **Language**: **Go 1.24** (Alpine-based build)
- **Framework**: `telebot v3`
- **Primary Data**: **PostgreSQL 15** for transactional metadata, stats, and auth.
- **Clinical Files**: **Markdown (.md)** mirrored filesystem storage for Obsidian compatibility.
- **Scheduling**: **Google Calendar Free/Busy API** (v3) for real-time availability.
- **Interfaces**:
  - **Telegram Bot**: Main interaction layer.
  - **TWA (Telegram Web App)**: Premium Clinical UI.
  - **WebDAV**: Clinical data sync server (CORS-enabled).

## 📁 Storage Structure (Clinical Storage 2.0)

The bot manages data in `/app/data/patients/` using a mirrored approach:

```text
data/patients/
└── Name (TelegramID)/          # Flexible folder name (suffix-tracked)
    ├── TelegramID.md           # Mirrored Medical Card (Markdown)
    ├── scans/                  # Categorized clinical documents
    ├── images/                 # MRI/X-Ray photos
    └── messages/               # Voice recordings (.ogg)
```

## ⚙️ Core Technical Services

### 1. Zero-Collision Scheduler

The `AppointmentService` leverages the official Google Free/Busy API. It queries all relevant calendars to find true available blocks.

- **JIT Verification**: A final Free/Busy check is performed *within* the booking confirmation transaction to prevent race conditions.

### 2. Automated Backup Worker (v5.0)

A background worker (24h ticker) that:

- Dumps the PostgreSQL database via `pg_dump`.
- Zips the database dump and the `/app/data/patients/` directory.
- Sends the resulting archive to the Admin's Telegram ID.
- Purges temporary files to maintain disk health.

### 3. Smart Forwarding & Admin Reply

Forwarded patient messages include a signed callback token. Admins can respond directly, and the entire "conversation loop" is automatically archived to the patient's Markdown record.

---

## 🚀 Development Workflow

### 1. External Dependencies

- **PostgreSQL**: Required for metadata.
- **Google Cloud Console**: Enable 'Calendar API', configure OAuth2.
- **Groq Cloud**: For Whisper transcription (`GROQ_API_KEY`).

### 2. Local Setup

```bash
go mod download
docker-compose build
```

---

## 📦 Deployment & Infrastructure

The production environment is managed via Docker Compose on the home server.

| File | Purpose |
| :--- | :--- |
| `docker-compose.yml` | **Master Controller**. |
| `deploy/docker-compose.prod.yml` | Production-optimized overrides. |
| `deploy/k8s/` | Kubernetes manifests. |
| `scripts/deploy_home_server.sh` | Deployment automation. |

---
*Created by Kirill Filin with Gemini Assistance. Build Version: 5.0.0-clinical.*
