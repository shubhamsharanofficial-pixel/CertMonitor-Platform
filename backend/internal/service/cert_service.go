package service

import (
	"cert-manager-backend/internal/model"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ListCertificates fetches certificates using Functional Options.
func (s *PostgresCertificateService) ListCertificates(ctx context.Context, userID string, opts ...FilterOption) (*model.PaginatedCerts, error) {
	// A. Apply Defaults
	filter := &CertFilter{
		Limit:  10,
		Offset: 0,
	}

	// B. Apply User Options
	for _, opt := range opts {
		opt(filter)
	}

	// C. Build Dynamic SQL
	baseQuery := `
        SELECT 
            ci.id, 
            ci.agent_id,
            a.hostname, 
            ci.source_uid, 
            ci.source_type,
            ci.current_status,
            c.subject_cn, c.subject_org, c.subject_ou,
            c.issuer_cn, c.issuer_org, c.issuer_ou,
            c.valid_from,
            c.valid_until, 
            ci.is_trusted,
            ci.trust_error,
            COUNT(*) OVER() as full_count 
        FROM certificate_instances ci
        JOIN agents a ON ci.agent_id = a.id
        JOIN certificates c ON ci.certificate_id = c.id
        WHERE a.user_id = $1
    `
	args := []interface{}{userID}
	argCounter := 2

	// Dynamic Filters
	if filter.AgentID != "" {
		baseQuery += fmt.Sprintf(" AND ci.agent_id = $%d", argCounter)
		args = append(args, filter.AgentID)
		argCounter++
	}

	if filter.SearchQuery != "" {
		// FIXED: Use distinct placeholders ($2, $3, $4, $5) and increment counter by 4.
		// This prevents "search string" arguments from bleeding into subsequent date filters.
		baseQuery += fmt.Sprintf(" AND (c.subject_cn ILIKE $%d OR c.issuer_cn ILIKE $%d OR a.hostname ILIKE $%d OR ci.source_uid ILIKE $%d)",
			argCounter, argCounter+1, argCounter+2, argCounter+3)

		val := "%" + filter.SearchQuery + "%"
		args = append(args, val, val, val, val)
		argCounter += 4
	}

	if filter.ValidAfter != nil {
		baseQuery += fmt.Sprintf(" AND c.valid_until >= $%d", argCounter)
		args = append(args, *filter.ValidAfter)
		argCounter++
	}

	if filter.ValidBefore != nil {
		baseQuery += fmt.Sprintf(" AND c.valid_until <= $%d", argCounter)
		args = append(args, *filter.ValidBefore)
		argCounter++
	}

	// Trust Filter
	if filter.IsTrusted != nil {
		baseQuery += fmt.Sprintf(" AND ci.is_trusted = $%d", argCounter)
		args = append(args, *filter.IsTrusted)
		argCounter++
	}

	// Status Filter (Active vs Missing)
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND ci.current_status = $%d", argCounter)
		args = append(args, filter.Status)
		argCounter++
	}

	// Sorting & Pagination
	baseQuery += fmt.Sprintf(" ORDER BY c.valid_until ASC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, filter.Limit, filter.Offset)

	// Execute
	rows, err := s.DB.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list certs: %w", err)
	}
	defer rows.Close()

	var list []model.CertResponse
	var total int

	for rows.Next() {
		var r model.CertResponse
		var sOrg, sOU, iOrg, iOU, tErr, sourceType, curStatus sql.NullString

		err := rows.Scan(
			&r.ID, &r.AgentID, &r.AgentHostname, &r.SourceUID,
			&sourceType, &curStatus,
			&r.Subject.CN, &sOrg, &sOU,
			&r.Issuer.CN, &iOrg, &iOU,
			&r.ValidFrom, &r.ValidUntil, &r.IsTrusted,
			&tErr,
			&total,
		)
		if err != nil {
			return nil, err
		}

		r.Subject.Org = sOrg.String
		r.Subject.OU = sOU.String
		r.Issuer.Org = iOrg.String
		r.Issuer.OU = iOU.String
		r.TrustError = tErr.String
		r.SourceType = sourceType.String
		r.CurrentStatus = curStatus.String

		// Logic to determine Priority Status
		now := time.Now()
		// Calculate thresholds based on "End of Day" logic, identical to ListCertificates
		y, m, d := now.Date()
		loc := now.Location()

		// Midnight tonight (The boundary for "Today")
		endOfToday := time.Date(y, m, d+1, 0, 0, 0, 0, loc)

		// Midnight tomorrow (The boundary for "Tomorrow")
		endOfTomorrow := time.Date(y, m, d+2, 0, 0, 0, 0, loc)

		if now.After(r.ValidUntil) {
			r.Status = model.StatusExpired
		} else if !r.IsTrusted {
			r.Status = model.StatusUntrusted
		} else if now.Before(r.ValidFrom) {
			r.Status = model.StatusNotYetValid
		} else if now.AddDate(0, 0, 30).After(r.ValidUntil) {
			if r.ValidUntil.Before(endOfToday) {
				r.Status = model.StatusExpiringToday
			} else if r.ValidUntil.Before(endOfTomorrow) {
				r.Status = model.StatusExpiringTomorrow
			} else if time.Until(r.ValidUntil).Hours() < 168 {
				r.Status = model.StatusExpiringThisWeek
			} else {
				r.Status = model.StatusExpiringSoon
			}
		} else {
			r.Status = model.StatusValid
		}

		list = append(list, r)
	}

	if list == nil {
		list = []model.CertResponse{}
	}

	return &model.PaginatedCerts{
		Data:  list,
		Total: total,
		Page:  (filter.Offset / filter.Limit) + 1,
		Limit: filter.Limit,
	}, nil
}

