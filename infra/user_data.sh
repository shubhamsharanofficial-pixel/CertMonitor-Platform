#!/bin/bash

# Log output for debugging
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1

echo "Starting CertMonitor Setup..."

# 1. Install Docker & Dependencies
apt-get update
apt-get install -y ca-certificates curl git

install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 2. Setup Swap (2GB) - Critical for 1GB RAM servers
fallocate -l 2G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab

# 3. Clone Repository
echo "Cloning Repository..."
if git clone https://github.com/shubhamsharanofficial-pixel/CertMonitor-Platform.git /home/ubuntu/app; then
    echo "✅ Git clone successful."
else
    echo "❌ CRITICAL ERROR: Git clone failed. Check repository URL or visibility."
    # Exit the script so we don't try to start a non-existent app
    exit 1
fi

# 4. Generate .env file from Terraform variables
cat <<EOF > /home/ubuntu/app/.env
DB_PASSWORD=${db_password}
JWT_SECRET=${jwt_secret}
SMTP_HOST=${smtp_host}
SMTP_PORT=${smtp_port}
SMTP_USER=${smtp_user}
SMTP_PASS=${smtp_pass}
SMTP_SENDER=${smtp_sender}
EOF

# 5. Start Application
cd /home/ubuntu/app
docker compose up -d --build

echo "Setup Complete!"