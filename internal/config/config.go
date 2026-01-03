package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Port          string `mapstructure:"port"`
	DatabaseURL   string `mapstructure:"database_url"`
	APIKeyHash    string `mapstructure:"api_key_hash"`
	GeminiAPIKey  string `mapstructure:"gemini_api_key"`
	YouTubeAPIKey string `mapstructure:"youtube_api_key"`
	LogLevel      string `mapstructure:"log_level"`
}

// Load reads configuration from environment variables
// Supports _FILE suffix pattern for reading secrets from files (Docker Swarm style)
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("port", "4500")
	v.SetDefault("log_level", "info")

	// Bind environment variables
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	// Map of config keys to their env var names
	envBindings := map[string]string{
		"port":            "PORT",
		"database_url":    "DATABASE_URL",
		"api_key_hash":    "API_KEY_HASH",
		"gemini_api_key":  "GEMINI_API_KEY",
		"youtube_api_key": "YOUTUBE_API_KEY",
		"log_level":       "LOG_LEVEL",
	}

	for key, envVar := range envBindings {
		if err := v.BindEnv(key, envVar); err != nil {
			return nil, fmt.Errorf("failed to bind env var %s: %w", envVar, err)
		}
	}

	cfg := &Config{}

	// Load each config value, checking for _FILE variants first
	cfg.Port = getConfigValue(v, "port", "PORT")
	cfg.DatabaseURL = getConfigValue(v, "database_url", "DATABASE_URL")
	cfg.APIKeyHash = getConfigValue(v, "api_key_hash", "API_KEY_HASH")
	cfg.GeminiAPIKey = getConfigValue(v, "gemini_api_key", "GEMINI_API_KEY")
	cfg.YouTubeAPIKey = getConfigValue(v, "youtube_api_key", "YOUTUBE_API_KEY")
	cfg.LogLevel = getConfigValue(v, "log_level", "LOG_LEVEL")

	// Validate required config
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.APIKeyHash == "" {
		return nil, fmt.Errorf("API_KEY_HASH is required")
	}

	return cfg, nil
}

// getConfigValue checks for FOO_FILE env var first, reads from file if exists,
// otherwise falls back to FOO env var
func getConfigValue(v *viper.Viper, key, envVar string) string {
	// Check for _FILE variant first
	fileEnvVar := envVar + "_FILE"
	if filePath := os.Getenv(fileEnvVar); filePath != "" {
		if data, err := os.ReadFile(filePath); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	// Fall back to regular env var via viper
	return v.GetString(key)
}