// DeleteInstance removes a specific certificate instance (Hard Delete).
func (s *PostgresCertificateService) DeleteInstance(ctx context.Context, userID, instanceID string) error {
	// We use a JOIN to ensure the instance actually belongs to an agent owned by the requesting user.
	// This prevents users from deleting each other's data by guessing IDs.
	query := `
		DELETE FROM certificate_instances ci
		USING agents a
		WHERE ci.agent_id = a.id
		  AND a.user_id = $1
		  AND ci.id = $2
	`

	result, err := s.DB.ExecContext(ctx, query, userID, instanceID)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("instance not found or access denied")
	}

	return nil
}

// DeleteAllMissingInstances removes ALL instances marked 'MISSING' for a user.
func (s *PostgresCertificateService) DeleteAllMissingInstances(ctx context.Context, userID string) (int64, error) {
	query := `
		DELETE FROM certificate_instances ci
		USING agents a
		WHERE ci.agent_id = a.id
		  AND a.user_id = $1
		  AND ci.current_status = 'MISSING'
	`

	result, err := s.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to prune missing instances: %w", err)
	}

	return result.RowsAffected()
}

// CleanupOrphanedCerts (Unchanged)
func (s *PostgresCertificateService) CleanupOrphanedCerts(ctx context.Context) (int64, error) {
	query := `
        DELETE FROM certificates 
        WHERE id NOT IN (
            SELECT DISTINCT certificate_id FROM certificate_instances
        )
    `
	result, err := s.DB.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup orphaned certs: %w", err)
	}
	return result.RowsAffected()
}

// GetExpiringCertificates fetches certificates expiring within the threshold.
func (s *PostgresCertificateService) GetExpiringCertificates(ctx context.Context, threshold time.Duration) ([]model.CertResponse, error) {
	cutoff := time.Now().Add(threshold)

	query := `
        SELECT 
            c.id, c.serial_number, c.valid_from, c.valid_until, 
            c.subject_cn, c.subject_org, c.subject_ou,
            c.issuer_cn, c.issuer_org, c.issuer_ou,
            ci.source_uid, ci.is_trusted,
            a.id AS agent_id, a.hostname, a.user_id
        FROM certificate_instances ci
        JOIN certificates c ON ci.certificate_id = c.id
        JOIN agents a ON ci.agent_id = a.id
        WHERE c.valid_until < $1 
          AND c.valid_until > NOW()
          AND ci.current_status = 'ACTIVE' 
    `

	rows, err := s.DB.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expiring certs: %w", err)
	}
	defer rows.Close()

	var alerts []model.CertResponse

	for rows.Next() {
		var cr model.CertResponse
		var subCN, subOrg, subOU sql.NullString
		var issCN, issOrg, issOU sql.NullString
		var serial sql.NullString

		err := rows.Scan(
			&cr.ID, &serial, &cr.ValidFrom, &cr.ValidUntil,
			&subCN, &subOrg, &subOU,
			&issCN, &issOrg, &issOU,
			&cr.SourceUID, &cr.IsTrusted,
			&cr.AgentID, &cr.AgentHostname, &cr.OwnerID,
		)
		if err != nil {
			return nil, err
		}

		cr.Subject = model.DN{CN: subCN.String, Org: subOrg.String, OU: subOU.String}
		cr.Issuer = model.DN{CN: issCN.String, Org: issOrg.String, OU: issOU.String}

		hoursUntil := time.Until(cr.ValidUntil).Hours()
		if hoursUntil < 24 {
			cr.Status = model.StatusExpiringToday
		} else if hoursUntil < 48 {
			cr.Status = model.StatusExpiringTomorrow
		} else if hoursUntil < 168 {
			cr.Status = model.StatusExpiringThisWeek
		} else {
			cr.Status = model.StatusExpiringSoon
		}

		alerts = append(alerts, cr)
	}

	return alerts, nil
}

// GetDashboardStats returns summary counts for the dashboard.
func (s *PostgresCertificateService) GetDashboardStats(ctx context.Context, userID string) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	// Uses 'current_status' = 'ACTIVE' to ensure we don't count missing certs
	queryCerts := `
        SELECT 
            COUNT(*) FILTER (WHERE ci.current_status = 'ACTIVE') AS total,
            COUNT(*) FILTER (WHERE c.valid_until < NOW() + INTERVAL '30 days' AND c.valid_until > NOW() AND ci.current_status = 'ACTIVE') AS expiring_soon,
            COUNT(*) FILTER (WHERE c.valid_until < NOW() AND ci.current_status = 'ACTIVE') AS expired
        FROM certificate_instances ci
        JOIN agents a ON ci.agent_id = a.id
        JOIN certificates c ON ci.certificate_id = c.id
        WHERE a.user_id = $1
    `
	err := s.DB.QueryRowContext(ctx, queryCerts, userID).Scan(&stats.TotalCerts, &stats.ExpiringSoon, &stats.Expired)
	if err != nil {
		return nil, fmt.Errorf("failed to count certs: %w", err)
	}

	// Query 2: Agent Counts
	offlineThreshold := time.Now().Add(-s.AgentOfflineThreshold)

	queryAgents := `
        SELECT 
            COUNT(*) AS total,
            COUNT(*) FILTER (WHERE last_seen_at > $2) AS online
        FROM agents 
        WHERE user_id = $1
    `
	err = s.DB.QueryRowContext(ctx, queryAgents, userID, offlineThreshold).Scan(&stats.TotalAgents, &stats.OnlineAgents)
	if err != nil {
		return nil, fmt.Errorf("failed to count agents: %w", err)
	}

	stats.OfflineAgents = stats.TotalAgents - stats.OnlineAgents

	return stats, nil
}
