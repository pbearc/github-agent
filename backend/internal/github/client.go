package github

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v43/github"
	// ADD THIS IMPORT (adjust path if your module name is different)
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
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &Client{
		client: client,
		logger: common.NewLogger(),
	}, nil
}

// --- Core GitHub Interaction Methods ---

// GetRepository fetches repository information
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to get repository")
	}
	return repository, nil
}

// GetFileContentText retrieves the content of a file as plain text
// NOTE: Now returns *models.FileContent as defined in your models package
func (c *Client) GetFileContentText(ctx context.Context, owner, repo, path, ref string) (*models.FileContent, error) {
	fileContent, _, resp, err := c.client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, common.NewError("file not found: " + path)
		}
		return nil, common.WrapError(err, fmt.Sprintf("failed to get file content for %s", path))
	}
	if fileContent == nil {
		return nil, common.NewError(fmt.Sprintf("received nil github.RepositoryContent for %s (possibly submodule?)", path))
	}
	// Check Encoding before Content, as GetContent relies on Encoding
	if fileContent.Encoding == nil {
 		// Directories might be returned by GetContents if path points to one
 		if fileContent.GetType() == "dir" {
 				return nil, common.NewError(fmt.Sprintf("path points to a directory, not a file: %s", path))
 		}
 		// Otherwise, unexpected nil encoding
 		c.logger.Warnf("Encoding is nil for: %s. Type: %s, Size: %d", path, fileContent.GetType(), fileContent.GetSize())
 		return nil, common.NewError(fmt.Sprintf("encoding is nil for: %s (Type: %s)", path, fileContent.GetType()))
	}
	if fileContent.Content == nil {
		// Log details from fileContent if possible
		c.logger.Warn(fmt.Sprintf("Content is nil for: %s. Type: %s, Size: %d", path, fileContent.GetType(), fileContent.GetSize()))
		if fileContent.GetSize() > 1*1024*1024 { // GitHub API limit is 1MB for contents
			return nil, common.NewError(fmt.Sprintf("file content likely too large for API: %s (size %d)", path, fileContent.GetSize()))
		}
 		return nil, common.NewError(fmt.Sprintf("content is nil for: %s (Type: %s)", path, fileContent.GetType()))
	}


	content, err := fileContent.GetContent() // Decodes base64
	if err != nil {
		return nil, common.WrapError(err, fmt.Sprintf("failed to decode file content for %s", path))
	}

	// Use Getters for nil safety, although we checked fileContent != nil above
	sha := fileContent.GetSHA()
	resolvedPath := fileContent.GetPath()
	if resolvedPath == "" {
		resolvedPath = path // Fallback to original path if GetPath returns empty
	}


	return &models.FileContent{ // Use the imported models.FileContent
		Path:    resolvedPath,
		Content: content,
		SHA:     sha, // Assuming models.FileContent *does* have an SHA field
	}, nil
}


// CreateOrUpdateFile creates or updates a file in a repository
func (c *Client) CreateOrUpdateFile(
	ctx context.Context,
	owner, repo, path, message, content string,
	branch string,
) error {
	var sha string
	// Use GetFileContentText which returns *models.FileContent now
	fileInfo, err := c.GetFileContentText(ctx, owner, repo, path, branch)

	// If file exists (no error or specifically not a 'not found' error)
	var isNotFoundError bool
	if err != nil {
		// Check if the error indicates 'not found'
		if commonErr, ok := err.(*common.AppError); ok && strings.Contains(commonErr.Message, "file not found") {
				isNotFoundError = true
		} else {
			// Log other errors but proceed (attempt create)
			c.logger.WithError(err).Warnf("Could not get existing file info for %s, attempting create/update", path)
		}
	}

	if err == nil && fileInfo != nil {
		sha = fileInfo.SHA // Get SHA from our model struct
	} else if !isNotFoundError {
		 // Log if error wasn't 'not found' but we still proceed
		 c.logger.WithError(err).Warnf("Proceeding without SHA for %s", path)
	}


	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
		SHA:     github.String(sha), // Pass SHA if non-empty (update), empty string is ignored (create)
	}

	_, _, err = c.client.Repositories.CreateFile(ctx, owner, repo, path, opts) // CreateFile handles both create/update based on SHA
	if err != nil {
		return common.WrapError(err, fmt.Sprintf("failed to create or update file %s", path))
	}
	return nil
}


