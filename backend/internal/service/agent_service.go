package service

import (
	"cert-manager-backend/internal/model"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PostgresAgentService struct {
	DB                    *sql.DB
	AgentOfflineThreshold time.Duration
}

// NewAgentService constructor now accepts the threshold config
func NewAgentService(db *sql.DB, offlineThreshold time.Duration) *PostgresAgentService {
	return &PostgresAgentService{
		DB:                    db,
		AgentOfflineThreshold: offlineThreshold,
	}
}

// ListAgents fetches all agents for a user.
func (s *PostgresAgentService) ListAgents(ctx context.Context, userID string) ([]model.AgentResponse, error) {
	// UPDATED: Added 'a.is_virtual' to the selection
	query := `
        SELECT 
            a.id, 
            a.hostname,
            COALESCE(a.ip_address, ''), 
            a.last_seen_at,
            a.is_virtual,
            COUNT(ci.id) as cert_count
        FROM agents a
        LEFT JOIN certificate_instances ci ON a.id = ci.agent_id
        WHERE a.user_id = $1
        GROUP BY a.id, a.hostname, a.ip_address, a.last_seen_at, a.is_virtual
        ORDER BY a.is_virtual DESC, a.last_seen_at DESC
    `
	// Note: ORDER BY is_virtual DESC puts Cloud Agents at the top of the list

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agents: %w", err)
	}
	defer rows.Close()

	var agents []model.AgentResponse
	for rows.Next() {
		var a model.AgentResponse

		// Ensure your model.AgentResponse has the IsVirtual field!
		err := rows.Scan(&a.ID, &a.Hostname, &a.IPAddress, &a.LastSeenAt, &a.IsVirtual, &a.CertCount)
		if err != nil {
			return nil, err
		}

		// Calculate Status Logic
		timeDiff := time.Since(a.LastSeenAt)
		if timeDiff < s.AgentOfflineThreshold {
			a.Status = model.StatusAgentOnline
		} else {
			a.Status = model.StatusAgentOffline
		}

		agents = append(agents, a)
	}

	if agents == nil {
		agents = []model.AgentResponse{}
	}

	return agents, nil
}

// DeleteAgent removes an agent. If it is a Virtual Agent, it cleans up monitoring targets.
func (s *PostgresAgentService) DeleteAgent(ctx context.Context, userID, agentID string) error {
	// 1. Start Transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 2. Check if Agent exists and is Virtual
	var isVirtual bool
	err = tx.QueryRowContext(ctx,
		"SELECT is_virtual FROM agents WHERE id = $1 AND user_id = $2",
		agentID, userID).Scan(&isVirtual)

	if err == sql.ErrNoRows {
		return fmt.Errorf("agent not found or access denied")
	} else if err != nil {
		return err
	}

	// 3. IF Virtual: Delete all Monitoring Rules (Stop the Worker for this user)
	if isVirtual {
		_, err := tx.ExecContext(ctx, "DELETE FROM monitored_targets WHERE user_id = $1", userID)
		if err != nil {
			return fmt.Errorf("failed to clean up cloud targets: %w", err)
		}
	}

	// 4. Delete the Agent
	// The DB "ON DELETE CASCADE" will automatically remove certificate_instances
	_, err = tx.ExecContext(ctx, "DELETE FROM agents WHERE id = $1", agentID)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return tx.Commit()
}

// CleanupDeadAgents deletes "Physical" agents that haven't been seen since the threshold.
// Virtual Agents are EXCLUDED from this cleanup to prevent configuration loss during inactive periods.
func (s *PostgresAgentService) CleanupDeadAgents(ctx context.Context, threshold time.Duration) (int64, error) {
	cutoff := time.Now().Add(-threshold)

	// UPDATED: Added 'AND is_virtual = FALSE'
	query := `DELETE FROM agents WHERE last_seen_at < $1 AND is_virtual = FALSE`

	result, err := s.DB.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup dead agents: %w", err)
	}

	return result.RowsAffected()
}
