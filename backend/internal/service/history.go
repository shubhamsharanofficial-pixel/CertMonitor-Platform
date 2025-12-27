package service

import (
	"cert-manager-backend/internal/model"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// PostgresHistoryService implements the HistoryService interface.
type PostgresHistoryService struct {
	db *sql.DB
}

// NewHistoryService returns a concrete implementation of the HistoryService interface.
func NewHistoryService(db *sql.DB) *PostgresHistoryService {
	return &PostgresHistoryService{
		db: db,
	}
}

// FilterByCertID implements the "Bulk Check" logic with a configurable cooldown.
// It takes a list of certs and returns ONLY the ones that haven't been sent within the cooldown window.
func (s *PostgresHistoryService) FilterByCertID(ctx context.Context, certs []model.CertResponse, alertType string, cooldown time.Duration) ([]model.CertResponse, error) {
	if len(certs) == 0 {
		return []model.CertResponse{}, nil
	}

	// 1. Extract IDs for the bulk query
	// We use string slice because lib/pq handles the conversion to UUID array
	var certIDs []string
	for _, c := range certs {
		certIDs = append(certIDs, c.ID)
	}

	// 2. Calculate the cutoff time in Go
	// Any alert sent AFTER this time is considered "recent" and should block a new alert.
	cutoff := time.Now().Add(-cooldown)

	// 3. Query DB: "Which of these IDs were logged AFTER the cutoff?"
	query := `
		SELECT certificate_id 
		FROM alert_history 
		WHERE alert_type = $1 
		  AND sent_at > $2
		  AND certificate_id = ANY($3::uuid[])
	`

	rows, err := s.db.QueryContext(ctx, query, alertType, cutoff, pq.Array(certIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to check alert history: %w", err)
	}
	defer rows.Close()

	// 4. Build a Blocklist Map (In-Memory)
	blockedIDs := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		blockedIDs[id] = true
	}

	// 5. Filter the original list
	var toSend []model.CertResponse
	for _, c := range certs {
		// If NOT in blocked list, add to "To Send"
		if !blockedIDs[c.ID] {
			toSend = append(toSend, c)
		}
	}

	return toSend, nil
}

// RecordSent implements the "Bulk Write" logic.
// It logs the sent notifications so they get blocked next time.
func (s *PostgresHistoryService) RecordSent(ctx context.Context, certs []model.CertResponse, alertType string) error {
	if len(certs) == 0 {
		return nil
	}

	// We use UNNEST to perform a bulk insert in a single query.
	// This is much faster than looping and running INSERT 100 times.
	// We cast the input arrays to ::uuid[] so Postgres knows how to handle the strings.
	query := `
		INSERT INTO alert_history (certificate_id, agent_id, alert_type)
		SELECT unnest($1::uuid[]), unnest($2::uuid[]), $3
	`

	var certIDs []string
	var agentIDs []string

	for _, c := range certs {
		certIDs = append(certIDs, c.ID)
		agentIDs = append(agentIDs, c.AgentID)
	}

	_, err := s.db.ExecContext(ctx, query, pq.Array(certIDs), pq.Array(agentIDs), alertType)
	if err != nil {
		return fmt.Errorf("failed to record alert history: %w", err)
	}

	return nil
}