// ListFiles lists files in a directory in a repository
func (c *Client) ListFiles(
	ctx context.Context,
	owner, repo, path, ref string,
) ([]*github.RepositoryContent, error) {
	_, directoryContent, _, err := c.client.Repositories.GetContents(
		ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return nil, common.WrapError(err, fmt.Sprintf("failed to list files for path %s", path))
	}
	return directoryContent, nil
}

// --- Structure and File Listing ---

// GetAllFiles gets all files and directories in a repository using the recursive tree endpoint
func (c *Client) GetAllFiles(ctx context.Context, owner, repo, ref string) ([]models.GitHubFile, error) {
	if ref == "" {
		repoInfo, err := c.GetRepository(ctx, owner, repo)
		if err != nil { return nil, common.WrapError(err, "failed to get repository info for default branch") }
		ref = repoInfo.GetDefaultBranch(); if ref == "" { return nil, common.NewError("default branch not found for repository") }
		c.logger.Info(fmt.Sprintf("No branch specified, using default branch: %s", ref))
	}

	tree, _, err := c.client.Git.GetTree(ctx, owner, repo, ref, true) // true for recursive
	if err != nil { return nil, common.WrapError(err, "failed to get repository tree recursively") }
	if tree == nil || tree.Entries == nil || len(tree.Entries) == 0 {
		c.logger.Warn(fmt.Sprintf("Repository tree is empty or null for %s/%s @ %s", owner, repo, ref)); return []models.GitHubFile{}, nil
	}

	var allFiles []models.GitHubFile
	for _, entry := range tree.Entries {
		if entry.Path == nil || entry.Type == nil { continue }

		fileType := entry.GetType() // Use getter for nil safety
		entryPath := *entry.Path

		// Construct the model using fields available and matching expected model structure (based on original code and compiler errors)
		file := models.GitHubFile{ // Use the imported models.GitHubFile
			Name:        filepath.Base(entryPath),
			Path:        entryPath,
			Size:        entry.GetSize(), // GetSize handles nil pointer
			Type:        fileType,
			// SHA:      entry.GetSHA(), // REMOVED - Compiler indicated this field doesn't exist in models.GitHubFile
			HTMLURL:     fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", owner, repo, ref, entryPath), // Construct HTML URL
			DownloadURL: "", // Not available directly from GetTree, set to empty
		}

		// Standardize type for consistency if needed by model
		if file.Type == "blob" { file.Type = "file" } else
		if file.Type == "tree" { file.Type = "dir" } else
		{ continue } // Skip submodules ("commit") etc.

		allFiles = append(allFiles, file)
	}
	c.logger.Info(fmt.Sprintf("GetAllFiles (recursive tree) found %d files/dirs.", len(allFiles)))
	return allFiles, nil
}


// GetRepositoryStructure retrieves the repository structure as a formatted string
func (c *Client) GetRepositoryStructure(ctx context.Context, owner, repo, ref string) (string, error) {
	// Reuses logic from GetAllFiles for efficiency
	allFiles, err := c.GetAllFiles(ctx, owner, repo, ref) // Use the improved GetAllFiles
	if err != nil {
		return "", common.WrapError(err, "failed to get file list for structure")
	}
	if len(allFiles) == 0 { return "Repository is empty.", nil }

	var sb strings.Builder
	structureMap := make(map[string][]string) // dir path -> list of children names (file or dir/)

	for _, file := range allFiles {
		dir := filepath.Dir(file.Path); if dir == "." { dir = "" } // Root is ""
		childName := file.Name
		if file.Type == "dir" { childName += "/" }
		structureMap[dir] = append(structureMap[dir], childName)
	}

	c.buildStructureString(&sb, structureMap, "", 0)
	return sb.String(), nil
}


