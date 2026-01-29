#!/bin/bash

# Configuration
APP_DIR="/opt/vera-bot-test"
SERVICE_NAME="massage-bot-test"
REPO_URL="https://github.com/kfilin/massage-bot.git" # Fallback if not set
# Use SSH remote if available using git remote get-url origin
REPO_URL=$(git remote get-url origin 2>/dev/null || echo "https://github.com/kfilin/massage-bot.git")

echo "🧪 Starting deployment on TEST Environment..."

# 0. Ensure Directory Exists and Clone if necessary
if [ ! -d "$APP_DIR" ]; then
    echo "📂 Creating test directory at $APP_DIR..."
    sudo mkdir -p "$APP_DIR"
    sudo chown $USER:$USER "$APP_DIR"
    
    echo "📥 Cloning repository..."
    git clone "$REPO_URL" "$APP_DIR"
else
    echo "found existing directory..."
fi

# 1. Pull latest changes
echo "📥 Pulling latest code from master..."
cd $APP_DIR || exit
git fetch origin master
git reset --hard origin/master

# 1.1 Ensure .env exists (Copy from .env.test if missing, or .env.example)
if [ ! -f ".env" ]; then
    echo "⚠️ .env not found! Attempting to create from .env.test..."
    if [ -f ".env.test" ]; then
        cp .env.test .env
        echo "✅ Created .env from .env.test"
    else
        echo "❌ .env.test not found. Please configure .env manually."
        exit 1
    fi
fi

# 2. Build and restart containers using TEST config (Override Strategy)
echo "🛠 Building TEST images and recreating containers..."
# Use standard docker-compose.yml + test override
docker compose -f docker-compose.yml -f deploy/docker-compose.test-override.yml -p massage-bot-test up -d --build --force-recreate

# 3. Check status
echo "📊 Test Environment Status:"
docker compose -f docker-compose.yml -f deploy/docker-compose.test-override.yml -p massage-bot-test ps

echo "📝 Recent Logs (Test):"
docker compose -f docker-compose.yml -f deploy/docker-compose.test-override.yml -p massage-bot-test logs --tail=20 $SERVICE_NAME

echo "✅ Test Deployment complete via Dual Folder Strategy!"
