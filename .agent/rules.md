# Massage Bot Project Rules

## 🌍 Environment & Networking
- **Home Server IP**: `192.168.1.102`
- **SSH Port**: `2222`
- **Username**: `kirill`
- **Root Path**: `/opt/vera-bot`
- **Caddy Network**: `caddy-test-net` (External)

## 🔄 Git & CI/CD Workflow
- **Master Branch**: Use `master` as the primary branch.
- **GitHub**: Primary remote for development (`github` or `origin`).
- **GitLab**: Remote for CI/CD and Registry (`gitlab`).
- **Automation**: GitHub mirrors to GitLab -> GitLab CI builds and deploys.
- **Rule**: Whenever pushing code, **ALWAYS** push to both `github` and `gitlab` remotes to ensure the mirror stays in sync.
- **Rule**: Never use `git push --force` on `master` unless specifically instructed to perform a project-wide reset.

## 🏗 Docker & Deployment
- **Method**: Use `docker compose` with `docker-compose.override.yml`.
- **Image**: Use `registry.gitlab.com/kfilin/massage-bot:latest` for production (Home Server).
- **Deployment Script**: Always use `scripts/deploy_home_server.sh` on the server to ensure consistent pull-based updates.

## 📝 Coding Standards
- **Health Checks**: The bot must expose a health server on port `8081`.
- **Logs**: Bot logs should be accessible via `docker compose logs`.
- **Sessions**: Use PostgreSQL for persistent session storage (`storage.PostgresSessionStorage`).
