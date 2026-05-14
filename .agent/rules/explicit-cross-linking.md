# Rule: Explicit Cross-Linking

**Scope**: Knowledge Base files, SOPs, and Rules.
**Goal**: Maintain a healthy, navigable Knowledge Base without forcing full-text searching.

1. **Hard Links Over Names**: Never reference a domain concept by name alone if it has a backing file. Instead of "Check the database expert skill", use explicit markdown links: `[Database Expert](../global-skills/database-expert.md)` or relative paths.
2. **Anchor Points**: Ensure that every new SOP, rule, or KI is anchored somewhere in a Hub. Normally, this means linking it inside `Project-Hub.md` or `startup.md` so it is discoverable simply by clicking or file parsing.
3. **No Orphans**: A file that exists but has no incoming markdown links from elsewhere in the system is considered an Orphan and a failure of explicit linking.
