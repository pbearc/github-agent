package services

import (
	"context"
	"fmt"
	"strings"

	internal_github "github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// CodeSearchService handles code searching
type CodeSearchService struct {
	githubClient *internal_github.Client
	llmClient    *llm.GeminiClient
	logger       *common.Logger
}

// NewCodeSearchService creates a new CodeSearchService instance
func NewCodeSearchService(githubClient *internal_github.Client, llmClient *llm.GeminiClient) *CodeSearchService {
	return &CodeSearchService{
		githubClient: githubClient,
		llmClient:    llmClient,
		logger:       common.NewLogger(),
	}
}

// SearchCode searches for code in a repository
func (s *CodeSearchService) SearchCode(ctx context.Context, owner, repo, branch, query string) (*models.CodeSearchResponse, error) {
	// Search code
	codeResults, err := s.githubClient.SearchCode(ctx, owner, repo, query)
	if err != nil {
		return nil, common.WrapError(err, "failed to search code")
	}

	// Convert to response format
	var searchItems []models.SearchItem
	var searchResultsStr strings.Builder
	
	// Process each code result
	for _, item := range codeResults {
		// Skip nil items or items without required fields
		if item == nil || item.Path == nil || item.HTMLURL == nil {
			continue
		}
		
		// Get file content for each result
		fileContent, err := s.githubClient.GetFileContentText(ctx, owner, repo, *item.Path, branch)
		if err != nil {
			s.logger.WithField("error", err).Warning("Failed to get file content for search result")
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

	// Use LLM to analyze search results if there are any
	var analysis string
	if len(searchItems) > 0 {
		analysis, err = s.llmClient.GenerateText(ctx, buildCodeSearchPrompt(query, searchResultsStr.String()))
		if err != nil {
			s.logger.WithField("error", err).Warning("Failed to generate analysis for search results")
		}
	}

	response := &models.CodeSearchResponse{
		Query:    query,
		Results:  searchItems,
		Analysis: analysis,
	}

	return response, nil
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