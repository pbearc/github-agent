package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pbearc/github-agent/backend/internal/api/handlers"
	"github.com/pbearc/github-agent/backend/internal/api/middleware"
	"github.com/pbearc/github-agent/backend/internal/config"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/graph"
	"github.com/pbearc/github-agent/backend/internal/llm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Set up Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORS())
	
	// Initialize clients
	githubClient, err := github.NewClient(cfg.GitHubToken)
	if err != nil {
		log.Fatalf("Failed to initialize GitHub client: %v", err)
	}

	llmClient, err := llm.NewGeminiClient(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini client: %v", err)
	}

	// Initialize Neo4j client
	neo4jClient, err := graph.NewNeo4jClient(cfg.Neo4jURI, cfg.Neo4jUsername, cfg.Neo4jPassword)
	if err != nil {
		log.Printf("Warning: Failed to initialize Neo4j client: %v", err)
		log.Printf("Continuing without Neo4j support")
		neo4jClient = nil
	} else {
		log.Println("Successfully connected to Neo4j database")
	}

	// Set up API handlers
	handlers.SetupRoutes(router, githubClient, llmClient, neo4jClient, cfg)
	
	// Set up the server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// Close Neo4j connection before shutting down
	if neo4jClient != nil {
		if err := neo4jClient.Close(); err != nil {
			log.Printf("Error closing Neo4j connection: %v", err)
		}
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}