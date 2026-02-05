# Developer Handoff

## Current Status

**Version**: v5.5.0 (Admin TWA Enhancements + HMAC Fix)
**Date**: 2026-02-05

## Critical Context

### 1. Admin TWA (NEW)

- **Features**: List all patients (autoload on empty query) + Override 72h cancellation rule.
- **Bot Integration**: Added `/patients` command for quick access.

### 2. Ops Improvements (FIXED)

- **Invalid Token**: Fixed HMAC whitespace mismatch. Backend and Frontend now trim IDs.
- **Documentation**: `Collaboration-Blueprint.md` consolidated to "Gold Standard". `files.md` updated.

## Next Steps

1. Monitor "Invalid Token" errors in logs (should be zero now).
2. Begin addressing **TWA UI Improvements** in Backlog (IDs 18-26).
3. Consider DB log verbosity reduction.

## Quick Commands

```bash
# Verify
make test && go build ./...

# Bot Admin
/patients   # List recent patients in bot
```
