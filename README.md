# 🏗️ Massage Bot - Telegram Appointment System

A production-ready Telegram bot for massage appointment booking with Google Calendar integration.

## 🚀 Quick Start

```bash
# Clone and build
git clone https://github.com/kfilin/massage-bot
cd massage-bot
go run cmd/bot/main.go

🔐 Environment Configuration
Required Environment Variables
bash

# Telegram Bot Configuration
BOT_TOKEN=your_telegram_bot_token_here
ADMIN_ID=your_telegram_user_id_here

# Health Server Configuration  
HEALTH_PORT=8080  # Port for health check endpoints (default: 8080)

# Google Calendar Configuration (choose one method)

Google Calendar Setup
Method 1: Environment Variables (Recommended for containers)
bash

# Google OAuth Credentials (JSON format)
GOOGLE_CREDENTIALS_JSON='{"web":{"client_id":"...","client_secret":"...","redirect_uris":["http://localhost:8080"]}}'

# Google OAuth Token (after initial authentication)
GOOGLE_TOKEN_JSON='{"access_token":"...","token_type":"Bearer",...}'

# Google Calendar ID
GOOGLE_CALENDAR_ID=your_calendar_id@gmail.com

Method 2: Local Files (for development)
bash

# Place credentials.json in project root
# Place token.json in project root (generated after first OAuth flow)

Getting Google OAuth Credentials:

    Go to Google Cloud Console

    Create a new project or select existing one

    Enable Google Calendar API

    Create OAuth 2.0 credentials (Web application)

    Set authorized redirect URIs to: http://localhost:8080

    Download credentials JSON or copy to environment variable

🏥 Health Endpoints

    GET /health - Application health status

    GET /ready - Readiness for traffic

    GET /live - Liveness probe

    GET / - Service information

The health server port can be configured via HEALTH_PORT environment variable.
bash

curl http://localhost:8080/health

🐳 Containerization
bash

# Build Docker image
docker build -t massage-bot:latest .

# Run container
docker run -d -p 8080:8080 \
  -e BOT_TOKEN=your_token \
  -e ADMIN_ID=your_id \
  -e HEALTH_PORT=8080 \
  massage-bot:latest

🔒 Security

Sensitive data is excluded from version control. Use environment variables in production.
📞 Support

Create GitHub issues for bugs and questions.
