---
description: Generate a structured session handoff for the next AI agent
---

When the user wants to end a session, run these steps:

1. **Review**: Look back at the conversation and the `git log` of today's work.
2. **Sync**: Ensure all work is pushed to both remotes using the `/sync` workflow first.
3. **Generate**: Create or update `.agent/last_session.md` using the standard handoff template.
4. **Summary**: Provide the user with a brief verbal summary of the handoff.
