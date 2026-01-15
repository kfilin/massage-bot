# 💆 Vera Massage Bot

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-Private-red)

A professional Telegram bot for scheduling massage appointments, managing medical records, and tracking patient history. Built for Vera massage studio in Fethiye.

## 🌟 Key Features

- **Smart Booking**: Interactive calendar for scheduling appointments.
- **Medical Records**: Automatically generates Markdown-based medical cards (`.md`) for each patient.
- **Document Storage**: Securely saves MRI, X-Ray, Videos, and Voice messages to the patient's record.
- **Blacklist System**: "Shadow ban" feature to politely block unwanted users.
- **Admin Dashboard**: Real-time notifications for bookings, cancellations, and file uploads.
- **Google Calendar Sync**: Two-way synchronization with the therapist's calendar.

## 📚 Documentation

Detailed documentation is available for both users and developers:

- **[📖 User Guide](USER_GUIDE.md)**  
  *For Patients*: How to book, cancel, and access your medical card.

- **[🛠 Developer Guide](DEVELOPER.md)**  
  *For Maintainers*: System architecture, configuration, testing, and deployment instructions.

## 🚀 Quick Start (Admin)

1. **Configure**: Ensure `.env` is set up with your `TG_BOT_TOKEN` and `TG_ADMIN_ID`.
2. **Run**:
   ```bash
   make run
   ```
3. **Backup**:
   Use `/backup` in the bot to download a ZIP of all patient data.

---
*Created by Kirill Filin & AntiGravity AI.*
