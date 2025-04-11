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
	
	// Pinecone configuration
	PineconeAPIKey      string
	PineconeEnvironment string
	PineconeIndexName   string
	
	// Neo4j configuration
	Neo4jURI      string
	Neo4jUsername string
	Neo4jPassword string
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
	
	// Pinecone API key is required
	pineconeAPIKey := os.Getenv("PINECONE_API_KEY")
	if pineconeAPIKey == "" {
		return nil, common.NewError("PINECONE_API_KEY environment variable is required")
	}

	return &Config{
		Port:                port,
		Environment:         getEnvOrDefault("ENVIRONMENT", "development"),
		GitHubToken:         githubToken,
		GeminiAPIKey:        geminiAPIKey,
		GeminiModel:         getEnvOrDefault("GEMINI_MODEL", "gemini-1.5-pro"),
		PineconeAPIKey:      pineconeAPIKey,
		PineconeEnvironment: getEnvOrDefault("PINECONE_ENVIRONMENT", "gcp-starter"),
		PineconeIndexName:   getEnvOrDefault("PINECONE_INDEX_NAME", "github-agent"),
		Neo4jURI:            getEnvOrDefault("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUsername:       getEnvOrDefault("NEO4J_USERNAME", "neo4j"),
		Neo4jPassword:       getEnvOrDefault("NEO4J_PASSWORD", "password"),
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