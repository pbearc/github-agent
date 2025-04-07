package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/internal/pinecone"
	"github.com/pbearc/github-agent/backend/internal/services"
)

// IndexCodebase handles codebase indexing requests
func (h *Handler) IndexCodebase(c *gin.Context) {
	var req models.CodebaseIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	// Parse the GitHub URL
	owner, repo, err := github.ParseRepoURL(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid GitHub URL",
			Details: err.Error(),
		})
		return
	}

	// Use the specified branch or default to main
	branch := req.Branch
	if branch == "" {
		// Get repository info to determine the default branch
		ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
		defer cancel()
		
		repoInfo, err := h.GithubClient.GetRepositoryInfo(ctx, owner, repo)
		if err != nil {
			branch = "main" // Fallback to main if unable to determine
		} else {
			branch = repoInfo.DefaultBranch
		}
	}

	// Set a timeout for the operation (indexing can take time)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 600*time.Second) // 10 minutes
	defer cancel()

	// Create pinecone client
	pineconeClient, err := pinecone.NewClient(
		h.Config.PineconeAPIKey,
		h.Config.PineconeEnvironment,
		h.Config.PineconeIndexName,
		h.LLMClient.GetEmbeddingDimension(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create Pinecone client",
			Details: err.Error(),
		})
		return
	}

	// Ensure index exists
	err = pineconeClient.EnsureIndex(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to ensure Pinecone index exists",
			Details: err.Error(),
		})
		return
	}

	// Create indexer service
	indexerService := services.NewIndexerService(h.GithubClient, pineconeClient, h.LLMClient)

	// Index repository
	namespace, err := indexerService.IndexRepository(ctx, owner, repo, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to index repository",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.CodebaseIndexResponse{
		Message:   "Repository indexed successfully",
		Status:    "completed",
		Namespace: namespace,
	})
}

// NavigateCodebase handles codebase Q&A requests
func (h *Handler) NavigateCodebase(c *gin.Context) {
	var req models.CodebaseNavigatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	// Parse the GitHub URL
	owner, repo, err := github.ParseRepoURL(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid GitHub URL",
			Details: err.Error(),
		})
		return
	}

	// Use the specified branch or default to main
	branch := req.Branch
	if branch == "" {
		// Get repository info to determine the default branch
		ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
		defer cancel()
		
		repoInfo, err := h.GithubClient.GetRepositoryInfo(ctx, owner, repo)
		if err != nil {
			branch = "main" // Fallback to main if unable to determine
		} else {
			branch = repoInfo.DefaultBranch
		}
	}

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	// Create pinecone client
	pineconeClient, err := pinecone.NewClient(
		h.Config.PineconeAPIKey,
		h.Config.PineconeEnvironment,
		h.Config.PineconeIndexName,
		h.LLMClient.GetEmbeddingDimension(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create Pinecone client",
			Details: err.Error(),
		})
		return
	}

	// Create navigator service
	navigatorService := services.NewNavigatorService(
		h.GithubClient,
		pineconeClient,
		h.LLMClient,
	)

	// Answer the question
	response, err := navigatorService.AnswerQuestion(ctx, owner, repo, branch, req.Question, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to answer question",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}