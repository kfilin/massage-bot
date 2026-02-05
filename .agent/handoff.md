# Developer Handoff

## Current Status

**Version**: v5.5.0 (GitLab CI/CD + TWA InitData Auth)  
**Commit**: `1962a7b`  
**Date**: 2026-02-05

## Critical Context

### CI/CD Flow (NEW)

1. Push to GitHub → triggers `ci.yml` (tests only) + `mirror.yml` (syncs to GitLab)
2. GitLab receives push → triggers full pipeline (test → build → deploy)
3. GitHub deploys **disabled** - GitLab handles all deployments

### TWA Authentication (FIXED)

- Cancel uses `window.Telegram.WebApp.initData` (not URL tokens)
- Validated server-side with `validateInitData()` using bot token
- Session-based, never expires - immune to deploy-related token issues

### Required Secrets

- **GitHub**: `GITLAB_TOKEN` (GitLab PAT with `write_repository` scope)
- **GitLab**: `HOME_SERVER_IP`, `SSH_PRIVATE_KEY` (for deploy SSH)

## Technical Debt

1. DB logs very verbose - consider `log_statement=none` for production
2. TWA UI improvements backlogged (IDs 18-26 in `backlog.md`)

## Quick Commands

```bash
# Deploy
./scripts/deploy_home_server.sh    # Production
./scripts/deploy_test_server.sh    # Test

# Verify
make test && go build ./...
```
