#!/bin/bash

# Configuration
APP_DIR="/opt/vera-bot"
SERVICE_NAME="app"

echo "🚀 Starting deployment on Home Server..."

# 1. Pull latest changes
echo "📥 Pulling latest code from master..."
cd $APP_DIR || exit
git fetch origin master
git reset --hard origin/master

# 2. Build and restart containers
echo "🛠 Building latest images (No Cache) and recreating containers..."
docker compose -f docker-compose.yml -f deploy/docker-compose.prod.yml build --no-cache --pull
docker compose -f docker-compose.yml -f deploy/docker-compose.prod.yml up -d --force-recreate

# 3. Check status
echo "📊 Deployment Status:"
docker compose -f docker-compose.yml -f deploy/docker-compose.prod.yml ps

echo "📝 Recent Logs:"
docker compose -f docker-compose.yml -f deploy/docker-compose.prod.yml logs --tail=20 $SERVICE_NAME

echo "✅ Deployment complete!"
