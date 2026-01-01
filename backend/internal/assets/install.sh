#!/bin/bash
set -e

# --- Configuration (Injected by Backend) ---
BASE_URL="{{BASE_URL}}"
# -------------------------------------------

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

API_KEY=""
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/cert-agent"
STATE_DIR="/var/lib/cert-agent"
SCAN_PATHS="/etc/ssl,/var/www,/usr/local/share/ca-certificates" 
SERVICE_FILE=""

show_help() {
    echo "Usage: curl ... | sudo bash -s -- -k <KEY>"
    # ...
}

while [[ "$#" -gt 0 ]]; do
    case $1 in
        -k|--key) API_KEY="$2"; shift ;;
        -s|--state-dir) STATE_DIR="$2"; shift ;;
        -c|--config-dir) CONFIG_DIR="$2"; shift ;;
        -p|--paths) SCAN_PATHS="$2"; shift ;;
        -h|--help) show_help; exit 0 ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

if [ -z "$API_KEY" ]; then
    echo -e "${RED}Error: API Key is required (-k)${NC}"; exit 1
fi

# 1. Detect OS & Arch
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)  BINARY_ARCH="amd64" ;;
    aarch64|arm64) BINARY_ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported arch: $ARCH${NC}"; exit 1 ;;
esac

if [ "$OS" = "Linux" ]; then
    BINARY_OS="linux"
elif [ "$OS" = "Darwin" ]; then
    BINARY_OS="darwin"
else
    echo -e "${RED}Error: OS $OS is not supported.${NC}"; exit 1
fi

echo "Detected: $OS ($BINARY_ARCH)"

# 2. Prepare Directories
echo "Creating directories..."
mkdir -p "$CONFIG_DIR"
mkdir -p "$STATE_DIR"
chmod 700 "$STATE_DIR"

# 3. Download Binary
BINARY_NAME="agent-${BINARY_OS}-${BINARY_ARCH}"
DOWNLOAD_URL="${BASE_URL}/downloads/${BINARY_NAME}"
TARGET_BIN="$INSTALL_DIR/cert-agent"

echo "Downloading $BINARY_NAME from $DOWNLOAD_URL..."
if command -v curl >/dev/null 2>&1; then
    curl -f -sL -o "$TARGET_BIN" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$TARGET_BIN" "$DOWNLOAD_URL"
else
    echo -e "${RED}Error: Need curl or wget${NC}"; exit 1
fi
chmod +x "$TARGET_BIN"
echo "Binary saved to: $TARGET_BIN"

# Remove macOS quarantine attribute if present (fixes "killed" or "cannot check for malicious software" errors)
if [ "$OS" = "Darwin" ] && command -v xattr >/dev/null 2>&1; then
    xattr -d com.apple.quarantine "$TARGET_BIN" 2>/dev/null || true
fi

# 4. Generate Config
CONFIG_FILE="$CONFIG_DIR/config.yaml"
echo "Generating config at $CONFIG_FILE..."

# Default log path inside state dir
LOG_FILE="$STATE_DIR/agent.log"

cat > "$CONFIG_FILE" <<EOF
backend_url: "${BASE_URL}/certs"
api_key: "${API_KEY}"
state_path: "${STATE_DIR}"

scan_interval_minutes: 60

log_path: "${LOG_FILE}"
log_console: true

cert_paths:
EOF

IFS=',' read -ra PATH_ADDR <<< "$SCAN_PATHS"
for i in "${PATH_ADDR[@]}"; do
    echo "  - \"$i\"" >> "$CONFIG_FILE"
done

# Append Network Scans configuration
cat >> "$CONFIG_FILE" <<EOF

network_scans:
  - "google.com:443"
EOF

# 5. Service Installation (Linux Only)
if [ "$OS" = "Linux" ] && [ -d "/etc/systemd/system" ] && command -v systemctl >/dev/null 2>&1; then
    echo "Installing system service..."
    SERVICE_FILE="/etc/systemd/system/cert-agent.service"

    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=CertMonitor Agent
After=network.target

[Service]
ExecStart=$TARGET_BIN -config $CONFIG_FILE
WorkingDirectory=$STATE_DIR
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable cert-agent
    systemctl restart cert-agent
    echo -e "${GREEN}==> Installation Complete! Service started.${NC}"
else
    echo -e "${YELLOW}Warning: Automatic service installation not supported on $OS / non-systemd.${NC}"
    echo -e "${GREEN}==> Installation Complete (Binary Only)${NC}"
    echo "To run the agent manually:"
    echo -e "${GREEN}sudo $TARGET_BIN -config $CONFIG_FILE${NC}"
fi

# 6. Log Rotation (Linux Only)
# Check if /etc/logrotate.d exists (Standard on almost all Linux distros)
if [ "$OS" = "Linux" ]; then
    if [ -d "/etc/logrotate.d" ]; then
        echo "Configuring log rotation..."
        cat > "/etc/logrotate.d/cert-agent" <<EOF
${LOG_FILE} {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
}
EOF
        echo "Log rotation config created at: /etc/logrotate.d/cert-agent"
    else
        echo -e "${YELLOW}Warning: /etc/logrotate.d not found. Log rotation not configured.${NC}"
        echo "Please manually manage log file size: ${LOG_FILE}"
    fi
fi

echo ""
echo "================================================"
echo "           INSTALLATION SUMMARY"
echo "================================================"
echo "Binary:    $TARGET_BIN"
echo "Config:    $CONFIG_FILE"
echo "Logs:      $LOG_FILE"
if [ -n "$SERVICE_FILE" ]; then
    echo "Service:   $SERVICE_FILE"
    echo ""
    echo "To edit config:  sudo nano $CONFIG_FILE"
    echo "To restart:      sudo systemctl restart cert-agent"
fi
echo "================================================"