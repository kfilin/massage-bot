---
description: Perform a technical "Memory Flush" to save context before it fades
---

// turbo-all
When the user sends `/checkpoint`, follow these steps:

1. **Summarize State**: Analyze the current chat history to identify all decisions made and code changes implemented.
2. **Commit Status**: Check the current git status and branch.
3. **Log Progress**: Update `.agent/last_session.md` (or a dedicated `checkpoint_log.md`) with a technical summary.
4. **Clean Slate**: Provide a "Compressed Context" message to the user—a single-paragraph summary of "Where we are" and "What's next" that effectively "reboots" the context.
5. **Verify**: Ensure all environment variables or temporary secrets discussed are noted.
