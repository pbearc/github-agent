// internal/api/handlers/smart_navigate.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/internal/services"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// SmartNavigateRequest represents a request to the smart navigation endpoint
type SmartNavigateRequest struct {
	models.RepositoryRequest
	Question string `json:"question" binding:"required"`
}

// SmartNavigateResponse represents a response from the smart navigation endpoint
type SmartNavigateResponse struct {
	Answer            string                  `json:"answer"`
	APIType           string                  `json:"api_type"`
	RelevantFiles     []models.RelevantFile   `json:"relevant_files,omitempty"`
	FollowupQuestions []string                `json:"followup_questions,omitempty"`
	ExtraData         map[string]interface{}  `json:"extra_data,omitempty"`
}

// SmartNavigate handles intelligent routing of questions to the appropriate API
func (h *Handler) SmartNavigate(c *gin.Context) {
	var req SmartNavigateRequest
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

	// Set a timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	// Route the question to the appropriate API
	routerResponse, err := h.LLMClient.RouteQuestion(ctx, req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to route question",
			Details: err.Error(),
		})
		return
	}

	h.Logger.WithField("api_type", routerResponse.APIType).WithField("keywords", routerResponse.Keywords).Info("Routed question")

	// Handle different API types
	var response *SmartNavigateResponse
	
	switch routerResponse.APIType {
	case llm.APITypeCodeSearch:
		// Use existing code navigation system
		resp, err := h.handleCodeSearchQuestion(ctx, owner, repo, req.Branch, req.Question, routerResponse.Keywords)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question with code search",
				Details: err.Error(),
			})
			return
		}
		response = resp

	case llm.APITypeCommits:
		resp, err := h.handleCommitsQuestion(ctx, owner, repo, req.Branch, req.Question, routerResponse.Keywords)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question about commits",
				Details: err.Error(),
			})
			return
		}
		response = resp

	case llm.APITypePulls:
		resp, err := h.handlePullsQuestion(ctx, owner, repo, req.Question, routerResponse.Keywords)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question about pull requests",
				Details: err.Error(),
			})
			return
		}
		response = resp

	case llm.APITypeIssues:
		resp, err := h.handleIssuesQuestion(ctx, owner, repo, req.Question, routerResponse.Keywords)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question about issues",
				Details: err.Error(),
			})
			return
		}
		response = resp

	case llm.APITypeReleases:
		resp, err := h.handleReleasesQuestion(ctx, owner, repo, req.Question, routerResponse.Keywords)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question about releases",
				Details: err.Error(),
			})
			return
		}
		response = resp

	case llm.APITypeStats:
		resp, err := h.handleStatsQuestion(ctx, owner, repo, req.Question, routerResponse.Keywords)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question about repository statistics",
				Details: err.Error(),
			})
			return
		}
		response = resp

	case llm.APITypeUsers:
		resp, err := h.handleUsersQuestion(ctx, owner, repo, req.Question, routerResponse.Keywords)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question about users",
				Details: err.Error(),
			})
			return
		}
		response = resp

	case llm.APITypeRepos:
		resp, err := h.handleReposQuestion(ctx, owner, repo, req.Question, routerResponse.Keywords)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question about repository",
				Details: err.Error(),
			})
			return
		}
		response = resp

	default:
		// Fallback to code search if routing fails
		resp, err := h.handleCodeSearchQuestion(ctx, owner, repo, req.Branch, req.Question, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to answer question",
				Details: err.Error(),
			})
			return
		}
		response = resp
	}

	c.JSON(http.StatusOK, response)
}

// handleCodeSearchQuestion handles questions about code using the existing code navigation system
func (h *Handler) handleCodeSearchQuestion(ctx context.Context, owner, repo, branch, question string, keywords []string) (*SmartNavigateResponse, error) {
	// Create the navigation service
	navigationService := services.NewCodeNavigationService(h.GithubClient, h.LLMClient, h.Neo4jClient)

	// Use the existing AnswerCodebaseQuestion method but pass the keywords if available
	answer, err := navigationService.AnswerCodebaseQuestion(ctx, owner, repo, branch, question, keywords)
	if err != nil {
		return nil, common.WrapError(err, "failed to answer question")
	}

	// Convert to smart navigate response
	return &SmartNavigateResponse{
		Answer:            answer.Answer,
		APIType:           llm.APITypeCodeSearch,
		RelevantFiles:     answer.RelevantFiles,
		FollowupQuestions: answer.FollowupQuestions,
	}, nil
}

