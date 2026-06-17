#!/bin/bash
#
# Thin deployment wrapper for the massage-bot stack.
# Replaces the legacy /opt/vera-bot/deploy.sh that used bare `docker run`
# with hard-coded ports. This script uses docker compose with the
# project's actual compose files, runs a pre-flight port collision check,
# and is parameterized for test/prod environments.
#
# Usage: deploy.sh <test|prod>

set -euo pipefail

ENV="${1:-}"

case "$ENV" in
  test)
    APP_DIR="/opt/vera-bot-test"
    SKIP_PORT_CHECK=1  # Test env is allowed to collide with prod
    ;;
  prod)
    APP_DIR="/opt/vera-bot"
    SKIP_PORT_CHECK=0
    ;;
  *)
    echo "ERROR: Unknown env '$ENV'. Usage: $0 <test|prod>"
    exit 1
    ;;
esac

COMPOSE_FILES=(-f docker-compose.yml -f deploy/docker-compose.prod.yml)
SERVICE="app"

echo "🚀 Deploying massage-bot to ${ENV} (${APP_DIR})"

# Pre-flight: refuse to deploy prod if webapp port is already in use.
# (Bots across other repos have historically collided on 8082.)
if [[ "$SKIP_PORT_CHECK" -eq 0 ]]; then
  # Read the target port from .env (with a sane default).
  # grep returns 1 if the key is missing; under `set -e` + `pipefail` that
  # would abort the whole script, so guard with `|| true` here.
  PORT=""
  if [[ -f "${APP_DIR}/.env" ]]; then
    # shellcheck disable=SC1090
    PORT=$(grep -E '^HOST_WEBAPP_PORT=' "${APP_DIR}/.env" | cut -d= -f2 | tr -d '"' | tr -d "'" || true)
  fi
  PORT="${PORT:-8082}"

  # Smart pre-flight: a normal `docker compose up -d --force-recreate` keeps
  # the OLD container bound to the port during the atomic swap, so a naive
  # `ss` check always fires on a healthy prod. Instead, only abort when the
  # port is bound by something OUTSIDE our compose project. This preserves
  # the original P0-incident protection (catching rogue bots squatting the
  # port) while allowing routine deploys.
  #
  # 1. If one of our own containers already exposes PORT, allow.
  #    `{{.Ports}}` can return a single published port, a published range
  #    (e.g. `127.0.0.1:8082-8083->8082-8083/tcp`), an exposed-but-unbound
  #    port (`8082/tcp`), or several comma-separated entries. We only need
  #    to recognise PORT as a *host-side* port in any of these forms.
  set +e
  OUR_BINDING=$(docker ps --filter "label=com.docker.compose.project=vera-bot" \
                       --format '{{.Ports}}' 2>/dev/null | \
                  awk -F'->' '{print $1}' | \
                  grep -E "(^|:|[-])${PORT}([-]|/|\$)")
  set -e
  if [[ -n "$OUR_BINDING" ]]; then
    echo "✅ Port ${PORT} is bound by our own container — proceeding."
  else
    # 2. Port is not ours. Is it bound at all?
    if command -v ss >/dev/null 2>&1 && \
       ss -tlnH 2>/dev/null | awk '{print $4}' | grep -qE "[:.]${PORT}\$"; then
      echo "ERROR: Port ${PORT} is in use by a process outside the vera-bot project."
      echo "       Refusing to deploy. Stop the conflicting container, or change"
      echo "       HOST_WEBAPP_PORT in ${APP_DIR}/.env, or set SKIP_PORT_CHECK=1."
      echo "       Bound port detail:"
      ss -tlnH 2>/dev/null | awk '{print "         ", $4}' | grep -E "[:.]${PORT}\$" || true
      exit 1
    fi
    # 3. Port is free — fine.
    echo "✅ Port ${PORT} is free — proceeding."
  fi
fi

cd "$APP_DIR" || { echo "ERROR: $APP_DIR does not exist"; exit 1; }

echo "📥 Pulling latest code from origin/master..."
git fetch origin master
git reset --hard origin/master

echo "🛠 Building images (no cache) and recreating containers..."
docker compose "${COMPOSE_FILES[@]}" build --no-cache --pull
docker compose "${COMPOSE_FILES[@]}" up -d --force-recreate

echo "📊 Status:"
docker compose "${COMPOSE_FILES[@]}" ps

echo "📝 Recent logs from ${SERVICE}:"
docker compose "${COMPOSE_FILES[@]}" logs --tail=20 "$SERVICE"

echo "✅ Deploy to ${ENV} complete"
