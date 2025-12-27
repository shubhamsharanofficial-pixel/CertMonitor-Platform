# **CertMonitor Platform**

**A self-hosted, agent-based SSL/TLS certificate monitoring solution.**

CertMonitor provides a centralized dashboard to track certificate expiry, verify trust chains, and monitor infrastructure health. Unlike external scanners, it uses a lightweight agent to discover internal certificates (files & ports) securely without opening firewall ports.

## **🚀 Features**

* **Agent-Based Discovery:** Scans /etc/ssl, /var/www, and internal ports (e.g., localhost:8443) for certificates.  
* **Ghost Detection:** Automatically flags certificates that have disappeared from a server ("Soft Delete").  
* **Smart Alerting:** Sends deduplicated email alerts via SMTP (Brevo/SendGrid) for expiring certs.  
* **Live Dashboard:** Real-time inventory with 30-second polling updates.  
* **Secure Auth:** JWT-based authentication with secure API Key management (hashes stored only).  
* **Account Security:** Email verification and password reset flows included.

## **🏗 Architecture & Design**

CertMonitor follows a **Hub-and-Spoke** architecture designed for security and ease of deployment.

### **1\. The Hub (Server Platform)**

The central platform runs as a **3-Container Docker Cluster**:

* **Container A: The Gateway (Frontend)**  
  * **Tech:** Nginx \+ React.  
  * **Role:** The "Front Door." It is the **only** container exposed to the internet (Port 80).  
  * **Function:** Serves the UI and acts as a **Reverse Proxy**, routing /api/... traffic to the backend. This hides the internal topology from the outside world.  
* **Container B: The Brain (Backend)**  
  * **Tech:** Go (Golang).  
  * **Role:** The logic layer. It is isolated in the internal Docker network.  
  * **Function:** Handles data ingestion, authentication (JWT), and background workers (Janitor for cleanup, Alerter for emails).  
* **Container C: The Vault (Database)**  
  * **Tech:** PostgreSQL.  
  * **Role:** Persistent storage.  
  * **Function:** Deeply isolated. Only the Backend can talk to it. Data is persisted to a Docker Volume so it survives container restarts.

### **2\. The Spokes (Agents)**

* **Tech:** Standalone Go binary (agent-linux-amd64, etc.).  
* **Role:** Runs on your remote servers.  
* **Function:** Wakes up periodically (default: 60 mins), scans for certificates, and **pushes** data to the Hub. It requires **no open inbound ports**, making it firewall-friendly.

## **🔄 How It Works (Data Flow)**

1. **Discovery:** The Agent scans local paths and network ports. It compiles a JSON report and POSTs it to domain.com/api/certs.  
2. **Ingestion:** Nginx receives the request and proxies it to the Backend.  
3. **Processing:** The Backend opens a transaction:  
   * **Deduplicates** certificates (storing distinct certs once, linking them to multiple agents).  
   * **Detects Ghosts:** If an agent report is missing a previously known certificate, it is marked as MISSING (Soft Delete).  
4. **Visualization:** You open the Dashboard. The React app polls the API every 30 seconds. You see live status updates (Green/Red) immediately.  
5. **Alerting:** A background worker checks for certificates expiring within 30 days and sends a consolidated email alert via your SMTP provider.

## **⚡️ Quick Start (Production)**

### **Prerequisites**

* Docker & Docker Compose installed.  
* An SMTP provider (e.g., free Brevo account) for emails.

### **1\. Clone & Prepare**

git clone https://github.com/shubhamsharanofficial-pixel/CertMonitor-Platform.git

cd cert-monitor-platform

### **2\. Configure Environment**

Create a .env file in the root directory. You can copy the example below:

\# Create the file  
nano .env

**Paste the following configuration:**

\# \--- Database Secrets \---  
\# Set a strong, unique password for the internal Postgres DB  
DB\_PASSWORD=change\_me\_to\_something\_secure

\# \--- App Secrets \---  
\# Random string used to sign JWT login tokens  
JWT\_SECRET=change\_me\_to\_a\_long\_random\_string

\# \--- SMTP / Email (Brevo Recommended) \---  
\# Required for Alerts, Verification, and Password Resets  
SMTP\_HOST=smtp-relay.brevo.com  
SMTP\_PORT=587  
SMTP\_USER=your\_brevo\_login\_email  
SMTP\_PASS=your\_brevo\_smtp\_key  
SMTP\_SENDER=alerts@yourdomain.com

### **3\. Launch**

Build and start the containers in detached mode:

docker-compose up \-d \--build

### **4\. Access**

Open your browser and visit:  
http://localhost (or your server's IP/Domain).

## **📦 Agent Installation**

Once the platform is running:

1. Log in to the Dashboard.  
2. Click **"Generate API Key"** in the navigation bar.  
3. Copy the **Auto-Install Command** (curl/wget).  
4. Run it on your Linux servers to install the agent service.

*Alternatively, visit the /downloads page to download binaries manually.*

## **🛡 Security Notes**

* **API Keys:** Stored as SHA-256 hashes. If you lose a key, you must regenerate it.  
* **Database:** Not exposed to the public internet (internal network only).  
* **SSL:** The Nginx container listens on Port 80 by default. For production usage over the internet, we highly recommend putting **Cloudflare** or a **Reverse Proxy (Nginx Proxy Manager)** in front to handle HTTPS.

## **🔧 Development**

To run the stack locally for development (with hot-reloading):

**Backend:**

cd backend  
go run cmd/server/main.go

**Frontend:**

cd frontend  
npm install  
npm run dev

*Note: You will need a local Postgres instance running for the backend dev mode.*