// handleCommitsQuestion handles questions about commits
func (h *Handler) handleCommitsQuestion(ctx context.Context, owner, repo, branch, question string, keywords []string) (*SmartNavigateResponse, error) {
	// Get commit data - implement this method in your GitHub client
	commits, err := h.GithubClient.GetCommits(ctx, owner, repo, branch, keywords)
	if err != nil {
		return nil, common.WrapError(err, "failed to get commits")
	}

	// Convert commits to a format suitable for the LLM
	commitsJSON, err := json.Marshal(commits)
	if err != nil {
		return nil, common.WrapError(err, "failed to marshal commits")
	}

	// Generate prompt for LLM
	prompt := fmt.Sprintf(`
You are an expert GitHub analyst. Please answer the following question about repository commits:

Question: %s

Here are the relevant commits:
%s

Please provide:
1. A comprehensive answer to the question
2. Include references to specific commits when relevant
3. Any additional context that would help understand the answer
4. If appropriate, suggest 2-3 follow-up questions

Format your response in a clear, structured manner with markdown formatting.
`, question, string(commitsJSON))

	// Generate response using LLM
	answer, err := h.LLMClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate answer")
	}

	// Extract follow-up questions
	followups := extractFollowupQuestions(answer)

	return &SmartNavigateResponse{
		Answer:            answer,
		APIType:           llm.APITypeCommits,
		FollowupQuestions: followups,
		ExtraData: map[string]interface{}{
			"commits": commits,
		},
	}, nil
}

// handlePullsQuestion handles questions about pull requests
func (h *Handler) handlePullsQuestion(ctx context.Context, owner, repo, question string, keywords []string) (*SmartNavigateResponse, error) {
	// Get pull request data - implement this method in your GitHub client
	pullRequests, err := h.GithubClient.GetPullRequests(ctx, owner, repo, keywords)
	if err != nil {
		return nil, common.WrapError(err, "failed to get pull requests")
	}

	// Convert pull requests to a format suitable for the LLM
	pullsJSON, err := json.Marshal(pullRequests)
	if err != nil {
		return nil, common.WrapError(err, "failed to marshal pull requests")
	}

	// Generate prompt for LLM
	prompt := fmt.Sprintf(`
You are an expert GitHub analyst. Please answer the following question about repository pull requests:

Question: %s

Here are the relevant pull requests:
%s

Please provide:
1. A comprehensive answer to the question
2. Include references to specific pull requests when relevant
3. Any additional context that would help understand the answer
4. If appropriate, suggest 2-3 follow-up questions

Format your response in a clear, structured manner with markdown formatting.
`, question, string(pullsJSON))

	// Generate response using LLM
	answer, err := h.LLMClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate answer")
	}

	// Extract follow-up questions
	followups := extractFollowupQuestions(answer)

	return &SmartNavigateResponse{
		Answer:            answer,
		APIType:           llm.APITypePulls,
		FollowupQuestions: followups,
		ExtraData: map[string]interface{}{
			"pull_requests": pullRequests,
		},
	}, nil
}

// handleIssuesQuestion handles questions about issues
func (h *Handler) handleIssuesQuestion(ctx context.Context, owner, repo, question string, keywords []string) (*SmartNavigateResponse, error) {
	// Get issues data - implement this method in your GitHub client
	issues, err := h.GithubClient.GetIssues(ctx, owner, repo, keywords)
	if err != nil {
		return nil, common.WrapError(err, "failed to get issues")
	}

	// Convert issues to a format suitable for the LLM
	issuesJSON, err := json.Marshal(issues)
	if err != nil {
		return nil, common.WrapError(err, "failed to marshal issues")
	}

	// Generate prompt for LLM
	prompt := fmt.Sprintf(`
You are an expert GitHub analyst. Please answer the following question about repository issues:

Question: %s

Here are the relevant issues:
%s

Please provide:
1. A comprehensive answer to the question
2. Include references to specific issues when relevant
3. Any additional context that would help understand the answer
4. If appropriate, suggest 2-3 follow-up questions

Format your response in a clear, structured manner with markdown formatting.
`, question, string(issuesJSON))

	// Generate response using LLM
	answer, err := h.LLMClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate answer")
	}

	// Extract follow-up questions
	followups := extractFollowupQuestions(answer)

	return &SmartNavigateResponse{
		Answer:            answer,
		APIType:           llm.APITypeIssues,
		FollowupQuestions: followups,
		ExtraData: map[string]interface{}{
			"issues": issues,
		},
	}, nil
}

