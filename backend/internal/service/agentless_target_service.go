package service

import (
	"cert-manager-backend/internal/model"
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strings"
)

type PostgresAgentLessTargetService struct {
	DB                            *sql.DB
	Scanner                       NetworkScanner
	CertSvc                       CertificateService
	UserScanDefaultFrequencyHours int
}

func NewAgentLessTargetService(db *sql.DB, sc NetworkScanner, certSvc CertificateService, userScanDefaultFrequencyHours int) *PostgresAgentLessTargetService {
	return &PostgresAgentLessTargetService{
		DB:                            db,
		Scanner:                       sc,
		CertSvc:                       certSvc,
		UserScanDefaultFrequencyHours: userScanDefaultFrequencyHours,
	}
}

// --- User Facing Logic ---
func (s *PostgresAgentLessTargetService) AddTarget(ctx context.Context, userID, rawURL string, frequency int) (*model.Target, error) {
	// frequency Validation
	// Default to 12 hours if user sends 0 or omits field
	if frequency <= 0 {
		frequency = s.UserScanDefaultFrequencyHours
	}

	// 1. Normalize
	targetAddr, err := normalizeTarget(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target format: %v", err)
	}

	// 2. Scan-on-Create (Immediate Feedback)
	scanResults, scanErr := s.Scanner.Scan(ctx, targetAddr)
	if scanErr != nil {
		return nil, fmt.Errorf("scan failed (target unreachable?): %v", scanErr)
	}

	for i := range scanResults {
		scanResults[i].SourceType = "CLOUD"
	}

	// 3. Save to DB
	t := &model.Target{
		UserID:         userID,
		TargetURL:      targetAddr,
		FrequencyHours: frequency,
		Status:         "SUCCESS",
	}

	query := `
        INSERT INTO monitored_targets (user_id, target_url, frequency_hours, last_scanned_at, last_status)
        VALUES ($1, $2, $3, NOW(), 'SUCCESS')
        RETURNING id, created_at
    `
	err = s.DB.QueryRowContext(ctx, query, userID, targetAddr, t.FrequencyHours).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("you are already monitoring this target")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 4. Ingest Results
	if err := s.CertSvc.IngestScanResults(ctx, userID, scanResults); err != nil {
		// COMPENSATING ACTION:
		// The ingest failed, so the target is useless. Delete it immediately.
		s.DB.ExecContext(ctx, "DELETE FROM monitored_targets WHERE id = $1", t.ID)

		return nil, fmt.Errorf("failed to ingest certificate data: %v", err)
	}

	return t, nil
}

func (s *PostgresAgentLessTargetService) UpdateTarget(ctx context.Context, userID, targetID string, frequency int) error {
	// frequency Validation
	if frequency < 1 {
		return fmt.Errorf("frequency must be at least 1 hour")
	}

	// We check user_id to ensure ownership
	query := `
        UPDATE monitored_targets 
        SET frequency_hours = $1 
        WHERE id = $2 AND user_id = $3
    `

	result, err := s.DB.ExecContext(ctx, query, frequency, targetID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("target not found or access denied")
	}

	return nil
}

