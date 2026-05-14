# Rule: Context Compaction

**Scope**: Multi-tasking sessions.
**Goal**: Prevent AI brain-fog and context drift.

1. **Monitor Depth**: When the conversation context grows dense (e.g., after completing a major task or diagnosing a long bug), prepare for compaction.
2. **Proactive Interrupts**: You are authorized and encouraged to actively suggest wrapping up. Ask the user if they'd like to run `/handoff` to clear active token buffers.
3. **No Unprompted Resets**: Do not run workflows entirely autonomously without permission, but repeatedly push the user toward it when the context gets heavy.
