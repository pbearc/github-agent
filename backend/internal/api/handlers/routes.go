package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pbearc/github-agent/backend/internal/config"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/graph"
	"github.com/pbearc/github-agent/backend/internal/llm"
)

// SetupRoutes sets up all API routes
func SetupRoutes(router *gin.Engine, githubClient *github.Client, llmClient *llm.GeminiClient, neo4jClient *graph.Neo4jClient, cfg *config.Config) {
    handler := NewHandler(githubClient, llmClient, neo4jClient, cfg)

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

        // Code search routes
        search := api.Group("/search")
        {
            search.POST("/code", handler.SearchCode)
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
            navigate.POST("/walkthrough", handler.GenerateCodeWalkthrough)
            navigate.POST("/function", handler.ExplainFunction)
            navigate.POST("/architecture", handler.VisualizeArchitecture)
            navigate.POST("/architecture-graph", handler.GetArchitectureGraph)
            navigate.POST("/explain-architecture", handler.ExplainArchitectureGraph)

        }

        // Push routes
        push := api.Group("/push")
        {
            push.POST("/file", handler.PushFile)
        }

        pr := api.Group("/pr")
        {
            pr.POST("/summary", handler.GetPRSummary)
        }

        llmNavigate := api.Group("/llm-navigate")
        {
            llmNavigate.POST("/index", handler.IndexCodebaseForNavigation)
            llmNavigate.POST("/question", handler.NavigateCodebaseWithLLM)
        }

        // Smart Navigation
        api.POST("/smart-navigate", handler.SmartNavigate)

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