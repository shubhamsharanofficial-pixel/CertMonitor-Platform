package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	FrontendURL string

	// Thresholds
	AgentOfflineMinutes time.Duration
	JanitorCleanupHours time.Duration
	AgentTTL            time.Duration
	MissingCertTTL      time.Duration

	// Alerter Configuration
	AlerterInterval     time.Duration
	AlerterExpiryWindow time.Duration

	// Email / SMTP Settings
	SMTP SMTPConfig
}

type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	SenderAddr string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DB_CONN", "postgres://postgres:postgres@localhost:5432/certdb?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "SUPER_SECRET_KEY_CHANGE_ME_IN_PROD"),
		Port:        getEnv("PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost"),

		// Parse integers from env, convert to Duration
		AgentOfflineMinutes: time.Duration(getEnvInt("AGENT_OFFLINE_MINUTES", 120)) * time.Minute,
		JanitorCleanupHours: time.Duration(getEnvInt("JANITOR_CLEANUP_HOURS", 24)) * time.Hour,
		AgentTTL:            time.Duration(getEnvInt("AGENT_TTL_HOURS", 24)) * time.Hour,

		// Default: 7 Days before hard deleting a missing cert
		MissingCertTTL: time.Duration(getEnvInt("MISSING_CERT_TTL_DAYS", 7)) * 24 * time.Hour,

		// New Alerter Configs
		AlerterInterval: time.Duration(getEnvInt("ALERTER_INTERVAL_HOURS", 24)) * time.Hour,
		// AlerterInterval:     time.Duration(getEnvInt("ALERTER_INTERVAL_SECONDS", 1*60)) * time.Second,
		AlerterExpiryWindow: time.Duration(getEnvInt("ALERTER_EXPIRY_DAYS", 150)) * 24 * time.Hour,

		// Initialize the SMTP struct here!
		SMTP: SMTPConfig{
			Host:       getEnv("SMTP_HOST", "smtp-relay.brevo.com"),
			Port:       getEnvInt("SMTP_PORT", 587),
			Username:   getEnv("SMTP_USER", ""),
			Password:   getEnv("SMTP_PASS", ""),
			SenderAddr: getEnv("SMTP_SENDER", "noreply@certmanager.local"),
		},
	}
}

// Helpers
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}
	if val, err := strconv.Atoi(strValue); err == nil {
		return val
	}
	return fallback
}
