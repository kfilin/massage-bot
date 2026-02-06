---
description: The Twin-Environment Release Cycle
---
# 🚀 CI/CD Pipeline & Deployment Guide

This document describes the automated deployment pipeline and the workflows to move code from local development to **Production**.

## 🔄 Deployment Overview

The project uses a **Git-Ops** workflow:

1. **Code Source**: GitHub (`kfilin/massage-bot`)
2. **Deployment Engine**: GitLab CI/CD (triggered via Mirroring)
3. **Target**: Home Server (`/opt/vera-bot`)

---

## 🔄 The "Twin" Concept

* **Local PC**: The *Editor*. Where code is written. (Needs to stay synced with GitHub).
* **Test Container (`massage-bot-test`)**: The *Lab*. Where code is run and verified against real-world networking (Caddy/TWA).
* **Production (`massage-bot`)**: The *Clinic*. Where the proven code serves real patients.

---

## 1. 🛠️ Develop (Local)

1. **Sync First**: Always ensure your local code is up to date before starting.

    ```bash
    git pull origin master
    ```

2. **Edit & Commit**: Make your changes locally.

    ```bash
    git add .
    git commit -m "feat: my new feature"
    ```

3. **Push to Source of Truth**:

    ```bash
    git push origin master
    ```

## 2. 🧪 Verify (Staging)

Deployed to `vera-bot-test.kfilin.icu`.

1. **Trigger Deployment**:
    Run the script from your local machine. It connects to the server and pulls `origin/master`.

    ```bash
    ./scripts/deploy_test_server.sh
    ```

2. **Test**: Open the Test TWA or interact with the Test Bot.
3. **Debug**: If it fails, fix locally, commit, push, and re-deploy.

## 3. 🚢 Promote (Production)

Deployed to `vera-bot.kfilin.icu`.

Once the feature is verified on Test, propagate it to Production.

### A. The "GitLab Trigger" (Automatic)

Since GitLab controls the Production Pipeline, it must receive the new code.
**We have a GitHub Action (`.github/workflows/mirror.yml`) that automatically mirrors `master` to GitLab.**

1. Push to GitHub:

   ```bash
   git push origin master
   ```

2. **Wait**: The GitHub Action will detect the push, sync to GitLab, and trigger the GitLab CI/CD pipeline.
   * **Result**: Builds Docker image -> Pushes to Registry -> Auto-deploys to Prod.

### B. Manual Override (Emergency)

If GitLab is down or slow, you can manually deploy `origin/master` to Prod.

```bash
./scripts/deploy_home_server.sh
```

---

## 🛑 Critical Rules

1. **Immutable Server**: NEVER edit files inside `/opt/vera-bot` or `/opt/vera-bot-test` directly.
    * *Why?* The deployment scripts run `git reset --hard`. Your changes will be deleted.
2. **Test First**: Always verify on `vera-bot-test` before pushing to `gitlab`.
