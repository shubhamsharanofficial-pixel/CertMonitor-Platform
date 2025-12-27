package worker

import (
	"cert-manager-backend/internal/service"
	"context"
	"log"
	"time"
)

// StartJanitor launches the background cleanup.
// Updated to accept missingCertTTL
func StartJanitor(svc *service.IngestService, interval time.Duration, agentTTL time.Duration, missingCertTTL time.Duration) {
	go func() {
		log.Printf("🧹 Janitor started. Cleanup Every: %v | Agent TTL: %v | Missing Cert TTL: %v", interval, agentTTL, missingCertTTL)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()

			// 1. Cleanup Dead Agents
			deletedAgents, err := svc.CleanupDeadAgents(ctx, agentTTL)
			if err != nil {
				log.Printf("⚠️ Janitor Error (Agents): %v", err)
			} else if deletedAgents > 0 {
				log.Printf("🧹 Janitor: Removed %d dead agents", deletedAgents)
			}

			// 2. Cleanup Missing Instances (Soft Deleted -> Hard Deleted)
			deletedInstances, err := svc.CleanupMissingInstances(ctx, missingCertTTL)
			if err != nil {
				log.Printf("⚠️ Janitor Error (Missing Instances): %v", err)
			} else if deletedInstances > 0 {
				log.Printf("🧹 Janitor: Removed %d missing certificate links (Grace period expired)", deletedInstances)
			}

			// 3. Cleanup Orphaned Definitions (Now that instances are gone)
			deletedCerts, err := svc.CleanupOrphanedCerts(ctx)
			if err != nil {
				log.Printf("⚠️ Janitor Error (Orphans): %v", err)
			} else if deletedCerts > 0 {
				log.Printf("🧹 Janitor: Removed %d orphaned certificate definitions", deletedCerts)
			}
		}
	}()
}
