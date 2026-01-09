# 🔐 Google OAuth Token Renewal Guide

## 📅 Renewal History
- **Last Renewal**: 2026-01-09
- **Next Due**: ~2026-07-09
- **Frequency**: Every ~6 months

## 🚨 Symptoms of Expired Token
- Bot shows time slots but fails to create calendar events
- "invalid_grant" errors in logs
- Calendar integration silent failure
- Error: "Token has been expired or revoked"

## 🛠️ Renewal Procedure

### Step 1: Generate Authorization URL
```bash
# Use your actual client ID from Google Cloud Console
CLIENT_ID="451987724111-3p6vs3dvk96sh42gaeuplcp26t9t3998.apps.googleusercontent.com"

echo "Open this URL in your browser:"
echo "https://accounts.google.com/o/oauth2/auth?client_id=${CLIENT_ID}&redirect_uri=http://localhost&response_type=code&scope=https://www.googleapis.com/auth/calendar&access_type=offline&prompt=consent"

### Step 2: Get Authorization Code

1. Open the URL in browser
    
2. Log in with the Google account that owns the calendar
    
3. Accept the calendar permissions
    
4. You'll be redirected to an error page (localhost can't be reached)
    
5. Copy the `code` parameter from the URL bar:
    
    text
    

Example: http://localhost/?code=4/0AThS5b3jYlC...&scope=https://www.googleapis.com/auth/calendar
Copy: 4/0AThS5b3jYlC...

### Step 3: Exchange Code for Token

bash

# Replace with your actual values
CLIENT_ID="451987724111-3p6vs3dvk96sh42gaeuplcp26t9t3998.apps.googleusercontent.com"
CLIENT_SECRET="GOCSPX-YOUR_CLIENT_SECRET_HERE"
CODE="PASTE_THE_CODE_FROM_STEP_2"

curl -d "client_id=${CLIENT_ID}&client_secret=${CLIENT_SECRET}&code=${CODE}&redirect_uri=http://localhost&grant_type=authorization_code" https://oauth2.googleapis.com/token

### Step 4: Update Production Configuration

The response will be JSON like:

json

{
  "access_token": "ya29.a0...",
  "expires_in": 3599,
  "refresh_token": "1//09...",
  "scope": "https://www.googleapis.com/auth/calendar",
  "token_type": "Bearer"
}

Update `/opt/vera-bot/.env` on the production server:

bash

# SSH to production server
ssh kirill@debian-server

# Update the token (use single quotes to preserve JSON)
cd /opt/vera-bot
cp .env .env.backup
sed -i "s/GOOGLE_TOKEN_JSON=.*/GOOGLE_TOKEN_JSON='PASTE_FULL_JSON_HERE'/" .env

# Restart the container
docker-compose restart

### Step 5: Verify Fix

1. Test the bot: `/start` → Book appointment
    
2. Check logs: `docker-compose logs --tail=20`
    
3. Verify Google Calendar has new event
    
4. Confirm "Google Calendar event created" appears in logs
    

## 📋 Quick Renewal Script

bash

#!/bin/bash
# renew_token.sh - Google OAuth Token Renewal Helper

CLIENT_ID="451987724111-3p6vs3dvk96sh42gaeuplcp26t9t3998.apps.googleusercontent.com"
CLIENT_SECRET="GOCSPX-YOUR_CLIENT_SECRET_HERE"

echo "🔐 Google OAuth Token Renewal"
echo "=============================="
echo ""
echo "1. Generate authorization URL:"
echo "https://accounts.google.com/o/oauth2/auth?client_id=${CLIENT_ID}&redirect_uri=http://localhost&response_type=code&scope=https://www.googleapis.com/auth/calendar&access_type=offline&prompt=consent"
echo ""
echo "2. After getting code, run:"
echo "curl -d 'client_id=${CLIENT_ID}&client_secret=${CLIENT_SECRET}&code=PASTE_CODE&redirect_uri=http://localhost&grant_type=authorization_code' https://oauth2.googleapis.com/token"
echo ""
echo "3. Update /opt/vera-bot/.env with the new token JSON"

## 🎯 Critical Fix Applied: 2026-01-09

**Problem**: Bot was silently failing to create calendar events due to expired OAuth token.

**Root Cause**: Google OAuth refresh token expired after ~6 months of inactivity (original token issued ~Nov 2025).

**Solution**:

1. Generated new OAuth authorization URL with `prompt=consent`
    
2. Obtained fresh authorization code
    
3. Exchanged code for new token valid for ~6 months
    
4. Updated production `.env` file
    
5. Restarted container
    

**Result**: Bot fully operational, creating calendar events successfully.

## ⏰ Reminder Setup

Set a calendar reminder for token renewal:

- Date: ~2026-07-01 (2 weeks before expiry)
    
- Action: Run renewal procedure
    

## 🆘 Troubleshooting

- **"invalid_client"**: Check client_id and client_secret
    
- **"invalid_grant"**: Code expired or already used (get new code)
    
- **"access_denied"**: User didn't grant permissions
    
- No error but no events: Check `GOOGLE_CALENDAR_ID` setting