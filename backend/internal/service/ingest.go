package service

import (
	"cert-manager-backend/internal/model"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

type PostgresCertificateService struct {
	DB                    *sql.DB
	AgentOfflineThreshold time.Duration
}

func NewCertificateService(db *sql.DB, offlineThreshold time.Duration) *PostgresCertificateService {
	return &PostgresCertificateService{
		DB:                    db,
		AgentOfflineThreshold: offlineThreshold,
	}
}

// --- 1. External Agent Ingestion (Physical) ---
func (s *PostgresCertificateService) ProcessReport(ctx context.Context, report model.AgentReport, ipAddress string) error {
	// 1. Auth Check
	if report.APIKey == "" {
		return fmt.Errorf("missing api_key")
	}

	hash := sha256.Sum256([]byte(report.APIKey))
	apiKeyHash := hex.EncodeToString(hash[:])

	var userID string
	err := s.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE api_key_hash = $1", apiKeyHash).Scan(&userID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("invalid api_key: authentication failed")
	} else if err != nil {
		return fmt.Errorf("auth check failed: %w", err)
	}

	// 2. Start Transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	batchTime := time.Now()

	// 3. Upsert Physical Agent
	// Note: is_virtual defaults to FALSE here
	queryAgent := `
        INSERT INTO agents (id, user_id, hostname, last_seen_at, is_virtual, ip_address)
        VALUES ($1, $2, $3, $4, FALSE, $5)
        ON CONFLICT (id) DO UPDATE 
        SET last_seen_at = EXCLUDED.last_seen_at, 
            hostname = EXCLUDED.hostname,
            user_id = EXCLUDED.user_id,
			ip_address = EXCLUDED.ip_address;
    `

	_, err = tx.ExecContext(ctx, queryAgent, report.AgentID, userID, report.Hostname, batchTime, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to upsert agent: %w", err)
	}
	// 4. Process Certificates (Shared Logic)
	if err := s.upsertCertificates(ctx, tx, report.AgentID, report.Certificates, batchTime); err != nil {
		return err
	}

	// 5. Soft Delete Ghosts (Only for Physical Agents)
	// Cloud agents perform partial scans, so we CANNOT assume missing items are deleted.
	_, err = tx.ExecContext(ctx, `
        UPDATE certificate_instances 
        SET current_status = 'MISSING'
        WHERE agent_id = $1 
        AND scanned_at != $2
    `, report.AgentID, batchTime)

	if err != nil {
		return fmt.Errorf("failed to mark missing certificates: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	log.Printf("✅ Processed report from %s (User: %s)", report.Hostname, userID)
	return nil
}

// --- 2. Internal Cloud Ingestion (Virtual) ---

func (s *PostgresCertificateService) IngestScanResults(ctx context.Context, userID string, certs []model.Certificate) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	batchTime := time.Now()

	// 1. Get or Create Virtual Agent
	// We use a deterministic UUID based on UserID to ensure 1 Cloud Agent per User.
	// Or we can query for one. Let's query/create for safety.
	agentID, err := s.ensureVirtualAgent(ctx, tx, userID, batchTime)
	if err != nil {
		return fmt.Errorf("failed to ensure virtual agent: %w", err)
	}

	// 2. Process Certificates (Shared Logic)
	// Note: We do NOT perform Ghost Pruning here because cloud scans are partial updates.
	if err := s.upsertCertificates(ctx, tx, agentID, certs, batchTime); err != nil {
		return err
	}

	return tx.Commit()
}

// --- 3. Shared Helpers ---

// ensureVirtualAgent finds the existing "Cloud Monitor" agent for this user or creates one.
func (s *PostgresCertificateService) ensureVirtualAgent(ctx context.Context, tx *sql.Tx, userID string, seenAt time.Time) (string, error) {
	var agentID string

	// Check for existing virtual agent
	err := tx.QueryRowContext(ctx, `
        SELECT id FROM agents 
        WHERE user_id = $1 AND is_virtual = TRUE 
        LIMIT 1
    `, userID).Scan(&agentID)

	if err == sql.ErrNoRows {
		// Create new Virtual Agent
		// We can generate a UUID in Go or let DB do it.
		// Since we need the ID immediately, let's generate in SQL or use a placeholder.
		// Assuming Postgres 'gen_random_uuid()' is available, but we need the ID back.
		err = tx.QueryRowContext(ctx, `
            INSERT INTO agents (id, user_id, hostname, last_seen_at, is_virtual)
            VALUES (gen_random_uuid(), $1, 'Cloud Monitor', $2, TRUE)
            RETURNING id
        `, userID, seenAt).Scan(&agentID)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else {
		// Update last_seen_at
		_, err = tx.ExecContext(ctx, "UPDATE agents SET last_seen_at = $1 WHERE id = $2", seenAt, agentID)
		if err != nil {
			return "", err
		}
	}

	return agentID, nil
}

// upsertCertificates handles the core logic of saving Definitions and Instances
func (s *PostgresCertificateService) upsertCertificates(ctx context.Context, tx *sql.Tx, agentID string, certs []model.Certificate, batchTime time.Time) error {
	for _, cert := range certs {
		var certID string

		// A. Deduplicate Certificate Definition
		err := tx.QueryRowContext(ctx,
			`SELECT id FROM certificates 
             WHERE serial_number = $1 
             AND issuer_cn = $2
             AND COALESCE(issuer_org, '') = $3
             AND COALESCE(issuer_ou, '') = $4`,
			cert.Serial, cert.Issuer.CN, cert.Issuer.Org, cert.Issuer.OU,
		).Scan(&certID)

		if err == sql.ErrNoRows {
			err = tx.QueryRowContext(ctx, `
                INSERT INTO certificates 
                (serial_number, issuer_cn, issuer_org, issuer_ou, subject_cn, subject_org, subject_ou, valid_from, valid_until, signature_algo)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
                RETURNING id
            `,
				cert.Serial,
				cert.Issuer.CN, cert.Issuer.Org, cert.Issuer.OU,
				cert.Subject.CN, cert.Subject.Org, cert.Subject.OU,
				cert.ValidFrom, cert.ValidUntil, cert.SignatureAlgo,
			).Scan(&certID)

			if err != nil {
				return fmt.Errorf("failed to insert cert definition %s: %w", cert.Serial, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to query cert existence: %w", err)
		}

		// B. Upsert Instance
		// Logic: Always mark as ACTIVE and update scanned_at
		sourceType := cert.SourceType
		if sourceType == "" {
			sourceType = "FILE"
		}
		sourceUID := cert.SourceUID
		if sourceUID == "" {
			continue // Skip invalid items
		}

		_, err = tx.ExecContext(ctx, `
            INSERT INTO certificate_instances (agent_id, certificate_id, source_uid, source_type, is_trusted, trust_error, current_status, scanned_at)
            VALUES ($1, $2, $3, $4, $5, $6, 'ACTIVE', $7)
            ON CONFLICT (agent_id, source_uid) DO UPDATE
            SET certificate_id = EXCLUDED.certificate_id,
                source_type = EXCLUDED.source_type,
                is_trusted = EXCLUDED.is_trusted,
                trust_error = EXCLUDED.trust_error,
                current_status = 'ACTIVE',
                scanned_at = EXCLUDED.scanned_at;
        `, agentID, certID, sourceUID, sourceType, cert.IsTrusted, cert.TrustError, batchTime)

		if err != nil {
			return fmt.Errorf("failed to link instance %s: %w", sourceUID, err)
		}
	}
	return nil
}

// CleanupMissingInstances deletes instances that have been MISSING for too long.
func (s *PostgresCertificateService) CleanupMissingInstances(ctx context.Context, gracePeriod time.Duration) (int64, error) {
	cutoff := time.Now().Add(-gracePeriod)

	query := `
        DELETE FROM certificate_instances 
        WHERE current_status = 'MISSING' 
        AND scanned_at < $1
    `

	result, err := s.DB.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup missing instances: %w", err)
	}

	return result.RowsAffected()
}
