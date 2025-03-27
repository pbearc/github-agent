package services

import (
	"context"

	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// RefactorService handles code refactoring
type RefactorService struct {
	githubClient *github.Client
	llmClient    *llm.GeminiClient
	logger       *common.Logger
}

// NewRefactorService creates a new RefactorService instance
func NewRefactorService(githubClient *github.Client, llmClient *llm.GeminiClient) *RefactorService {
	return &RefactorService{
		githubClient: githubClient,
		llmClient:    llmClient,
		logger:       common.NewLogger(),
	}
}

// GenerateRefactoredCode generates refactored code for a file
func (s *RefactorService) GenerateRefactoredCode(ctx context.Context, owner, repo, branch, path, instructions string) (string, error) {
	// Get file content
	fileContent, err := s.githubClient.GetFileContentText(ctx, owner, repo, path, branch)
	if err != nil {
		return "", common.WrapError(err, "failed to get file content")
	}

	// Detect language from file extension
	language := detectLanguageFromPath(path)

	// Generate refactored code
	refactoredCode, err := s.llmClient.GenerateCodeRefactor(ctx, fileContent.Content, language, instructions)
	if err != nil {
		return "", common.WrapError(err, "failed to generate refactored code")
	}

	return refactoredCode, nil
}

// PushRefactoredFile generates refactored code for a file and pushes it to the repository
func (s *RefactorService) PushRefactoredFile(ctx context.Context, owner, repo, branch, path, instructions string) error {
	// Generate refactored code
	refactoredCode, err := s.GenerateRefactoredCode(ctx, owner, repo, branch, path, instructions)
	if err != nil {
		return err
	}

	// If branch is not specified, get the default branch
	if branch == "" {
		repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
		if err != nil {
			return common.WrapError(err, "failed to get repository info")
		}
		branch = repoInfo.DefaultBranch
	}

	// Push the refactored file
	commitMessage := "Refactor " + path + " via GitHub Agent"
	err = s.githubClient.CreateOrUpdateFile(ctx, owner, repo, path, commitMessage, refactoredCode, branch)
	if err != nil {
		return common.WrapError(err, "failed to push refactored file")
	}

	return nil
}

// RefactorMultipleFiles refactors multiple files based on the same instructions
func (s *RefactorService) RefactorMultipleFiles(ctx context.Context, owner, repo, branch string, paths []string, instructions string) (map[string]string, error) {
	results := make(map[string]string)
	
	for _, path := range paths {
		refactoredCode, err := s.GenerateRefactoredCode(ctx, owner, repo, branch, path, instructions)
		if err != nil {
			s.logger.WithField("error", err).WithField("path", path).Error("Failed to refactor file")
			results[path] = "Error: " + err.Error()
			continue
		}
		
		// Push the refactored file
		commitMessage := "Refactor " + path + " via GitHub Agent"
		err = s.githubClient.CreateOrUpdateFile(ctx, owner, repo, path, commitMessage, refactoredCode, branch)
		if err != nil {
			s.logger.WithField("error", err).WithField("path", path).Error("Failed to push refactored file")
			results[path] = "Error pushing: " + err.Error()
			continue
		}
		
		results[path] = "Successfully refactored and pushed"
	}
	
	return results, nil
}

// GenerateRefactoringPlan generates a plan for refactoring a repository
func (s *RefactorService) GenerateRefactoringPlan(ctx context.Context, owner, repo, branch string) (string, error) {
	// Get repository info
	repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		return "", common.WrapError(err, "failed to get repository info")
	}

	// Get repository structure
	repoStructure, err := s.githubClient.GetRepositoryStructure(ctx, owner, repo, branch)
	if err != nil {
		return "", common.WrapError(err, "failed to get repository structure")
	}

	// Build a prompt for generating a refactoring plan
	prompt := `
You are an expert software architect and developer. Your task is to create a refactoring plan for a GitHub repository.

Repository Information:
- Name: ` + repoInfo.Name + `
- Owner: ` + repoInfo.Owner + `
- Description: ` + repoInfo.Description + `
- Primary Language: ` + repoInfo.Language + `

Repository Structure:
` + repoStructure + `

Please provide a comprehensive refactoring plan that includes:
1. Overall architecture recommendations
2. Key files/components that should be refactored
3. Specific refactoring recommendations for each file/component
4. Priority order (which files to refactor first)
5. Potential benefits of the refactoring
6. Any additional recommendations for improving code quality

Format your response as a detailed markdown document.
`

	plan, err := s.llmClient.GenerateText(ctx, prompt)
	if err != nil {
		return "", common.WrapError(err, "failed to generate refactoring plan")
	}

	return plan, nil
}