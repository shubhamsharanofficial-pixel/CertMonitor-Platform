package model

import "time"

// --- CONSTANTS ---

type CertStatus string

const (
	StatusValid            CertStatus = "Valid"
	StatusNotYetValid      CertStatus = "Not Yet Valid"
	StatusExpired          CertStatus = "Expired"
	StatusExpiringSoon     CertStatus = "Expiring Soon"
	StatusExpiringThisWeek CertStatus = "Expiring This Week"
	StatusExpiringTomorrow CertStatus = "Expiring Tomorrow"
	StatusExpiringToday    CertStatus = "Expiring Today"
	StatusUntrusted        CertStatus = "Untrusted"
)

type InstanceStatus string

const (
	InstanceActive  InstanceStatus = "ACTIVE"
	InstanceMissing InstanceStatus = "MISSING"
)

type AgentStatus string

const (
	StatusAgentOnline  AgentStatus = "Online"
	StatusAgentOffline AgentStatus = "Offline"
)

// Notification Policies
// These are shared across all notifiers and the history service.
type AlertType string

const (
	AlertTypeEmail AlertType = "EMAIL"
	// AlertTypeWhatsapp AlertType = "WHATSAPP"
	// AlertTypeSlack    AlertType = "SLACK"
)

// Standard Frequencies for cooldowns
const (
	FrequencyDaily  = 24 * time.Hour
	FrequencyWeekly = 7 * 24 * time.Hour
)

// --- MODELS ---

type AgentReport struct {
	AgentID      string        `json:"agent_id"`
	APIKey       string        `json:"api_key"`
	Hostname     string        `json:"hostname"`
	ScannedAt    time.Time     `json:"scanned_at"`
	Certificates []Certificate `json:"certificates"`
}

type Certificate struct {
	SourceUID     string    `json:"source_uid"`
	SourceType    string    `json:"source_type"`
	Serial        string    `json:"serial"`
	Subject       DN        `json:"subject"`
	Issuer        DN        `json:"issuer"`
	SignatureAlgo string    `json:"signature_algo"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidUntil    time.Time `json:"valid_until"`
	DNSNames      []string  `json:"dns_names"`
	IsTrusted     bool      `json:"is_trusted"`
	TrustError    string    `json:"trust_error,omitempty"`
}

type DN struct {
	CN  string `json:"cn"`
	Org string `json:"org,omitempty"`
	OU  string `json:"ou,omitempty"`
}

type CertResponse struct {
	ID            string     `json:"id"`
	AgentID       string     `json:"agent_id"`
	AgentHostname string     `json:"agent_hostname"`
	SourceUID     string     `json:"source_uid"`
	SourceType    string     `json:"source_type"`
	CurrentStatus string     `json:"current_status"`
	Subject       DN         `json:"subject"`
	Issuer        DN         `json:"issuer"`
	ValidFrom     time.Time  `json:"valid_from"`
	ValidUntil    time.Time  `json:"valid_until"`
	IsTrusted     bool       `json:"is_trusted"`
	TrustError    string     `json:"trust_error,omitempty"`
	Status        CertStatus `json:"status"`

	// The Link to the User.
	OwnerID string `json:"owner_id"`
}

type Target struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id,omitempty"`
	TargetURL      string    `json:"target_url"`
	FrequencyHours int       `json:"frequency_hours"`
	LastScannedAt  time.Time `json:"last_scanned_at,omitempty"`
	Status         string    `json:"last_status,omitempty"`
	LastError      string    `json:"last_error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PaginatedCerts struct {
	Data  []CertResponse `json:"data"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type AgentResponse struct {
	ID         string      `json:"id"`
	Hostname   string      `json:"hostname"`
	IPAddress  string      `json:"ip_address"`
	LastSeenAt time.Time   `json:"last_seen_at"`
	IsVirtual  bool        `json:"is_virtual"`
	Status     AgentStatus `json:"status"`
	CertCount  int         `json:"cert_count"`
}

// DashboardStats holds the counts for the summary cards
type DashboardStats struct {
	TotalCerts    int `json:"total_certs"`
	ExpiringSoon  int `json:"expiring_soon"`
	Expired       int `json:"expired"`
	TotalAgents   int `json:"total_agents"`
	OnlineAgents  int `json:"online_agents"`
	OfflineAgents int `json:"offline_agents"`
}
