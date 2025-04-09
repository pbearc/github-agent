// internal/api/handlers/codenavigation.go
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/internal/services"
)

// IndexCodebase handles codebase indexing requests
func (h *Handler) IndexCodebaseForNavigation(c *gin.Context) {
    var req models.RepositoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid request",
            Details: err.Error(),
        })
        return
    }

    // Parse the GitHub URL
    owner, repo, err := github.ParseRepoURL(req.URL)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid GitHub URL",
            Details: err.Error(),
        })
        return
    }

    // Set a timeout for the GitHub API request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 600*time.Second)
    defer cancel()

    // Get repository info
    repoInfo, err := h.GithubClient.GetRepositoryInfo(ctx, owner, repo)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to get repository info",
            Details: err.Error(),
        })
        return
    }

    // Get repository structure
    structure, err := h.GithubClient.GetRepositoryStructure(ctx, owner, repo, req.Branch)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to get repository structure",
            Details: err.Error(),
        })
        return
    }

    // Create the navigation service
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient)

    // Generate default walkthrough
    walkthrough, err := navigationService.GenerateCodeWalkthrough(
        ctx, 
        owner, 
        repo, 
        req.Branch,
        3, // Default depth
        "", // No specific focus path
        []string{}, // Auto-detect entry points
    )
    if err != nil {
        h.Logger.WithField("error", err).Warning("Failed to generate walkthrough")
    }

    // Generate architecture visualization
    architecture, err := navigationService.VisualizeArchitecture(
        ctx,
        owner,
        repo,
        req.Branch,
        "medium", // Default detail level
        []string{}, // No specific focus paths
    )
    if err != nil {
        h.Logger.WithField("error", err).Warning("Failed to visualize architecture")
    }

    // Generate best practices guide
    bestPractices, err := navigationService.GenerateBestPracticesGuide(
        ctx,
        owner,
        repo,
        req.Branch,
        "full", // Full repository scope
        "", // No specific path
    )
    if err != nil {
        h.Logger.WithField("error", err).Warning("Failed to generate best practices guide")
    }

    response := gin.H{
        "owner":          repoInfo.Owner,
        "name":           repoInfo.Name,
        "structure":      structure,
        "walkthrough":    walkthrough,
        "architecture":   architecture,
        "best_practices": bestPractices,
    }

    c.JSON(http.StatusOK, response)
}

// NavigateCodebase handles codebase Q&A requests
func (h *Handler) NavigateCodebaseWithLLM(c *gin.Context) {
    var req struct {
        models.RepositoryRequest
        Question string `json:"question" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid request",
            Details: err.Error(),
        })
        return
    }

    // Parse the GitHub URL
    owner, repo, err := github.ParseRepoURL(req.URL)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid GitHub URL",
            Details: err.Error(),
        })
        return
    }

    // Set a timeout for the GitHub API request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
    defer cancel()

    // Create the navigation service
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient)

    // Get answer to the question
    answer, err := navigationService.AnswerCodebaseQuestion(
        ctx,
        owner,
        repo,
        req.Branch,
        req.Question,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to answer question",
            Details: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, answer)
}

// GenerateCodeWalkthrough handles code walkthrough generation requests
func (h *Handler) GenerateCodeWalkthrough(c *gin.Context) {
    var req models.CodeWalkthroughRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid request",
            Details: err.Error(),
        })
        return
    }

    // Parse the GitHub URL
    owner, repo, err := github.ParseRepoURL(req.URL)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid GitHub URL",
            Details: err.Error(),
        })
        return
    }

    // Set a timeout for the GitHub API request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
    defer cancel()

    // Create the navigation service
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient)

    // Generate code walkthrough
    walkthrough, err := navigationService.GenerateCodeWalkthrough(
        ctx,
        owner,
        repo,
        req.Branch,
        req.Depth,
        req.FocusPath,
        req.EntryPoints,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to generate code walkthrough",
            Details: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, walkthrough)
}

// ExplainFunction handles function explanation requests
func (h *Handler) ExplainFunction(c *gin.Context) {
    var req models.FunctionExplainerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid request",
            Details: err.Error(),
        })
        return
    }

    // Parse the GitHub URL
    owner, repo, err := github.ParseRepoURL(req.URL)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid GitHub URL",
            Details: err.Error(),
        })
        return
    }

    // Set a timeout for the GitHub API request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
    defer cancel()

    // Create the navigation service
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient)

    // Generate function explanation
    explanation, err := navigationService.ExplainFunction(
        ctx,
        owner,
        repo,
        req.Branch,
        req.FilePath,
        req.FunctionName,
        req.LineStart,
        req.LineEnd,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to explain function",
            Details: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, explanation)
}

// VisualizeArchitecture handles architecture visualization requests
func (h *Handler) VisualizeArchitecture(c *gin.Context) {
    var req models.ArchitectureVisualizerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid request",
            Details: err.Error(),
        })
        return
    }

    // Parse the GitHub URL
    owner, repo, err := github.ParseRepoURL(req.URL)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid GitHub URL",
            Details: err.Error(),
        })
        return
    }

    // Set a timeout for the GitHub API request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
    defer cancel()

    // Create the navigation service
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient)

    // Generate architecture visualization
    architecture, err := navigationService.VisualizeArchitecture(
        ctx,
        owner,
        repo,
        req.Branch,
        req.Detail,
        req.FocusPaths,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to visualize architecture",
            Details: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, architecture)
}

// GenerateBestPracticesGuide handles best practices guide generation requests
func (h *Handler) GenerateBestPracticesGuide(c *gin.Context) {
    var req models.BestPracticesRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid request",
            Details: err.Error(),
        })
        return
    }

    // Parse the GitHub URL
    owner, repo, err := github.ParseRepoURL(req.URL)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: "Invalid GitHub URL",
            Details: err.Error(),
        })
        return
    }

    // Set a timeout for the GitHub API request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
    defer cancel()

    // Create the navigation service
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient)

    // Generate best practices guide
    bestPractices, err := navigationService.GenerateBestPracticesGuide(
        ctx,
        owner,
        repo,
        req.Branch,
        req.Scope,
        req.Path,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to generate best practices guide",
            Details: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, bestPractices)
}