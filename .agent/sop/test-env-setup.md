---
description: How to maintain the parallel Test Environment on the Home Server
---

# Test Environment Workflow (Twin Strategy)

We run a "Twin" environment on the Home Server to test changes before going live. We use a **Separate Folder** strategy for safety.

## 1. Architecture

| Feature | **Production** (`/opt/vera-bot`) | **Test** (`/opt/vera-bot-test`) |
| :--- | :--- | :--- |
| **Domain** | `app.massagemni.com` | `vera-bot-test.kfilin.icu` |
| **Ports** | `8082`, `8083` | `9082`, `9083` |
| **Database** | `massage-bot-db` | `massage-bot-db-test` |
| **Command** | `docker compose up -d` | `docker compose up -d` |

## 2. Setup (One-Time Migration)

1. **Clone to Test Folder**:

    ```bash
    cd /opt
    git clone https://github.com/kfilin/massage-bot.git vera-bot-test
    cd vera-bot-test
    ```

2. **Configure Secrets**:

    ```bash
    cp .env.example .env
    nano .env
    # 1. Set TG_BOT_TOKEN to your Test Token
    # 2. Set WEBAPP_URL to vera-bot-test.kfilin.icu
    # 3. Set DB_NAME=massage_bot_test
    ```

3. **Apply Port Overrides**:
    Copy the override file to valid `docker-compose.override.yml`:

    ```bash
    cp deploy/docker-compose.test-override.yml docker-compose.override.yml
    ```

    *Docker automatically merges `docker-compose.yml` + `docker-compose.override.yml`.*

4. **Start Test Env**:

    ```bash
    docker compose up -d
    ```

## 3. Usage

Since we use separate folders, standard commands work safely:

* **Logs**: `docker compose logs -f`
* **Stop**: `docker compose down` (Only affects Test!)
* **Update**: `git pull && docker compose build --no-cache && docker compose up -d`
