# **CertMonitor Platform**

**A self-hosted, agent-based SSL/TLS certificate monitoring solution.**

CertMonitor provides a centralized dashboard to track certificate expiry, verify trust chains, and monitor infrastructure health. Unlike external scanners, it uses a lightweight agent to discover internal certificates (files & ports) securely without opening firewall ports.

## **🚀 Features**

* **Agent-Based Discovery:** Scans /etc/ssl, /var/www, and internal ports (e.g., localhost:8443) for certificates.  
* **Ghost Detection:** Automatically flags certificates that have disappeared from a server ("Soft Delete").  
* **Smart Alerting:** Sends deduplicated email alerts via SMTP (Brevo/SendGrid) for expiring certs.  
* **Live Dashboard:** Real-time inventory with 30-second polling updates.  
* **Secure Auth:** JWT-based authentication with secure API Key management (hashes stored only).  
* **Infrastructure as Code:** Fully automated AWS deployment via Terraform.

## **🏗 Architecture**

CertMonitor runs as a **4-Container Docker Cluster** orchestrated via Docker Compose:

1. **Nginx Proxy Manager (Gateway):**  
   * The only entry point (Ports 80/443).  
   * Handles **SSL Termination** (Let's Encrypt) and auto-renewal.  
   * Proxies traffic to the Frontend.  
2. **Frontend (Nginx \+ React):**  
   * Serves the UI and acts as an internal Reverse Proxy for API requests.  
3. **Backend (Go):**  
   * REST API listening internally. Handles ingestion, auth, and background workers.  
   * **Stateless:** Pulls pre-built images from GitHub Container Registry (GHCR).  
4. **Database (PostgreSQL):**  
   * Persistent storage for users and certificate data.

## **☁️ Deployment Option A: Automated (Terraform)**

**Recommended.** Deploys a production-ready server on **AWS Free Tier (t3.micro)** in minutes.

### **Prerequisites**

* AWS CLI configured.  
* Terraform installed (brew install terraform).

### **1\. Setup Secrets**

Navigate to the infrastructure folder and create your secrets file:

cd infra  
nano secrets.tfvars

**Paste and fill:**

db\_password  \= "strong\_db\_password"  
jwt\_secret   \= "long\_random\_string"  
smtp\_user    \= "your\_brevo\_email"  
smtp\_pass    \= "your\_brevo\_smtp\_key"  
smtp\_sender  \= "alerts@yourdomain.com"

### **2\. Deploy**

terraform init  
terraform apply \-var-file="secrets.tfvars"

**What this does:**

1. Provisions an EC2 Instance (Ubuntu 24.04).  
2. Allocates and attaches a Static IP (Elastic IP).  
3. Configures Security Groups (Firewall).  
4. **Auto-Bootstraps:** Installs Docker, pulls images from GHCR, and starts the stack.

## **🐳 Deployment Option B: Manual (Any VPS)**

Use this for DigitalOcean, Hetzner, or local testing.

### **1\. Clone & Config**

git clone \[https://github.com/shubhamsharanofficial-pixel/CertMonitor-Platform.git\](https://github.com/shubhamsharanofficial-pixel/CertMonitor-Platform.git)  
cd cert-monitor-platform

Create a .env file in the root:

DB\_PASSWORD=secure\_pass  
JWT\_SECRET=random\_string  
SMTP\_HOST=smtp-relay.brevo.com  
SMTP\_PORT=587  
SMTP\_USER=brevo\_email  
SMTP\_PASS=brevo\_key  
SMTP\_SENDER=alerts@domain.com  
FRONTEND\_URL=http://YOUR\_SERVER\_IP

### **2\. Run (Image Based)**

This pulls pre-compiled images from the registry. No build tools required.

\# Rename the production compose file  
cp docker-compose.prod.yml docker-compose.yml

\# Start  
docker compose up \-d

## **🔒 Post-Deployment: Enable SSL**

Once your server is running (Option A or B):

1. **Point your Domain:** Create an A-Record (DNS) pointing your-domain.com to your Server IP.  
2. **Access Admin Panel:** Open http://YOUR\_IP:81.  
   * Default: admin@example.com / changeme  
3. **Create Proxy Host:**  
   * Domain: your-domain.com  
   * Forward Hostname: cert\_frontend  
   * Forward Port: 80  
   * **SSL Tab:** Request a new Let's Encrypt Certificate.

## **📦 Agent Installation**

### **Option A: Auto-Install Script (Recommended)**

1. Log in to the Dashboard.  
2. Click **"Generate API Key"** (or Rotate Key) in the top navigation bar.  
3. Copy the provided curl or wget command.  
4. Run it on your server. It automatically downloads the binary, configures it with your key, and sets up a systemd service.

### **Option B: Manual Binary**

1. Log in to your new Dashboard and go to the **Downloads** page.  
2. Download the binary for your OS (Linux/macOS).  
3. Download the config.yaml template and add your API Key.  
4. Run it:  
   ./agent-linux-amd64 \-config config.yaml

### **Option C: Docker (Containerized)**

1. Download docker-compose.agent.yml and config.yaml from the **Downloads** page.  
2. Edit config.yaml:  
   * Add your API Key.  
   * **Important:** Update paths to match container mounts (e.g., set state\_path: "/app/data").  
3. Run the container:  
   docker compose \-f docker-compose.agent.yml up \-d

## **🔧 Development (Local)**

To contribute to the code:

**Backend:**

cd backend  
go run cmd/server/main.go

**Frontend:**

cd frontend  
npm install  
npm run dev  
