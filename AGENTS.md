# AGENTS.md — Cold-Start Guide

## What is this?

A clinical Telegram bot for massage therapists. Go 1.24 + PostgreSQL 15 + Docker.

## Repo Layout

- `cmd/bot/` — Entry point (`main.go`, `webapp.go`)
- `internal/delivery/telegram/` — Bot handlers and routing
- `internal/delivery/web/` — TWA (Telegram Web App) HTTP handlers + templates
- `internal/repository/postgres/` — Database layer (queries, migrations)
- `internal/storage/` — Filesystem clinical storage (markdown med-cards)
- `scripts/` — Deployment and utility scripts
- `deploy/` — Docker Compose configs and monitoring stack

## Domain Context (load on demand)

- **Database work** → load `.agent/skills/database-expert/SKILL.md`
- **TWA frontend** → load `.agent/skills/twa-aesthetics/SKILL.md`
- **Deployment/infra** → load `.agent/skills/devops-harness/SKILL.md`
- **AI integration** → load `.agent/skills/ai-integration-expert/SKILL.md`

## Hard Constraints (always active)

- See `.agent/rules/` — all rules are mandatory
- Key ones: `hypothesis-first`, `no-server-commits`, `pii-shield`

## Session Management

- Current state: `.agent/Project-Hub.md`
- Handoff context: `.agent/handoff.md`
- Workflows: `/checkpoint`, `/changelog`, `/release`, `/report`, `/review`

## Build & Test

- Build: `go build ./cmd/bot/`
- Test: `go test ./...`
- Vet: `go vet ./...`
- Local dev: see `.agent/sop/local-dev.md`

## Model Roles

- **Coding & Tasks**: Gemini 3 / Sonnet (primary implementers)
- **Review**: Opus 4.6 (skeptical reviewer — see `/review` workflow)
