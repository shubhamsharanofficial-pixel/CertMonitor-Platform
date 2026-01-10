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
	AgentTTL            time.Duration
	MissingCertTTL      time.Duration

	// Cron Schedules
	JanitorSchedule string // e.g., "0 0 * * *"
	AlerterSchedule string // e.g., "0 9 * * *"

	// Alerter Configuration
	AlerterExpiryWindow time.Duration

	// Cloud Monitor Configs
	CloudScannerInterval            time.Duration
	CloudScannerTimeout             time.Duration
	CloudScannerConcurrency         int
	CloudScannerUserDefaultScanHour int

	// Log Notifier Configuration
	EnableLogAlerts bool

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
		FrontendURL: getEnv("FRONTEND_URL", "https://certmonitor.systems"),

		// Parse integers from env, convert to Duration
		AgentOfflineMinutes: time.Duration(getEnvInt("AGENT_OFFLINE_MINUTES", 360)) * time.Minute,
		AgentTTL:            time.Duration(getEnvInt("AGENT_TTL_HOURS", 24*3)) * time.Hour,
		// Default: 7 Days before hard deleting a missing cert
		MissingCertTTL: time.Duration(getEnvInt("MISSING_CERT_TTL_DAYS", 7)) * 24 * time.Hour,

		// 1. Janitor: 00:00 IST = 18:30 UTC
		// We set minute to 30 and hour to 18
		JanitorSchedule: getEnv("JANITOR_CRON", "30 18 * * *"),

		// 2. Alerter: 09:00 AM IST = 03:30 AM UTC
		// We set minute to 30 and hour to 3
		AlerterSchedule: getEnv("ALERTER_CRON", "30 3 * * *"),

		// Alerter Configs
		AlerterExpiryWindow: time.Duration(getEnvInt("ALERTER_EXPIRY_DAYS", 30)) * 24 * time.Hour,

		EnableLogAlerts: getEnvBool("ENABLE_LOG_NOTIFIER_ALERTS", false),

		// Cloud Monitor Configs
		CloudScannerInterval:            time.Duration(getEnvInt("CLOUD_SCANNER_INTERVAL_SECONDS", 300)) * time.Second,
		CloudScannerTimeout:             time.Duration(getEnvInt("CLOUD_SCANNER_TIMEOUT_SECONDS", 10)) * time.Second,
		CloudScannerConcurrency:         getEnvInt("CLOUD_SCANNER_CONCURRENCY", 2),
		CloudScannerUserDefaultScanHour: getEnvInt("CLOUD_SCANNER_USER_DEFAULT_SCAN_TIMER", 12),

		// Initialize the SMTP struct
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

func getEnvBool(key string, defaultVal bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
