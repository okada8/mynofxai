#!/bin/bash

# Configuration
DB_HOST="127.0.0.1"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="m2DvxgvMzZUn1Bx91GaJZ"
DB_NAME="nofx"
BACKUP_DIR="/Users/tom/Desktop/nofi/nofx/data"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/nofx_backup_$DATE.sql"
TELEGRAM_BOT_TOKEN="8359421812:AAFiXHmspm9bhu6RUBm1c7CsE5Glh7e1UGs"
TELEGRAM_CHAT_ID="1654754281"

# Function to send Telegram message
send_telegram() {
    local message="$1"
    curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage" \
        -d chat_id="$TELEGRAM_CHAT_ID" \
        -d text="$message" > /dev/null
}

# Ensure backup directory exists
if [ ! -d "$BACKUP_DIR" ]; then
    mkdir -p "$BACKUP_DIR"
fi

# Export password for pg_dump
export PGPASSWORD="$DB_PASSWORD"

# Check if pg_dump is available
if ! command -v pg_dump &> /dev/null; then
    echo "pg_dump could not be found"
    send_telegram "❌ Database backup failed: pg_dump command not found"
    exit 1
fi

# Perform backup
if pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" > "$BACKUP_FILE"; then
    echo "Backup successful: $BACKUP_FILE"
    
    # Delete backups older than 7 days
    find "$BACKUP_DIR" -name "nofx_backup_*.sql" -type f -mtime +7 -delete
    
    # Send success notification
    send_telegram "✅ Database backup successful! File: $(basename "$BACKUP_FILE")"
else
    echo "Backup failed!"
    # Send failure notification
    send_telegram "❌ Database backup failed!"
    exit 1
fi

# Unset password
unset PGPASSWORD
