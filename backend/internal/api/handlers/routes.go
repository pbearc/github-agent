package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pbearc/github-agent/backend/internal/config"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// Handler contains all the dependencies for API handlers
type Handler struct {
	GithubClient *github.Client
	LLMClient    *llm.GeminiClient
	Config       *config.Config
	Logger       *common.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(githubClient *github.Client, llmClient *llm.GeminiClient, cfg *config.Config) *Handler {
	return &Handler{
		GithubClient: githubClient,
		LLMClient:    llmClient,
		Config:       cfg,
		Logger:       common.NewLogger(),
	}
}

// SetupRoutes sets up all API routes
func SetupRoutes(router *gin.Engine, githubClient *github.Client, llmClient *llm.GeminiClient, cfg *config.Config) {
    handler := NewHandler(githubClient, llmClient, cfg)

    // Health check
    router.GET("/health", handler.HealthCheck)

    // API routes
    api := router.Group("/api")
    {
        // Repository routes
        repo := api.Group("/repo")
        {
            repo.POST("/info", handler.GetRepositoryInfo)
            repo.POST("/file", handler.GetFileContent)
            repo.POST("/files", handler.ListFiles)
        }

        // Generation routes
        generate := api.Group("/generate")
        {
            generate.POST("/readme", handler.GenerateReadme)
            generate.POST("/dockerfile", handler.GenerateDockerfile)
            generate.POST("/comments", handler.GenerateComments)
            generate.POST("/refactor", handler.RefactorCode)
        }

        // Navigator routes (replacing search)
        navigate := api.Group("/navigate")
        {
            navigate.POST("/index", handler.IndexCodebase)
            navigate.POST("/question", handler.NavigateCodebase)
        }

        // Push routes
        push := api.Group("/push")
        {
            push.POST("/file", handler.PushFile)
        }

        // LLM operation route
        api.POST("/llm/operation", handler.ProcessLLMOperation)
    }
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
		"service": "github-agent",
	})
}