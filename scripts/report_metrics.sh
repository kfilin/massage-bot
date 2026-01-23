#!/bin/bash

# Vera Massage Bot - Metrics Reporter
# Fetches Prometheus metrics and formats them for human readability.

METRICS_URL="${1:-http://localhost:8083/metrics}"

# Fetch metrics
data=$(curl -s "$METRICS_URL")

if [ -z "$data" ]; then
    echo "Error: Could not fetch metrics from $METRICS_URL"
    echo "Make sure the bot is running and the port is accessible."
    exit 1
fi

echo "===================================================="
echo "📊 VERA MASSAGE BOT - BUSINESS INTELLIGENCE SUMMARY"
echo "===================================================="
echo "Report Generated: $(date)"
echo "----------------------------------------------------"

echo ""
echo "📈 BOOKING STATISTICS"
echo "-------------------"
bookings=$(echo "$data" | grep "^vera_bookings_total" | awk -F' ' '{sum+=$2} END {print sum+0}')
returning=$(echo "$data" | grep 'vera_appointment_type_total{type="returning"}' | awk '{print $2+0}')
first_visit=$(echo "$data" | grep 'vera_appointment_type_total{type="first_visit"}' | awk '{print $2+0}')
cancellations=$(echo "$data" | grep "^vera_cancellations_total" | awk -F' ' '{sum+=$2} END {print sum+0}')

printf "Total Completed Bookings:  %d\n" "$bookings"
printf "Returning Patients:        %d\n" "$returning"
printf "First-Time Patients:       %d\n" "$first_visit"
printf "Total Cancellations:       %d\n" "$cancellations"

echo ""
echo "💆 SERVICE POPULARITY"
echo "-------------------"
echo "$data" | grep "^vera_service_bookings_total" | sed -E 's/vera_service_bookings_total\{service_name="([^"]+)"\} (.*)/\1: \2/' | sort -k2 -nr

echo ""
echo "🎙️ CLINICAL ENGAGEMENT"
echo "-------------------"
note_len=$(echo "$data" | grep "^vera_clinical_note_length_chars" | awk '{print $2+0}')
printf "Avg Clinical Note Depth:   %d chars\n" "$note_len"

echo ""
echo "💻 TECHNICAL HEALTH"
echo "-------------------"
active_sessions=$(echo "$data" | grep "^vera_active_sessions" | awk '{print $2+0}')
token_expiry=$(echo "$data" | grep "^vera_token_expiry_days" | awk '{print $2+0}')
db_errors=$(echo "$data" | grep "^vera_db_errors_total" | awk -F' ' '{sum+=$2} END {print sum+0}')

printf "Active User Sessions:      %d\n" "$active_sessions"
printf "OAuth Token Valid for:     %.1f days\n" "$token_expiry"
printf "Total Database Errors:     %d\n" "$db_errors"

echo ""
echo "🌐 API DEPENDENCIES (Requests/Status)"
echo "-------------------"
echo "$data" | grep "^vera_api_requests_total" | sed -E 's/vera_api_requests_total\{operation="([^"]+)",provider="([^"]+)",status="([^"]+)"\} (.*)/\2 [\1] \3: \4/' | sort

echo "----------------------------------------------------"
echo "End of Report"
echo "===================================================="
