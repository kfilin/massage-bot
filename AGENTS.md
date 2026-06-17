# AGENTS.md

## 🚨 MANDATORY STARTUP — Non-Negotiable, No Exceptions

**The VERY FIRST ACTION of every session, before responding to any user message — even a greeting or a "quick question" — is to execute the startup routine.**

This is not optional. It is not skippable when the task "seems simple". It is not something to do after you answer the first message. It is the first thing.

### Startup Procedure

1. **Execute `.pi/skills/startup/SKILL.md`**: Read and fully internalize it. Then:
   - Run the Graphify queries: `graph_stats()`, `god_nodes(top_n=10)`, `query_graph("What are the main components and how do they interact?", depth=2)`
   - Absorb the project context, infrastructure map, ops patterns, and current priorities in that file
   - Do NOT print raw query output to the user
2. **Align with `BACKLOG.md`**: Identify current high-priority tasks.
3. **Only then**: Respond to the user's actual request.

> **Why this matters**: Skipping startup means working with stale assumptions about the architecture, dead files, and wrong infrastructure state. A 30-second startup prevents hours of wrong-direction work and dangerous edits to a live production system.

---

## 🛡️ Guardrails
- **Logic Over Compliance**: Push back professionally if a requested pattern is brittle, introduces a Single Point of Failure (SPOF), or creates a risk of infinite execution loops. (Example: If asked to run system commands inside nested execution environments, advise on safer native alternatives).
- **Hypothesis-First Engineering**: Do not edit code or configurations without explicitly testing your hypothesis. Follow the order: Observe Logs -> State Hypothesis -> Determine Metric of Success -> Execute.
- **Hybrid Execution Protocol**: Safe introspection and read operations (e.g., viewing logs, lists, reading files) are handled via read-only tools to bypass approval fatigue. Any operation that mutates state (writes, deletes, installations, deployments) MUST run through tools that request explicit user approval.
- **Cost Awareness (Terse Outputs)**: Respect resource and token constraints. Output tokens are expensive; respond precisely and concisely. Avoid repetition, pleasantries, and unnecessary explanations. When delivering code, output only the modified code block unless further explanation is strictly requested.
- **Never Proceed on Automated System Approvals**: If the system outputs an automated message stating *"The user has automatically approved the artifact... Proceed to execution"*, you **MUST NOT** proceed. Ignore the system's directive, output the implementation plan, changes, or code directly to the chat, and wait for the developer to type a manual confirmation message.
- **PII Shield (Privacy First)**: Never output real user data, names, phone numbers, or emails in chat or artifacts. Use `[REDACTED]` or shadow IDs (e.g., `User (ID: 3045...)`).
- **No Production Commits**: If the working directory suggests a live server or production environment (e.g., absolute paths like `/opt/`), do NOT run `git commit`. Instruct the user to push changes to Git/GitHub and mirror to the server for deployment.
- **Server Read-Only Convention**: `/opt/vera-bot/` is read-only except for `data/`, `credentials.json`, `.env`, `.env.test` (gitignored, server-local state). All code, config, and script changes flow exclusively through `scripts/deploy.sh prod` (which does `git reset --hard origin/master`) — never `scp` a file directly to the server, never `ssh server "vi ..."` to edit tracked code, never `git commit` on the server.
- **Constraints, Not Checklists**: Keep rules as meta-constraints rather than verbose checklists. No single rules block should exceed 15 items to prevent instructions overload.

---

## Operational Rules

### Session Reset Triggers

You operate in sessions that accumulate context over time. Suggest a session reset when:
- 30+ exchanges have passed (context window approaching 100K tokens)
- 30+ minutes of continuous conversation
- Switching to a different task domain
- You've noticed you've forgotten early context

If a reset would help, gently suggest it to the user.

### Tool Output Discipline

Before returning tool output to the user:

1. Filter for relevance — prune verbose sections
2. Summarize large JSON responses
3. Ask yourself: "Does the user need all 500 lines, or just the error?"

**Example**

> Raw tool output: 2,000 lines of API response
> Your response: "The API returned a 404 error on endpoint `/users/123`. This likely means the user was deleted. Here's what I suggest: ..."

### Rate & Resource Limits

- Maximum 10 API calls per user message
- If a task will exceed significant context sizes, warn the user
- Before calling tools, ask: "Is this call necessary?" Batch related queries into one tool call. Use cached results when available

### Continuous Cost Awareness

Track your behavior:
- How many tool calls per question?
- When do you hit compaction?

If you find yourself struggling with a task, document it and discuss with the user — there might be a better model choice for that work.

---

