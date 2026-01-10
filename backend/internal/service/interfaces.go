package service

import (
	"cert-manager-backend/internal/model"
	"context"
	"time"
)

// AuthService
type AuthService interface {
	Register(ctx context.Context, req model.SignupRequest) error
	Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error)
	RegenerateAPIKey(ctx context.Context, userID string) (string, error)

	// Bulk fetch users to support the Alerter lookaside pattern
	GetUsersByIDs(ctx context.Context, userIDs []string) (map[string]model.User, error)

	// Profile Management
	GetProfile(ctx context.Context, userID string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID string, req model.UpdateProfileRequest) error

	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// EmailService Interface
// This abstracts the email sending provider (Brevo) from the business logic
type EmailService interface {
	SendVerificationEmail(toEmail, token string) error
	SendPasswordResetEmail(toEmail, token string) error
}

// CertificateService
type CertificateService interface {
	// 1. External Agents (Physical)
	ProcessReport(ctx context.Context, report model.AgentReport, ipAddress string) error

	// 2. Internal Cloud Worker (Virtual) - NEW
	// Ingests a batch of certs for a user without needing an API Key.
	IngestScanResults(ctx context.Context, userID string, certs []model.Certificate) error

	CleanupOrphanedCerts(ctx context.Context) (int64, error)
	CleanupMissingInstances(ctx context.Context, gracePeriod time.Duration) (int64, error)

	// Uses Functional Options for flexible filtering
	ListCertificates(ctx context.Context, userID string, opts ...FilterOption) (*model.PaginatedCerts, error)

	// Returns a flat list of certs. Grouping happens in the Notifier.
	GetExpiringCertificates(ctx context.Context, threshold time.Duration) ([]model.CertResponse, error)

	// GetDashboardStats calculates summary counts for the dashboard
	GetDashboardStats(ctx context.Context, userID string) (*model.DashboardStats, error)

	// Delete a specific certificate instance (Individual Delete)
	DeleteInstance(ctx context.Context, userID, instanceID string) error

	// Delete all MISSING instances for a user (Bulk Prune)
	DeleteAllMissingInstances(ctx context.Context, userID string) (int64, error)
}

// NetworkScanner defines the capability to perform remote TLS scans.
// This interface allows us to mock the scanner in tests.
type NetworkScanner interface {
	// Scan performs the handshake and returns the certificate chain.
	// target should be in "host:port" format.
	Scan(ctx context.Context, target string) ([]model.Certificate, error)
}

type AgentLessTargetService interface {
	// --- User Facing ---
	// AddTarget: Validates, Scans immediately, Saves to DB, Ingests result.
	AddTarget(ctx context.Context, userID, rawURL string, frequency int) (*model.Target, error)
	UpdateTarget(ctx context.Context, userID, targetID string, frequency int) error
	ListTargets(ctx context.Context, userID string) ([]model.Target, error)
	DeleteTarget(ctx context.Context, userID, targetID string) error

	// --- Worker Facing ---
	GetStaleTargets(ctx context.Context) ([]model.Target, error)
	UpdateTargetStatus(ctx context.Context, targetID, status, errStr string) error
}

// AgentService
type AgentService interface {
	ListAgents(ctx context.Context, userID string) ([]model.AgentResponse, error)
	DeleteAgent(ctx context.Context, userID, agentID string) error
	CleanupDeadAgents(ctx context.Context, threshold time.Duration) (int64, error)
}

// HistoryService handles alert deduplication and logging.
type HistoryService interface {
	// FilterByCertID checks which certs have recently triggered an alert of the given type for a cooldown duration.
	FilterByCertID(ctx context.Context, certs []model.CertResponse, alertType string, cooldown time.Duration) ([]model.CertResponse, error)

	// RecordSent logs that an alert was sent for the given certificates.
	RecordSent(ctx context.Context, certs []model.CertResponse, alertType string) error
}

// Notifier Interface
// Accepts the Data (Certs) and the Context (UserMap) separately.
type Notifier interface {
	Notify(ctx context.Context, certs []model.CertResponse, users map[string]model.User) error
}
