# Rule: No Server Commits

**Scope**: Global.
**Goal**: Enforce the Mirroring Deployment flow.

1. **Detection**: If the working directory suggests a "Home Server" or "Production" environment (e.g., absolute paths like `/opt/massage-bot/`), do NOT run `git commit`.
2. **Flow**: Instruct the user that changes must be pushed to GitHub (`origin`) and mirrored to the server for deployment.
3. **Sync**: If the server is dirty, use `git reset --hard origin/master` to sync with the source of truth.
