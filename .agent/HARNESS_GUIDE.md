# Universal Project Harness: Master Guide

This document serves as the definitive manual for the `.agent/` and `global-skills/` collaboration harness. It is designed to be read by both Human Project Managers and AI Agents to ensure strict discipline, high-velocity automation, and seamless context sharing.

---

## Part 1: Architecture & Philosophy

The Universal Project Harness is an "Agentic Operating System." It blends rigorous Test-Driven Development (TDD) and debugging constraints with high-speed `// turbo` bash automations. 

**The Hydration Concept**
To solve the AI "Cold Start" problem (where generic prompts cause hallucinations), the harness uses a "Hydration" pattern. The `.agent/` folder acts as a universal template containing variables like `massage-bot`. When deployed to a new repository, it must be "hydrated" to hardcode those variables, permanently turning the generic template into a project-specific brain.

---

## Part 2: The User Guide (Human-Facing)

### Bootstrapping a New Project
1. **(Human Action)**: Copy the `.agent/` and `global-skills/` directories into your new repository.
2. **(Human Action)**: Manually open `.agent/project-config.env` and fill in the variables (`PROJECT_NAME`, `GIT_MAIN_BRANCH`). Ensure `HYDRATED=false`.
3. **(Human Action)**: Start a session with your AI agent and type the command: `/hydrate`.
4. **(Agent Action)**: The AI will automatically replace all template variables across the harness and commit the changes.

### Session Management
- **Command:** `/handoff` (or "wrap up")
- **When to use:** At the end of your workday, or when you are switching to a completely different task domain. 
- **What it does:** Updates the `startup.md` status, writes a timestamped note to your Obsidian vault, and automatically commits the session to Git.

### Automated Workflows
- **Release Checkpoints:** Type `/changelog` to have the AI analyze your recent git commits, write a human-readable `CHANGELOG.md` entry, and push the release.
- **Codebase Audits:** Type `/review` to have the AI run native linters, check test coverage, identify dead code, and append technical debt to your `BACKLOG.md`.

---

## Part 3: The Developer Guide (AI-Facing)

If you are an AI Agent operating in this harness, you are bound by the following laws:

### 1. The Nag System
When you boot up, you MUST read `.agent/project-config.env`. If `HYDRATED=false`, you must actively refuse to do normal coding tasks and prompt the user to run `/hydrate`. (You may nag up to 3 times per session before relenting).

### 2. The Strict Debugging Loop
You are forbidden from "guess and check" coding. You must follow the strict loop defined in `quality-gates.md`:
1. **Observe:** Prove the bug exists (read logs/trace).
2. **Hypothesize:** State the exact root cause.
3. **Verify:** **Wait for user acknowledgment** before modifying any code.

### 3. Core Constraints
- **PII Shield:** Never output real user data, names, or emails in chat or artifacts. Use `[REDACTED]` or IDs.
- **Interactive Verification:** Any UI or frontend changes you make MUST be presented to the user via a `walkthrough.md` artifact containing screenshots or browser recordings.
- **Explicit Cross-Linking:** When writing knowledge items or SOPs, always use hard markdown links to other files so the Knowledge Base remains traversable without full-text search.

---

## Part 4: Configuration Reference

The `.agent/project-config.env` file tracks the state of the harness.

- `HYDRATED`: (boolean) `true` if the project template has been initialized. If `false`, triggers the Nag System.
- `PROJECT_NAME`: (string) Used to ground the AI's context and format automated git commit messages.
- `GIT_MAIN_BRANCH`: (string) Used by the `// turbo` automations (like `/handoff` and `/changelog`) to know where to push code on the remote server.
