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

// GetPRSummary handles pull request summary requests
func (h *Handler) GetPRSummary(c *gin.Context) {
	var req models.PRSummaryRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	// Parse the GitHub PR URL
	owner, repo, prNumber, err := github.ParsePullRequestURL(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid GitHub pull request URL",
			Details: err.Error(),
		})
		return
	}

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	// Create PR summary service
	prSummaryService := services.NewPRSummaryService(h.GithubClient, h.LLMClient)

	// Generate summary
	summary, err := prSummaryService.GenerateSummary(ctx, owner, repo, prNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate pull request summary",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PRSummaryResponse{
		Summary: *summary,
	})
}