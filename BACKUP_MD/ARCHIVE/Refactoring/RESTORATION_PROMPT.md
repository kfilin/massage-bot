# 🆘 Restoration Procedure Prompt

**Context**:
You are entering a project that has undergone a "Documentation-First" refactoring.
A previous session fully designed and documented a 10-part refactoring roadmap, but due to a sync issue, **the actual code changes for Parts 2 through 10 are missing**.

**Current State**:

- **Codebase**: Reflects the state after **Part 1** (Config Centralization) ONLY.
- **Documentation**: `docs/Refactoring/` contains `Part 1.md` through `Part 10.md`.
- **The Gap**: Parts 2-10 are described in the docs as if they were done (containing detailed "Solution" sections), but the code is NOT there.
- **Status**: Parts 2-10 have been marked as "**Pending Implementation**" in their respective files.

**Your Mission**:
You are to re-implement the missing features one by one, using the existing `Part X.md` files as your **Blueprints**. Do NOT reinvent the architecture; follow the "Solution" details already written in the markdown files to ensure the implementation matches the documentation.

**Step-by-Step Instructions**:

1. **Analyze**: Start by reading `docs/Refactoring/Refactoring Proposals.md` to get the high-level roadmap.
2. **Execute Loop**: For each Part (starting from **Part 2**), perform the following:
    - **Read Blueprint**: Open `docs/Refactoring/Part X.md`.
    - **Verify Gap**: Check the codebase to confirm the feature is indeed missing (e.g., for Part 2, check if secrets are still hardcoded).
    - **Implement**: Write the code *exactly* as described in the "Solution" or "Implementation Plan" section of that markdown file.
    - **Verify**: Run tests to ensure it works.
    - **Update Status**: Change the header in `Part X.md` from "Pending Implementation" back to "Completed".

**Priority Order**:

1. **Part 2**: Eliminate Hard-Coded Secrets.
2. **Part 3**: Raise Test Coverage.
3. **Part 4**: Structured Logging (Crucial for debugging).
4. **Part 5**: Free-Busy Cache (Performance).
5. **Part 6**: CI & Documentation.
6. **Part 7**: Service Refactoring (SlotEngine).
7. **Part 8**: PII Redaction.
8. **Part 9**: Graceful Shutdown.
9. **Part 10**: Final Audit.

**Start Prompt**:
"I am ready to restore the missing refactoring work. I have read the Restoration Procedure. Please confirm if I should start immediately with **Part 2: Eliminate Hard-Coded Secrets**."
