package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/types"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// PRSummaryService handles pull request summary generation
type PRSummaryService struct {
	githubClient *github.Client
	llmClient    *llm.GeminiClient
	logger       *common.Logger
}

// NewPRSummaryService creates a new PRSummaryService
func NewPRSummaryService(githubClient *github.Client, llmClient *llm.GeminiClient) *PRSummaryService {
	return &PRSummaryService{
		githubClient: githubClient,
		llmClient:    llmClient,
		logger:       common.NewLogger(),
	}
}

// Internal FileGroup for processing
type fileGroup struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
	Importance  int      `json:"importance"`
}

// Internal summary structure for LLM response parsing
type llmSummaryResponse struct {
	Description        string      `json:"description"`
	MainPoints         []string    `json:"main_points"`
	KeyChanges         []string    `json:"key_changes"`
	FileGroups         []fileGroup `json:"file_groups"`
	PotentialImpact    string      `json:"potential_impact"`
	SuggestedReviewers []string    `json:"suggested_reviewers"`
	TechnicalDetails   string      `json:"technical_details"`
}

// GenerateSummary generates a summary for a pull request
func (s *PRSummaryService) GenerateSummary(ctx context.Context, owner, repo string, prNumber int) (*types.PRSummary, error) {
	// Get PR details
	pr, err := s.githubClient.GetPullRequest(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, common.WrapError(err, "failed to get pull request details")
	}

	// Group related files
	fileGroups := s.groupFiles(pr.Files)

	// Generate summary using LLM
	llmResponse, err := s.generateLLMSummary(ctx, pr, fileGroups)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate summary")
	}
	
	// Convert to shared type
	summary := &types.PRSummary{
		Title:              pr.Title,
		Description:        llmResponse.Description,
		MainPoints:         llmResponse.MainPoints,
		KeyChanges:         llmResponse.KeyChanges,
		PotentialImpact:    llmResponse.PotentialImpact,
		SuggestedReviewers: llmResponse.SuggestedReviewers,
		TechnicalDetails:   llmResponse.TechnicalDetails,
		PRNumber:           pr.Number,
		PRURL:              pr.URL,
		Repository:         fmt.Sprintf("%s/%s", owner, repo),
		Author:             pr.User,
		ChangedFiles:       pr.ChangedFiles,
		Additions:          pr.Additions,
		Deletions:          pr.Deletions,
	}
	
	// Convert file groups
	summary.FileGroups = make([]types.FileGroup, len(llmResponse.FileGroups))
	for i, g := range llmResponse.FileGroups {
		summary.FileGroups[i] = types.FileGroup{
			Name:        g.Name,
			Description: g.Description,
			Files:       g.Files,
			Importance:  g.Importance,
		}
	}
	
	// If LLM didn't provide good file groups, use our simple ones
	if len(summary.FileGroups) == 0 {
		summary.FileGroups = s.convertFileGroups(fileGroups)
	} else {
		// Ensure file groups have files assigned
		s.assignFilesToGroups(summary.FileGroups, pr.Files)
	}

	return summary, nil
}

// groupFiles groups related files based on directory structure and file types
func (s *PRSummaryService) groupFiles(files []github.FileChange) []fileGroup {
	// Simple grouping by top-level directory
	dirGroups := make(map[string][]string)
	
	for _, file := range files {
		parts := strings.Split(file.Filename, "/")
		var dir string
		if len(parts) > 1 {
			dir = parts[0]
		} else {
			dir = "root"
		}
		
		dirGroups[dir] = append(dirGroups[dir], file.Filename)
	}
	
	// Convert to FileGroup slice
	var groups []fileGroup
	for dir, fileList := range dirGroups {
		group := fileGroup{
			Name:  dir,
			Files: fileList,
			// We'll fill in descriptions with LLM
		}
		groups = append(groups, group)
	}
	
	return groups
}

