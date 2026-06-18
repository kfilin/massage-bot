---
description: The Twin-Environment Release Cycle
---
# 🚀 CI/CD Pipeline & Deployment Guide

This document describes the pipeline that moves code from development to production.

## 🔄 Deployment Overview

```
Local PC ──push──> GitHub ──mirror──> GitLab CI ──deploy──> Production
                                               └──deploy──> Test
```

1. **Code Source**: GitHub (`kfilin/massage-bot`)
2. **Deployment Engine**: GitLab CI/CD (triggered via GitHub Actions mirror)
3. **Target**: Home Server (test: `/opt/vera-bot-test/`, prod: `/opt/vera-bot/`)

---

## 🔄 The "Twin" Concept

- **Local PC** — The editor where code is written.
- **Test Container (`massage-bot-test`)** — Lab environment. Test against real networking (Caddy/TWA).
- **Production (`massage-bot`)** — The clinic where proven code serves patients.

---

## 1. 🛠️ Develop (Local)

```bash
# Stay up to date
git pull origin master

# Edit, test, commit
go test ./cmd/... ./internal/...
git add .
git commit -m "feat: my new feature"

# Push to GitHub (source of truth)
git push origin master
```

> [!NOTE]
> Use `go test ./cmd/... ./internal/...` instead of `go test ./...` to avoid `postgres_data` permission errors.

## 2. 🧪 Verify (Staging)

Deploy to test environment for real-world verification:

```bash
./scripts/deploy.sh test
```

The deploy script:
1. Runs port-collision pre-flight (skipped for test — allowed to share ports with prod).
2. SSHes to the server, pulls `origin/master`, rebuilds images, restarts.
3. Test environment at `vera-bot-test.kfilin.icu` (port 8086).

## 3. 🚢 Promote (Production)

### A. GitLab CI (Automatic)

Push to GitHub → GitHub Action mirrors to GitLab → GitLab CI builds image, pushes to registry, deploys to prod.

```bash
git push origin master
```

### B. Manual Deploy (Emergency / Direct)

```bash
./scripts/deploy.sh prod
```

> [!WARNING]
> The prod deploy does `git reset --hard origin/master` on the server. Any uncommitted or unpushed changes on the server are destroyed.

---

## 🛑 Critical Rules

1. **Immutable Server**: NEVER edit files directly inside `/opt/vera-bot/` or `/opt/vera-bot-test/`. Deploy scripts run `git reset --hard`. Your changes vanish.
2. **Allowed server-local state**: Only `data/`, `credentials.json`, `.env`, `.env.test`. Everything else flows through `scripts/deploy.sh`.
3. **Test First**: Always verify on test before promoting to production.
4. **Doc-only changes** (AGENTS.md, BACKLOG.md, CHANGELOG.md, etc.) can be synced with `ssh server "cd /opt/vera-bot && git pull --ff-only"` — never `scp` a tracked file directly.

---

## 📊 CI Pipeline Details (`.gitlab-ci.yml`)

| Stage | Image | Description |
|-------|-------|-------------|
| `test` | `golang:1.25.3-alpine` | Runs `go test ./cmd/... ./internal/...`, `go vet`, `golangci-lint` |
| `build` | `docker:27-cli` (Docker-in-Docker) | Builds production image, tags as `registry.gitlab.com/kfilin/massage-bot` |
| `deploy-test` | `alpine:3.21` | SSH to server, pull image, deploy test compose |
| `deploy-prod` | `alpine:3.21` (manual gate) | SSH to server, pull image, deploy prod compose |

All base images are pinned to specific versions (no `:latest`).

---
*Last updated: 2026-06-18.*