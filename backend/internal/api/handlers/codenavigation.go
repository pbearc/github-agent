// internal/api/handlers/codenavigation.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pbearc/github-agent/backend/internal/config"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/graph"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/internal/services"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

type Handler struct {
    GithubClient *github.Client
    LLMClient    *llm.GeminiClient
    Neo4jClient  *graph.Neo4jClient
    Config       *config.Config
    Logger       *common.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(githubClient *github.Client, llmClient *llm.GeminiClient, neo4jClient *graph.Neo4jClient, cfg *config.Config) *Handler {
    return &Handler{
        GithubClient: githubClient,
        LLMClient:    llmClient,
        Neo4jClient:  neo4jClient,
        Config:       cfg,
        Logger:       common.NewLogger(),
    }
}

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
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient, h.Neo4jClient)

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

    

    response := gin.H{
        "owner":          repoInfo.Owner,
        "name":           repoInfo.Name,
        "structure":      structure,
        "walkthrough":    walkthrough,
        "architecture":   architecture,

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
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient, h.Neo4jClient)

    // Get answer to the question
    answer, err := navigationService.AnswerCodebaseQuestion(
        ctx,
        owner,
        repo,
        req.Branch,
        req.Question,
        nil,
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
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient, h.Neo4jClient)

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
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient, h.Neo4jClient)

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

// GetArchitectureGraph handles architecture graph data requests
func (h *Handler) GetArchitectureGraph(c *gin.Context) {
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

    // Check if Neo4j client is available
    if h.Neo4jClient == nil {
        c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
            Error: "Neo4j service not available",
            Details: "The Neo4j database is not configured or not accessible",
        })
        return
    }

    // Set a timeout for the GitHub API request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
    defer cancel()

    // Create the navigation service
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient, h.Neo4jClient)
    
    // Store the codebase structure in Neo4j
    err = navigationService.StoreCodebaseInNeo4j(ctx, owner, repo, req.Branch)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to store codebase structure",
            Details: err.Error(),
        })
        return
    }

    // Get graph data
    graphData, err := h.Neo4jClient.GetCodebaseGraph(ctx, owner, repo, req.Branch)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Failed to get architecture graph",
            Details: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, graphData)
}

// ExplainArchitectureGraph generates an explanation of the architecture graph
func (h *Handler) ExplainArchitectureGraph(c *gin.Context) {
    var req struct {
        URL    string `json:"url" binding:"required"`
        Branch string `json:"branch"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request: " + err.Error(),
        })
        return
    }

    // Parse the GitHub URL
    owner, repo, err := github.ParseRepoURL(req.URL)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid GitHub URL: " + err.Error(),
        })
        return
    }

    // Set a timeout for the request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
    defer cancel()

    // Get architecture graph data
    graphData, err := h.Neo4jClient.GetCodebaseGraph(ctx, owner, repo, req.Branch)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get architecture graph: " + err.Error(),
        })
        return
    }

    // Convert graphData to a format suitable for the LLM
    graphDataJSON, err := json.Marshal(graphData)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to process graph data: " + err.Error(),
        })
        return
    }

    // Create a prompt for the LLM
    prompt := fmt.Sprintf(`
You are an expert software architect. Analyze this codebase structure and provide a clear explanation.

Here is the graph data representing the codebase structure:
%s

Please provide:
1. A high-level overview of the architecture
2. Identification of key files and their roles in the system
3. Explanation of important relationships between files
4. Any notable patterns or architectural decisions evident from the graph

Your response should be well-structured and easy to read when rendered as Markdown.
`, string(graphDataJSON))

    // Generate explanation using LLM
    explanation, err := h.LLMClient.GenerateCompletion(ctx, prompt, 0.7, 1024)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to generate architecture explanation: " + err.Error(),
        })
        return
    }

    // Return both the graph data and the explanation
    c.JSON(http.StatusOK, gin.H{
        "graph": graphData,
        "explanation": explanation,
    })
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
    navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient, h.Neo4jClient)

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
