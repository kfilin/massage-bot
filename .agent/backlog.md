# Project Backlog

## 📋 Observation & Improvement Ideas

### 1. "History of Visits" Data Accuracy & Utility

- **Issue**: Cancelled appointments in Google Calendar might still influence the "First/Last Visit" dates or appear in the history table even if the total count decreases correctly.
- **Goal**: Decide if the "History of Visits" block is needed. If it is kept, ensure cancelled/deleted events are filtered out completely from the chronological calculations.
- **Context**: The user noticed that while the visit count badge updates correctly, the history table sometimes retains cancelled future entries.

### 2. Robust TWA Authentication (InitData)

- **Issue**: Opening the Web App from the Telegram "Menu Button" or static links results in a "Missing ID or Token" error because the current logic requires query parameters.
- **Goal**: Implement validation using `window.Telegram.WebApp.initData`. This allows identifying the user securely regardless of how the app was opened.
- **Context**: The user sent a screenshot showing the "Missing id or token" error when presumably opening the app from a non-dynamic source.

### 3. WebDAV & VPN Clinical Privacy

- **Issue**: WebDAV is currently accessible via the public internet (Basic Auth). While functional, it exposes medical records to brute-force attempts.
- **Goal**: Restrict WebDAV access to the **SoftEther SSTP VPN** local network. Update Caddy/Docker configuration to only allow connections from VPN IP ranges.
- **Context**: The user runs a home-based SSTP VPN, which is the "Gold Standard" for securing clinical data at rest.

---
*Last updated: 2026-01-23 20:25*
