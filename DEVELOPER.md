# 🛠 Vera Massage Bot - Developer Guide

Technical documentation for maintainers and developers.

## 🏗 Architecture
- **Language**: Go 1.22+
- **Framework**: `gopkg.in/telebot.v3` for Telegram API interaction.
- **Calendar**: Google Calendar API v3 for appointment scheduling.
- **Storage**: Filesystem-based JSON database (NoSQL-lite).
    - `data/patients/{id}/patient.json`: User profile.
    - `data/patients/{id}/documents/`: Medical files.
    - `data/blacklist.txt`: Banned user list.

## 🔧 Prerequisites
1. **Go**: Install Go (1.22 or newer).
2. **Google Cloud Project**:
    - Enable Google Calendar API.
    - Download `credentials.json` (OAuth2 Client ID).
3. **Telegram Bot**:
    - Create a bot via @BotFather.
    -Get the `TG_BOT_TOKEN`.

## ⚙️ Configuration
The application uses environment variables or a `.env` file for configuration.

| Variable | Description | Example |
| :--- | :--- | :--- |
| `TG_BOT_TOKEN` | Telegram Bot Token | `12345:ABCdef...` |
| `TG_ADMIN_ID` | Primary Admin ID | `304528450` |
| `ALLOWED_TELEGRAM_IDS` | Comma-separated Admin IDs | `304528450,5331880756` |
| `GOOGLE_CALENDAR_ID` | Calendar ID to write to | `primary` |
| `DATA_DIR` | Custom data directory | `/app/data` |

## 🚀 Running Locally

1. **Clone the repo**:
   ```bash
   git clone <repo-url>
   cd massage-bot
   ```

2. **Setup Env**:
   Create a `.env` file in the root directory with the variables above.

3. **Run**:
   ```bash
   make run
   # Or directly: go run ./cmd/bot
   ```

## 🧪 Testing
The project includes a suite of unit and integration tests.

```bash
make test
```

## Deployment (Docker Compose)

The project includes a `docker-compose.yml` file for easy deployment.

### 1. Prerequisites
- Docker & Docker Compose installed.
- `.env` file populated with production credentials.
- `google_credentials.json` (if using file-based auth) or JSON content in `.env`.

### 2. Build & Run
```bash
# Build the image (if not pulling from registry)
docker build -t registry.gitlab.com/kfilin/massage-bot:latest .

# Start the service
docker-compose up -d
```

### 3. Verify
Check the logs to ensure the bot started and connected to Google Calendar:
```bash
docker-compose logs -f vera-bot
```

### 4. Updating
```bash
docker-compose pull
docker-compose up -d --force-recreate
```

> [!IMPORTANT]
> The `docker-compose.yml` mounts `./data` and `./logs` to ensure patient records and logs persist across restarts.

## 📁 Project Structure
- `cmd/bot/`: Entry point (`main.go`).
- `internal/`: Core application code.
    - `delivery/telegram/`: Bot logic and handlers.
    - `services/appointment/`: Business logic (scheduling, validation).
    - `adapters/googlecalendar/`: Google API integration.
    - `storage/`: File-based persistence.
    - `domain/`: Data models and interfaces.

## 🔐 Admin Commands

| Command | Description |
| :--- | :--- |
| `/backup` | Download complete patient data as ZIP |
| `/block` | Block time slots for personal matters (gym, lunch, etc.) |
| `/ban {id}` | Shadow ban a user |
| `/unban {id}` | Remove user from blacklist |
| `/status` | Check bot health and metrics |

### Using `/block`
1. Send `/block` to the bot
2. Select duration (30min, 1h, 1.5h, 2h, or all day)
3. Pick date from calendar
4. Select time slot
5. Confirm - creates "⛔ Blocked" event in Google Calendar
