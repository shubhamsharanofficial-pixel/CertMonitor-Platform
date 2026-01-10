package worker

import (
	"cert-manager-backend/internal/service"
	"context"
	"log"
	"time"
)

// NewJanitorJob returns a function that performs the cleanup logic ONCE.
// It is designed to be scheduled by a Cron runner.
func NewJanitorJob(
	agentSvc service.AgentService,
	certSvc service.CertificateService,
	agentTTL time.Duration,
	missingCertTTL time.Duration,
) func() {
	return func() {
		log.Println("🧹 Janitor: Starting scheduled cleanup...")
		ctx := context.Background()

		// 1. Cleanup Dead Agents
		deletedAgents, err := agentSvc.CleanupDeadAgents(ctx, agentTTL)
		if err != nil {
			log.Printf("⚠️ Janitor Error (Agents): %v", err)
		} else if deletedAgents > 0 {
			log.Printf("🧹 Janitor: Removed %d dead agents", deletedAgents)
		}

		// 2. Cleanup Missing Instances
		deletedInstances, err := certSvc.CleanupMissingInstances(ctx, missingCertTTL)
		if err != nil {
			log.Printf("⚠️ Janitor Error (Missing Instances): %v", err)
		} else if deletedInstances > 0 {
			log.Printf("🧹 Janitor: Removed %d missing certificate links", deletedInstances)
		}

		// 3. Cleanup Orphaned Definitions
		deletedCerts, err := certSvc.CleanupOrphanedCerts(ctx)
		if err != nil {
			log.Printf("⚠️ Janitor Error (Orphans): %v", err)
		} else if deletedCerts > 0 {
			log.Printf("🧹 Janitor: Removed %d orphaned certificate definitions", deletedCerts)
		}

		log.Println("🧹 Janitor: Cleanup complete.")
	}
}
