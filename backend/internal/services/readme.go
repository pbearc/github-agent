package services

import (
	"context"

	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// ReadmeService handles README generation
type ReadmeService struct {
	githubClient *github.Client
	llmClient    *llm.GeminiClient
	logger       *common.Logger
}

// NewReadmeService creates a new ReadmeService instance
func NewReadmeService(githubClient *github.Client, llmClient *llm.GeminiClient) *ReadmeService {
	return &ReadmeService{
		githubClient: githubClient,
		llmClient:    llmClient,
		logger:       common.NewLogger(),
	}
}

// GenerateReadme generates a README.md file for a repository
func (s *ReadmeService) GenerateReadme(ctx context.Context, owner, repo, branch string, includeFiles bool) (string, error) {
	// Get repository info
	repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		return "", common.WrapError(err, "failed to get repository info")
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

	// If branch is not specified, use the default branch
	if branch == "" {
		branch = repoInfo.DefaultBranch
	}

	// Get file list if requested
	var files []string
	if includeFiles {
		repoStructure, err := s.githubClient.GetRepositoryStructure(ctx, owner, repo, branch)
		if err != nil {
			s.logger.WithField("error", err).Warning("Failed to get repository structure")
		} else {
			// Split the structure into lines
			files = splitLines(repoStructure)
		}
	}

	// Generate README
	readmeContent, err := s.llmClient.GenerateReadme(ctx, repoInfoMap, files)
	if err != nil {
		return "", common.WrapError(err, "failed to generate README")
	}

	return readmeContent, nil
}

// PushReadme generates and pushes a README.md file to a repository
func (s *ReadmeService) PushReadme(ctx context.Context, owner, repo, branch string, includeFiles bool) error {
	// Generate README
	readmeContent, err := s.GenerateReadme(ctx, owner, repo, branch, includeFiles)
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

	// Push the README
	err = s.githubClient.CreateReadmeFile(ctx, owner, repo, readmeContent, branch)
	if err != nil {
		return common.WrapError(err, "failed to push README")
	}

	return nil
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