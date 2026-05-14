---
name: generate-changelog
description: Analyze session changes, update the changelog, and push a new release checkpoint. Use when the user asks to generate a changelog or do a release checkpoint.
category: System
priority: 2
---

# Skill: Generate Changelog

## Objective
Analyze project changes since the last checkpoint, generate human-readable technical notes in `CHANGELOG.md`, and execute an automated git commit.

## When to Execute
- The user requests `/changelog` or explicitly asks to generate a changelog.
- A significant milestone or phase is completed, and it's time to snapshot the progress.

## Workflow

### Step 1 — Project Analysis
- **Trigger**: Run `git log` or `git diff` to analyze project changes since the last release or checkpoint.
- **List**: Enumerate implemented features, bug fixes, and architectural shifts.
- **Describe**: Provide concise engineering-focused descriptions for each item.

### Step 2 — Update `CHANGELOG.md`
Edit the project's `CHANGELOG.md` (or create one if it doesn't exist).
- Add a new dated entry (e.g., `## [YYYY-MM-DD]`).
- Insert the generated list of changes under appropriate subheadings (Features, Fixes, Architecture).

### Step 3 — Git Commit & Archival
Use the `// turbo` directive to automatically commit the changelog and push.
*Note: Before running the turbo block, explicitly ask the user: "Should I push this to the remote server, or commit locally only?" and adjust the `PUSH_TO_REMOTE` variable accordingly.*

// turbo
```bash
source .agent/project-config.env || true
go vet ./... || echo "⚠️ go vet warnings"
git add CHANGELOG.md
git commit -m "docs: release changelog $(date +%F)" || true
if [ "$PUSH_TO_REMOTE" = "true" ]; then
  git push origin ${GIT_MAIN_BRANCH:-master}
fi
```
