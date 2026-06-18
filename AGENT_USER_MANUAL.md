# 📖 Vera Massage Bot — Agent Collaboration Manual

This guide explains how AI agents interact with the **Vera Massage Bot** codebase — what files to read, which conventions to follow, and how the agent harness works.

---

## 🧠 Project Knowledge Structure

### 📏 Rules (`AGENTS.md`)

The root `AGENTS.md` is the project constitution. Every agent MUST read it at session start. Key rules:

- **PII Shield**: Never output real patient data (names, phone numbers, emails). Use `[REDACTED]`.
- **No Production Commits**: Never commit from `/opt/vera-bot/` — use `scripts/deploy.sh`.
- **Logic Over Compliance**: Challenge brittle patterns; propose better engineering alternatives.
- **Hypothesis-First Engineering**: Observe → Hypothesize → Propose → Execute.
- **Hybrid Execution**: Read-only tools bypass approval; mutations require user approval.
- **Graphify Mandatory**: Graph queries run every session (no fallback).
- **Server Read-Only Convention**: `/opt/vera-bot/` is RW only for `data/`, `credentials.json`, `.env`, `.env.test`. Everything else flows through `scripts/deploy.sh`.

### 🛠️ Skills (`.pi/skills/`)

Specialized instructions that agents "equip" based on task:

| Skill | When to Use |
|-------|-------------|
| `startup/SKILL.md` | **Mandatory** — every session start |
| `handoff/SKILL.md` | End-of-session / context compaction |
| `hydrate-harness/SKILL.md` | Project harness hydration (`/hydrate`) |

### 📋 Session State (`BACKLOG.md`)

Temporal tracking — active tasks, completed work, blockers, and bugs. The "Active Focus" section tells the next agent what to work on.

---

## 🗣️ Communication Protocol

### "Logic Over Compliance"

If you suggest a pattern that introduces brittleness or risk, I will:

1. **Acknowledge** the request.
2. **Explain** risks (Security, Scalability, Consistency).
3. **Propose** a better alternative.
4. **Wait** for your decision.

### "State Your Hypothesis"

Before modifying files, I follow this loop:

1. **Observe**: "I see symptoms X and Y."
2. **Hypothesize**: "I suspect the root cause is Z."
3. **Propose**: "I'll test by doing A."
4. **Execute**: Only after you say "Go."

---

## 🔄 Session Workflow

### Start of Session

1. Agent reads `.pi/skills/startup/SKILL.md`.
2. Runs graphify queries (graph_stats, god_nodes, query_graph).
3. Reads `BACKLOG.md` (tail).
4. Reports branch, tree, services, tests, blockers — then waits.

### End of Session (`/handoff`)

Run the handoff skill when:
- Context is getting large (30+ exchanges).
- Switching task domains.
- User says "wrap up", "new chat", "save state".

The handoff routine:
1. Updates startup status section.
2. Updates `BACKLOG.md`.
3. Writes checkpoint handoff note to Obsidian vault.
4. Verifies persistence.
5. Git commits (user confirms push or local only).

---

## 🛠️ Extending the Harness

- **New skill?** Add to `.pi/skills/<name>/SKILL.md` (Agent Skills standard format).
- **New rule?** Update `AGENTS.md` (durable contracts only; temporal tracking stays in `BACKLOG.md`).

---

## 📚 Docs Index

| Document | Audience | Purpose |
|----------|----------|---------|
| `AGENTS.md` | Agents | Project rules and guardrails |
| `README.md` | Everyone | Project overview, features, quick start |
| `USER_GUIDE.md` | Patients | How to book, cancel, use medical card (EN) |
| `USER_GUIDE_RU.md` | Patients | How to book, cancel, use medical card (RU) |
| `DEVELOPER.md` | Developers | Architecture, testing, deployment |
| `docs/VERA_GUIDE_RU.md` | Therapist | Clinical workflow, admin commands (RU) |
| `docs/API.md` | Developers | HTTP endpoints, metrics reference |
| `docs/CI_CD_Pipeline.md` | Developers | Deployment pipeline |
| `CHANGELOG.md` | Everyone | Release history |
| `BACKLOG.md` | Agents | Active tasks, bugs, session history |

---
*Last updated: 2026-06-18.*
