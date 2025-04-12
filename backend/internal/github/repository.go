package github

import (
	"context"
	"strings"

	"github.com/pbearc/github-agent/backend/pkg/common"
)

// RepositoryInfo contains essential information about a repository
type RepositoryInfo struct {
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	URL           string `json:"url"`
	Language      string `json:"language"`
	Stars         int    `json:"stars"`
	Forks         int    `json:"forks"`
	HasReadme     bool   `json:"has_readme"`
}



// CreateFile creates a new file in the repository
func (c *Client) CreateFile(ctx context.Context, owner, repo, path, message, content, branch string) error {
	return c.CreateOrUpdateFile(ctx, owner, repo, path, message, content, branch)
}

// UpdateFile updates an existing file in the repository
func (c *Client) UpdateFile(ctx context.Context, owner, repo, path, message, content, branch string) error {
	return c.CreateOrUpdateFile(ctx, owner, repo, path, message, content, branch)
}

// ParseRepoURL parses a GitHub URL into owner and repo
func ParseRepoURL(url string) (string, string, error) {
	// Remove trailing slashes and whitespace
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")
	
	// Simple owner/repo format (e.g., "owner/repo")
	if strings.Count(url, "/") == 1 && !strings.Contains(url, "github.com") {
		parts := strings.Split(url, "/")
		return parts[0], parts[1], nil
	}
	
	// GitHub URL format (e.g., "https://github.com/owner/repo")
	if strings.Contains(url, "github.com") {
		// Extract the part after github.com
		parts := strings.Split(url, "github.com/")
		if len(parts) != 2 {
			return "", "", common.NewError("invalid GitHub URL format")
		}
		
		// Split into owner and repo
		ownerRepo := strings.Split(parts[1], "/")
		if len(ownerRepo) < 2 {
			return "", "", common.NewError("invalid GitHub URL format")
		}
		
		owner := ownerRepo[0]
		repo := ownerRepo[1]
		
		// Remove any suffix like .git
		repo = strings.TrimSuffix(repo, ".git")
		
		return owner, repo, nil
	}
	
	// SSH URL format (e.g., "git@github.com:owner/repo.git")
	if strings.Contains(url, "git@github.com:") {
		parts := strings.Split(url, "git@github.com:")
		if len(parts) != 2 {
			return "", "", common.NewError("invalid GitHub SSH URL format")
		}
		
		// Split into owner and repo
		ownerRepo := strings.Split(parts[1], "/")
		if len(ownerRepo) < 2 {
			return "", "", common.NewError("invalid GitHub SSH URL format")
		}
		
		owner := ownerRepo[0]
		repo := ownerRepo[1]
		
		// Remove any suffix like .git
		repo = strings.TrimSuffix(repo, ".git")
		
		return owner, repo, nil
	}
	
	return "", "", common.NewError("unsupported GitHub URL format")
}