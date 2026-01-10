package worker

import (
	"cert-manager-backend/internal/model"
	"cert-manager-backend/internal/service"
	"context"
	"log"
	"time"
)

// NewAlerterJob returns a function that performs the notification logic ONCE.
func NewAlerterJob(
	certSvc service.CertificateService,
	authSvc service.AuthService,
	notifiers []service.Notifier,
	expiryWindow time.Duration,
) func() {
	return func() {
		log.Println("🔔 Alerter: Starting check...")
		ctx := context.Background()

		// --- PHASE 1: Fetch Data (The "What") ---
		certs, err := certSvc.GetExpiringCertificates(ctx, expiryWindow)
		if err != nil {
			log.Printf("⚠️ Alerter Error: Failed to fetch expiring certs: %v", err)
			return
		}

		if len(certs) == 0 {
			log.Println("✅ Alerter: No expiring certificates found.")
			return
		}

		// --- PHASE 2: Fetch Context (The "Who") ---
		// We need to resolve OwnerID -> User Profile (Email, Org, etc.)
		// allowing us to send 1 DB query instead of N queries inside the loop.

		uniqueUserIDs := getUniqueOwnerIDs(certs)

		userMap, err := authSvc.GetUsersByIDs(ctx, uniqueUserIDs)
		if err != nil {
			// We decide here: Do we abort? Or try to notify anyway?
			// Aborting is safer to avoid sending emails to "Unknown".
			log.Printf("⚠️ Alerter Error: Failed to fetch user context: %v", err)
			return
		}

		// --- PHASE 3: Notify (The "Action") ---
		log.Printf("🔔 Alerter: Processing %d certs for %d users across %d channels.",
			len(certs), len(uniqueUserIDs), len(notifiers))

		for _, n := range notifiers {
			// We pass both the raw data AND the user context map
			if err := n.Notify(ctx, certs, userMap); err != nil {
				log.Printf("⚠️ Alerter: A notifier failed: %v", err)
			}
		}
	}
}

// Helper to extract unique IDs from the certificate list
func getUniqueOwnerIDs(certs []model.CertResponse) []string {
	seen := make(map[string]bool)
	var ids []string

	for _, c := range certs {
		if c.OwnerID != "" && !seen[c.OwnerID] {
			seen[c.OwnerID] = true
			ids = append(ids, c.OwnerID)
		}
	}
	return ids
}
