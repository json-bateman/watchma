package app

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	JELLYFIN_API_KEY  = "JELLYFIN_API_KEY"
	JELLYFIN_BASE_URL = "JELLYFIN_BASE_URL"
	OPENAI_API_KEY    = "OPENAI_API_KEY"
	PORT              = "PORT"
	LOG_LEVEL         = "LOG_LEVEL"
	IS_DEV            = "IS_DEV"
)

type Settings struct {
	// Don't log the api keys
	JellyfinApiKey  string `json:"-"` // Exclude from JSON Marshalling
	JellyfinBaseURL string
	UseDummyData    bool // Use dummy data when Jellyfin credentials not available

	OpenAIApiKey string `json:"-"` // Exclude from JSON Marshalling

	Port     int
	LogLevel slog.Level
	IsDev    bool
}

func LoadSettings() *Settings {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using environment variables instead")
	}

	config := &Settings{
		// Once again, don't log the api keys!
		JellyfinApiKey:  os.Getenv(JELLYFIN_API_KEY),
		JellyfinBaseURL: os.Getenv(JELLYFIN_BASE_URL),
		UseDummyData:    os.Getenv(JELLYFIN_API_KEY) == "" || os.Getenv(JELLYFIN_BASE_URL) == "",
		LogLevel:        parseLogLevel(os.Getenv(LOG_LEVEL)),

		OpenAIApiKey: os.Getenv(OPENAI_API_KEY),

		Port:  getEnvAsInt(PORT, 58008),
		IsDev: strings.ToLower(os.Getenv(IS_DEV)) == "true",
	}

	if err := config.validate(); err != nil {
		slog.Error("Configuration validation failed", "error", err)
		os.Exit(1)
	}

	return config
}

// Validate checks for essential .env variables
func (a *Settings) validate() error {
	// Only require Jellyfin credentials if not using dummy data
	if !a.UseDummyData {
		if a.JellyfinApiKey == "" {
			return fmt.Errorf("required environment variable %s is not set", JELLYFIN_API_KEY)
		}
		if a.JellyfinBaseURL == "" {
			return fmt.Errorf("required environment variable %s is not set", JELLYFIN_BASE_URL)
		}
	}
	if a.Port < 1 || a.Port > 65535 {
		return fmt.Errorf("invalid port: %d", a.Port)
	}
	return nil
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
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
