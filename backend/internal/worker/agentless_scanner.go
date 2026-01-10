package worker

import (
	"cert-manager-backend/internal/model"
	"cert-manager-backend/internal/service"
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

// StartAgentlessScanner launches the background poller.
func StartAgentlessScanner(
	targetSvc service.AgentLessTargetService,
	sc service.NetworkScanner,
	certSvc service.CertificateService,
	interval time.Duration,
	concurrency int,
) {
	// 1. Define the work logic as a reusable function
	runScanBatch := func() {
		ctx := context.Background()

		// Fetch Work
		targets, err := targetSvc.GetStaleTargets(ctx)
		if err != nil {
			log.Printf("⚠️ Agentless Worker: Failed to fetch targets: %v", err)
			return
		}

		if len(targets) == 0 {
			return
		}

		log.Printf("☁️ Agentless Worker: Scanning %d targets...", len(targets))

		// Process Batch with Worker Pool
		var wg sync.WaitGroup
		sem := make(chan struct{}, concurrency) // Semaphore

		for _, t := range targets {
			wg.Add(1)
			sem <- struct{}{}

			go func(target model.Target) {
				defer wg.Done()
				defer func() { <-sem }()

				// --- PANIC RECOVERY BLOCK ---
				defer func() {
					if r := recover(); r != nil {
						// 1. Log the crash immediately
						errMsg := fmt.Sprintf("Internal Scanner Panic: %v", r)
						log.Printf("🔥 CRITICAL PANIC in Agentless Worker [Target: %s]: %v", target.TargetURL, r)
						fmt.Println(string(debug.Stack()))

						// 2. Safe Cleanup (Nested Recovery)
						// We wrap the DB call in a func to catch a "Double Panic"
						func() {
							defer func() {
								if r2 := recover(); r2 != nil {
									log.Printf("💀 FATAL: Failed to update DB status during panic recovery. targetSvc might be nil. Error: %v", r2)
								}
							}()

							// Use fresh context for cleanup
							updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
							defer cancel()

							if err := targetSvc.UpdateTargetStatus(updateCtx, target.ID, "FAILED", errMsg); err != nil {
								log.Printf("⚠️ Failed to mark target %s as FAILED after panic: %v", target.TargetURL, err)
							}
						}()
					}
				}()
				// -----------------------------

				processTarget(ctx, target, sc, certSvc, targetSvc)
			}(t)
		}

		wg.Wait()
		log.Printf("☁️ Agentless Worker: Batch complete.")
	}

	// 2. Launch the loop
	go func() {
		log.Printf("☁️ Agentless Scanner started. Interval: %v | Concurrency: %d", interval, concurrency)

		// A. Run IMMEDIATELY on startup (Don't wait for first tick)
		runScanBatch()

		// B. Then run on schedule
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runScanBatch()
		}
	}()
}

func processTarget(
	ctx context.Context,
	t model.Target,
	sc service.NetworkScanner,
	certSvc service.CertificateService,
	targetSvc service.AgentLessTargetService,
) {
	// A. Perform Scan (with 10s timeout)
	scanCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	results, err := sc.Scan(scanCtx, t.TargetURL)

	status := "SUCCESS"
	errStr := ""

	// B. Handle Result
	if err != nil {
		status = "FAILED"
		errStr = err.Error()
		log.Printf("❌ Failed to scan %s: %v", t.TargetURL, err)
	} else {
		// --- DEFENSIVE FIX: Ensure SourceType is CLOUD ---
		// Even if the scanner sets it, we enforce it here to be safe.
		for i := range results {
			results[i].SourceType = "CLOUD"
		}

		// C. Ingest if successful
		if ingestErr := certSvc.IngestScanResults(ctx, t.UserID, results); ingestErr != nil {
			log.Printf("⚠️ Scanned %s but ingest failed: %v", t.TargetURL, ingestErr)
			// Status remains SUCCESS because the scan worked, but we log the system error
		}
	}

	// D. Update Target State
	if updateErr := targetSvc.UpdateTargetStatus(ctx, t.ID, status, errStr); updateErr != nil {
		log.Printf("⚠️ Failed to update status for %s: %v", t.ID, updateErr)
	}
}