func (s *PostgresAgentLessTargetService) ListTargets(ctx context.Context, userID string) ([]model.Target, error) {
	query := `
        SELECT id, target_url, frequency_hours, last_scanned_at, last_status, last_error
        FROM monitored_targets
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []model.Target
	for rows.Next() {
		var t model.Target
		var lastScanned sql.NullTime
		var lastErr sql.NullString
		if err := rows.Scan(&t.ID, &t.TargetURL, &t.FrequencyHours, &lastScanned, &t.Status, &lastErr); err != nil {
			return nil, err
		}
		if lastScanned.Valid {
			t.LastScannedAt = lastScanned.Time
		}
		t.LastError = lastErr.String
		targets = append(targets, t)
	}
	return targets, nil
}

func (s *PostgresAgentLessTargetService) DeleteTarget(ctx context.Context, userID, targetID string) error {
	// 1. Start Transaction (Atomic Delete)
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 2. Resolve Target URL & User's Virtual Agent ID
	// We need the Target URL because that is how we tagged the certificate instances (SourceUID).
	var targetURL string
	err = tx.QueryRowContext(ctx, `
		SELECT target_url FROM monitored_targets 
		WHERE id = $1 AND user_id = $2
	`, targetID, userID).Scan(&targetURL)

	if err == sql.ErrNoRows {
		return fmt.Errorf("target not found or access denied")
	} else if err != nil {
		return err
	}

	// 3. Find the User's Virtual Agent ID
	var agentID string
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM agents WHERE user_id = $1 AND is_virtual = TRUE
	`, userID).Scan(&agentID)

	// If agent exists, delete the specific instances associated with this target
	if err == nil {
		// This removes the "Leaf", "Intermediate", and "Root" entries for this specific URL
		_, err = tx.ExecContext(ctx, `
			DELETE FROM certificate_instances 
			WHERE agent_id = $1 AND source_uid = $2
		`, agentID, targetURL)

		if err != nil {
			return fmt.Errorf("failed to cleanup certificate instances: %w", err)
		}
	}

	// 4. Delete the Monitored Target
	_, err = tx.ExecContext(ctx, `
		DELETE FROM monitored_targets 
		WHERE id = $1 AND user_id = $2
	`, targetID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete target: %w", err)
	}

	// 5. Commit
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// --- Worker Facing Logic ---

func (s *PostgresAgentLessTargetService) GetStaleTargets(ctx context.Context) ([]model.Target, error) {
	// Select targets that have NEVER been scanned OR are past their frequency interval
	query := `
        SELECT id, user_id, target_url, frequency_hours
		FROM monitored_targets
		WHERE last_scanned_at IS NULL
		OR (last_scanned_at + (frequency_hours * INTERVAL '1 hour')) < NOW()
		ORDER BY last_scanned_at ASC NULLS FIRST
		LIMIT 50
    `
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []model.Target
	for rows.Next() {
		var t model.Target
		if err := rows.Scan(&t.ID, &t.UserID, &t.TargetURL, &t.FrequencyHours); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (s *PostgresAgentLessTargetService) UpdateTargetStatus(ctx context.Context, targetID, status, errStr string) error {
	query := `
        UPDATE monitored_targets
        SET last_scanned_at = NOW(),
            last_status = $1,
            last_error = $2
        WHERE id = $3
    `
	var dbErr sql.NullString
	if errStr != "" {
		dbErr = sql.NullString{String: errStr, Valid: true}
	}
	_, err := s.DB.ExecContext(ctx, query, status, dbErr, targetID)
	return err
}

func normalizeTarget(input string) (string, error) {
	// 1. Clean whitespace (Handle accidental copy-paste spaces)
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("target URL cannot be empty")
	}

	// 2. Add Scheme if missing (Crucial for url.Parse)
	// If the user typed "www.google.com", url.Parse sees it as a "Path", not a "Host".
	// Adding "https://" forces the parser to recognize the hostname.
	if !strings.Contains(input, "://") {
		input = "https://" + input
	}

	// 3. Parse the URL
	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// 4. Extract Hostname
	// u.Hostname() is smart:
	// - It handles "www.name.com" -> "www.name.com"
	// - It handles "http://name.com" -> "name.com"
	// - It strips paths like "/login" automatically
	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("could not identify a valid hostname")
	}

	// 5. Extract or Default Port
	// u.Port() returns the port only if explicitly defined (e.g., :8443).
	// If empty, we DEFAULT to 443 (HTTPS), even if the input started with http://.
	port := u.Port()
	if port == "" {
		port = "443"
	}

	// 6. Return strict "host:port" format
	return net.JoinHostPort(host, port), nil
}
