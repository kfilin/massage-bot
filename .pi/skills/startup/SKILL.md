---
name: startup
description: Project-agnostic session-start skill for the Agentic OS. Injects a thin orientation skeleton in <5K tokens, then asks the project to provide its own context (AGENTS.md, README, docs/, source layout). Use this as the base skill in any project the OS hydrates.
category: System
priority: 0
---

# Skill: Session Startup — Project-Agnostic Skeleton

## Objective

Orient the agent in any project the Agentic OS hydrates. This skill provides the **skeleton** — the steps every session must take, the output format, and the constraints. The **content** (architecture, services, gotchas, paths) is **discovered at runtime** from the project itself, not hardcoded here. That keeps this file small and reusable.

**The cardinal rule:** never edit this file to add project-specific knowledge. Put that in the project's `AGENTS.md`, `README.md`, or `.agent/project-config.env`. The OS will read those in Step 2 below.

---

## Mandatory: Execute on Session Start

> [!IMPORTANT]
> **This routine applies to ALL agents in this workspace — including IDE agents, CLI agents, and the OS's own subagents.**
> If you are an agent that joined mid-session and has not run this routine yet: **stop what you are doing and run it now.** Do not answer the user's request first.

**You MUST execute these steps automatically at the start of every conversation, BEFORE responding to any user task.**

### Step 0 — Hydration Check (3-Strike Nag)

1. Read `.agent/project-config.env`. If the file is absent, skip to Step 1 (legacy project).
2. If `HYDRATED=false`:
   - **Tell the user**: *"This project harness has not been hydrated. Please fill out `.agent/project-config.env` and run `/hydrate`."*
   - **3-Strike Rule**: If the user persists with a normal task, remind them up to **3 times per session** before relenting (prevents hard blockages). After 3 strikes, proceed.
3. If `HYDRATED=true`, continue to Step 1.

### Step 1 — Graph Awareness (Codebase Understanding)

If a code-graph MCP server is available (`graphify`, `repomix`, `code-graph`, etc.), run three queries silently. Do not print raw output:

```
1. graph_stats()              → understand scale (file count, node count)
2. god_nodes(top_n=10)        → identify core abstractions
3. query_graph("What are the main components and how do they interact?",
              depth=2, mode="bfs")
```

If no graph server is available, skip silently. The discovery steps below cover the gap.

Store results mentally. Use them to inform your responses throughout the session.

---

### Step 2 — Discover Project Context

**This is where the project's own knowledge lives. Find it; do not invent it.**

Run all of these in one command:

```bash
cd <project-root> && \
  echo "=== ROOT FILES ===" && ls -la 2>/dev/null | head -30 && \
  echo "=== PROJECT CONFIG ===" && \
    [ -f .agent/project-config.env ] && cat .agent/project-config.env || echo "  (no project-config.env)" && \
  echo "=== ENTRY POINTS ===" && \
    for f in AGENTS.md CLAUDE.md README.md CONTRIBUTING.md docs/INDEX.md; do \
      [ -f "$f" ] && echo "  ✓ $f ($(wc -l < $f) lines)" || echo "  ✗ $f"; \
    done && \
  echo "=== LANGUAGES / BUILD ===" && \
    [ -f go.mod ]          && echo "  go: $(head -1 go.mod | awk '{print $2}')" || true && \
    [ -f package.json ]    && echo "  node: $(grep -m1 '"name"' package.json)" || true && \
    [ -f Cargo.toml ]      && echo "  rust: $(grep -m1 '^name' Cargo.toml)" || true && \
    [ -f pyproject.toml ]  && echo "  python: $(grep -m1 '^name' pyproject.toml)" || true && \
    [ -f pom.xml ]         && echo "  java (maven)" || true && \
    [ -f build.gradle* ]   && echo "  java (gradle)" || true && \
  echo "=== SERVICE MANAGER ===" && \
    [ -f docker-compose.yml ]   && echo "  compose: $(grep -c '^  [a-z]' docker-compose.yml) services" || true && \
    [ -f docker-compose.yaml ]  && echo "  compose: $(grep -c '^  [a-z]' docker-compose.yaml) services" || true && \
    [ -f Dockerfile ]           && echo "  docker: single container" || true && \
    [ -f Chart.yaml ]           && echo "  helm chart" || true && \
    [ -d .kube ] || [ -d k8s ]  && echo "  k8s manifests" || true && \
    [ -f fly.toml ]             && echo "  fly.io" || true && \
  echo "=== JOURNAL ===" && \
    [ -f BACKLOG.md ]    && echo "  BACKLOG.md: $(wc -l < BACKLOG.md) lines" || echo "  no BACKLOG.md" && \
    [ -f CHANGELOG.md ]  && echo "  CHANGELOG.md: present" || true && \
  echo "=== DOCS ===" && \
    [ -d docs ] && ls docs/ 2>/dev/null | head -20 || echo "  no docs/ dir"
```

**Read in this order (each is optional — skip if missing):**

1. `.agent/project-config.env` — if present, defines `PROJECT_NAME`, `GIT_MAIN_BRANCH`, `HYDRATED`, etc. (Agentic OS convention)
2. `AGENTS.md` (or `CLAUDE.md`) — the project's mandatory rules, conventions, and "do not touch" lists
3. `README.md` — one-paragraph project summary
4. `docs/INDEX.md` (if present) — points to deeper docs
5. The project's own `.agent/skills/` (if present) — project-specific skills that **extend** this skeleton

**Then run the universal state commands:**

#### 2.a — Git state
```bash
echo "=== BRANCH ===" && git branch --show-current && \
  echo "=== HEAD ==="   && git log -1 --pretty='%h %s  %ai' && \
  echo "=== TREE ==="   && git status --short && \
  echo "=== UNTRACKED ===" && git ls-files --others --exclude-standard | head -5
```

