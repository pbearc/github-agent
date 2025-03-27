package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/models"
)

// GenerateReadme handles README generation requests
func (h *Handler) GenerateReadme(c *gin.Context) {
	var req models.GenerateReadmeRequest
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

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Get repository info
	repoInfo, err := h.GithubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get repository info",
			Details: err.Error(),
		})
		return
	}

	// Convert repository info to a map for the LLM
	repoInfoMap := map[string]interface{}{
		"owner":          repoInfo.Owner,
		"name":           repoInfo.Name,
		"description":    repoInfo.Description,
		"default_branch": repoInfo.DefaultBranch,
		"url":            repoInfo.URL,
		"language":       repoInfo.Language,
		"stars":          repoInfo.Stars,
		"forks":          repoInfo.Forks,
		"has_readme":     repoInfo.HasReadme,
	}

	// Get file list if requested
	var files []string
	if req.IncludeFiles {
		// Use the specified branch or default
		branch := req.Branch
		if branch == "" {
			branch = repoInfo.DefaultBranch
		}

		// Get repository structure
		repoStructure, err := h.GithubClient.GetRepositoryStructure(ctx, owner, repo, branch)
		if err != nil {
			h.Logger.WithField("error", err).Warning("Failed to get repository structure")
		} else {
			files = splitLines(repoStructure)
		}
	}

	// Generate README
	readmeContent, err := h.LLMClient.GenerateReadme(ctx, repoInfoMap, files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate README",
			Details: err.Error(),
		})
		return
	}

	response := models.GenerateResponse{
		Content: readmeContent,
	}

	c.JSON(http.StatusOK, response)
}

// GenerateDockerfile handles Dockerfile generation requests
func (h *Handler) GenerateDockerfile(c *gin.Context) {
	var req models.GenerateDockerfileRequest
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

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Get repository info
	repoInfo, err := h.GithubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get repository info",
			Details: err.Error(),
		})
		return
	}

	// Use provided language or fallback to repo language
	language := req.Language
	if language == "" {
		language = repoInfo.Language
	}

	// Convert repository info to a map for the LLM
	repoInfoMap := map[string]interface{}{
		"owner":          repoInfo.Owner,
		"name":           repoInfo.Name,
		"description":    repoInfo.Description,
		"default_branch": repoInfo.DefaultBranch,
		"url":            repoInfo.URL,
		"language":       repoInfo.Language,
	}

	// Generate Dockerfile
	dockerfileContent, err := h.LLMClient.GenerateDockerfile(ctx, repoInfoMap, language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate Dockerfile",
			Details: err.Error(),
		})
		return
	}

	response := models.GenerateResponse{
		Content: dockerfileContent,
	}

	c.JSON(http.StatusOK, response)
}

// GenerateComments handles code commenting requests
func (h *Handler) GenerateComments(c *gin.Context) {
	var req models.CodeCommentsRequest
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
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		repoInfo, err := h.GithubClient.GetRepositoryInfo(ctx, owner, repo)
		if err != nil {
			branch = "main" // Fallback to main if unable to determine
		} else {
			branch = repoInfo.DefaultBranch
		}
	}

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Get file content
	fileContent, err := h.GithubClient.GetFileContentText(ctx, owner, repo, req.FilePath, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get file content",
			Details: err.Error(),
		})
		return
	}

	// Detect language from file extension
	language := detectLanguageFromPath(req.FilePath)

	// Generate comments
	commentedCode, err := h.LLMClient.GenerateCodeComments(ctx, fileContent.Content, language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate comments",
			Details: err.Error(),
		})
		return
	}

	response := models.GenerateResponse{
		Content: commentedCode,
	}

	c.JSON(http.StatusOK, response)
}

// RefactorCode handles code refactoring requests
func (h *Handler) RefactorCode(c *gin.Context) {
	var req models.CodeRefactorRequest
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
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		repoInfo, err := h.GithubClient.GetRepositoryInfo(ctx, owner, repo)
		if err != nil {
			branch = "main" // Fallback to main if unable to determine
		} else {
			branch = repoInfo.DefaultBranch
		}
	}

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Get file content
	fileContent, err := h.GithubClient.GetFileContentText(ctx, owner, repo, req.FilePath, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get file content",
			Details: err.Error(),
		})
		return
	}

	// Detect language from file extension
	language := detectLanguageFromPath(req.FilePath)

	// Generate refactored code
	refactoredCode, err := h.LLMClient.GenerateCodeRefactor(ctx, fileContent.Content, language, req.Instructions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to refactor code",
			Details: err.Error(),
		})
		return
	}

	response := models.GenerateResponse{
		Content: refactoredCode,
	}

	c.JSON(http.StatusOK, response)
}

// ProcessLLMOperation handles generic LLM operation requests
func (h *Handler) ProcessLLMOperation(c *gin.Context) {
	var req models.LLMOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	// Validate the operation
	if err := llm.ValidateOperation(&req.Operation); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid operation",
			Details: err.Error(),
		})
		return
	}

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Process the operation
	result, err := h.LLMClient.ProcessOperation(ctx, &req.Operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to process operation",
			Details: err.Error(),
		})
		return
	}

	response := models.GenerateResponse{
		Content: result,
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions

// detectLanguageFromPath detects the programming language from a file path
func detectLanguageFromPath(path string) string {
	// Extract extension from path
	ext := ""
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			ext = path[i+1:]
			break
		}
		if path[i] == '/' {
			break
		}
	}

	// Map extension to language
	switch ext {
	case "py":
		return "python"
	case "js":
		return "javascript"
	case "ts":
		return "typescript"
	case "go":
		return "go"
	case "java":
		return "java"
	case "c":
		return "c"
	case "cpp", "cc", "cxx":
		return "c++"
	case "cs":
		return "c#"
	case "rb":
		return "ruby"
	case "php":
		return "php"
	case "swift":
		return "swift"
	case "kt", "kts":
		return "kotlin"
	case "rs":
		return "rust"
	case "sh":
		return "shell"
	case "html":
		return "html"
	case "css":
		return "css"
	case "md":
		return "markdown"
	default:
		return "unknown"
	}
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	var line []rune

	for _, r := range s {
		if r == '\n' {
			lines = append(lines, string(line))
			line = line[:0]
		} else {
			line = append(line, r)
		}
	}

	if len(line) > 0 {
		lines = append(lines, string(line))
	}

	return lines
}