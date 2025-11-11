# 🧘 Vera Massage Bot - Professional Booking System

![Go Version](https://img.shields.io/badge/Go-1.21+-blue)
![License](https://img.shields.io/badge/License-MIT-green)

A production-ready Telegram bot for massage appointment booking with Google Calendar integration.

## ✨ Features

- **📅 Smart Booking**: Real-time availability checking with overbooking prevention
- **🇷🇺 Russian Interface**: Complete localization for Russian-speaking clients  
- **📱 Telegram Integration**: Seamless booking experience via Telegram
- **🗓️ Google Calendar Sync**: Automatic synchronization with business calendar
- **🛡️ Professional**: Clean architecture, proper error handling, health checks

## 🏗️ Architecture

- **Go 1.21+** with modern patterns
- **Clean Architecture** with ports/adapters
- **Telegram Bot API** integration
- **Google Calendar API** for appointment management
- **Health checks** on port 8080 (`/health`, `/ready`)

## 🚀 Quick Start

```bash
# Set up environment
export TG_BOT_TOKEN="your_telegram_bot_token"

# Run with Google Calendar integration
go run cmd/bot/main.go

# Or run with mock calendar for testing  
USE_MOCK_CALENDAR=true go run cmd/bot/main.go
