# Last Session: Stability & Authentication Crisis Resolution

**Date**: 2026-02-01
**Focus**: Diagnosing "Invalid Token" errors and crash loops.

## 🎯 Accomplishments

- **Diagnosed and Fixed Crash Loop**: Identified that `telebot` API timeouts were causing the entire process to exit every 30s. Modified `main.go` to run `StartBot` async and block the main thread, ensuring the WebApp server stays alive 24/7.
- **Resolved DNS Collision**: Discovered that the Test environment (`massage-bot-test`) was responding to Production traffic due to a shared service name alias on the `caddy-test-net` network. Stopping the Test environment fixed the routing.
- **Enhanced Logging**: Added request tracing to `webapp.go` and retry logic to `bot.go`.
- **Force Rebuild**: Overcame stale Docker cache issues by forcing a code change and using `--no-cache`.

## 🚧 Challenges

- **"Works then doesn't"**: This confusing symptom was caused by the dual root cause (Crash Loop + DNS Load Balancing).
- **Stale Code**: Local `git pull` was failing silently (no upstream), leading to deployment of old binaries.

## 📝 Next Steps

- **Monitor**: Watch logs for `DEBUG_RETRY` to ensure the new resilience logic handles future network glitches gracefully.
- **Test Env**: When bringing the Test environment back, ensure it uses a DISTINCT Docker Service Name alias to avoid conflicting with Prod on the shared network.

## 🧠 Context for Next Agent

- **Current State**: TWA is updated on Test (manual deploy) and pipeline is running for the same commit (`7468de0`).
- **Critical**: `caddy-test-net` is `external: true`.
- **Docs**: `Project-Hub.md` is updated with v5.3.0.
