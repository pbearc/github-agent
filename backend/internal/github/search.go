package github

import (
	"context"

	"github.com/google/go-github/v43/github"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// SearchCode searches for code in a repository
func (c *Client) SearchCode(ctx context.Context, owner, repo, query string) ([]*github.CodeResult, error) {
    // Construct the search query to include the repository filter
    searchQuery := query + " repo:" + owner + "/" + repo

    // Search for code
    result, _, err := c.client.Search.Code(ctx, searchQuery, nil)
    if err != nil {
        return nil, common.WrapError(err, "failed to search code")
    }

    // The go-github/v43 library uses CodeResults, not Items
    return result.CodeResults, nil
}