// buildStructureString recursively builds the structure string from the map
func (c *Client) buildStructureString(sb *strings.Builder, structureMap map[string][]string, currentDir string, depth int) {
    children, ok := structureMap[currentDir]
    if !ok { return }
    // Consider sorting children alphabetically here if desired
    for _, childName := range children {
		for i := 0; i < depth; i++ { sb.WriteString("  ") }
		sb.WriteString("- "); sb.WriteString(childName); sb.WriteString("\n")
		if strings.HasSuffix(childName, "/") {
			childDirName := strings.TrimSuffix(childName, "/")
			nextDirPath := filepath.Join(currentDir, childDirName)
			// Prevent infinite loops in case of weird data
			if depth < 20 { // Limit recursion depth for safety
				c.buildStructureString(sb, structureMap, nextDirPath, depth+1)
			}
		}
    }
}

// --- Import Map Generation with AST for Go ---

// Enhanced GetImportMap to better handle Go imports and other relationships
func (c *Client) GetImportMap(ctx context.Context, owner, repo, ref string) (map[string][]string, error) {
    c.logger.Info(fmt.Sprintf("Starting GetImportMap for %s/%s @ %s", owner, repo, ref))
    
    // Get all files first
    allFiles, err := c.GetAllFiles(ctx, owner, repo, ref)
    if err != nil { 
        return nil, common.WrapError(err, "failed to get all files for import map") 
    }

    // Build a map of file paths to file types for quick lookups
    existingFiles := make(map[string]string)
    existingPaths := make(map[string]bool) // For exact path matching
    
    for _, f := range allFiles {
        existingFiles[f.Path] = f.Type
        existingPaths[f.Path] = true
    }
    
    c.logger.Info(fmt.Sprintf("Found %d total files/dirs for import analysis.", len(existingFiles)))

    // Initialize the import map
    importMap := make(map[string][]string)
    filesProcessed := 0
    
    // First pass: analyze Go imports using AST
    for _, file := range allFiles {
        // Skip directories and non-source files
        if file.Type != "file" || !isGoSourceFile(file.Path) {
            continue
        }
        
        filesProcessed++
        
        // Get the file content
        content, err := c.GetFileContentText(ctx, owner, repo, file.Path, ref)
        if err != nil {
            c.logger.WithField("error", err).Warning("Skipping import extraction: Failed to get content for file: " + file.Path)
            continue
        }
        
        // Extract Go imports using AST
        imports := extractGoImportsUsingAST(content.Content, c.logger)
        if len(imports) == 0 {
            continue
        }
        
        c.logger.Debug(fmt.Sprintf("Found %d imports in %s (Go)", len(imports), file.Path))
        
        // Resolve import paths to actual files in the repository
        resolvedImports := []string{}
        
        // For each import, try to resolve it to a file in the repository
        for _, imp := range imports {
            // Skip standard library imports and third-party packages
            if !strings.HasPrefix(imp, ".") && !strings.HasPrefix(imp, "github.com/"+owner+"/"+repo) {
                continue
            }
            
            // If it's an internal import (from the same repo), try to resolve it
            if strings.HasPrefix(imp, "github.com/"+owner+"/"+repo) {
                // Remove the repository prefix to get the relative path
                relPath := strings.TrimPrefix(imp, "github.com/"+owner+"/"+repo+"/")
                
                // Check if the file exists in our list
                for existingPath := range existingPaths {
                    if strings.HasPrefix(existingPath, relPath) && existingFiles[existingPath] == "file" {
                        resolvedImports = append(resolvedImports, existingPath)
                        break
                    }
                }
            } else if strings.HasPrefix(imp, ".") {
                // For relative imports
                sourceDir := filepath.Dir(file.Path)
                if sourceDir == "." {
                    sourceDir = ""
                }
                
                // Construct potential paths
                potentialPath := filepath.Clean(filepath.Join(sourceDir, imp))
                
                // Check if it's a directory
                if _, ok := existingFiles[potentialPath]; ok && existingFiles[potentialPath] == "dir" {
                    // Look for a Go file in this directory
                    for existingPath := range existingPaths {
                        dirPath := filepath.Dir(existingPath)
                        if dirPath == potentialPath && filepath.Ext(existingPath) == ".go" {
                            resolvedImports = append(resolvedImports, existingPath)
                            break
                        }
                    }
                } else {
                    // Try with .go extension
                    potentialPath = potentialPath + ".go"
                    if _, ok := existingPaths[potentialPath]; ok {
                        resolvedImports = append(resolvedImports, potentialPath)
                    }
                }
            }
        }
        
        // Add the resolved imports to the import map
        if len(resolvedImports) > 0 {
            // Remove duplicates
            uniqueImports := make(map[string]bool)
            var finalImports []string
            
            for _, imp := range resolvedImports {
                if !uniqueImports[imp] {
                    uniqueImports[imp] = true
                    finalImports = append(finalImports, imp)
                }
            }
            
            importMap[file.Path] = finalImports
        }
    }
    
    // Second pass: analyze other languages and infer relationships based on naming patterns
    // (This is a heuristic approach for other file types)
    for _, file := range allFiles {
        // Skip directories
        if file.Type != "file" {
            continue
        }
        
        // Skip files already processed
        if _, ok := importMap[file.Path]; ok {
            continue
        }
        
        // Try to infer relationships based on naming patterns
        // For example, model.go might be related to model_test.go
        baseName := strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))
        dirPath := filepath.Dir(file.Path)
        
        // Check for test files
        if strings.HasSuffix(baseName, "_test") {
            // This is a test file, find the corresponding implementation file
            implName := strings.TrimSuffix(baseName, "_test") + filepath.Ext(file.Path)
            implPath := filepath.Join(dirPath, implName)
            
            if _, ok := existingPaths[implPath]; ok {
                if _, ok := importMap[file.Path]; !ok {
                    importMap[file.Path] = []string{implPath}
                } else {
                    importMap[file.Path] = append(importMap[file.Path], implPath)
                }
            }
        }
        
        // Check if it's a Go file in a package directory
        if filepath.Ext(file.Path) == ".go" {
            // Find other Go files in the same directory
            sameDirFiles := []string{}
            
            for existingPath := range existingPaths {
                if filepath.Dir(existingPath) == dirPath && 
                   existingPath != file.Path && 
                   filepath.Ext(existingPath) == ".go" {
                    sameDirFiles = append(sameDirFiles, existingPath)
                }
            }
            
            if len(sameDirFiles) > 0 {
                if _, ok := importMap[file.Path]; !ok {
                    importMap[file.Path] = sameDirFiles
                } else {
                    importMap[file.Path] = append(importMap[file.Path], sameDirFiles...)
                }
            }
        }
    }
    
    c.logger.Info(fmt.Sprintf("Processed %d source files for imports.", filesProcessed))
    c.logger.Info(fmt.Sprintf("Final import map contains %d source files with resolved imports.", len(importMap)))
    
    return importMap, nil
}