// generateLLMSummary uses the LLM to generate a PR summary
func (s *PRSummaryService) generateLLMSummary(ctx context.Context, pr *github.PullRequest, fileGroups []fileGroup) (*llmSummaryResponse, error) {
	// Create a context of the PR for the LLM
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString("Please analyze this GitHub pull request and generate a concise summary:\n\n")
	
	// PR metadata
	promptBuilder.WriteString(fmt.Sprintf("PR Title: %s\n", pr.Title))
	promptBuilder.WriteString(fmt.Sprintf("PR Description: %s\n\n", pr.Description))
	promptBuilder.WriteString(fmt.Sprintf("Changes: %d files changed with %d additions and %d deletions\n\n", 
		pr.ChangedFiles, pr.Additions, pr.Deletions))
	
	// Select the most significant files to include in the prompt
	var significantFiles []github.FileChange
	
	// Sort files by the sum of additions and deletions
	sort.Slice(pr.Files, func(i, j int) bool {
		return (pr.Files[i].Additions + pr.Files[i].Deletions) > 
			   (pr.Files[j].Additions + pr.Files[j].Deletions)
	})
	
	// Take top 10 files or all if less than 10
	maxFiles := 10
	if len(pr.Files) < maxFiles {
		maxFiles = len(pr.Files)
	}
	
	significantFiles = pr.Files[:maxFiles]
	
	// Add file changes to prompt
	promptBuilder.WriteString("Most significant file changes:\n\n")
	
	for _, file := range significantFiles {
		promptBuilder.WriteString(fmt.Sprintf("File: %s\n", file.Filename))
		promptBuilder.WriteString(fmt.Sprintf("Status: %s\n", file.Status))
		promptBuilder.WriteString(fmt.Sprintf("Changes: +%d -%d\n", file.Additions, file.Deletions))
		
		// Include patch for smaller changes
		if file.Patch != "" && len(file.Patch) < 1500 {
			promptBuilder.WriteString("Patch:\n```\n")
			promptBuilder.WriteString(file.Patch)
			promptBuilder.WriteString("\n```\n")
		}
		
		promptBuilder.WriteString("\n")
	}
	
	// Add instructions for what we want in the summary
	promptBuilder.WriteString(`
Based on the PR information above, please generate a comprehensive summary with the following components:

1. A clear 1-2 sentence description of what this PR does
2. 3-5 main points highlighting the most important changes
3. A list of key technical changes introduced
4. File groupings with meaningful names and descriptions
5. Assessment of potential impact (e.g., performance, security, user experience)
6. Suggested areas for reviewers to focus on

Format your response as JSON with the following structure:
{
  "description": "Brief description of the PR",
  "main_points": ["Point 1", "Point 2", "Point 3"],
  "key_changes": ["Technical change 1", "Technical change 2"],
  "file_groups": [
    {
      "name": "Group name",
      "description": "What these files do",
      "importance": 1-10 scale
    }
  ],
  "potential_impact": "Description of potential impact",
  "suggested_reviewers": ["backend", "frontend", "security"],
  "technical_details": "Additional technical details that reviewers should be aware of"
}
`)

	// Send to LLM
	llmOutput, err := s.llmClient.GenerateText(ctx, promptBuilder.String())
	if err != nil {
		return nil, common.WrapError(err, "failed to generate LLM summary")
	}
	
	// Parse LLM response as JSON
	summary, err := s.parseSummaryResponse(llmOutput)
	if err != nil {
		return nil, common.WrapError(err, "failed to parse LLM response")
	}
	
	return summary, nil
}

// parseSummaryResponse parses the LLM response into a structured PRSummary
func (s *PRSummaryService) parseSummaryResponse(response string) (*llmSummaryResponse, error) {
	// Find JSON within the response (LLM might add additional text)
	startIndex := strings.Index(response, "{")
	endIndex := strings.LastIndex(response, "}")
	
	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		return nil, common.NewError("invalid JSON response from LLM")
	}
	
	jsonStr := response[startIndex:endIndex+1]
	
	// Parse JSON
	var summary llmSummaryResponse
	err := json.Unmarshal([]byte(jsonStr), &summary)
	if err != nil {
		return nil, common.WrapError(err, "failed to unmarshal LLM response")
	}
	
	return &summary, nil
}

// convertFileGroups converts internal file groups to the shared type
func (s *PRSummaryService) convertFileGroups(groups []fileGroup) []types.FileGroup {
	result := make([]types.FileGroup, len(groups))
	for i, g := range groups {
		result[i] = types.FileGroup{
			Name:        g.Name,
			Description: g.Description,
			Files:       g.Files,
			Importance:  g.Importance,
		}
	}
	return result
}

// assignFilesToGroups assigns files to the LLM-generated groups
func (s *PRSummaryService) assignFilesToGroups(groups []types.FileGroup, files []github.FileChange) {
	// Only do this if file lists are empty
	hasFiles := false
	for _, g := range groups {
		if len(g.Files) > 0 {
			hasFiles = true
			break
		}
	}
	
	if hasFiles {
		return
	}
	
	// Simple heuristic: assign files to groups based on name matching
	for _, file := range files {
		bestGroup := 0
		bestScore := -1
		
		filename := strings.ToLower(file.Filename)
		
		for i, group := range groups {
			// Calculate a simple relevance score based on substring matching
			groupName := strings.ToLower(group.Name)
			score := 0
			
			// Check if filename contains group name
			if strings.Contains(filename, groupName) {
				score += 3
			}
			
			// Check if path components match group name
			parts := strings.Split(filename, "/")
			for _, part := range parts {
				if strings.Contains(part, groupName) || strings.Contains(groupName, part) {
					score += 2
					break
				}
			}
			
			if score > bestScore {
				bestScore = score
				bestGroup = i
			}
		}
		
		// If we found a decent match, add the file to that group
		if bestScore > 0 {
			groups[bestGroup].Files = append(groups[bestGroup].Files, file.Filename)
		} else {
			// If no good match, add to the first group (or create a misc group)
			if len(groups) > 0 {
				groups[0].Files = append(groups[0].Files, file.Filename)
			}
		}
	}
}