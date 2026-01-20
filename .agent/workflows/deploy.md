---
description: Trigger the professional automated CI/CD deployment
---

1. **Commit & Push**: Push all changes to the primary repository.
// turbo
```bash
git push github master && git push gitlab master
```

2. **Monitor GitLab**: Advise the user to check the GitLab Pipeline for the build status.
3. **Verify Server**: Once the pipeline passes, verify the bot status on the server.
// turbo
```bash
ssh -p 2222 kirill@192.168.1.102 "cd /opt/vera-bot && docker compose logs --tail=10 massage-bot"
```
