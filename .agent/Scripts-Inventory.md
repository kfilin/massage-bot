# 🛠️ Scripts Inventory

This document tracks the automation scripts available in `scripts/`. Use these tools to manage the lifecycle of the bot.

## 🚀 Deployment

| Script | Description | Usage |
| :--- | :--- | :--- |
| `deploy_home_server.sh` | **Deploy to Production**. Pulls from Git, rebuilds containers, and restarts the service. | `./scripts/deploy_home_server.sh` |
| `deploy_test_server.sh` | **Deploy to Test Env**. Deploys to the isolated test folder/port. | `./scripts/deploy_test_server.sh` |

## 🛡️ Maintenance & Operations

| Script | Description | Usage |
| :--- | :--- | :--- |
| `backup_data.sh` | **Backup Data**. Creates a ZIP archive of the `data/` directory. | `./scripts/backup_data.sh` |
| `renew_token.sh` | **Token Rotation**. Helper to update secrets/tokens safely. | `./scripts/renew_token.sh` |
| `report_metrics.sh` | **Health Check**. Gathers basic metrics and logs. | `./scripts/report_metrics.sh` |
| `refactor_logging.sh` | **Refactoring Helper**. Mass-updates logging statements to the new standard. | `./scripts/refactor_logging.sh` |

## ⚠️ Deprecated / Removed

* `mirror.sh`: **Removed**. Replaced by GitHub Actions (`.github/workflows/mirror.yml`).

## 📝 Best Practices

1. **Run from Root**: Always execute scripts from the project root (e.g., `./scripts/deploy.sh`).
2. **Check Permissions**: Ensure scripts are executable (`chmod +x scripts/*.sh`).
3. **Logs**: Most scripts output to stdout. Redirect to a file if needed.
