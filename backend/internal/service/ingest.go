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

type IngestService struct {
	DB                    *sql.DB
	AgentOfflineThreshold time.Duration
}

func NewIngestService(db *sql.DB, offlineThreshold time.Duration) *IngestService {
	return &IngestService{
		DB:                    db,
		AgentOfflineThreshold: offlineThreshold,
	}
}

// ProcessReport handles the transaction for an incoming agent report
func (s *IngestService) ProcessReport(ctx context.Context, report model.AgentReport) error {
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

	// 3. Upsert Agent
	queryAgent := `
        INSERT INTO agents (id, user_id, hostname, last_seen_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE 
        SET last_seen_at = EXCLUDED.last_seen_at, 
            hostname = EXCLUDED.hostname,
            user_id = EXCLUDED.user_id;
    `
	_, err = tx.ExecContext(ctx, queryAgent, report.AgentID, userID, report.Hostname, batchTime)
	if err != nil {
		return fmt.Errorf("failed to upsert agent: %w", err)
	}

	// 4. Process Certificates
	for _, cert := range report.Certificates {
		var certID string

		// Deduplicate Certificate Definition
		err = tx.QueryRowContext(ctx,
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
				return fmt.Errorf("failed to insert cert %s: %w", cert.Serial, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to query cert existence: %w", err)
		}

		// Determine Source Type & UID
		sourceType := cert.SourceType
		if sourceType == "" {
			sourceType = "FILE"
		}

		// Fallback for legacy agents sending 'Path' mapped to SourceUID
		sourceUID := cert.SourceUID
		if sourceUID == "" {
			// If Agent sent nothing, skip it or error out.
			// Ideally agent should be updated to send SourceUID.
			// Assuming 'cert.SourceUID' maps to the JSON field 'source_uid'
			// which used to be 'path'.
			continue
		}

		// Link Instance (Upsert with current_status = ACTIVE)
		// RENAMED: file_path -> source_uid
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
        `, report.AgentID, certID, sourceUID, sourceType, cert.IsTrusted, cert.TrustError, batchTime)

		if err != nil {
			return fmt.Errorf("failed to link instance %s: %w", sourceUID, err)
		}
	}

	// 5. Soft Delete Ghosts
	// Update current_status = 'MISSING' for any cert belonging to this agent
	// that was NOT updated in this batch (scanned_at != batchTime).
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

// CleanupMissingInstances deletes instances that have been MISSING for too long.
func (s *IngestService) CleanupMissingInstances(ctx context.Context, gracePeriod time.Duration) (int64, error) {
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
