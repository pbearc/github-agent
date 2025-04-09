package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// PullRequest represents details of a GitHub pull request
type PullRequest struct {
	Number       int          `json:"number"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	State        string       `json:"state"`
	User         string       `json:"user"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
	Commits      int          `json:"commits"`
	Additions    int          `json:"additions"`
	Deletions    int          `json:"deletions"`
	ChangedFiles int          `json:"changed_files"`
	Files        []FileChange `json:"files"`
	URL          string       `json:"url"`
}

// FileChange represents a file that was changed in a PR
type FileChange struct {
	Filename    string `json:"filename"`
	Status      string `json:"status"` // "added", "modified", "removed"
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	Changes     int    `json:"changes"`
	Patch       string `json:"patch,omitempty"`
	BlobURL     string `json:"blob_url"`
	ContentType string `json:"content_type,omitempty"`
}

// GetPullRequest gets details of a pull request
func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, common.WrapError(err, "failed to get pull request")
	}

	// Get the files changed in this PR
	opts := &github.ListOptions{
		PerPage: 100,
	}
	
	var allFiles []*github.CommitFile
	for {
		files, resp, err := c.client.PullRequests.ListFiles(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, common.WrapError(err, "failed to list files in pull request")
		}
		
		allFiles = append(allFiles, files...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Convert GitHub API objects to our own types
	fileChanges := make([]FileChange, 0, len(allFiles))
	for _, file := range allFiles {
		fileChange := FileChange{
			Filename:  *file.Filename,
			Status:    *file.Status,
			Additions: *file.Additions,
			Deletions: *file.Deletions,
			Changes:   *file.Changes,
			BlobURL:   *file.BlobURL,
		}
		
		// Patch might be too large, so only include it for smaller changes
		if file.Patch != nil && len(*file.Patch) < 10000 {
			fileChange.Patch = *file.Patch
		}
		
		// Detect file type
		fileChange.ContentType = detectContentType(fileChange.Filename)
		
		fileChanges = append(fileChanges, fileChange)
	}

	return &PullRequest{
		Number:       *pr.Number,
		Title:        *pr.Title,
		Description:  *pr.Body,
		State:        *pr.State,
		User:         *pr.User.Login,
		CreatedAt:    pr.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    pr.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Commits:      *pr.Commits,
		Additions:    *pr.Additions,
		Deletions:    *pr.Deletions,
		ChangedFiles: *pr.ChangedFiles,
		Files:        fileChanges,
		URL:          *pr.HTMLURL,
	}, nil
}

// detectContentType tries to determine the content type of a file based on its extension
func detectContentType(filename string) string {
	if !strings.Contains(filename, ".") {
		return "unknown"
	}
	
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	
	switch ext {
	case "go":
		return "go"
	case "js", "jsx", "ts", "tsx":
		return "javascript"
	case "py":
		return "python"
	case "java":
		return "java"
	case "c", "cpp", "h", "hpp":
		return "c++"
	case "cs":
		return "c#"
	case "php":
		return "php"
	case "rb":
		return "ruby"
	case "html", "htm":
		return "html"
	case "css":
		return "css"
	case "json":
		return "json"
	case "yml", "yaml":
		return "yaml"
	case "md":
		return "markdown"
	case "sql":
		return "sql"
	case "sh", "bash":
		return "shell"
	case "xml":
		return "xml"
	case "txt":
		return "text"
	default:
		return "other"
	}
}

// ParsePullRequestURL extracts owner, repo, and PR number from a GitHub pull request URL
func ParsePullRequestURL(url string) (string, string, int, error) {
	// Match patterns like:
	// https://github.com/owner/repo/pull/123
	// https://github.com/owner/repo/pulls/123
	
	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return "", "", 0, common.NewError("invalid GitHub pull request URL format")
	}
	
	// Check if it contains github.com
	if !strings.Contains(parts[2], "github.com") {
		return "", "", 0, common.NewError("URL must be a GitHub pull request URL")
	}
	
	// Check if it contains pull or pulls
	if parts[5] != "pull" && parts[5] != "pulls" {
		return "", "", 0, common.NewError("URL must be a GitHub pull request URL")
	}
	
	owner := parts[3]
	repo := parts[4]
	
	// Parse PR number
	var prNumber int
	_, err := fmt.Sscanf(parts[6], "%d", &prNumber)
	if err != nil {
		return "", "", 0, common.WrapError(err, "invalid pull request number")
	}
	
	return owner, repo, prNumber, nil
}