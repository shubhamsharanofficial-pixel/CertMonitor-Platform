# **CertMonitor Agent**

**The lightweight, cross-platform data collector for the CertMonitor Platform.**

This agent runs on distributed infrastructure (Linux/macOS), scans for SSL/TLS certificates in file systems and network ports, and securely reports metadata to the centralized CertMonitor backend.

## **🚀 Features**

* **File Discovery:** Recursively scans specified directories (e.g., /etc/ssl) for .pem, .crt, and .cer files.  
* **Network Scanning:** Performs TLS handshakes on specified ports (e.g., google.com:443, localhost:8443) to capture leaf certificates.  
* **Container Ready:** Official Docker image available for Kubernetes, ECS, or Docker Compose setups.  
* **Enterprise Friendly:** Supports loading **Private Root CAs** for internal corporate networks.  
* **Secure Identity:** Uses a persistent UUID and API Key authentication.

## **⚙️ Configuration**

The agent requires a config.yaml file to run.

\# \==========================================  
\# CertMonitor Agent Configuration  
\# \==========================================

\# \--- 1\. Backend Connection \---  
\# The URL where the agent sends reports (Nginx Proxy).  
backend\_url: "\[https://monitor.your-domain.com/api/certs\](https://monitor.your-domain.com/api/certs)"

\# Your unique API Key (Generate this in the Dashboard)  
api\_key: "crt\_live\_YOUR\_KEY\_HERE"

\# Optional: Override Agent Name  
\# Default: Uses OS Hostname (or Container ID).  
\# hostname: "web-server-01"

\# \--- 2\. File Scanning \---  
\# Directories to recursively scan for certificate files.  
cert\_paths:  
  \- "/etc/ssl/certs"  
  \- "/etc/pki/tls/certs"

\# \--- 3\. Network Scanning \---  
\# List of "host:port" to perform TLS handshakes with.  
network\_scans:  
  \- "google.com:443"  
  \- "localhost:8443"

\# \--- 4\. Private Certificates (Optional) \---  
\# Folder containing extra Root CA files (.pem/.crt) to trust.  
\# Use this if scanning internal apps signed by a Private CA.  
\# extra\_certs\_path: "/app/certs"

\# \--- 5\. Agent State \---  
\# Directory to store the Agent's identity (UUID).  
state\_path: "."

\# \--- 6\. Behavior \---  
\# Scan interval in minutes (Default: 60\)  
scan\_interval\_minutes: 60  
log\_console: true  
log\_path: "./agent.log"

## **🐳 Option A: Running via Docker (Recommended)**

We provide a public image compatible with Linux (AMD64/ARM64) and macOS.

### **1\. Quick Start (Docker Compose)**

Create a docker-compose.yml file:

version: '3.8'

services:  
  cert-agent:  
    image: ghcr.io/shubhamsharanofficial-pixel/cert-agent:latest  
    container\_name: cert-agent  
    restart: unless-stopped  
    network\_mode: host \# Allows scanning localhost ports on the host

    volumes:  
      \# 1\. Config: Map your local config file  
      \- ./config.yaml:/app/config.yaml:ro  
        
      \# 2\. Persistence: Keep Agent ID safe across restarts  
      \- agent\_data:/app/data  
        
      \# 3\. Targets: Map host directories to scan  
      \- /etc/ssl/certs:/app/scan\_target/etc-ssl:ro

      \# 4\. (Optional) Private CAs: Map your custom root certs  
      \# \- ./my-private-ca:/app/certs:ro

    environment:  
      \- TZ=UTC

volumes:  
  agent\_data:

**Note:** If using Docker, update your config.yaml paths to match the container paths (e.g., state\_path: "/app/data").

## **🏃‍♂️ Option B: Running Binary (Manual)**

### **1\. Download**

Download the latest binary from the **Downloads** page of your CertMonitor dashboard, or compile it yourself.

### **2\. Run**

\# Basic run  
./agent-linux-amd64 \-config config.yaml

\# Run with dry-run (Prints report to console, does not send)  
./agent-linux-amd64 \-config config.yaml \-dry-run

## **📦 Building from Source**

To compile the agent for different architectures (Cross-Compilation):

**Linux (AMD64 / Standard Servers):**

GOOS=linux GOARCH=amd64 go build \-o bin/agent-linux-amd64 ./cmd/agent

**Linux (ARM64 / AWS Graviton / Raspberry Pi):**

GOOS=linux GOARCH=arm64 go build \-o bin/agent-linux-arm64 ./cmd/agent

**macOS (Apple Silicon / M1+):**

GOOS=darwin GOARCH=arm64 go build \-o bin/agent-darwin-arm64 ./cmd/agent

## **🤝 Contribution & Development**

### **Project Structure**

* **cmd/agent**: Main entry point. Handles CLI flags and config loading.  
* **pkg/scanner**: Logic for walking file trees and parsing x509 certificates.  
* **pkg/network**: Logic for dialing TCP ports and grabbing TLS chains.  
* **pkg/reporter**: Logic for packaging the JSON payload and POSTing to the backend.

### **Running Tests**

go test ./...  
