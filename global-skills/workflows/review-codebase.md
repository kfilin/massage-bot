---
name: review-codebase
description: Explicitly audit the codebase for tech debt, dead code, and test coverage. Use when user runs /review or requests a memory lint.
category: System
priority: 2
---

# Skill: Review Codebase

## Objective
Provide an automated, focused routine for identifying dead code, checking test coverage, and updating project debt.

## When to Execute
- The user runs `/review`, `/lint-memory`, or explicitly asks to audit the codebase.
- Before a major release or when starting a refactoring phase.

## Workflow

### Step 1 — Linting and Coverage
- **Action**: Run the project's native linters (e.g., `go vet ./...`, `npm run lint`).
- **Action**: Run the test suite and check test coverage.
- **Action**: Search for obvious dead code, `TODO` markers, or placeholder comments.

### Step 2 — Analysis
- Compile the findings into a structured list.
- Highlight critical issues that violate `core-principles.md`.
- Identify opportunities for refactoring.

### Step 3 — Update `BACKLOG.md`
- Automatically append actionable technical debt items to the project's `BACKLOG.md` under a `## Technical Debt` section.

### Step 4 — Report to User
Present the findings to the user. Do NOT fix the issues automatically unless the user explicitly approves the remediation plan, following the **Hypothesis-First Engineering** strict loop in `core-principles.md`.
