---
description: How to maintain the parallel Test Environment on the Home Server
---

# Test Environment Workflow (Twin Strategy)

We run a "Twin" environment on the Home Server to test changes before going live. Both environments co-exist in the **same folder** (`/opt/vera-bot`) but use different configuration files.

## 1. Architecture

| Feature | **Production (Prod)** | **Test (Twin)** |
| :--- | :--- | :--- |
| **Domain** | `app.massagemni.com` | `vera-bot-test.kfilin.icu` |
| **Bot Token** | Real Vera Bot | Test Bot (from @BotFather) |
| **App Port** | `8082` | `9082` |
| **Health Port** | `8083` | `9083` |
| **Database** | `massage-bot-db` (Prod Data) | `massage-bot-db-test` (Empty/Test Data) |
| **Config** | `.env` | `.env.test` |
| **Compose** | `docker-compose.yml` (Implicit) | `docker-compose.test.yml` |
| **Data Dir** | `data/` | `data_test/` |

## 2. Setup (One-Time)

1. **Switch to User**: `su - kirill` (or your user).
2. **Pull Code**: `cd /opt/vera-bot && git pull origin master`.
3. **Create Secrets**:

    ```bash
    cp .env.test.example .env.test
    nano .env.test # Add TEST_BOT_TOKEN and WEBAPP_URL
    ```

4. **Configure Caddy**:
    Add the contents of `deploy/caddy_test_config.snippet` to `/etc/caddy/Caddyfile` and run `sudo service caddy reload`.

## 3. Usage

### Deploying to Test

To update the test bot with the latest code from `master`:

```bash
./scripts/deploy_test_server.sh
```

### Viewing Logs

```bash
docker compose -f docker-compose.test.yml -p massage-bot-test logs -f --tail=50 massage-bot-test
```

### Stopping Test Environment

```bash
docker compose -f docker-compose.test.yml -p massage-bot-test down
```
