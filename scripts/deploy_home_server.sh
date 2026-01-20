#!/bin/bash

# Configuration
APP_DIR="/opt/vera-bot"
SERVICE_NAME="massage-bot"

echo "🚀 Starting deployment on Home Server..."

# 1. Pull latest changes
echo "📥 Pulling latest code from master..."
cd $APP_DIR || exit
git fetch origin master
git reset --hard origin/master

# 2. Build and restart containers
echo "🛠 Building and recreating containers..."
# Using --build to ensure local source changes are compiled
docker compose up -d --build --force-recreate

# 3. Cleanup unused images
echo "🧹 Cleaning up old images..."
docker image prune -f

# 4. Check status
echo "📊 Deployment Status:"
docker compose ps

echo "📝 Recent Logs:"
docker compose logs --tail=20 $SERVICE_NAME

echo "✅ Deployment complete!"
