package services

import (
	"context"

	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// DockerfileService handles Dockerfile generation
type DockerfileService struct {
	githubClient *github.Client
	llmClient    *llm.GeminiClient
	logger       *common.Logger
}

// NewDockerfileService creates a new DockerfileService instance
func NewDockerfileService(githubClient *github.Client, llmClient *llm.GeminiClient) *DockerfileService {
	return &DockerfileService{
		githubClient: githubClient,
		llmClient:    llmClient,
		logger:       common.NewLogger(),
	}
}

// GenerateDockerfile generates a Dockerfile for a repository
func (s *DockerfileService) GenerateDockerfile(ctx context.Context, owner, repo, branch, language string) (string, error) {
	// Get repository info
	repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		return "", common.WrapError(err, "failed to get repository info")
	}

	// Use provided language or fallback to repo language
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
	dockerfileContent, err := s.llmClient.GenerateDockerfile(ctx, repoInfoMap, language)
	if err != nil {
		return "", common.WrapError(err, "failed to generate Dockerfile")
	}

	return dockerfileContent, nil
}

// PushDockerfile generates and pushes a Dockerfile to a repository
func (s *DockerfileService) PushDockerfile(ctx context.Context, owner, repo, branch, language string) error {
	// Generate Dockerfile
	dockerfileContent, err := s.GenerateDockerfile(ctx, owner, repo, branch, language)
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

	// Push the Dockerfile
	err = s.githubClient.CreateDockerfile(ctx, owner, repo, dockerfileContent, branch)
	if err != nil {
		return common.WrapError(err, "failed to push Dockerfile")
	}

	return nil
}