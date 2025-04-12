// internal/llm/prompt_router.go
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// QuestionRouterResponse represents the response from the question router
type QuestionRouterResponse struct {
	APIType     string   `json:"api_type"`
	Explanation string   `json:"explanation"`
	Keywords    []string `json:"keywords"`
}

// RouterAPIType represents the type of GitHub API to use
const (
	APITypeCodeSearch = "code_search"
	APITypeCommits    = "commits"
	APITypePulls      = "pulls"
	APITypeIssues     = "issues"
	APITypeReleases   = "releases"
	APITypeStats      = "stats"
	APITypeUsers      = "users"
	APITypeRepos      = "repos"
)

// buildQuestionRouterPrompt builds a prompt for routing a question to the appropriate API
func buildQuestionRouterPrompt(question string) string {
	return fmt.Sprintf(`
You are an API router for GitHub-related questions. Your task is to analyze a user's question and determine which GitHub API it relates to.

User Question: %s

Choose the most appropriate GitHub API from the following options:
1. code_search - For questions about the codebase, code structure, how specific features are implemented, or any question requiring examination of source code
2. commits - For questions about commit history, specific commits, or authors of changes
3. pulls - For questions about pull requests, reviews, or merge status
4. issues - For questions about issues, bug reports, or feature requests
5. releases - For questions about software releases, versions, or release notes
6. stats - For questions about repository statistics, contributor activities, or code frequency
7. users - For questions about GitHub users, their contributions, or profiles
8. repos - For questions about repository metadata, settings, or general information

Analyze the question carefully and extract 2-5 keywords that would be useful for searching or filtering with the chosen API.

Return your response as a JSON object with these fields:
{
  "api_type": "THE_CHOSEN_API_TYPE",
  "explanation": "A brief explanation of why this API is most appropriate",
  "keywords": ["keyword1", "keyword2", "keyword3"] 
}

Make sure the api_type is exactly one of: code_search, commits, pulls, issues, releases, stats, users, repos

For different API types, focus on extracting these types of keywords:
- code_search: Technical terms, library names, function names, implementation concepts
- commits: Commit messages, authors, date ranges, feature names
- pulls: PR titles, status (open/closed/merged), authors, reviewers
- issues: Issue titles, labels, status (open/closed), assignees
- releases: Version numbers, release names, features, milestones
- stats: Specific metrics, time periods, contributors
- users: Usernames, roles, contribution types
- repos: Repository attributes, settings, configurations

IMPORTANT: Always return a valid JSON object with all three fields: api_type, explanation, and keywords.
`, question)
}

// RouteQuestion determines which GitHub API is most appropriate for a question
func (c *GeminiClient) RouteQuestion(ctx context.Context, question string) (*QuestionRouterResponse, error) {
	prompt := buildQuestionRouterPrompt(question)
	
	responseText, err := c.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to route question: %w", err)
	}
	
	// Parse JSON response
	var response QuestionRouterResponse
	err = json.Unmarshal([]byte(responseText), &response)
	if err != nil {
		// Try to extract structured data from unstructured response
		extractedResponse := extractRouterResponse(responseText, question)
		return &extractedResponse, nil
	}
	
	// Ensure we have keywords
	if response.Keywords == nil || len(response.Keywords) == 0 {
		// Extract keywords from the question if none were provided
		response.Keywords = extractKeywordsFromQuestion(question)
	}
	
	return &response, nil
}

// extractRouterResponse attempts to extract a structured response from text
func extractRouterResponse(text string, question string) QuestionRouterResponse {
	response := QuestionRouterResponse{
		APIType:     APITypeCodeSearch, // Default to code search
		Explanation: "Fallback to code search due to parsing issues",
		Keywords:    extractKeywordsFromQuestion(question),
	}
	
	// Look for API type indicators
	apiTypeMap := map[string]string{
		"code search": APITypeCodeSearch,
		"codebase":    APITypeCodeSearch,
		"source code": APITypeCodeSearch,
		"implement":   APITypeCodeSearch,
		"commit":      APITypeCommits,
		"pull request": APITypePulls,
		"pr":          APITypePulls,
		"issue":       APITypeIssues,
		"bug":         APITypeIssues,
		"release":     APITypeReleases,
		"version":     APITypeReleases,
		"statistic":   APITypeStats,
		"contributor": APITypeStats,
		"user":        APITypeUsers,
		"profile":     APITypeUsers,
		"repository info": APITypeRepos,
		"repo setting":    APITypeRepos,
	}
	
	// Try to find API type in text
	for keyword, apiType := range apiTypeMap {
		if containsIgnoreCase(text, keyword) {
			response.APIType = apiType
			response.Explanation = fmt.Sprintf("Detected reference to %s", keyword)
			break
		}
	}
	
	// Try to extract keywords from JSON-like structures
	keywordsRegex := `"keywords"\s*:\s*\[(.*?)\]`
	matches := regexp.MustCompile(keywordsRegex).FindStringSubmatch(text)
	if len(matches) > 1 {
		keywordsStr := matches[1]
		// Remove quotes and split by commas
		keywordsStr = strings.ReplaceAll(keywordsStr, "\"", "")
		keywordsStr = strings.ReplaceAll(keywordsStr, "'", "")
		keywords := strings.Split(keywordsStr, ",")
		
		// Clean up keywords
		for i, k := range keywords {
			keywords[i] = strings.TrimSpace(k)
		}
		
		// Filter out empty keywords
		var filteredKeywords []string
		for _, k := range keywords {
			if k != "" {
				filteredKeywords = append(filteredKeywords, k)
			}
		}
		
		if len(filteredKeywords) > 0 {
			response.Keywords = filteredKeywords
		}
	}
	
	return response
}

// extractKeywordsFromQuestion extracts keywords from a question as a fallback
func extractKeywordsFromQuestion(question string) []string {
	// Remove common words
	commonWords := map[string]bool{
		"a": true, "an": true, "the": true, "this": true, "that": true,
		"is": true, "are": true, "am": true, "was": true, "were": true,
		"be": true, "being": true, "been": true,
		"have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true,
		"in": true, "on": true, "at": true, "by": true, "to": true,
		"for": true, "with": true, "about": true, "from": true,
		"how": true, "what": true, "when": true, "where": true, "who": true, "why": true,
		"and": true, "or": true, "but": true, "if": true, "then": true,
		"no": true, "not": true, "all": true, "any": true, "each": true,
		"you": true, "your": true, "can": true, "could": true, "would": true, "should": true,
	}
	
	// Split words
	words := strings.Fields(strings.ToLower(question))
	
	// Filter out common words and short words
	var keywords []string
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,;:!?\"'()[]{}")
		
		if !commonWords[word] && len(word) > 2 {
			keywords = append(keywords, word)
		}
	}
	
	// Limit to top 5 keywords
	if len(keywords) > 5 {
		keywords = keywords[:5]
	}
	
	return keywords
}

// containsIgnoreCase checks if a string contains a substring (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}