package github

import (
	"context"
	"encoding/base64"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// FileContent represents the content of a file
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	SHA     string `json:"sha"`
}

// GetDirectoryContent gets the content of a directory in a repository
func (c *Client) GetDirectoryContent(ctx context.Context, owner, repo, path, ref string) ([]*github.RepositoryContent, error) {
	_, directoryContent, _, err := c.client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return nil, common.WrapError(err, "failed to get directory content")
	}
	return directoryContent, nil
}

// // GetFileContentText gets the text content of a file in a repository
// func (c *Client) GetFileContentText(ctx context.Context, owner, repo, path, ref string) (*FileContent, error) {
//     fileContent, _, resp, err := c.client.Repositories.GetContents(
//         ctx,
//         owner,
//         repo,
//         path,
//         &github.RepositoryContentGetOptions{Ref: ref},
//     )
    
//     if err != nil {
//         if resp != nil && resp.StatusCode == 404 {
//             return nil, common.NewError("file not found: " + path)
//         }
//         return nil, common.WrapError(err, "failed to get file content")
//     }
    
//     // Check if fileContent is nil
//     if fileContent == nil {
//         return nil, common.NewError("received nil content for: " + path)
//     }
    
//     // Check if required fields are nil
//     if fileContent.Encoding == nil {
//         return nil, common.NewError("encoding is nil for: " + path)
//     }
    
//     content, err := fileContent.GetContent()
//     if err != nil {
//         return nil, common.WrapError(err, "failed to decode file content")
//     }
    
//     // Check if required fields for return are nil
//     if fileContent.Path == nil {
//         return nil, common.NewError("path is nil for: " + path)
//     }
    
//     if fileContent.SHA == nil {
//         return nil, common.NewError("SHA is nil for: " + path)
//     }
    
//     return &FileContent{
//         Path:    *fileContent.Path,
//         Content: content,
//         SHA:     *fileContent.SHA,
//     }, nil
// }

// CreateReadmeFile creates or updates the README.md file in a repository
func (c *Client) CreateReadmeFile(ctx context.Context, owner, repo, content, branch string) error {
	message := "Add README.md via GitHub Agent"
	return c.CreateOrUpdateFile(ctx, owner, repo, "README.md", message, content, branch)
}

// CreateDockerfile creates or updates a Dockerfile in a repository
func (c *Client) CreateDockerfile(ctx context.Context, owner, repo, content, branch string) error {
	message := "Add Dockerfile via GitHub Agent"
	return c.CreateOrUpdateFile(ctx, owner, repo, "Dockerfile", message, content, branch)
}

// FindFilesByExtension finds all files with a specific extension in a repository
func (c *Client) FindFilesByExtension(ctx context.Context, owner, repo, extension, ref string) ([]string, error) {
	tree, _, err := c.client.Git.GetTree(ctx, owner, repo, ref, true)
	if err != nil {
		return nil, common.WrapError(err, "failed to get repository tree")
	}

	var files []string
	for _, entry := range tree.Entries {
		if entry.Type != nil && *entry.Type == "blob" && entry.Path != nil {
			if filepath.Ext(*entry.Path) == extension {
				files = append(files, *entry.Path)
			}
		}
	}

	return files, nil
}

// AddCommentToFile adds a comment to a file and commits the changes
func (c *Client) AddCommentToFile(
	ctx context.Context,
	owner, repo, path, comment, branch string,
) error {
	// Get current file content
	fileContent, err := c.GetFileContentText(ctx, owner, repo, path, branch)
	if err != nil {
		return err
	}

	// Determine the file type and appropriate comment syntax
	var commentPrefix, commentSuffix string
	ext := filepath.Ext(path)
	switch strings.ToLower(ext) {
	case ".py":
		commentPrefix = "# "
		// Handle multiline comments for Python
		if strings.Contains(comment, "\n") {
			commentPrefix = "\"\"\"\n"
			commentSuffix = "\n\"\"\"\n"
		}
	case ".js", ".ts", ".jsx", ".tsx", ".go", ".java", ".c", ".cpp", ".cs":
		commentPrefix = "// "
		// Handle multiline comments
		if strings.Contains(comment, "\n") {
			commentPrefix = "/*\n * "
			commentSuffix = "\n */\n"
			comment = strings.ReplaceAll(comment, "\n", "\n * ")
		}
	case ".html", ".xml":
		commentPrefix = "<!-- "
		commentSuffix = " -->\n"
	default:
		commentPrefix = "# "
	}

	// Add the comment to the beginning of the file
	updatedContent := commentPrefix + comment + commentSuffix + fileContent.Content

	// Commit the changes
	message := "Add comments via GitHub Agent"
	return c.CreateOrUpdateFile(ctx, owner, repo, path, message, updatedContent, branch)
}

// EncodeContent encodes content to base64
func EncodeContent(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

// DecodeContent decodes base64 content
// DecodeContent decodes base64 content
func DecodeContent(encodedContent string) (string, error) {
    if encodedContent == "" {
        return "", common.NewError("encoded content is empty")
    }
    
    decodedBytes, err := base64.StdEncoding.DecodeString(encodedContent)
    if err != nil {
        return "", common.WrapError(err, "failed to decode content")
    }
    return string(decodedBytes), nil
}