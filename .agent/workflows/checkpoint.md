---
description: Formal session checkpoint and developer handoff
---
# /checkpoint

Run this workflow whenever the user requests a checkpoint or at the end of a session to consolidate progress.

1. **Verify State**:
   - Ensure the current master branch is stable.
   - Run `git rev-parse HEAD` to get the latest commit hash.

2. **Update Project Hub**:
   - Open `.agent/Project-Hub.md`.
   - Update the **Version** if a new milestone was reached.
   - Update the **Gold Standard Checkpoint** section with the current hash, date, and a "v4.x.x stable" summary.

3. **Update Changelog**:
   - Open `CHANGELOG.md`.
   - Add a new `[vX.X.X] - YYYY-MM-DD` section.
   - Categorize changes into `### Added`, `### Changed`, and `### Fixed`.

4. **Prepare Handoff & Last Session**:
   - Open `.agent/last_session.md` and `.agent/handoff.md`.
   - Summarize "Accomplishments" (copy bits from the changelog).
   - List outstanding "Technical Debt" and "Next Steps".

5. **Commit & Push**:
// turbo
   - Run `git add -f .agent/*.md CHANGELOG.md && git commit -m "docs: establish vX.X.X stable checkpoint" && git push origin master && git push gitlab master`.
