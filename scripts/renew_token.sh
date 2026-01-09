#!/bin/bash
# Google OAuth Token Renewal Helper Script
# Usage: ./scripts/renew_token.sh

set -e

echo "🔐 Vera Massage Bot - Google Token Renewal Helper"
echo "=================================================="
echo ""
echo "This script helps you renew the Google OAuth token for calendar integration."
echo "Tokens expire every ~6 months. Last renewal: 2026-01-09"
echo ""

# Your credentials (update these with your actual values)
CLIENT_ID="451987724111-3p6vs3dvk96sh42gaeuplcp26t9t3998.apps.googleusercontent.com"
CLIENT_SECRET="GOCSPX-YOUR_CLIENT_SECRET_HERE"

echo "📋 Step 1: Generate Authorization URL"
echo "-------------------------------------"
echo ""
echo "Open this URL in your browser:"
echo ""
echo "https://accounts.google.com/o/oauth2/auth?client_id=${CLIENT_ID}&redirect_uri=http://localhost&response_type=code&scope=https://www.googleapis.com/auth/calendar&access_type=offline&prompt=consent"
echo ""
echo "📋 Step 2: Get the Authorization Code"
echo "--------------------------------------"
echo "1. Log in with your Google account"
echo "2. Accept the calendar permissions"
echo "3. You'll see 'localhost refused to connect' error"
echo "4. Copy the code from the URL (looks like: 4/0AThS5b3jYlC...)"
echo ""
read -p "📝 Paste the authorization code here: " AUTH_CODE

if [ -z "$AUTH_CODE" ]; then
    echo "❌ No code provided. Exiting."
    exit 1
fi

echo ""
echo "📋 Step 3: Exchange Code for Token"
echo "-----------------------------------"
echo "Making request to Google OAuth server..."
echo ""

# Make the request
RESPONSE=$(curl -s -d "client_id=${CLIENT_ID}&client_secret=${CLIENT_SECRET}&code=${AUTH_CODE}&redirect_uri=http://localhost&grant_type=authorization_code" https://oauth2.googleapis.com/token)

if echo "$RESPONSE" | grep -q "error"; then
    echo "❌ Error from Google:"
    echo "$RESPONSE"
    exit 1
fi

echo "✅ Success! New token obtained:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
echo ""
echo "📋 Step 4: Update Production Server"
echo "------------------------------------"
echo "1. Copy the entire JSON response above"
echo "2. SSH to production server: ssh kirill@debian-server"
echo "3. Update /opt/vera-bot/.env:"
echo "   sed -i \"s/GOOGLE_TOKEN_JSON=.*/GOOGLE_TOKEN_JSON='PASTE_JSON_HERE'/\" .env"
echo "4. Restart: docker-compose restart"
echo "5. Test the bot"
echo ""
echo "🎉 Renewal complete! Next renewal due: ~2026-07-09"
