package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cert-manager-backend/internal/api"
	"cert-manager-backend/internal/config"
	"cert-manager-backend/internal/db"
	"cert-manager-backend/internal/notify"
	"cert-manager-backend/internal/scanner"
	"cert-manager-backend/internal/service"
	"cert-manager-backend/internal/worker"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	// =========================================================================
	// 1. Load Config & Environment
	// =========================================================================
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system env vars")
	}

	// 2. Load Configuration
	cfg := config.LoadConfig()

	fmt.Printf("✅ Config Loaded: Port=%s | OfflineThreshold=%v | CloudScanInterval=%v | CloudScannerTimeout=%v | CloudScannerUserDefaultScanHour=%v\n",
		cfg.Port, cfg.AgentOfflineMinutes, cfg.CloudScannerInterval, cfg.CloudScannerTimeout, cfg.CloudScannerUserDefaultScanHour)

	fmt.Printf("✅ Port=%s | AgentTTL=%v | MissingCertTTL=%v | AlerterExpiryWindow=%v | FrontendURL=%v\n",
		cfg.Port, cfg.AgentTTL, cfg.MissingCertTTL, cfg.AlerterExpiryWindow, cfg.FrontendURL)

	// =========================================================================
	// 2. Database Setup
	// =========================================================================
	store, err := db.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer store.Conn.Close()

	if err := store.InitSchema(); err != nil {
		log.Fatalf("Failed to init schema: %v", err)
	}

	// =========================================================================
	// 3. Service Initialization
	// =========================================================================

	// A. Core Services
	// AgentService: Handles Agent Lifecycle (List, Delete, Cleanup)
	agentSvc := service.NewAgentService(store.Conn, cfg.AgentOfflineMinutes)

	// CertificateService: Handles Ingestion (ProcessReport), Cleanup, Listing, Stats
	// (Previously split between IngestService and CertService)
	certSvc := service.NewCertificateService(store.Conn, cfg.AgentOfflineMinutes)

	// HistoryService: Needed for Alerter logs
	historySvc := service.NewHistoryService(store.Conn)

	// B. Cloud Monitoring Components
	tlsScanner := scanner.NewTLSScanner(cfg.CloudScannerTimeout)
	// CloudService: Orchestrates the "Scan -> Save" flow. Consumes certSvc for ingestion.
	cloudSvc := service.NewAgentLessTargetService(store.Conn, tlsScanner, certSvc, cfg.CloudScannerUserDefaultScanHour)

	// C. Auth & Notifications
	emailNotifier := notify.NewEmailNotifier(cfg.SMTP, cfg.FrontendURL, historySvc)
	authSvc := service.NewAuthService(store.Conn, cfg.JWTSecret, emailNotifier)

	// =========================================================================
	// 4. Handler Wiring
	// =========================================================================

	authHandler := api.NewAuthHandler(authSvc)
	agentHandler := api.NewAgentHandler(agentSvc)

	// CertHandler now handles BOTH ingestion (POST) and listing (GET)
	certHandler := api.NewCertHandler(certSvc)

	cloudHandler := api.NewCloudHandler(cloudSvc)

	// =========================================================================
	// 5.1 Background Workers
	// =========================================================================

	// A. Notifier List
	activeNotifiers := []service.Notifier{
		emailNotifier,
	}

	// Conditionally add the Log Notifier
	if cfg.EnableLogAlerts {
		activeNotifiers = append([]service.Notifier{notify.NewLogNotifier()}, activeNotifiers...)
		log.Println("✅ Log Alerts Enabled")
	}

	// B. Cloud Scanner (Agentless Engine)
	worker.StartAgentlessScanner(
		cloudSvc,   // Source of Targets (GetStaleTargets)
		tlsScanner, // The Network Tool
		certSvc,    // The Data Sink (IngestScanResults)
		cfg.CloudScannerInterval,
		cfg.CloudScannerConcurrency,
	)

	// ==========================================j
	// 5.2 Background Workers (CRON SCHEDULER)
	// ==========================================

	// Create the Scheduler
	c := cron.New()

	// C. Schedule Janitor
	// Uses cfg.JanitorSchedule (default: "0 0 * * *")
	_, janitorErr := c.AddFunc(cfg.JanitorSchedule, worker.NewJanitorJob(
		agentSvc,
		certSvc,
		cfg.AgentTTL,
		cfg.MissingCertTTL,
	))
	if janitorErr != nil {
		log.Fatalf("❌ Failed to schedule Janitor: %v", janitorErr)
	}
	log.Printf("✅ Janitor scheduled: %s", cfg.JanitorSchedule)

	// D. Schedule Alerter
	// Uses cfg.AlerterSchedule (default: "0 9 * * *")
	_, alerterErr := c.AddFunc(cfg.AlerterSchedule, worker.NewAlerterJob(
		certSvc,
		authSvc,
		activeNotifiers,
		cfg.AlerterExpiryWindow,
	))
	if alerterErr != nil {
		log.Fatalf("❌ Failed to schedule Alerter: %v", alerterErr)
	}
	log.Printf("✅ Alerter scheduled: %s", cfg.AlerterSchedule)

	// Start the Scheduler (runs in its own goroutine)
	c.Start()

	// =========================================================================
	// 6. Router Setup
	// =========================================================================

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Update for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// --- Public Routes ---
	r.Post("/api/signup", authHandler.HandleSignup)
	r.Post("/api/login", authHandler.HandleLogin)
	r.Post("/api/auth/verify", authHandler.HandleVerifyEmail)
	r.Post("/api/auth/forgot-password", authHandler.HandleForgotPassword)
	r.Post("/api/auth/reset-password", authHandler.HandleResetPassword)

	// Ingestion (Physical Agents)
	// Now mapped to CertHandler because CertService has ProcessReport
	r.Post("/api/certs", certHandler.HandleIngest)

	// Downloads
	r.Get("/api/agent/install", agentHandler.HandleGetInstallScript)
	workDir, _ := os.Getwd()
	downloadsDir := filepath.Join(workDir, "public", "downloads")
	FileServer(r, "/api/downloads", http.Dir(downloadsDir))

	// --- Protected Routes ---
	r.Group(func(r chi.Router) {
		r.Use(api.MakeAuthMiddleware(cfg.JWTSecret))

		// Certificates
		r.Get("/api/certs", certHandler.HandleListCerts)
		r.Delete("/api/certs/{id}", certHandler.HandleDeleteInstance)
		r.Delete("/api/certs/missing", certHandler.HandlePruneMissing)
		r.Get("/api/stats", certHandler.HandleGetStats)

		// Agents
		r.Post("/api/key/regenerate", authHandler.HandleRegenerateKey)
		r.Get("/api/agents", agentHandler.HandleListAgents)
		r.Delete("/api/agents/{agentID}", agentHandler.HandleDeleteAgent)
		r.Post("/api/key/regenerate", authHandler.HandleRegenerateKey)

		// Profile
		r.Get("/api/profile", authHandler.HandleGetProfile)
		r.Put("/api/profile", authHandler.HandleUpdateProfile)

		// ☁️ Cloud Monitoring (Agentless)
		r.Post("/api/cloud/targets", cloudHandler.HandleAddTarget)
		r.Get("/api/cloud/targets", cloudHandler.HandleListTargets)
		r.Delete("/api/cloud/targets/{id}", cloudHandler.HandleDeleteTarget)
		r.Put("/api/cloud/targets/{id}", cloudHandler.HandleUpdateTarget)
	})

	// =========================================================================
	// 7. Start Server
	// =========================================================================
	log.Printf("🚀 Server starting on %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// FileServer Helper
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
