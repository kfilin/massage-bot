---
description: Synchronize all environments (Local, GitHub, GitLab, Server)
---

To ensure total project consistency across all 4 "pillars", run these steps:

1. **Local Cleanup**: Commit or stash any local changes.
2. **Push Remotes**:
// turbo
```bash
git push github master && git push gitlab master
```

3. **Align Server**:
// turbo
```bash
ssh -p 2222 kirill@192.168.1.102 "cd /opt/vera-bot && git fetch --all && git reset --hard origin/master"
```

4. **Verify**: Check `git log` on both Local and Server to ensure the commit hashes match.