// Helper function to check if a file is a Go source file
func isGoSourceFile(path string) bool {
    return filepath.Ext(path) == ".go"
}

// resolveImportPath tries to resolve a raw import string to a full file path within the repo.
func (c *Client) resolveImportPath(sourceDir, rawImport, language string, existingFiles map[string]string) string {
	rawImport = strings.TrimSpace(rawImport)
	if rawImport == "" { return "" }
	if !strings.HasPrefix(rawImport, ".") { return "" } // Ignore non-relative for now

	potentialTargetPath := filepath.Join(sourceDir, rawImport)
	potentialTargetPath = filepath.Clean(potentialTargetPath)

	var possibleExtensions []string; var indexFileNames []string
	switch language {
	case "JavaScript", "TypeScript":
		possibleExtensions = []string{".js", ".jsx", ".ts", ".tsx", ".json"}
		indexFileNames = []string{"index.js", "index.jsx", "index.ts", "index.tsx"}
	case "Python":
		possibleExtensions = []string{".py"}
		indexFileNames = []string{"__init__.py"}
	case "C", "C++":
		possibleExtensions = []string{".h", ".hpp"} // Primarily headers are included relatively
	} // Add others if needed

	// A. Direct file match
	if ftype, ok := existingFiles[potentialTargetPath]; ok && ftype == "file" { return potentialTargetPath }
	// B. Match with extension
	for _, ext := range possibleExtensions {
		pathWithExt := potentialTargetPath + ext
		if ftype, ok := existingFiles[pathWithExt]; ok && ftype == "file" { return pathWithExt }
	}
	// C. Match directory + index file
	if dtype, ok := existingFiles[potentialTargetPath]; ok && dtype == "dir" {
		for _, indexFile := range indexFileNames {
			indexPath := filepath.Join(potentialTargetPath, indexFile)
			if ftype, ok := existingFiles[indexPath]; ok && ftype == "file" { return indexPath }
		}
	}
	return "" // No resolution found
}


