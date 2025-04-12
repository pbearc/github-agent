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

// GetRepositoryInfo gets detailed information about a repository
// func (c *Client) GetRepositoryInfo(ctx context.Context, owner, repo string) (*RepositoryInfo, error) {
// 	repository, err := c.GetRepository(ctx, owner, repo)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Check if README exists
// 	hasReadme := false
// 	_, _, _, err = c.client.Repositories.GetContents(
// 		ctx,
// 		owner,
// 		repo,
// 		"README.md",
// 		&github.RepositoryContentGetOptions{},
// 	)
// 	if err == nil {
// 		hasReadme = true
// 	} else {
// 		// Try lowercase readme
// 		_, _, _, err = c.client.Repositories.GetContents(
// 			ctx,
// 			owner,
// 			repo,
// 			"readme.md",
// 			&github.RepositoryContentGetOptions{},
// 		)
// 		if err == nil {
// 			hasReadme = true
// 		}
// 	}

// 	description := ""
// 	if repository.Description != nil {
// 		description = *repository.Description
// 	}

// 	language := ""
// 	if repository.Language != nil {
// 		language = *repository.Language
// 	}

// 	return &RepositoryInfo{
// 		Owner:         owner,
// 		Name:          repository.GetName(),
// 		Description:   description,
// 		DefaultBranch: repository.GetDefaultBranch(),
// 		URL:           repository.GetHTMLURL(),
// 		Language:      language,
// 		Stars:         repository.GetStargazersCount(),
// 		Forks:         repository.GetForksCount(),
// 		HasReadme:     hasReadme,
// 	}, nil
// }

// // GetRepositoryStructure gets the file structure of a repository
// func (c *Client) GetRepositoryStructure(ctx context.Context, owner, repo, ref string) (string, error) {
// 	tree, _, err := c.client.Git.GetTree(ctx, owner, repo, ref, true)
// 	if err != nil {
// 		return "", common.WrapError(err, "failed to get repository structure")
// 	}

// 	var sb strings.Builder
// 	for _, entry := range tree.Entries {
// 		sb.WriteString(*entry.Path)
// 		sb.WriteString("\n")
// 	}

// 	return sb.String(), nil
// }

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