package config

import (
	"os"
)

// Database configuration
var (
	DBHost     = getEnvWithDefault("MYSQL_HOST", "localhost")
	DBUser     = getEnvWithDefault("MYSQL_USER", "root")
	DBPassword = getEnvWithDefault("MYSQL_PASSWORD", "")
	DBName     = getEnvWithDefault("MYSQL_DATABASE", "codehub")
)

// Application configuration
var (
	SecretKey          = getEnvWithDefault("SECRET_KEY", "default-secret-key-change-in-production")
	DashboardURL       = getEnvWithDefault("DASHBOARD_URL", "http://localhost:8080")
	PterodactylURL     = getEnvWithDefault("PTERODACTYL_URL", "")
	PterodactylAdminKey = getEnvWithDefault("PTERODACTYL_ADMIN_KEY", "")
	AutodeployNestID   = getEnvWithDefault("AUTODEPLOY_NEST_ID", "1")
	Port               = getEnvWithDefault("PORT", "8080")
)

// getEnvWithDefault gets an environment variable or returns a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
