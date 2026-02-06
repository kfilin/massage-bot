# Part 9: Add a Graceful‑Shutdown Path

## Goal

Implement a graceful shutdown mechanism to ensure the application cleans up resources (DB connections, background workers) and flushes logs/metrics before exiting.

## Sub-tasks

1. **Analyze & Critique**: The app currently uses `select{}` blocking and `ListenAndServe`, which precludes graceful exit.
2. **Signal Handling**: Listen for SIGINT/SIGTERM.
3. **Orchestration**: Use `sync.WaitGroup` to wait for components.
4. **Component Refactor**: Update HTTP servers and Bot to accept Context cancellation.

---

## 9.1 Solution Status

### **Completed**

> [!NOTE]
> **Status**: Restored and Verified.
> **Verification**:
>
> - `cmd/bot/main.go` uses `signal.NotifyContext` and `sync.WaitGroup`.
> - `webapp.go` and `health.go` accept cancellation Contexts.
> - `bot.go` stops polling on shutdown.

## 9.2 Analyze & Critique Solution

Critique identified "zombie workers" and hard kills.

- **Plan**: Replace `select{}` with `signal.NotifyContext`. Pass `ctx` and `WaitGroup` to all servers and workers.

## 9.2 Implementation Summary

### 1. Central Signal Management

Updated `cmd/bot/main.go` to use `signal.NotifyContext`, listening for `SIGINT` and `SIGTERM`.

### 2. Orchestrated Shutdown

Introduced `sync.WaitGroup` to coordinate lifecycle:

- **Health Server**: Gracefully shuts down with a 5-second timeout.
- **WebApp Server**: Gracefully shuts down with a 10-second timeout.
- **Telegram Bot**: Stops long-polling immediately via `b.Stop()`.

### 3. Background Worker Safety

The **Daily Backup Worker** was refactored to use a `select` block listening for the shutdown signal, ensuring it exits immediately during a system restart instead of sleeping.

### 4. Verification

The shutdown sequence respects the hierarchy: Signal → Context Cancel → Component Shutdown → WaitGroup Wait → Main Exit.
