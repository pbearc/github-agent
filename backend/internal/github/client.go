package github

import (
	"context"

	"github.com/google/go-github/v43/github"
	"github.com/pbearc/github-agent/backend/pkg/common"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client
type Client struct {
	client *github.Client
	logger *common.Logger
}

// NewClient creates a new GitHub client with authentication
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, common.NewError("GitHub token is required")
	}

	// Create an OAuth2 client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	// Create the GitHub client
	client := github.NewClient(tc)

	return &Client{
		client: client,
		logger: common.NewLogger(),
	}, nil
}

// GetRepository fetches repository information
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to get repository")
	}
	return repository, nil
}

// GetFileContent fetches content of a file in a repository
func (c *Client) GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, error) {
	fileContent, _, _, err := c.client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return nil, common.WrapError(err, "failed to get file content")
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, common.WrapError(err, "failed to decode file content")
	}

	return []byte(content), nil
}

// CreateOrUpdateFile creates or updates a file in a repository
func (c *Client) CreateOrUpdateFile(
	ctx context.Context,
	owner, repo, path, message, content string,
	branch string,
) error {
	// Get the current file to get its SHA (needed for update)
	var sha string
	fileContent, _, _, err := c.client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		&github.RepositoryContentGetOptions{Ref: branch},
	)
	if err == nil && fileContent != nil {
		sha = *fileContent.SHA
	}

	// Create the file update options
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
	}

	// If SHA exists, we're updating an existing file
	if sha != "" {
		opts.SHA = github.String(sha)
	}

	// Commit the changes
	_, _, err = c.client.Repositories.CreateFile(ctx, owner, repo, path, opts)
	if err != nil {
		return common.WrapError(err, "failed to create or update file")
	}

	return nil
}

// ListFiles lists files in a directory in a repository
func (c *Client) ListFiles(
	ctx context.Context,
	owner, repo, path, ref string,
) ([]*github.RepositoryContent, error) {
	_, directoryContent, _, err := c.client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return nil, common.WrapError(err, "failed to list files")
	}

	return directoryContent, nil
}