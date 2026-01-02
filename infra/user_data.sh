#!/bin/bash

# Log output for debugging
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1

echo "Starting CertMonitor Setup..."

# 1. Install Docker & Dependencies
apt-get update
apt-get install -y ca-certificates curl

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

# 3. Prepare App Directory
mkdir -p /home/ubuntu/app
cd /home/ubuntu/app

# 4. Download Production Compose File
# We fetch the raw file from GitHub and save it as standard 'docker-compose.yml'
echo "Downloading Deployment Config..."
curl -o docker-compose.yml https://raw.githubusercontent.com/shubhamsharanofficial-pixel/CertMonitor-Platform/main/docker-compose.prod.yml

# 5. Generate .env file
# We use the IP passed from Terraform directly
cat <<EOF > /home/ubuntu/app/.env
DB_PASSWORD=${db_password}
JWT_SECRET=${jwt_secret}
SMTP_HOST=${smtp_host}
SMTP_PORT=${smtp_port}
SMTP_USER=${smtp_user}
SMTP_PASS=${smtp_pass}
SMTP_SENDER=${smtp_sender}
FRONTEND_URL=http://${public_ip}
EOF

# 6. Start Application
echo "Pulling images and starting..."
docker compose up -d

echo "Setup Complete!"