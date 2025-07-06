package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	JellyfinApiKey string
	BaseUrl        string
}

var Config AppConfig

const J_API_KEY = "JELLYFIN_API_KEY"
const J_BASE_URL = "JELLYFIN_BASE_URL"

func LoadConfig() {
	// Load jellyfin settings from .env and others into memory at the entrypoint of the app
	err := godotenv.Load()
	if err != nil {
		slog.Warn(".env file not found")
		slog.Warn(fmt.Sprintf("Make sure %s and %s are set in running environment", J_API_KEY, J_BASE_URL))
	}

	Config.JellyfinApiKey = os.Getenv(J_API_KEY)
	Config.BaseUrl = os.Getenv(J_BASE_URL)

	slog.Info(fmt.Sprintf("%+v", Config))
}
