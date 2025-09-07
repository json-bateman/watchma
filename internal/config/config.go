package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	J_API_KEY  = "JELLYFIN_API_KEY"
	J_BASE_URL = "JELLYFIN_BASE_URL"
	PORT       = "PORT"
	LOG_LEVEL  = "LOG_LEVEL"
	ENV        = "ENVIRONMENT"
)

type Config struct {
	// Don't log the api key
	JellyfinApiKey  string `json:"-"` // Exclude from JSON Marshalling
	JellyfinBaseURL string

	Port     int
	LogLevel slog.Level
	Env      string

	// Timeout users who are inactive maybe?
	SessionTimeout time.Duration
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using environment variables instead")
	}

	config := &Config{
		// Once again, don't log the api key!
		JellyfinApiKey:  os.Getenv(J_API_KEY),
		JellyfinBaseURL: os.Getenv(J_BASE_URL),
		LogLevel:        parseLogLevel(os.Getenv(LOG_LEVEL)),

		Port:           getEnvAsInt(PORT, 8080),
		Env:            getEnvAsString(ENV, "development"),
		SessionTimeout: getEnvAsDuration("SESSION_TIMEOUT", 30*time.Minute),
	}

	if err := config.validate(); err != nil {
		slog.Error("Configuration validation failed", "error", err)
		os.Exit(1)
	}

	config.logConfig()
	return config
}

// Validate checks for essential .env variables
func (c *Config) validate() error {
	if c.JellyfinApiKey == "" {
		return fmt.Errorf("required environment variable %s is not set", J_API_KEY)
	}
	if c.JellyfinBaseURL == "" {
		return fmt.Errorf("required environment variable %s is not set", J_BASE_URL)
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	return nil
}

func (c *Config) logConfig() {
	fmt.Println("------------------------------------------------------------------")
	slog.Info("Configuration loaded")
	slog.Info("JELLYFIN_URL", "url", c.JellyfinBaseURL)
	if c.JellyfinApiKey != "" {
		slog.Info("JELLYFIN_API_KEY", "status", "loaded") // Confirm existance only
	}
	slog.Info("PORT", "port", c.Port)
	slog.Info("LOG_LEVEL", "level", c.LogLevel)
	slog.Info("ENVIRONMENT", "env", c.Env)
	fmt.Println("------------------------------------------------------------------")
}

func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
