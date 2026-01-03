package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Port          string
	DatabaseURL   string
	APIKeyHash    string
	GeminiAPIKey  string
	YouTubeAPIKey string
	LogLevel      string
	SecureCookies bool
}

// Load reads configuration from environment variables.
// Supports _FILE suffix pattern for reading secrets from files (Docker Swarm style).
func Load() (*Config, error) {
	var err error
	cfg := &Config{}

	if cfg.Port, err = getEnv("PORT", "4500"); err != nil {
		return nil, err
	}
	if cfg.DatabaseURL, err = getEnv("DATABASE_URL", ""); err != nil {
		return nil, err
	}
	if cfg.APIKeyHash, err = getEnv("API_KEY_HASH", ""); err != nil {
		return nil, err
	}
	if cfg.GeminiAPIKey, err = getEnv("GEMINI_API_KEY", ""); err != nil {
		return nil, err
	}
	if cfg.YouTubeAPIKey, err = getEnv("YOUTUBE_API_KEY", ""); err != nil {
		return nil, err
	}
	if cfg.LogLevel, err = getEnv("LOG_LEVEL", "info"); err != nil {
		return nil, err
	}

	// Secure cookies enabled by default (production), set SECURE_COOKIES=false for local dev
	secureCookiesStr, err := getEnv("SECURE_COOKIES", "true")
	if err != nil {
		return nil, err
	}
	cfg.SecureCookies = secureCookiesStr != "false"

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.APIKeyHash == "" {
		return nil, fmt.Errorf("API_KEY_HASH is required")
	}

	return cfg, nil
}

// getEnv checks for FOO_FILE env var first, reads from file if exists,
// otherwise falls back to FOO env var, then to the default value.
// Returns an error if _FILE is set but the file cannot be read.
func getEnv(key, defaultVal string) (string, error) {
	// Check for _FILE variant first (Docker Swarm secrets pattern)
	if filePath := os.Getenv(key + "_FILE"); filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read %s_FILE (%s): %w", key, filePath, err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	if val := os.Getenv(key); val != "" {
		return val, nil
	}

	return defaultVal, nil
}
