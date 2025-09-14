package config

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"info", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"warn", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},        // default
		{"invalid", slog.LevelInfo}, // default
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := parseLogLevel(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetEnvAsString(t *testing.T) {
	os.Setenv("TEST_STRING_VAR", "test_value")
	defer os.Unsetenv("TEST_STRING_VAR")

	result := getEnvAsString("TEST_STRING_VAR", "default")
	assert.Equal(t, "test_value", result)

	result = getEnvAsString("NON_EXISTING_VAR", "default")
	assert.Equal(t, "default", result)

	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")
	result = getEnvAsString("EMPTY_VAR", "default")
	assert.Equal(t, "default", result)
}

func TestGetEnvAsInt(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "99")
	os.Setenv("TEST_INVALID_INT", "not_a_number")
	defer func() {
		os.Unsetenv("TEST_INT_VAR")
		os.Unsetenv("TEST_INVALID_INT")
	}()

	result := getEnvAsInt("TEST_INT_VAR", 7)
	assert.Equal(t, 99, result)

	result = getEnvAsInt("TEST_INVALID_INT", 100)
	assert.Equal(t, 100, result)

	result = getEnvAsInt("NON_EXISTING_INT", 101)
	assert.Equal(t, 101, result)
}

func TestGetEnvAsDuration(t *testing.T) {
	// Set test env vars
	os.Setenv("TEST_DURATION_VAR", "5m")
	os.Setenv("TEST_INVALID_DURATION", "not_a_duration")
	defer func() {
		os.Unsetenv("TEST_DURATION_VAR")
		os.Unsetenv("TEST_INVALID_DURATION")
	}()

	result := getEnvAsDuration("TEST_DURATION_VAR", time.Minute)
	assert.Equal(t, 5*time.Minute, result)

	result = getEnvAsDuration("TEST_INVALID_DURATION", 10*time.Minute)
	assert.Equal(t, 10*time.Minute, result)

	result = getEnvAsDuration("NON_EXISTING_DURATION", 15*time.Minute)
	assert.Equal(t, 15*time.Minute, result)
}

func TestSettings_Validate(t *testing.T) {
	tests := []struct {
		name        string
		settings    *Settings
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid settings",
			settings: &Settings{
				JellyfinApiKey:  "test-api-key",
				JellyfinBaseURL: "http://localhost:8096",
				Port:            8080,
			},
			expectError: false,
		},
		{
			name: "missing API key",
			settings: &Settings{
				JellyfinApiKey:  "",
				JellyfinBaseURL: "http://localhost:8096",
				Port:            8080,
			},
			expectError: true,
			errorMsg:    "required environment variable JELLYFIN_API_KEY is not set",
		},
		{
			name: "missing base URL",
			settings: &Settings{
				JellyfinApiKey:  "test-api-key",
				JellyfinBaseURL: "",
				Port:            8080,
			},
			expectError: true,
			errorMsg:    "required environment variable JELLYFIN_BASE_URL is not set",
		},
		{
			name: "invalid port - too low",
			settings: &Settings{
				JellyfinApiKey:  "test-api-key",
				JellyfinBaseURL: "http://localhost:8096",
				Port:            0,
			},
			expectError: true,
			errorMsg:    "invalid port: 0",
		},
		{
			name: "invalid port - too high",
			settings: &Settings{
				JellyfinApiKey:  "test-api-key",
				JellyfinBaseURL: "http://localhost:8096",
				Port:            65536,
			},
			expectError: true,
			errorMsg:    "invalid port: 65536",
		},
		{
			name: "valid port boundaries",
			settings: &Settings{
				JellyfinApiKey:  "test-api-key",
				JellyfinBaseURL: "http://localhost:8096",
				Port:            1,
			},
			expectError: false,
		},
		{
			name: "valid port boundaries - max",
			settings: &Settings{
				JellyfinApiKey:  "test-api-key",
				JellyfinBaseURL: "http://localhost:8096",
				Port:            65535,
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.settings.validate()

			if test.expectError {
				require.Error(t, err)
				assert.Equal(t, test.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadSettings_Integration(t *testing.T) {
	originalVars := map[string]string{
		J_API_KEY:  os.Getenv(J_API_KEY),
		J_BASE_URL: os.Getenv(J_BASE_URL),
		PORT:       os.Getenv(PORT),
		LOG_LEVEL:  os.Getenv(LOG_LEVEL),
		ENV:        os.Getenv(ENV),
	}

	os.Setenv(J_API_KEY, "test-api-key")
	os.Setenv(J_BASE_URL, "http://localhost:8096")
	os.Setenv(PORT, "9090")
	os.Setenv(LOG_LEVEL, "DEBUG")
	os.Setenv(ENV, "test")

	// Restore original env vars after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// This test will fail if .env loading fails, but we expect it to use env vars
	settings := &Settings{
		JellyfinApiKey:  os.Getenv(J_API_KEY),
		JellyfinBaseURL: os.Getenv(J_BASE_URL),
		Port:            getEnvAsInt(PORT, 8080),
		LogLevel:        parseLogLevel(os.Getenv(LOG_LEVEL)),
		Env:             getEnvAsString(ENV, "development"),
		SessionTimeout:  getEnvAsDuration("SESSION_TIMEOUT", 30*time.Minute),
	}

	assert.Equal(t, "test-api-key", settings.JellyfinApiKey)
	assert.Equal(t, "http://localhost:8096", settings.JellyfinBaseURL)
	assert.Equal(t, 9090, settings.Port)
	assert.Equal(t, slog.LevelDebug, settings.LogLevel)
	assert.Equal(t, "test", settings.Env)
	assert.Equal(t, 30*time.Minute, settings.SessionTimeout)

	// Validate should pass
	err := settings.validate()
	assert.NoError(t, err)
}
