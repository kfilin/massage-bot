#!/bin/bash

# Configuration
APP_DIR="/opt/vera-bot"
SERVICE_NAME="massage-bot-test"

echo "🧪 Starting deployment on TEST Environment..."

# 1. Pull latest changes
echo "📥 Pulling latest code from master..."
cd $APP_DIR || exit
git fetch origin master
git reset --hard origin/master

# 2. Build and restart containers using TEST config
echo "🛠 Building TEST images and recreating containers..."
# We explicitly specify the project name 'massage-bot-test' to avoid orphan conflicts with Prod
docker compose -f docker-compose.test.yml -p massage-bot-test build --no-cache --pull
docker compose -f docker-compose.test.yml -p massage-bot-test up -d --force-recreate

# 3. Check status
echo "📊 Test Environment Status:"
docker compose -f docker-compose.test.yml -p massage-bot-test ps

echo "📝 Recent Logs (Test):"
docker compose -f docker-compose.test.yml -p massage-bot-test logs --tail=20 $SERVICE_NAME

echo "✅ Test Deployment complete!"