// handleReleasesQuestion handles questions about releases
func (h *Handler) handleReleasesQuestion(ctx context.Context, owner, repo, question string, keywords []string) (*SmartNavigateResponse, error) {
	// Get releases data - implement this method in your GitHub client
	releases, err := h.GithubClient.GetReleases(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to get releases")
	}

	// Convert releases to a format suitable for the LLM
	releasesJSON, err := json.Marshal(releases)
	if err != nil {
		return nil, common.WrapError(err, "failed to marshal releases")
	}

	// Generate prompt for LLM
	prompt := fmt.Sprintf(`
You are an expert GitHub analyst. Please answer the following question about repository releases:

Question: %s

Here are the releases:
%s

Please provide:
1. A comprehensive answer to the question
2. Include references to specific releases when relevant
3. Any additional context that would help understand the answer
4. If appropriate, suggest 2-3 follow-up questions

Format your response in a clear, structured manner with markdown formatting.
`, question, string(releasesJSON))

	// Generate response using LLM
	answer, err := h.LLMClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate answer")
	}

	// Extract follow-up questions
	followups := extractFollowupQuestions(answer)

	return &SmartNavigateResponse{
		Answer:            answer,
		APIType:           llm.APITypeReleases,
		FollowupQuestions: followups,
		ExtraData: map[string]interface{}{
			"releases": releases,
		},
	}, nil
}

// handleStatsQuestion handles questions about repository statistics
func (h *Handler) handleStatsQuestion(ctx context.Context, owner, repo, question string, keywords []string) (*SmartNavigateResponse, error) {
	// Get repository stats - implement this method in your GitHub client
	stats, err := h.GithubClient.GetRepositoryStats(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to get repository stats")
	}

	// Convert stats to a format suitable for the LLM
	statsJSON, err := json.Marshal(stats)
	if err != nil {
		return nil, common.WrapError(err, "failed to marshal repository stats")
	}

	// Generate prompt for LLM
	prompt := fmt.Sprintf(`
You are an expert GitHub analyst. Please answer the following question about repository statistics:

Question: %s

Here are the repository statistics:
%s

Please provide:
1. A comprehensive answer to the question
2. Include references to specific statistics when relevant
3. Any additional context that would help understand the answer
4. If appropriate, suggest 2-3 follow-up questions

Format your response in a clear, structured manner with markdown formatting.
`, question, string(statsJSON))

	// Generate response using LLM
	answer, err := h.LLMClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate answer")
	}

	// Extract follow-up questions
	followups := extractFollowupQuestions(answer)

	return &SmartNavigateResponse{
		Answer:            answer,
		APIType:           llm.APITypeStats,
		FollowupQuestions: followups,
		ExtraData: map[string]interface{}{
			"stats": stats,
		},
	}, nil
}

// handleUsersQuestion handles questions about users
func (h *Handler) handleUsersQuestion(ctx context.Context, owner, repo, question string, keywords []string) (*SmartNavigateResponse, error) {
	// Get repository contributors - implement this method in your GitHub client
	contributors, err := h.GithubClient.GetContributors(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to get contributors")
	}

	// Convert contributors to a format suitable for the LLM
	contributorsJSON, err := json.Marshal(contributors)
	if err != nil {
		return nil, common.WrapError(err, "failed to marshal contributors")
	}

	// Generate prompt for LLM
	prompt := fmt.Sprintf(`
You are an expert GitHub analyst. Please answer the following question about repository users:

Question: %s

Here are the repository contributors:
%s

Please provide:
1. A comprehensive answer to the question
2. Include references to specific contributors when relevant
3. Any additional context that would help understand the answer
4. If appropriate, suggest 2-3 follow-up questions

Format your response in a clear, structured manner with markdown formatting.
`, question, string(contributorsJSON))

	// Generate response using LLM
	answer, err := h.LLMClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate answer")
	}

	// Extract follow-up questions
	followups := extractFollowupQuestions(answer)

	return &SmartNavigateResponse{
		Answer:            answer,
		APIType:           llm.APITypeUsers,
		FollowupQuestions: followups,
		ExtraData: map[string]interface{}{
			"contributors": contributors,
		},
	}, nil
}

