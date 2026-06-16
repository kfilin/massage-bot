---
name: hydrate-harness
description: Hydrate the universal project harness by replacing placeholders with actual values from project-config.env. Use when user runs /hydrate.
category: System
priority: 2
---

# Skill: Hydrate Harness

## Objective
Convert the universal project harness template into a project-specific brain. It reads `.agent/project-config.env` and hardcodes the variables across the `.agent/` and `global-skills/` directories so you don't have to juggle them in your context window.

## When to Execute
- The user runs `/hydrate`.
- During the `startup.md` Step 0 check if the user asks to proceed with hydration.

## Workflow

### Step 1 — Verify Config
Ensure `.agent/project-config.env` is filled out and `HYDRATED` is currently `false`.

### Step 2 — Hydration Script
Use the `// turbo` directive to execute the hydration script. This script will replace all instances of the literal string `\Antigravity_on_steroids` with the real project name.

// turbo
```bash
source .agent/project-config.env || true

# Replace PROJECT_NAME placeholder
find .agent global-skills -type f -name "*.md" -exec sed -i "s/\Antigravity_on_steroids/$PROJECT_NAME/g" {} +
find .agent global-skills -type f -name "*.md" -exec sed -i "s/\${PROJECT_NAME}/$PROJECT_NAME/g" {} +

# Replace GIT_MAIN_BRANCH placeholder
find .agent global-skills -type f -name "*.md" -exec sed -i "s/\${GIT_MAIN_BRANCH}/$GIT_MAIN_BRANCH/g" {} +

# Set HYDRATED=true
sed -i "s/HYDRATED=false/HYDRATED=true/g" .agent/project-config.env

# Commit the hydration locally
git add .agent/ global-skills/
git commit -m "chore: hydrate universal project harness for $PROJECT_NAME" || true
```

### Step 3 — Confirm
Print a confirmation message that the harness has been successfully hydrated and is ready to use for development.