#### 2.b — Journal (tail only, NOT head)
```bash
[ -f BACKLOG.md ] && tail -60 BACKLOG.md || echo "  no BACKLOG.md"
```

#### 2.c — Source layout (auto-discover, never read)
```bash
find . -maxdepth 3 -type d \( -name node_modules -o -name .git -o -name vendor -o -name target -o -name dist -o -name build -o -name __pycache__ \) -prune -o -type d -print 2>/dev/null | head -40
```

#### 2.d — Services (auto-detect manager)
```bash
if [ -f docker-compose.yml ] || [ -f docker-compose.yaml ]; then
  docker-compose ps 2>/dev/null | awk 'NR==1 || /Up|Exit/'
elif command -v kubectl >/dev/null 2>&1 && ([ -d .kube ] || [ -d k8s ]); then
  kubectl get pods 2>/dev/null | head -20
elif command -v docker >/dev/null 2>&1; then
  docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null
else
  echo "  no recognized service manager"
fi
```

#### 2.e — Test status (auto-detect runner)
```bash
if [ -f go.mod ]; then
  go test ./... -count=1 2>&1 | tail -3
elif [ -f package.json ]; then
  jq -r '.scripts // {} | to_entries[] | select(.key | test("test|lint|type"; "i")) | "  \(.key): \(.value)"' package.json 2>/dev/null
elif [ -f Cargo.toml ]; then
  cargo test 2>&1 | tail -3
elif [ -f pyproject.toml ]; then
  if grep -q 'pytest' pyproject.toml 2>/dev/null; then
    pytest 2>&1 | tail -3
  fi
elif [ -f Makefile ]; then
  grep -E '^(test|check|lint):' Makefile
fi
```

The summary line is what matters. We do **not** need the full PASS/FAIL names at startup.

**Full test suites are pre-commit / pre-deploy checks, not session-start.**

---

### Step 3 — Output Format (6 lines, then stop)

```
project:   <name> @ <hash>  "<commit subject>"  (<date>)
tree:      clean | dirty: <list of files>
journal:   <one-line summary of latest BACKLOG entry or "no journal">
services:  N/M up  |  <list names if any down>  |  no service manager
tests:     green | red | unknown | no test runner
blocker:   <one line if any, else "none">
```

If a field does not apply (no `BACKLOG.md`, no service manager, no test runner), say so explicitly with `no <thing>` rather than omitting. The user can scan 6 lines in 2 seconds and know the full state.

After the report, **stop and wait for the user**. If the user gave a task before startup finished, finish steps 1-2 first, then do the task — never skip startup.

---

## MUST NOT at startup

- ❌ Read every file in the project. The discovery steps above are enough.
- ❌ Run the full test suite (`npx vitest run`, `cargo test`, `playwright test`, `tsc --noEmit`, `cargo check`, etc.). Step 2.e summary is sufficient.
- ❌ `find` / `grep` the entire codebase to "orient." Use the source-layout summary from 2.c.
- ❌ Read whole files without `offset/limit` when files are >200 lines.
- ❌ Re-read files from earlier turns. The harness keeps history; re-reading is pure waste.
- ❌ Propose work, suggest refactors, or start implementing before the user gives a task. **Startup = report and wait.**
- ❌ Edit this file to add project-specific knowledge. Project knowledge belongs in the project's `AGENTS.md`, `README.md`, or `.agent/project-config.env`.

---

## When to read more on demand (lazy load)

| Need to know about | Read |
|---|---|
| Project rules and conventions | `AGENTS.md` (already read in Step 2) |
| Detailed project description | `README.md` |
| Deeper docs index | `docs/INDEX.md` (if present) |
| Phase / milestone plan | `<project>/docs/<plan>.md` (auto-discovered in 2.a) |
| Source layout for a subsystem | `ls <subdir>/` (don't read a file) |
| Project-specific skills | `<project>/.agent/skills/` (if present) |
| Reference conventions from Agentic OS | `~/Projects/agentic-os/` or the OS's own docs |
| Other global methodology skills | `global-skills/` (TDD, code review, debugging, etc.) |

---

## Project Extension Points (where to put project-specific knowledge)

If a project needs to **add** to this skeleton without modifying it, the project can provide:

| File | Purpose |
|---|---|
| `.agent/project-config.env` | `PROJECT_NAME=...`, `GIT_MAIN_BRANCH=...`, `HYDRATED=true` — read first in Step 2 |
| `AGENTS.md` | Project rules, conventions, hard "do not" lists |
| `README.md` | One-paragraph project summary |
| `docs/INDEX.md` | Map of `docs/` — what to read when |
| `<project>/.agent/skills/<name>.md` | Project-specific skills that extend (not replace) the OS skills |
| `BACKLOG.md` | Session journal — OS reads the tail (last 60 lines) at startup |

The project's `AGENTS.md` is the canonical place for **project-specific gotchas** (e.g. "we use `docker-compose` not `docker compose`", "this monorepo has 4 services on ports 8088/50052/50051/50053"). Do not duplicate them here.

---

## Guidelines

1. **Don't hallucinate paths or configs** — always verify with `ls`, `cat`, or the equivalent before acting.
2. **Verify before reporting success** — run the thing, check the output.
3. **Infrastructure is live** — changes to service / deploy configs may affect production. Be careful.
4. **Use the graph** — query the code-graph MCP server when working on code to understand impact before changing things.
5. **Read existing skills** — check `global-skills/` for established patterns (TDD, code review, debugging) before creating new ones.
6. **Stay project-agnostic** — this file must work in any project the OS hydrates. If a piece of knowledge is specific to one project, it does not belong here.
