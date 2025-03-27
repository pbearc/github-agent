package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pbearc/github-agent/backend/pkg/common"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port        int
	Environment string

	// GitHub configuration
	GitHubToken string

	// Gemini configuration
	GeminiAPIKey string
	GeminiModel  string
}

// New creates a new Config instance from environment variables
func New() (*Config, error) {
	port, err := strconv.Atoi(getEnvOrDefault("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return nil, common.NewError("GITHUB_TOKEN environment variable is required")
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		return nil, common.NewError("GEMINI_API_KEY environment variable is required")
	}

	return &Config{
		Port:         port,
		Environment:  getEnvOrDefault("ENVIRONMENT", "development"),
		GitHubToken:  githubToken,
		GeminiAPIKey: geminiAPIKey,
		GeminiModel:  getEnvOrDefault("GEMINI_MODEL", "gemini-1.5-pro"),
	}, nil
}

// getEnvOrDefault retrieves an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}