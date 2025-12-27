package service

import (
	"cert-manager-backend/internal/model"
	"context"
	"fmt"
	"time"
)

// ListAgents fetches all agents for a user.
func (s *IngestService) ListAgents(ctx context.Context, userID string) ([]model.AgentResponse, error) {
	query := `
		SELECT 
			a.id, 
			a.hostname, 
			COALESCE(a.ip_address, ''), 
			a.last_seen_at,
			COUNT(ci.id) as cert_count
		FROM agents a
		LEFT JOIN certificate_instances ci ON a.id = ci.agent_id
		WHERE a.user_id = $1
		GROUP BY a.id, a.hostname, a.ip_address, a.last_seen_at
		ORDER BY a.last_seen_at DESC
	`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agents: %w", err)
	}
	defer rows.Close()

	var agents []model.AgentResponse
	for rows.Next() {
		var a model.AgentResponse

		err := rows.Scan(&a.ID, &a.Hostname, &a.IPAddress, &a.LastSeenAt, &a.CertCount)
		if err != nil {
			return nil, err
		}

		// Calculate Status Logic using ENUMS
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

// DeleteAgent removes an agent and (via CASCADE) all its certificate links.
func (s *IngestService) DeleteAgent(ctx context.Context, userID, agentID string) error {
	query := `DELETE FROM agents WHERE id = $1 AND user_id = $2`

	result, err := s.DB.ExecContext(ctx, query, agentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found or access denied")
	}

	return nil
}

// CleanupDeadAgents deletes agents that haven't been seen since the threshold.
func (s *IngestService) CleanupDeadAgents(ctx context.Context, threshold time.Duration) (int64, error) {
	cutoff := time.Now().Add(-threshold)

	query := `DELETE FROM agents WHERE last_seen_at < $1`

	result, err := s.DB.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup dead agents: %w", err)
	}

	return result.RowsAffected()
}
