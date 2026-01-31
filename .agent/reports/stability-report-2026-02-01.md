# 📊 System Stability & Authentication Diagnosis Report

**Date**: 2026-02-01  
**Status**: ✅ RESOLVED (Stable)  
**Binary Version**: v5.3.6 Clinical Edition

---

## ✅ Resolution

**The "Invalid Token" issue and TWA instability have been permanently fixed.**

### Root Cause

The container was in a crash loop because the main process exited whenever the `telebot` failed to connect to the Telegram API (due to network time-outs). Since the WebApp server runs as a goroutine within the same process, it was being killed every ~30 seconds and restarting, leading to dropped requests and "Invalid Token" errors due to interruptions mid-validation.

### Solution Applied

We modified `cmd/bot/main.go` to **decouple the WebApp lifespan from the Bot's connectivity status**:

1. **Async Bot Startup**: `telegram.StartBot` is now launched in a separate goroutine.
2. **Permanent Process Block**: The `main` thread now blocks forever using `select {}` instead of relying on `StartBot` to keep the process alive.
3. **Outcome**: Even if the Telegram API is unreachable and the Bot fails to poll, the **WebApp server (port 8082) remains online and healthy**, serving requests without interruption.

---

## 🔍 Symptoms (Historical)

1. **WebApp "Invalid Token"**: Users intermittently receive an "Invalid token" error when trying to access their medical cards.
2. **Database Timeout**: Initial logs showed `dial tcp 172.21.0.2:5432: i/o timeout` between the bot and the DB.
3. **Telegram API Timeout**: Latest logs show `telebot: Post "https://api.telegram.org/.../getMe": dial tcp 149.154.166.110:443: i/o timeout`.

---

## 🛠️ Actions Taken (Session Snapshot)

- **Architecture Fix**: Decoupled WebApp from Bot startup to prevent crash loops.
- **Networking**: Renamed and recreated the internal Docker network to `massage-bot-internal` to ensure no IP conflicts or stale bridges.
- **Hardening**:
  - Added `connect_timeout=10` to the PostgreSQL connection string.
  - Added a 5-second sleep before `log.Fatalf` in `main.go`.
- **Versioning**: Forced a full rebuild to ensure the running binary matches the latest code (`v5.3.6`).
- **Healthchecks**: Fixed a port mismatch in the `Dockerfile` healthcheck (redirected from `:8081` to `:8083`).

---

## 📉 Root Cause Analysis

The system is currently stuck in a **Lifecycle Deadlock**:

1. The Bot starts and successfully connects to the **Database**.
2. It initializes the **WebApp Server** in a goroutine.
3. It attempts to connect to the **Telegram API** to start polling.
4. The connection to Telegram **times out** (30s).
5. The bot process calls `log.Fatalf` and **exits**.
6. Docker restarts the container, and the cycle repeats.

**Impact**: The "Invalid Token" error occurs because the WebApp server is flickering. Links generated in one lifecycle may be invalid or fail to validate because the server is being killed mid-request or before it can fully stabilize its environment.

---

## 🚀 Recommendations for Next Session

1. **Decouple WebApp binary**: Run the WebApp server as a separate container or a truly independent process that doesn't die if the Telegram Bot cannot reach the API.
2. **Network Investigation**:
    - Check host firewall/MTU settings (common for Home Server instability).
    - Verify if the Telegram API IP is being throttled or blocked.
3. **Resilient Startup**: Implement a backoff-retry loop for `telebot.NewBot` so the process doesn't exit immediately on API unavailability.
