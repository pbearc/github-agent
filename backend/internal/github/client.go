package github

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/pbearc/github-agent/backend/internal/models"
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

// buildStructureTree recursively builds a tree representation of the repository structure
func (c *Client) buildStructureTree(sb *strings.Builder, ctx context.Context, owner, repo string, contents []*github.RepositoryContent, path string, ref string, depth int) error {
	for _, content := range contents {
		if content == nil || content.Name == nil || content.Type == nil {
			continue
		}

		// Add indentation
		for i := 0; i < depth; i++ {
			sb.WriteString("  ")
		}

		// Add the item
		sb.WriteString("- ")
		sb.WriteString(*content.Name)
		sb.WriteString("\n")

		// Recursively process directories, up to a reasonable depth
		if *content.Type == "dir" && depth < 5 {
			contentPath := *content.Path
			subContents, err := c.ListFiles(ctx, owner, repo, contentPath, ref)
			if err != nil {
				c.logger.WithField("error", err).Warning("Failed to list files for path: " + contentPath)
				continue
			}

			err = c.buildStructureTree(sb, ctx, owner, repo, subContents, contentPath, ref, depth+1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetFunctionCode extracts a function code from a file
func (c *Client) GetFunctionCode(ctx context.Context, owner, repo, path, functionName, ref string) (string, int, int, error) {
	content, err := c.GetFileContentText(ctx, owner, repo, path, ref)
	if err != nil {
		return "", 0, 0, common.WrapError(err, "failed to get file content")
	}

	lines := strings.Split(content.Content, "\n")
	startLine, endLine := -1, -1
	inFunction := false
	braceCount := 0

	// This is a simple function finder - in practice, you'd want to use a proper parser
	// for the specific language, but this gives the basic idea
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Look for function declaration with the given name
		if !inFunction && (strings.Contains(trimmedLine, "func "+functionName+"(") || 
						   strings.Contains(trimmedLine, "function "+functionName+"(") ||
						   strings.Contains(trimmedLine, functionName+" = function(") ||
						   strings.Contains(trimmedLine, "def "+functionName+"(")) {
			startLine = i
			inFunction = true
			braceCount = strings.Count(trimmedLine, "{") - strings.Count(trimmedLine, "}")
			continue
		}

		if inFunction {
			braceCount += strings.Count(trimmedLine, "{") - strings.Count(trimmedLine, "}")
			
			// For languages using braces
			if braceCount == 0 && strings.Contains(trimmedLine, "}") {
				endLine = i
				break
			}
			
			// For Python-like languages
			if (trimmedLine == "" || !strings.HasPrefix(trimmedLine, " ")) && startLine != i-1 {
				endLine = i - 1
				break
			}
		}
	}

	if startLine == -1 {
		return "", 0, 0, common.NewError("function not found: " + functionName)
	}

	if endLine == -1 {
		endLine = len(lines) - 1
	}

	// Extract the function code
	functionLines := lines[startLine : endLine+1]
	return strings.Join(functionLines, "\n"), startLine + 1, endLine + 1, nil
}

// GetImportMap analyzes the codebase to find import relationships
func (c *Client) GetImportMap(ctx context.Context, owner, repo, ref string) (map[string][]string, error) {
	// Get all files
	allFiles, err := c.GetAllFiles(ctx, owner, repo, ref)
	if err != nil {
		return nil, common.WrapError(err, "failed to get all files")
	}

	importMap := make(map[string][]string)
	
	// Process each file to find imports
	for _, file := range allFiles {
		if file.Type != "file" || !isSourceFile(file.Path) {
			continue
		}
		
		content, err := c.GetFileContentText(ctx, owner, repo, file.Path, ref) // Returns *FileContent
		if err != nil {
			c.logger.WithField("error", err).Warning("Failed to get content for file: " + file.Path)
			continue
		}
		
		imports := extractImports(content.Content, getLanguageFromPath(file.Path))
		if len(imports) > 0 {
			importMap[file.Path] = imports
		}
	}
	
	return importMap, nil
}

// GetAllFiles gets all files in a repository
func (c *Client) GetAllFiles(ctx context.Context, owner, repo, ref string) ([]models.GitHubFile, error) {
	rootContent, err := c.ListFiles(ctx, owner, repo, "", ref)
	if err != nil {
		return nil, common.WrapError(err, "failed to list root files")
	}

	var allFiles []models.GitHubFile
	err = c.collectAllFiles(&allFiles, ctx, owner, repo, rootContent, ref)
	if err != nil {
		return nil, common.WrapError(err, "failed to collect all files")
	}

	return allFiles, nil
}

// collectAllFiles recursively collects all files in a repository
func (c *Client) collectAllFiles(files *[]models.GitHubFile, ctx context.Context, owner, repo string, contents []*github.RepositoryContent, ref string) error {
	for _, content := range contents {
		if content == nil || content.Name == nil || content.Type == nil || content.Path == nil {
			continue
		}

		var downloadURL string
		if content.DownloadURL != nil {
			downloadURL = *content.DownloadURL
		}

		file := models.GitHubFile{
			Name:        *content.Name,
			Path:        *content.Path,
			Size:        *content.Size,
			Type:        *content.Type,
			HTMLURL:     *content.HTMLURL,
			DownloadURL: downloadURL,
		}

		*files = append(*files, file)

		// Recursively process directories
		if *content.Type == "dir" {
			subContents, err := c.ListFiles(ctx, owner, repo, *content.Path, ref)
			if err != nil {
				c.logger.WithField("error", err).Warning("Failed to list files for path: " + *content.Path)
				continue
			}

			err = c.collectAllFiles(files, ctx, owner, repo, subContents, ref)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper functions
func isSourceFile(path string) bool {
	ext := filepath.Ext(path)
	sourceExts := map[string]bool{
		".go":    true,
		".java":  true,
		".js":    true,
		".jsx":   true,
		".ts":    true,
		".tsx":   true,
		".py":    true,
		".rb":    true,
		".php":   true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".cs":    true,
		".swift": true,
		".kt":    true,
	}
	return sourceExts[ext]
}

func getLanguageFromPath(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "Go"
	case ".java":
		return "Java"
	case ".js", ".jsx":
		return "JavaScript"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".py":
		return "Python"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".c", ".h":
		return "C"
	case ".cpp":
		return "C++"
	case ".cs":
		return "C#"
	case ".swift":
		return "Swift"
	case ".kt":
		return "Kotlin"
	default:
		return "Unknown"
	}
}

func extractImports(content, language string) []string {
	var imports []string
	lines := strings.Split(content, "\n")

	switch language {
	case "Go":
		inImportBlock := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import (") {
				inImportBlock = true
				continue
			}
			if inImportBlock && line == ")" {
				inImportBlock = false
				continue
			}
			if inImportBlock || strings.HasPrefix(line, "import ") {
				imp := extractGoImport(line)
				if imp != "" {
					imports = append(imports, imp)
				}
			}
		}
	case "JavaScript", "TypeScript":
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "require(") {
				imp := extractJSImport(line)
				if imp != "" {
					imports = append(imports, imp)
				}
			}
		}
	case "Python":
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
				imp := extractPythonImport(line)
				if imp != "" {
					imports = append(imports, imp)
				}
			}
		}
	}

	return imports
}

func extractGoImport(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "import ") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			imp := strings.Trim(parts[len(parts)-1], "\"")
			return imp
		}
	} else if line != "" && !strings.HasPrefix(line, "//") {
		imp := strings.Trim(line, "\"")
		return imp
	}
	return ""
}

func extractJSImport(line string) string {
	if strings.HasPrefix(line, "import ") {
		// Handle ES6 import
		fromIndex := strings.Index(line, " from ")
		if fromIndex != -1 {
			importPath := line[fromIndex+6:]
			return strings.Trim(importPath, "\"';")
		}
	} else if strings.HasPrefix(line, "require(") {
		// Handle CommonJS require
		start := strings.Index(line, "(") + 1
		end := strings.Index(line, ")")
		if start != -1 && end != -1 && end > start {
			importPath := line[start:end]
			return strings.Trim(importPath, "\"'")
		}
	}
	return ""
}

func extractPythonImport(line string) string {
	if strings.HasPrefix(line, "import ") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			return parts[1]
		}
	} else if strings.HasPrefix(line, "from ") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return ""
}