package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/models"
)

// GetRepositoryInfo handles repository info requests
func (h *Handler) GetRepositoryInfo(c *gin.Context) {
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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
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

	// Get repository languages
	languages, err := h.GithubClient.GetRepositoryLanguages(ctx, owner, repo)
	if err != nil {
		h.Logger.WithField("error", err).Warning("Failed to get repository languages")
	}

	// Convert languages to a slice
	languageSlice := make([]string, 0, len(languages))
	for lang := range languages {
		languageSlice = append(languageSlice, lang)
	}

	response := models.RepositoryInfoResponse{
		Owner:         repoInfo.Owner,
		Name:          repoInfo.Name,
		Description:   repoInfo.Description,
		DefaultBranch: repoInfo.DefaultBranch,
		URL:           repoInfo.URL,
		Language:      repoInfo.Language,
		Languages:     languageSlice,
		Stars:         repoInfo.Stars,
		Forks:         repoInfo.Forks,
		HasReadme:     repoInfo.HasReadme,
	}

	c.JSON(http.StatusOK, response)
}

// GetFileContent handles file content requests
func (h *Handler) GetFileContent(c *gin.Context) {
	var req struct {
		models.RepositoryRequest
		Path string `json:"path" binding:"required"`
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

	// Set a timeout for the GitHub API request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	// Get file content
	fileContent, err := h.GithubClient.GetFileContentText(ctx, owner, repo, req.Path, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to get file content",
			Details: err.Error(),
		})
		return
	}

	response := models.FileContentResponse{
		Path:    fileContent.Path,
		Content: fileContent.Content,
	}

	c.JSON(http.StatusOK, response)
}

// ListFiles handles file listing requests
func (h *Handler) ListFiles(c *gin.Context) {
	var req struct {
		models.RepositoryRequest
		Path string `json:"path"`
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

	// Use default path if not specified
	path := req.Path
	if path == "" {
		path = "/"
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

	// Set a timeout for the GitHub API request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	// List files
	files, err := h.GithubClient.ListFiles(ctx, owner, repo, path, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to list files",
			Details: err.Error(),
		})
		return
	}

	// Convert to response format
	var fileList []models.GitHubFile
	for _, file := range files {
		// Handle nil pointer for DownloadURL
		var downloadURL string
		if file.DownloadURL != nil {
			downloadURL = *file.DownloadURL
		}
		
		fileList = append(fileList, models.GitHubFile{
			Name:        *file.Name,
			Path:        *file.Path,
			Size:        *file.Size,
			Type:        *file.Type,
			HTMLURL:     *file.HTMLURL,
			DownloadURL: downloadURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"files": fileList,
		"path":  path,
	})
}

// SearchCode handles code search requests
func (h *Handler) SearchCode(c *gin.Context) {
	var req models.CodeSearchRequest
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

	// Search code
	results, err := h.GithubClient.SearchCode(ctx, owner, repo, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to search code",
			Details: err.Error(),
		})
		return
	}

	// Convert to response format
	var searchItems []models.SearchItem
	var searchResultsStr strings.Builder
	
	// Check if we got valid results
	if len(results) > 0 && results[0] != nil {
		// Process the items in the result
		for _, item := range results {
			// Skip nil items
			if item == nil || item.Path == nil || item.HTMLURL == nil {
				continue
			}
			
			// Get file content for each result
			fileContent, err := h.GithubClient.GetFileContentText(ctx, owner, repo, *item.Path, req.Branch)
			if err != nil {
				h.Logger.WithField("error", err).Warning("Failed to get file content for search result")
				continue
			}
			
			searchItem := models.SearchItem{
				Path:     *item.Path,
				Content:  fileContent.Content,
				HTMLLink: *item.HTMLURL,
			}
			searchItems = append(searchItems, searchItem)
			
			// Add to string builder for LLM analysis
			searchResultsStr.WriteString("File: ")
			searchResultsStr.WriteString(*item.Path)
			searchResultsStr.WriteString("\n\n")
			searchResultsStr.WriteString(fileContent.Content)
			searchResultsStr.WriteString("\n\n---\n\n")
		}
	}

	// Use LLM to analyze search results if there are any
	var analysis string
	if len(searchItems) > 0 {
		analysis, err = h.LLMClient.GenerateText(ctx, buildCodeSearchPrompt(req.Query, searchResultsStr.String()))
		if err != nil {
			h.Logger.WithField("error", err).Warning("Failed to generate analysis for search results")
		}
	}

	response := models.CodeSearchResponse{
		Query:    req.Query,
		Results:  searchItems,
		Analysis: analysis,
	}

	c.JSON(http.StatusOK, response)
}

// PushFile handles file pushing requests
func (h *Handler) PushFile(c *gin.Context) {
	var req struct {
		models.RepositoryRequest
		Path    string `json:"path" binding:"required"`
		Content string `json:"content" binding:"required"`
		Message string `json:"message"`
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

	// Use default commit message if not provided
	message := req.Message
	if message == "" {
		message = "Update " + req.Path + " via GitHub Agent"
	}

	// Set a timeout for the GitHub API request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	// Push the file
	err = h.GithubClient.CreateOrUpdateFile(ctx, owner, repo, req.Path, message, req.Content, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to push file",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "File pushed successfully",
		Data: gin.H{
			"path":   req.Path,
			"branch": branch,
		},
	})
}

// buildCodeSearchPrompt builds a prompt for analyzing code search results
func buildCodeSearchPrompt(query string, searchResults string) string {
	return fmt.Sprintf(`
You are an expert code analyzer. Your task is to analyze the following search results for the query "%s" and provide insights.

Here are the search results:
%s

Please provide:
1. A summary of what the search results reveal
2. Key code patterns or issues found
3. Suggestions for improvements if applicable
4. Any potential security or performance concerns based on these results

Format your response in a clear, structured manner with headings and bullet points where appropriate.
`, query, searchResults)
}