// extractImports uses AST for Go, basic regex for others.
func extractImports(content, language string, logger *common.Logger) []string {
	switch language {
	case "Go":
		return extractGoImportsUsingAST(content, logger)
	case "JavaScript", "TypeScript":
		logger.Debug("Using basic Regex extraction for JavaScript/TypeScript")
		return extractJSImportsRegex(content)
	case "Python":
		logger.Debug("Using basic Regex extraction for Python")
		return extractPythonImportsRegex(content)
	case "C", "C++":
		logger.Debug("Using basic Regex extraction for C/C++")
		return extractCImportsRegex(content)
	default:
		logger.Debug(fmt.Sprintf("Import extraction not implemented for language: %s", language))
		return []string{}
	}
}

// --- AST Parsing for Go ---

func extractGoImportsUsingAST(content string, logger *common.Logger) []string {
	fset := token.NewFileSet()
	// ParseFile may return partial AST on syntax errors, try to proceed if possible
	fileAst, err := parser.ParseFile(fset, "", content, parser.ImportsOnly | parser.ParseComments) // Include comments if needed later
	if err != nil {
		// Log error but don't necessarily stop if fileAst is partially valid
		logger.WithError(err).Warn("Failed to fully parse Go file, attempting to extract imports from partial AST")
		if fileAst == nil { return nil } // Cannot proceed if AST is totally nil
	}

	imports := []string{}
	if fileAst.Imports != nil { // Check if Imports slice exists
		for _, imp := range fileAst.Imports {
			if imp.Path != nil && imp.Path.Value != "" {
				importPath, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					logger.WithError(err).Warnf("Failed to unquote Go import path: %s", imp.Path.Value)
					continue
				}
				imports = append(imports, importPath)
			}
		}
	}
	return imports
}

// --- Basic Regex Import Extraction (Placeholders) ---

var (
    // Fix for jsImportRegex - replace backreference with character class
    jsImportRegex    = regexp.MustCompile(`(?m)^\s*import(?:["'\s]*(?:[\w*{}\n\r\s,./[\]]+from\s*)?)(["'])(?P<path>.*?[^\\])(["'])`)
    
    // Fix for jsRequireRegex
    jsRequireRegex   = regexp.MustCompile(`(?m)require\((["'])(?P<path>.*?[^\\])(["'])\)`)
    
    // Fix for jsDynImportRegex
    jsDynImportRegex = regexp.MustCompile(`(?m)import\((["'])(?P<path>.*?[^\\])(["'])\)`)
    
    // These regexes don't use backreferences, so they're fine
    pyImportRegex    = regexp.MustCompile(`(?m)^\s*import\s+((?:\.?\w+)(?:\s*,\s*\.?\w+)*)`)
    pyFromImportRegex= regexp.MustCompile(`(?m)^\s*from\s+((?:\.+)?[\w\.]+)\s+import`)
    cIncludeRegex    = regexp.MustCompile(`(?m)^\s*#include\s*(["<])(?P<path>[^>"]+)[">]`)
)