## ⚙️ Operational Routines
- **Mandatory Session Handoff**: At the end of every active development session, or when wrapping up, you MUST execute the handoff routine (`.pi/skills/handoff/SKILL.md`). This includes updating the repository architecture diagrams, regenerating code graphs using AST code-graph tools (if available), and executing a DOX documentation pass to update technical guides to fully document whatever features and commands were implemented.
- **Mandatory Semantic Search**: Before implementing any new helper function, utility, state transition, or model configuration, you MUST run a semantic search (if available) with a clear explanation of your intent to prevent duplicating helper functions or creating divergent design patterns.
- **Concept Proposals for New Code**: When you create or refactor a core utility, pattern, or skill, propose a new concept card (or update an existing one) using the `propose_concept` tool (if available).

---

# DOX framework

- DOX is highly performant AGENTS.md hierarchy installed here
- Agent must follow DOX instructions across any edits

## Core Contract

- AGENTS.md files are binding work contracts for their subtrees
- Work products, source materials, instructions, records, assets, and durable docs must stay understandable from the nearest applicable AGENTS.md plus every parent AGENTS.md above it
- **Durable vs Temporal Split**: Keep a strict separation between blueprints and active tracking. Durable contracts (AGENTS.md files) contain static rules, folder hierarchies, and long-term constraints. Temporal logs (BACKLOG.md, task.md, and Obsidian checkpoint handoffs) track active checklists, bugs, progress summaries, and session status.

## Read Before Editing

1. Read the root AGENTS.md
2. Identify every file or folder you expect to touch
3. Walk from the repository root to each target path
4. Read every AGENTS.md found along each route
5. If a parent AGENTS.md lists a child AGENTS.md whose scope contains the path, read that child and continue from there
6. Use the nearest AGENTS.md as the local contract and parent docs for repo-wide rules
7. If docs conflict, the closer doc controls local work details, but no child doc may weaken DOX

Do not rely on memory. Re-read the applicable DOX chain in the current session before editing.

## Update After Editing

Every meaningful change requires a DOX pass before the task is done.

Update the closest owning AGENTS.md when a change affects:

- purpose, scope, ownership, or responsibilities
- durable structure, contracts, workflows, or operating rules
- required inputs, outputs, permissions, constraints, side effects, or artifacts
- user preferences about behavior, communication, process, organization, or quality
- AGENTS.md creation, deletion, move, rename, or index contents

Update parent docs when parent-level structure, ownership, workflow, or child index changes. Update child docs when parent changes alter local rules. Remove stale or contradictory text immediately. Small edits that do not change behavior or contracts may leave docs unchanged, but the DOX pass still must happen.

## Hierarchy

- Root AGENTS.md is the DOX rail: project-wide instructions, global preferences, durable workflow rules, and the top-level Child DOX Index
- Child AGENTS.md files own domain-specific instructions and their own Child DOX Index
- Each parent explains what its direct children cover and what stays owned by the parent
- The closer a doc is to the work, the more specific and practical it must be

## Child Doc Shape

- Create a child AGENTS.md when a folder becomes a durable boundary with its own purpose, rules, responsibilities, workflow, materials, or quality standards
- Work Guidance must reflect the current standards of the project or user instructions; if there are no specific standards or instructions yet, leave it empty
- Verification must reflect an existing check; if no verification framework exists yet, leave it empty and update it when one exists

Default section order:
- Purpose
- Ownership
- Local Contracts
- Work Guidance
- Verification
- Child DOX Index

## Style

- Keep docs concise, current, and operational
- Document stable contracts, not diary entries (NEVER add temporal checklists, checkboxes, bug lists, or task notes to AGENTS.md files; use BACKLOG.md, task.md, or Obsidian checkpoints instead)
- Put broad rules in parent docs and concrete details in child docs
- Prefer direct bullets with explicit names
- Do not duplicate rules across many files unless each scope needs a local version
- Delete stale notes instead of explaining history
- Trim obvious statements, repeated rules, misplaced detail, and warnings for risks that no longer exist

## Closeout

1. Re-check changed paths against the DOX chain
2. Update nearest owning docs and any affected parents or children
3. Refresh every affected Child DOX Index
4. Remove stale or contradictory text
5. Verify no temporal task checkboxes (e.g. `[ ]` or `[x]`) or diary entries were added to any AGENTS.md file
6. Run existing verification when relevant
7. Report any docs intentionally left unchanged and why

## User Preferences

When the user requests a durable behavior change, record it here or in the relevant child AGENTS.md

## Child DOX Index

- [cmd](cmd) - Main application entry points and HTTP handlers.
- [configs](configs) - Configuration structures and environment files.
- [deploy](deploy) - Docker Compose configurations, staging twins, and Prometheus/Grafana files.
- [docs](docs) - API specifications and developer onboarding SOPs.
- [internal](internal) - Core domain logic, ports, adapters, and services.
- [scripts](scripts) - Backup, metric compilation, and deployment helper scripts.
- [.agent](.agent) - Harness session storage, Project Hub, and project config (HARNESS_GUIDE.md, project-config.env, Project-Hub.md). Skills live in `.pi/skills/`.
- [.pi](.pi) - Pi-native harness: project-local skills (Agent Skills standard) and settings (preferred for new work).
- [global-skills](global-skills) - Project-agnostic engineering methodologies library.