// handleReposQuestion handles questions about the repository itself
func (h *Handler) handleReposQuestion(ctx context.Context, owner, repo, question string, keywords []string) (*SmartNavigateResponse, error) {
	// Get repository info - implement this method in your GitHub client
	repoInfo, err := h.GithubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to get repository info")
	}

	// Get repository languages
	languages, err := h.GithubClient.GetRepositoryLanguages(ctx, owner, repo)
	if err != nil {
		// Non-critical error, just log and continue
		h.Logger.WithError(err).Warning("Failed to get repository languages")
	}

	// Get repository topics/tags
	topics, err := h.GithubClient.GetRepositoryTopics(ctx, owner, repo)
	if err != nil {
		// Non-critical error, just log and continue
		h.Logger.WithError(err).Warning("Failed to get repository topics")
	}

	// Combine all repository info
	repoData := map[string]interface{}{
		"info":      repoInfo,
		"languages": languages,
		"topics":    topics,
	}

	// Convert repository data to a format suitable for the LLM
	repoDataJSON, err := json.Marshal(repoData)
	if err != nil {
		return nil, common.WrapError(err, "failed to marshal repository data")
	}

	// Generate prompt for LLM
	prompt := fmt.Sprintf(`
You are an expert GitHub analyst. Please answer the following question about the repository:

Question: %s

Here is the repository information:
%s

Please provide:
1. A comprehensive answer to the question
2. Include references to specific repository data when relevant
3. Any additional context that would help understand the answer
4. If appropriate, suggest 2-3 follow-up questions

Format your response in a clear, structured manner with markdown formatting.
`, question, string(repoDataJSON))

	// Generate response using LLM
	answer, err := h.LLMClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate answer")
	}

	// Extract follow-up questions
	followups := extractFollowupQuestions(answer)

	return &SmartNavigateResponse{
		Answer:            answer,
		APIType:           llm.APITypeRepos,
		FollowupQuestions: followups,
		ExtraData: map[string]interface{}{
			"repository": repoData,
		},
	}, nil
}

// extractFollowupQuestions extracts follow-up questions from the LLM's answer
func extractFollowupQuestions(text string) []string {
	var followups []string
	lines := strings.Split(text, "\n")
	
	inFollowupSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for follow-up questions section
		if strings.Contains(strings.ToLower(line), "follow-up") || 
		   strings.Contains(strings.ToLower(line), "follow up") {
			inFollowupSection = true
			continue
		}
		
		// Extract questions from the follow-up section
		if inFollowupSection {
			// Look for bullet points or numbered lists
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || 
			   strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") || 
			   strings.HasPrefix(line, "3.") {
				
				// Extract the question
				question := strings.TrimPrefix(line, "-")
				question = strings.TrimPrefix(question, "*")
				question = strings.TrimPrefix(question, "1.")
				question = strings.TrimPrefix(question, "2.")
				question = strings.TrimPrefix(question, "3.")
				question = strings.TrimSpace(question)
                
                // Only include the question if it ends with a question mark or seems like a question
                if strings.HasSuffix(question, "?") || 
                   strings.HasPrefix(strings.ToLower(question), "how") || 
                   strings.HasPrefix(strings.ToLower(question), "what") || 
                   strings.HasPrefix(strings.ToLower(question), "why") || 
                   strings.HasPrefix(strings.ToLower(question), "when") || 
                   strings.HasPrefix(strings.ToLower(question), "where") || 
                   strings.HasPrefix(strings.ToLower(question), "which") || 
                   strings.HasPrefix(strings.ToLower(question), "who") || 
                   strings.HasPrefix(strings.ToLower(question), "can") {
                    
                    followups = append(followups, question)
                }
            }
        }
    }
    
    // Limit to 3 follow-up questions
    if len(followups) > 3 {
        followups = followups[:3]
    }
    
    return followups
}