func removeComments(content, language string) string { // Basic comment removal
	var singleLineComment string
	switch language {
	case "Go", "JavaScript", "TypeScript", "Java", "C", "C++", "C#", "Swift", "Kotlin": singleLineComment = "//"
	case "Python", "Ruby": singleLineComment = "#"
	case "PHP": singleLineComment = "//"
	default: return content
	}

	var cleanedLines []string; lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if singleLineComment != "" && strings.HasPrefix(trimmedLine, singleLineComment) { continue }
		if language == "PHP" && strings.HasPrefix(trimmedLine, "#") { continue }
		if singleLineComment != "" {
			if idx := strings.Index(line, singleLineComment); idx != -1 {
				quoteCountBefore := strings.Count(line[:idx], "\"") + strings.Count(line[:idx], "'") + strings.Count(line[:idx], "`")
				if quoteCountBefore%2 == 0 { line = line[:idx] }
			}
		}
		// Simplified block comment removal (very basic)
		if language == "JavaScript" || language == "TypeScript" || language == "Go" || language == "Java" || language == "C" || language == "C++" || language == "C#" {
			if startIdx := strings.Index(line, "/*"); startIdx != -1 {
				if endIdx := strings.Index(line, "*/"); endIdx > startIdx { line = line[:startIdx] + line[endIdx+2:] } else { line = line[:startIdx] }
			}
		}
		cleanedLines = append(cleanedLines, line)
	}
	return strings.Join(cleanedLines, "\n")
}

func extractJSImportsRegex(content string) []string { // Basic JS/TS extraction
	content = removeComments(content, "JavaScript"); imports := make(map[string]bool)
	processMatches := func(re *regexp.Regexp) {
		matches := re.FindAllStringSubmatch(content, -1)
		pathIndex := re.SubexpIndex("path")
		if pathIndex == -1 { pathIndex = len(re.SubexpNames())-1 } // Fallback to last group
		for _, match := range matches { if len(match) > pathIndex && match[pathIndex] != "" { imports[match[pathIndex]] = true } }
	}
	processMatches(jsImportRegex); processMatches(jsRequireRegex); processMatches(jsDynImportRegex)
	result := []string{}; for imp := range imports { result = append(result, imp) }
	return result
}

func extractPythonImportsRegex(content string) []string { // Basic Python extraction
	content = removeComments(content, "Python"); imports := make(map[string]bool)
	matches := pyImportRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches { if len(match) > 1 { for _, mod := range strings.Split(match[1], ",") { imports[strings.TrimSpace(mod)] = true } } }
	matches = pyFromImportRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches { if len(match) > 1 { imports[match[1]] = true } } // Stores 'x' or '.x' from 'from x import y'

	result := []string{};
	for imp := range imports { // Convert module paths like 'a.b' or '.a.b' to relative paths
		leadingDots := 0; tempImp := imp
		for strings.HasPrefix(tempImp, ".") { leadingDots++; tempImp = tempImp[1:] }
		pathParts := strings.Split(tempImp, ".")
		// Reconstruct path: add leading "../" for each dot, join parts with "/"
		finalPath := strings.Repeat("../", leadingDots) + strings.Join(pathParts, "/")
		// Remove trailing slash if it was just dots (e.g., "from . import x" -> "../")
		if finalPath == strings.Repeat("../", leadingDots) && len(pathParts) == 1 && pathParts[0] == "" {
				if leadingDots > 0 { finalPath = strings.Repeat("../", leadingDots-1) } else { continue } // Skip empty result
		}
		if finalPath != "" {	result = append(result, finalPath) }
	}
	return result
}


func extractCImportsRegex(content string) []string { // Basic C/C++ extraction
	content = removeComments(content, "C"); imports := make(map[string]bool)
	matches := cIncludeRegex.FindAllStringSubmatch(content, -1)
	pathIndex := cIncludeRegex.SubexpIndex("path")
	if pathIndex == -1 { pathIndex = len(cIncludeRegex.SubexpNames())-1 }
	for _, match := range matches { if len(match) > pathIndex && match[pathIndex] != "" { imports[match[pathIndex]] = true } }
	result := []string{}; for imp := range imports { result = append(result, imp) }
	return result
}

