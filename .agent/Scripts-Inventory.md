# 🛠️ Scripts Inventory

Detailed registry of all automation and maintenance scripts located in the `scripts/` directory.

### 1. `report_metrics.sh`

- **Purpose**: Generates a human-readable Business Intelligence (BI) and Technical Health report.
- **What it does**:
  - Fetches raw Prometheus data from the bot's `:8083/metrics` endpoint.
  - Parses and aggregates data into groups: Bookings, Loyalty, Service Popularity, and API Latency.
  - Displays a clean CLI summary for quick decision-making.
- **Usage**: `./scripts/report_metrics.sh [metrics_url]`

### 2. `backup_data.sh`

- **Purpose**: Creates local backups of clinical data and logs.
- **What it does**:
  - Zips the entire `./data` directory.
  - Excludes existing backups to prevent recursive bloat.
  - Implements a retention policy: keeps only the last 30 backups.
- **Usage**: Typically run via `cron` on the production server.

### 3. `deploy_home_server.sh`

- **Purpose**: Automates the deployment pipeline on the Home Server.
- **What it does**:
  - Switches to the `/opt/vera-bot` directory.
  - Force-pulls the latest `master` branch from GitHub (mirroring to GitLab is now manual).
  - Triggers a full Docker Compose rebuild and restart (`--force-recreate`).
  - Outputs the current container status and tail-logs for verification.
- **Usage**: Triggered by the GitLab CI/CD pipeline or manually via SSH.

### 4. `renew_token.sh`

- **Purpose**: Manual helper for Google OAuth Refresh Token renewal.
- **What it does**:
  - Provides the exact URL for the therapist to authorize the app.
  - Exchanges the user's authorization code for a fresh Access/Refresh token pair using `curl`.
  - Outputs the JSON format required for the `GOOGLE_TOKEN_JSON` environment variable.
- **Usage**: Run manually every ~6 months or when logs show `invalid_grant`.
