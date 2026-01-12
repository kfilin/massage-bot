#!/bin/bash

# Configuration
DATA_DIR=${DATA_DIR:-"./data"}
BACKUP_DIR="${DATA_DIR}/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/cron_backup_${TIMESTAMP}.zip"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Check if zip command is available
if ! command -v zip &> /dev/null
then
    echo "Error: 'zip' command not found. Please install it."
    exit 1
fi

# Create zip of the data directory (excluding existing backups)
echo "Starting backup of ${DATA_DIR}..."
zip -r "$BACKUP_FILE" "$DATA_DIR" -x "${DATA_DIR}/backups/*"

if [ $? -eq 0 ]; then
    echo "Backup successfully created at ${BACKUP_FILE}"
    
    # Optional: Keep only last 30 backups
    ls -t "${BACKUP_DIR}"/* | tail -n +31 | xargs rm -f 2>/dev/null
else
    echo "Backup failed!"
    exit 1
fi