// --- Other Helper Functions ---

func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	sourceExts := map[string]bool{ ".go": true, ".java": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".py": true, ".rb": true, ".php": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true, ".cs": true, ".swift": true, ".kt": true, ".kts": true, }
	return sourceExts[ext]
}

func getLanguageFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go": return "Go"; case ".java": return "Java"; case ".js", ".jsx": return "JavaScript"; case ".ts", ".tsx": return "TypeScript"; case ".py": return "Python"; case ".rb": return "Ruby"; case ".php": return "PHP"; case ".c", ".h": return "C"; case ".cpp", ".hpp": return "C++"; case ".cs": return "C#"; case ".swift": return "Swift"; case ".kt", ".kts": return "Kotlin"
	default: return "Unknown"
	}
}

// GetRepositoryInfo fetches combined repository information
// NOTE: Uses the imported models.RepositoryInfo
func (c *Client) GetRepositoryInfo(ctx context.Context, owner, repo string) (*models.RepositoryInfo, error) {
	repository, err := c.GetRepository(ctx, owner, repo)
	if err != nil { return nil, common.WrapError(err, "failed to get base repository info") }

	languages, _, err := c.client.Repositories.ListLanguages(ctx, owner, repo)
	if err != nil { c.logger.WithField("error", err).Warning("Failed to list repository languages"); languages = make(map[string]int) }
	primaryLang := ""; maxBytes := 0
	for lang, bytes := range languages { if bytes > maxBytes { maxBytes = bytes; primaryLang = lang } }

	return &models.RepositoryInfo{ // Use imported models.RepositoryInfo
		Owner: owner, Name: repo, Description: repository.GetDescription(), DefaultBranch: repository.GetDefaultBranch(), URL: repository.GetHTMLURL(), Language: primaryLang, Languages: languages, Stars: repository.GetStargazersCount(), Forks: repository.GetForksCount(),
	}, nil
}


// GetFunctionCode extracts a function's code (remains basic)
func (c *Client) GetFunctionCode(ctx context.Context, owner, repo, path, functionName, ref string) (string, int, int, error) {
	// Keeping the previous basic (non-AST) implementation for brevity
	content, err := c.GetFileContentText(ctx, owner, repo, path, ref)
	if err != nil { return "", 0, 0, common.WrapError(err, "failed to get file content") }

	lines := strings.Split(content.Content, "\n"); startLine, endLine := -1, -1; inFunction := false; braceCount := 0
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if !inFunction && (strings.Contains(trimmedLine, "func "+functionName+"(") || strings.Contains(trimmedLine, "function "+functionName+"(") || strings.Contains(trimmedLine, functionName+" = function(") || strings.Contains(trimmedLine, "def "+functionName+"(")) {
			startLine = i; inFunction = true; braceCount = strings.Count(trimmedLine, "{") - strings.Count(trimmedLine, "}"); continue
		}
		if inFunction {
			braceCount += strings.Count(trimmedLine, "{") - strings.Count(trimmedLine, "}")
			if braceCount <= 0 && (strings.Contains(trimmedLine, "}") || (startLine == i && !strings.Contains(trimmedLine,"{")) ){ endLine = i; break }
			if startLine != -1 && i > startLine && trimmedLine != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				 if i > 0 { prevTrimmed := strings.TrimSpace(lines[i-1]); if prevTrimmed == "" || !strings.HasPrefix(lines[i-1], " ") && !strings.HasPrefix(lines[i-1], "\t") { endLine = i - 1; break }
				 } else { endLine = startLine; break }
			}
		}
	}
	if startLine == -1 { return "", 0, 0, common.NewError("function not found: " + functionName) }
	if endLine == -1 { endLine = len(lines) - 1 }
	functionLines := lines[startLine : endLine+1]
	return strings.Join(functionLines, "\n"), startLine + 1, endLine + 1, nil
}