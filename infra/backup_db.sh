#!/bin/bash

# --- Configuration ---
# 1. The SSH Key you created in AWS (Path relative to project root)
KEY_PATH="./cert-monitor-key.pem"

# 2. The User (Always ubuntu for Ubuntu AMIs)
USER="ubuntu"

# 3. The Container Name (From docker-compose.yml)
CONTAINER_NAME="cert_db"

# 4. Database Credentials (Internal defaults)
DB_USER="postgres"
DB_NAME="certdb"

# --- Usage Check ---
if [ -z "$1" ]; then
    echo "Usage: ./backup_db.sh <SERVER_IP_ADDRESS>"
    exit 1
fi

SERVER_IP="$1"
DATE=$(date +%Y-%m-%d_%H-%M-%S)
OUTPUT_FILE="backup_${DATE}.sql"

echo "📡 Connecting to $SERVER_IP..."
echo "📦 Dumping database '$DB_NAME' from container '$CONTAINER_NAME'..."

# --- The Magic Command ---
# 1. SSH into server
# 2. Run 'sudo docker exec' (Sudo is usually required on AWS Ubuntu)
# 3. Run 'pg_dump' to output SQL text
# 4. Pipe that output (>) directly to a file on your Mac
ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no "$USER@$SERVER_IP" \
    "sudo docker exec $CONTAINER_NAME pg_dump -U $DB_USER $DB_NAME" > "$OUTPUT_FILE"

# Check if file has content
if [ -s "$OUTPUT_FILE" ]; then
    echo "✅ Backup successful!"
    echo "📂 Saved to: $(pwd)/$OUTPUT_FILE"
    
    # Check size
    SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
    echo "📊 Size: $SIZE"
else
    echo "❌ Backup failed. File is empty."
    rm "$OUTPUT_FILE"
fi

# **Run the backup:**
# ./backup_db.sh 13.234.xx.xx (ip of ec2)