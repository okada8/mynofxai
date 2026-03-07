#!/bin/bash

# Configuration
DB_HOST="127.0.0.1"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="m2DvxgvMzZUn1Bx91GaJZ"
DB_NAME="nofx"
TELEGRAM_BOT_TOKEN="8359421812:AAFiXHmspm9bhu6RUBm1c7CsE5Glh7e1UGs"
TELEGRAM_CHAT_ID="1654754281"

# Function to send Telegram message
send_telegram() {
    local message="$1"
    curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage" \
        -d chat_id="$TELEGRAM_CHAT_ID" \
        -d text="$message" > /dev/null
}

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "psql could not be found"
    send_telegram "❌ Audit logs cleanup failed: psql command not found"
    exit 1
fi

# Calculate deletion cutoff date for logging (macOS syntax)
CUTOFF_DATE=$(date -v-15d +%Y-%m-%d)

echo "Cleaning up audit logs older than 15 days (before $CUTOFF_DATE)..."

# Execute cleanup
if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '15 days';"; then
    echo "Cleanup successful"
    send_telegram "✅ Audit logs cleanup successful! Deleted records older than 15 days (before $CUTOFF_DATE)."
else
    echo "Cleanup failed!"
    send_telegram "❌ Audit logs cleanup failed!"
    exit 1
fi

# Unset password
unset PGPASSWORD
