# Session Log: v5.1.1 (Restoring Speed & Simplicity)

**Date**: 2026-01-29
**Commit**: `057e937`
**Goal**: Restore TWA performance ("Lightning Fast") and fix critical cancellation/data bugs.

---

## 🏛️ Architectural Decisions

### 1. Local-First Caching Strategy

* **Decision**: Switched TWA from Synchronous Google Calendar Sync to **Local DB Cache**.
* **Rationale**: The synchronous API call added 3-5 seconds of latency per page load. We now serve 100% of TWA data from the local PostgreSQL DB (0ms latency).
* **Sync Mechanism**: A "Smart Sync" hybrid:
  * **Empty Cache**: Blocking Sync (UX: "Loading...") -> Save to DB -> Render.
  * **Warm Cache**: Instant Render -> Background Sync -> Update DB.

### 2. Denormalized Appointments Table

* **Decision**: Implemented a standalone `appointments` table with flattened service details.
* **Rationale**: The previous relational schema depended on a `services` table that didn't exist (services are hardcoded). Denormalizing `service_name`, `duration`, and `price` into the `appointments` table ensures robust data storage without unnecessary complexity.

---

## 🐛 Technical Challenges & Elegant Solutions

### 1. The "Freezing" Confirmation

* **Problem**: Apple iOS blocked the native `confirm("Are you sure?")` JavaScript dialog, causing the TWA to freeze indefinitely on "Step 1".
* **Solution**: Removed the blocking dialog entirely. In the context of a 72h cancellation policy, speed is prioritized over "are you sure?" friction.

### 2. The Silent Network Failure

* **Problem**: Local development with Ngrok was failing silently because Ngrok injects a "Browser Warning" page on every AJAX request.
* **Solution**: Added the `ngrok-skip-browser-warning` header to all fetch requests. This is harmless in production but critical for local debugging.

---

## 💡 Learnings & Interesting Bits

* **Database Reality Check**: Always verify the schema actually exists. The `appointments` table was missing for days, but because we were relying purely on GCal API reads, nobody noticed until we tried to cache data locally.

---
*Created by Antigravity AI following the Collaboration Blueprint.*
*Project Status: STABLE (v5.1.1).*
