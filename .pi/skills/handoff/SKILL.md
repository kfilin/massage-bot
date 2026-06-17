---
name: handoff
description: End-of-session routine that preserves all context for the next agent. Triggered by "/handoff", "wrap up", "new chat", "save state", or when context is getting large. Replaces the old checkpoint skill.
category: System
priority: 1
---

# Skill: Session Handoff

## Objective
Ensure zero context loss between sessions. When a session ends, everything the next agent needs to hit the ground running must be persisted — not in your memory (which dies), but in files that survive.

## When to Execute

Run this routine when ANY of these occur:
- User says `/handoff`, "wrap up", "let's move to new chat", "save state", or similar
- User is about to switch major task domains
- **Context Compaction**: Context window is getting large (30+ exchanges or you're noticing forgotten details). You must actively monitor context drift and proactively suggest running `/handoff` to the user when context grows dense.
- You proactively sense the session should end (suggest it, don't force it)

## Handoff Routine (Execute in Order)

### Step 1 — Update `startup.md` Status Section

Edit `.agent/skills/startup.md` → the `## 📋 Status & Priorities` section.

Update it to reflect reality RIGHT NOW:
- Move any newly completed items to "Recently Completed"
- Update "Active Focus" to reflect what matters next
- Add any new blockers or context the next agent needs

This is the most critical step — it's what the next session reads first.

### Step 2 — Update `BACKLOG.md`

Edit `BACKLOG.md`:
- Mark completed items as `[x]`
- Add any new tasks discovered during this session
- Add any bugs found but not fixed

### Step 3 — Write Handoff Note

Create a timestamped handoff note in the Obsidian vault:

**Path**: `~/Documents/my_obsidian_vault/Bridge/massage-bot-project/Checkpoints/Handoff-YYYY-MM-DD-HHmm.md`

**Template**:
```markdown
---
author: [Agent Name]
type: handoff
created: YYYY-MM-DD HH:mm
session_topic: [2-5 word topic]
tags: [handoff, session-log]
---

# Handoff: [Brief Title]

## What Was Done
- [Concrete accomplishment 1]
- [Concrete accomplishment 2]

## What Changed (Files)
- `path/to/file`: [what and why]

## Unfinished / In Progress
- [Task that was started but not completed]
- [Blocker that prevented completion]

## Critical Context for Next Session
> Anything non-obvious that the next agent MUST know.
> Environment state, temporary workarounds, pending user decisions.

## Decisions Made
- [Decision]: [Rationale — why this approach, not another]
```

### Step 4 — Verify Persistence

Before closing out, confirm:
```
✓ startup.md status section updated
✓ BACKLOG.md updated
✓ Handoff note written to Bridge/massage-bot-project/Checkpoints/
✓ Any modified config files saved
```

Print this checklist to the user with actual ✓/✗ marks.

### Step 5 — Git Commit & Archival

Use the `// turbo` directive to automatically commit the session and push. 
*Note: Before running the turbo block, explicitly ask the user: "Should I push this to the remote server, or commit locally only?" and adjust the `PUSH_TO_REMOTE` variable accordingly.*

// turbo
```bash
source .agent/project-config.env || true
git add .
git commit -m "docs: session handoff $(date +%F)" || true
if [ "$PUSH_TO_REMOTE" = "true" ]; then
  git push origin ${GIT_MAIN_BRANCH:-${GIT_MAIN_BRANCH}}
fi
```

## Guidelines

- **Be concrete, not vague.** "Fixed sync" is useless. "Flattened server vault from `/opt/obsidian/my_obsidian_vault/my_obsidian_vault/` to `/opt/obsidian/my_obsidian_vault/`, restored local `.obsidian/` dir" is useful.
- **Include the WHY.** Don't just list what was done — explain decisions so the next agent doesn't reverse them.
- **Critical context = things that break if forgotten.** Temporary SSH tunnels, running background processes, half-applied migrations, pending user actions.
- **Under 500 words** for the handoff note. Dense, not verbose.
- **Don't ask for permission** to run this routine — just do it when triggered. The user already asked by saying "wrap up."
