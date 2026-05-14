---
name: startup
description: Mandatory context injection at the start of every session. Gives the agent full awareness of the project architecture, infrastructure, ops patterns, and current priorities. Run this BEFORE any user task.
category: System
priority: 0
---

# Skill: Session Startup — Project Context Injection

## Objective
Load the complete project context so you can work effectively from the first message. This replaces cold-start guessing with deep awareness of the codebase, infrastructure, and conventions.

## Mandatory: Execute on Session Start

> [!IMPORTANT]
> **This routine applies to ALL agents in this workspace — including Antigravity IDE agents, not just the Telegram/Discord bot.**
> If you are an IDE agent (Antigravity, Goose, Claude Code, etc.) and you have not run this routine yet: **stop what you are doing and run it now.** Do not answer the user’s request first.

**You MUST execute these steps automatically at the start of every conversation, BEFORE responding to any user task.**

### Step 0 — Hydration Check (The Nag System)

1. Read `.agent/project-config.env`.
2. If `HYDRATED=false`, STOP everything and inform the user: *"This project harness has not been hydrated. Please fill out `.agent/project-config.env` and ask me to run `/hydrate`."*
3. **The 3-Strike Rule**: If the user tries to bypass this warning and gives you a normal task, you MUST remind them to hydrate first. You will do this up to 3 times per session before relenting (to prevent hard blockages).

### Step 1 — Graph Awareness (Codebase Understanding)

If the `graphify` MCP server is available, run these queries silently (do not print raw output to user):

```
1. graph_stats()           → understand scale
2. god_nodes(top_n=10)     → identify core abstractions
3. query_graph("What are the main components and how do they interact?", depth=2, mode="bfs")
```

Store results mentally. Use them to inform your responses throughout the session.

### Step 2 — Project & Infrastructure Context

Absorb the following project context. Do NOT summarize it back to the user. Use `graphify` (Step 1) for current file structure and code logic.

---

## 🖥 Infrastructure & Environment

### Server Access
- **SSH alias**: `server` (use `ssh server "command"` to run remote commands)
- **Domain**: `agenticlab.kfilin.icu` (Caddy reverse proxy with auto-TLS)
- **Primary IP**: `45.89.189.13`

### Key Paths & Services
| Path | Purpose |
|---|---|
| `/opt/massage-bot/` | Production deployment of this project |
| `/opt/obsidian/` | Obsidian vault data (WebDAV-served) |
| `/opt/caddy/` | Caddy reverse proxy config (`Caddyfile`) |

### Caddy Routes (agenticlab.kfilin.icu)
| Route | Target |
|---|---|
| `/sync/*` | WebDAV container (Obsidian sync) |
| `/webhook/tg` | Omnichannel Bridge (Telegram webhook) |

### Local Environment (PC)
- **Project**: `~/Projects/Antigravity_On_Steroids/massage-bot/`
- **Obsidian vault**: `~/Documents/my_obsidian_vault/`
- **Vault symlink**: `massage-bot/vault` → `~/Documents/my_obsidian_vault`

---

## ⚙️ Ops Patterns & Conventions

### Deploying Changes
1. Local development → test locally.
2. Push to server: `rsync` or `git pull` on server.
3. Rebuild: `ssh server "cd /opt/massage-bot && docker compose up -d --build <service>"`.

### Common Gotchas
- **File naming**: Android Obsidian cannot handle `*` in filenames. Avoid special characters in note names.
- **WebDAV nesting**: The vault structure must be flat inside `my_obsidian_vault/`. No double nesting.
- **Session Handoff**: When ending a session, execute the `handoff` skill (`global-skills/handoff.md`).

---

## 📋 Status & Priorities

> **Update this section as project priorities shift via the Handoff routine.**

### Recently Completed
- ✅ **Universal Project Harness**: Successfully ported strict TDD/Debugging constraints from `massage-bot`, implemented automated `/hydrate` and `/changelog` scripts, and authored the master `HARNESS_GUIDE.md`.
- ✅ Architectural Review: Established "Integration First" vision (AI as Executor reading from external tools like Notion/Drive) rather than building a custom multi-tenant RBAC system.
- ✅ Idea Generation: Created `IDEAS.md` and moved highly hypothetical Phase 12-14 items into it.
- ✅ Semantic Caching Strategy: Validated Redis as a Semantic Cache layer (with webhook invalidation) over using local LLMs for RAG routing.

### Active Focus
- **Integration Layer**: Planning external API connectors (OAuth/Service Accounts) to interface with established company tools.
- **Semantic Cache Layer**: Designing the Redis vector-cache logic to reduce LLM token overhead for repetitive departmental queries.

---

## Guidelines

1. **Don't hallucinate paths or configs** — verify with `ls`, `cat`, or `ssh server`.
2. **Verify before reporting success** — run the thing, check the output.
3. **Infrastructure is live** — changes to server configs affect production.
4. **Use the graph** — query Graphify to understand impact before changing code.
5. **Read existing skills** — check `global-skills/` for patterns.
