-- 1. Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 2. Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    organization_name TEXT NOT NULL,
    api_key_hash TEXT UNIQUE,
    email_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, 

    -- New Columns for Auth V2
    is_verified BOOLEAN DEFAULT FALSE,
    verification_token_hash TEXT,
    verification_token_expiry TIMESTAMP WITH TIME ZONE,
    reset_token_hash TEXT,
    reset_token_expiry TIMESTAMP WITH TIME ZONE
);

-- 3. Agents Table
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY, 
    user_id UUID REFERENCES users(id),
    hostname TEXT NOT NULL,
    ip_address TEXT,
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Certificates Table
CREATE TABLE IF NOT EXISTS certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    serial_number TEXT NOT NULL,
    issuer_cn TEXT NOT NULL, 
    issuer_org TEXT,
    issuer_ou TEXT,
    subject_cn TEXT NOT NULL,
    subject_org TEXT,
    subject_ou TEXT,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    signature_algo TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (serial_number, issuer_cn, issuer_org, issuer_ou)
);

CREATE INDEX IF NOT EXISTS idx_certs_valid_until ON certificates(valid_until);

-- 5. Certificate Instances
-- REFACTORED: 'file_path' -> 'source_uid' to support Network/Port scans
CREATE TABLE IF NOT EXISTS certificate_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    certificate_id UUID NOT NULL REFERENCES certificates(id) ON DELETE CASCADE,
    
    -- "source_uid" is the unique ID on the Agent side.
    -- For Files: "/etc/ssl/cert.pem"
    -- For Network: "google.com:443"
    source_uid TEXT NOT NULL,
    source_type TEXT DEFAULT 'FILE',        -- 'FILE' or 'NETWORK'

    -- "Ghost Pruning" / Soft Delete Logic
    current_status TEXT DEFAULT 'ACTIVE',   -- 'ACTIVE' or 'MISSING'

    is_trusted BOOLEAN DEFAULT FALSE,
    trust_error TEXT,
    scanned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- The Unique Constraint now uses source_uid
    UNIQUE (agent_id, source_uid)
);

-- 6. Alert History (The Deduplication Log)
CREATE TABLE IF NOT EXISTS alert_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Link to specific Certificate and Agent context
    certificate_id UUID NOT NULL REFERENCES certificates(id) ON DELETE CASCADE,
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE, 
    
    -- Channel Type (e.g., 'EMAIL', 'WHATSAPP')
    alert_type TEXT NOT NULL, 
    
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for "Bulk Check" performance
CREATE INDEX IF NOT EXISTS idx_alert_hist_cert ON alert_history(alert_type, certificate_id, sent_at);
CREATE INDEX IF NOT EXISTS idx_alert_hist_agent ON alert_history(alert_type, agent_id, sent_at);