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
	"cert-manager-backend/internal/service"
	"cert-manager-backend/internal/worker"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system env vars")
	}

	// 2. Load Configuration
	cfg := config.LoadConfig()

	fmt.Printf("✅ Config Values: AgentOffline=%v, JanitorCleanup=%v, AgentTTL=%v, AlerterInterval=%v, AlerterExpiryWindow=%v\n",
		cfg.AgentOfflineMinutes, cfg.JanitorCleanupHours, cfg.AgentTTL, cfg.AlerterInterval, cfg.AlerterExpiryWindow)

	// 3. Database Setup
	store, err := db.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer store.Conn.Close()

	if err := store.InitSchema(); err != nil {
		log.Fatalf("Failed to init schema: %v", err)
	}

	// 4. Service Wiring
	ingestService := service.NewIngestService(store.Conn, cfg.AgentOfflineMinutes)

	// Create History Service (Needed for Email Notifier)
	historyService := service.NewHistoryService(store.Conn)

	// Create Email Notifier (Serves as both Notifier AND EmailService)
	// UPDATED: Now passing cfg.FrontendURL
	emailNotifier := notify.NewEmailNotifier(cfg.SMTP, cfg.FrontendURL, historyService)

	// Inject EmailService into AuthService
	authService := service.NewAuthService(store.Conn, cfg.JWTSecret, emailNotifier)

	// 5. Notifier Setup (For Alerter)
	activeNotifiers := []service.Notifier{
		notify.NewLogNotifier(),
		emailNotifier, // This now works as a regular notifier too
	}

	// 6. Handler Wiring
	authHandler := api.NewAuthHandler(authService)
	certHandler := api.NewCertHandler(ingestService)
	agentHandler := api.NewAgentHandler(ingestService)

	// 7. Router Setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Public Routes
	r.Post("/api/signup", authHandler.HandleSignup)
	r.Post("/api/login", authHandler.HandleLogin)
	r.Post("/api/certs", certHandler.HandleIngest)
	r.Get("/api/agent/install", agentHandler.HandleGetInstallScript)

	// NEW: Public Auth Routes (Phase 2)
	r.Post("/api/auth/verify", authHandler.HandleVerifyEmail)
	r.Post("/api/auth/forgot-password", authHandler.HandleForgotPassword)
	r.Post("/api/auth/reset-password", authHandler.HandleResetPassword)

	// Serve Static Binaries (e.g., /api/downloads/agent-linux-amd64)
	workDir, _ := os.Getwd()
	downloadsDir := filepath.Join(workDir, "public", "downloads")
	FileServer(r, "/api/downloads", http.Dir(downloadsDir))

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(api.MakeAuthMiddleware(cfg.JWTSecret))

		// Certificates
		r.Get("/api/certs", certHandler.HandleListCerts)
		r.Delete("/api/certs/{id}", certHandler.HandleDeleteInstance)
		r.Delete("/api/certs/missing", certHandler.HandlePruneMissing)

		// Agents
		r.Post("/api/key/regenerate", authHandler.HandleRegenerateKey)
		r.Get("/api/agents", agentHandler.HandleListAgents)
		r.Delete("/api/agents/{agentID}", agentHandler.HandleDeleteAgent)

		// Profile Management Routes
		r.Get("/api/profile", authHandler.HandleGetProfile)
		r.Put("/api/profile", authHandler.HandleUpdateProfile)

		// Dashboard Stats
		r.Get("/api/stats", certHandler.HandleGetStats)
	})

	// 8. Background Workers
	worker.StartJanitor(
		ingestService,
		cfg.JanitorCleanupHours,
		cfg.AgentTTL,
		cfg.MissingCertTTL,
	)

	worker.StartAlerter(
		ingestService,
		authService,
		activeNotifiers,
		cfg.AlerterInterval,
		cfg.AlerterExpiryWindow,
	)

	// 9. Start Server
	log.Printf("🚀 Server starting on %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// Helper function to serve static files from a directory
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
