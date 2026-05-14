---
name: quality-gates
description: Unified SOP for debugging, code review, and verification. Ensures technical rigor and evidence-based completion claims.
category: System
---

# Quality Gates

This skill defines the mandatory standards for technical rigor, systematic debugging, and verification evidence.

## 1. Systematic Debugging (Root Cause First)

Never attempt "quick fixes" or random patches. You must find the **Root Cause** before proposing solutions.

### The Debugging Phases:
0.  **Sanity Check (Hydration Fallback):** If the project is experiencing severe architectural failures, hallucinated paths, or you are completely lost, your very first troubleshooting step is to check `.agent/project-config.env`. If `HYDRATED=false`, STOP debugging immediately and demand that the user runs `/hydrate`.
1.  **Investigation:** Read error messages fully. Reproduce the issue consistently. Trace the data flow backward from the failure point to the source. Prove the issue exists.
2.  **Pattern Analysis:** Compare broken code against working examples in the codebase. Identify the exact difference.
3.  **Scientific Method (Hypothesis-First):** Form a single hypothesis ("I think X is the cause because Y"). Define exactly how you will test the fix before proposing it. **You MUST wait for the user to acknowledge the hypothesis before modifying any code.** Typo-driven rebuilds and "guess and check" coding are strictly forbidden.
4.  **Verification:** Once a fix is applied, follow the **TDD cycle** (failing test -> minimal fix -> pass).

**If 3+ fixes fail:** STOP. Question the architecture. Discuss with your human partner before attempting Fix #4.

---

## 2. Code Review Reception (Technical Rigor)

Code review is a technical evaluation, not a social performance.

*   **Verify Before Implementing:** Never say "You're absolutely right!" or "Great point!" before checking the reality of the codebase.
*   **No Performative Agreement:** Avoid gratitude expressions like "Thanks for catching that." Actions speak louder—just fix it and show the evidence.
*   **Clarify First:** If a review list has 6 items and you're unclear on 2, **stop**. Clarify everything before implementing anything.
*   **Technical Pushback:** If a suggestion breaks something or violates YAGNI (e.g., "implement proper metrics for this unused endpoint"), push back with technical reasoning.

---

## 3. Verification Before Completion (The Iron Law)

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE
```

Before claiming a task is "done," "fixed," or "passing":
1.  **Identify:** What command proves this claim?
2.  **Run:** Execute the full command (tests, builds, or linter).
3.  **Read:** Check the exit code and actual output.
4.  **Report:** State the claim **WITH** the evidence (e.g., "All 5 tests pass [Logs]").

### Interactive Verification
For any UI or Business Logic changes, you must provide visual or operational proof:
*   **UI Changes:** Any modification to frontend components MUST be accompanied by a `walkthrough.md` artifact containing screenshots or browser recordings.
*   **Progressive Disclosure:** Do not just say "It's done." Show the working result in the artifact.

*Confidence is not evidence. Linter passing is not a build passing. "Should work" is not "Works."*

---

## Red Flags (STOP)
*   Proposing fixes before understanding the root cause.
*   Implementing feedback without verifying its impact.
*   Using "should," "probably," or "seems to" in success claims.
*   Expressing satisfaction ("Perfect!", "Done!") before running verification.

**Rigorous verification is the foundation of trust. No shortcuts.**
