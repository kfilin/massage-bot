---
name: implementation-workflow
description: Unified development workflow for implementing features or bugfixes using TDD and subagent delegation.
category: Development
---

# Implementation Workflow

This skill provides a unified process for executing implementation plans with high discipline and quality. It combines Test-Driven Development (TDD) with structured execution and (optional) subagent delegation.

## Core Principles

1.  **Plan Before Action:** Never start without a clear implementation plan.
2.  **TDD First:** No production code without a failing test first.
3.  **Bite-Sized Tasks:** Execute in small, independent, verifiable steps.
4.  **Verify Everything:** Fresh evidence is required for every completion claim.

---

## Phase 1: Planning & Setup

1.  **Create/Read Plan:** Use `writing-plans.md` to establish a step-by-step roadmap.
2.  **Isolate Workspace:** Use `using-git-worktrees.md` to ensure a clean branch.
3.  **Create Todo List:** Use `TodoWrite` (if available) to track task progress.

---

## Phase 2: Execution (The TDD Cycle)

For each task in the plan, follow the **Red-Green-Refactor** cycle:

### 1. RED — Write a Failing Test
*   Write a minimal test that proves the feature is missing or the bug exists.
*   **Watch it fail:** Run the test and confirm it fails for the *correct* reason.
*   *Rationalization check:* "It's too simple to test" or "I'll test after" means you've skipped this step.

### 2. GREEN — Minimal Implementation
*   Write the *absolute minimum* code required to pass the test.
*   Avoid "future-proofing" or adding unrequested features (YAGNI).
*   **Watch it pass:** Run all tests. Confirm everything is green.

### 3. REFACTOR — Clean Up
*   Improve variable names, remove duplication, and optimize structure.
*   Keep tests green during refactoring.

---

## Phase 3: Subagent Delegation (Optional)

If subagents are available, delegate independent tasks to them to preserve your context.

1.  **Dispatch Implementer:** Provide the subagent with the exact task text and relevant file context.
2.  **Review Cycles:**
    *   **Spec Review:** Confirm the subagent's changes match the plan exactly.
    *   **Quality Review:** Ensure implementation follows project patterns and is clean.
3.  **Merge & Verify:** Once the subagent is DONE, verify the changes in your main session.

---

## Phase 4: Quality Gates & Completion

Before marking a task as DONE:
1.  **Fresh Verification:** Run the command that proves the claim.
2.  **No Completion Claims Without Evidence:** Provide logs or test results in your response.
3.  **Final Review:** Use `finishing-a-development-branch.md` once all tasks are complete.

---

## Red Flags (STOP and Start Over)
*   Implementation code written before the test.
*   Test passes on the first run (means you're testing existing behavior).
*   Ignoring test failures or "assuming" they will pass later.
*   Skipping verifications because "it's a small change."

**If you catch yourself skipping TDD: Delete the code. Start over. This is non-negotiable.**
