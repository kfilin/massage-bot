# 🛠 Vera Massage Bot - Developer Guide (v4.1.0)

Technical documentation for maintainers and developers. This project has been rebuilt on the stable **v3.x backbone** to eliminate PDF-complexity while retaining advanced clinical features.

## 🏗 Architecture Overview

- **Language**: **Go 1.24** (Alpine-based build)
- **Framework**: `telebot v3`
- **Primary Data**: **PostgreSQL 15** for transactional metadata, stats, and auth.
- **Clinical Files**: **Markdown (.md)** mirrored filesystem storage for Obsidian compatibility.
- **Interfaces**:
  - **Telegram Bot**: Main interaction layer.
  - **TWA (Telegram Web App)**: Premium Clinical UI (Auth via HMAC-SHA256).
  - **WebDAV**: Clinical data sync server (CORS-enabled).

## 📁 Storage Structure (Clinical Storage 2.0)

The bot manages data in `/app/data/patients/` using a mirrored approach:

```text
data/patients/
└── Name (TelegramID)/          # Flexible folder name (suffix-tracked)
    ├── TelegramID.md           # Mirrored Medical Card (Markdown)
    ├── scans/                  # Categorized clinical documents
    │   └── DD.MM.YY/*.pdf
    ├── images/                 # MRI/X-Ray photos
    └── messages/               # Voice recordings (.ogg)
```

## ⚙️ Configuration (Environment Variables)

| Variable | Description |
| :--- | :--- |
| `DB_URL` | PostgreSQL connection string |
| `WEBAPP_SECRET` | Used for both TWA HMAC validation and WebDAV password |
| `TZ` | System timezone (Default: `Europe/Istanbul`) |
| `DATA_DIR` | Directory for clinical Markdown files |
| `GOOGLE_CALENDAR_ID` | Targeted GCal ID |

## 🚀 Development Workflow

### 1. External Dependencies

- **PostgreSQL**: Required for metadata.
- **Google Cloud Console**: Enable 'Google Calendar API', configure OAuth2, and place `credentials.json` in root.

### 2. Local Setup

```bash
# Install dependencies
go mod download

# Build check
docker-compose build massage-bot
```

### 3. Database Resilience

The system includes a **DB Retry Loop**. When the backend starts, it will attempt to connect to PostgreSQL 5 times with exponential backoff. This is crucial for home-server deployments where the DB container might start slower than the app.

## 💉 WebDAV & TWA Integration

- **WebDAV**: Mounted at `/webdav/`. Uses **Basic Auth** (Username: TelegramID, Password: `WEBAPP_SECRET` based HMAC).
- **TWA**: Hosted on port `8082`. Uses **HMAC-SHA256** validation of `initData` provided by Telegram.

## 📦 Deployment (Docker)

The production image is a multi-stage `Dockerfile` (Builder -> Runtime) resulting in a minimal footprint (<50MB).

```bash
# Deploy to home server
./scripts/deploy_home_server.sh
```

### Resource Guards

`docker-compose.yml` includes specific limits to prevent host thrashing:

- `cpus: 0.5`
- `memory: 256M`

---
*Created by AntiGravity AI. Build Version: 4.1.0-clinical.*
