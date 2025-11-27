# Telegram Massage Bot 🤖

A sophisticated Telegram bot for managing massage appointments with built-in scheduling, user management, and admin notifications.

## 🚀 Features

- **Appointment Booking**: Easy time slot selection and booking  
- **User Management**: Session-based user state management
- **Admin Notifications**: Instant alerts for new appointments
- **Health Monitoring**: Built-in health endpoints for DevOps
- **SQLite Database**: Persistent data storage

## 🛠️ Setup

### Prerequisites
- Go 1.21+
- Telegram Bot Token from [@BotFather](https://t.me/botfather)

### Installation

1. **Clone and configure**
   ```bash
   git clone https://github.com/kfilin/massage-bot.git
   cd massage-bot
   cp config.example.yml config.yml
   # Edit config.yml with your actual token and admin ID

Run the bot
bash

go run cmd/bot/main.go

🏥 Health Endpoints

    GET /health - System health with database status

    GET /ready - Readiness checks

    GET /live - Liveness probes

bash

curl http://localhost:8080/health

🔒 Security

Sensitive data is excluded from version control. Use environment variables in production.
📞 Support

Create GitHub issues for bugs and questions.
