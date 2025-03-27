package models

import "github.com/pbearc/github-agent/backend/internal/llm"

// RepositoryRequest contains the request data for repository operations
type RepositoryRequest struct {
	URL     string `json:"url" binding:"required"`
	Branch  string `json:"branch"`
}

// GenerateReadmeRequest contains the request data for README generation
type GenerateReadmeRequest struct {
	RepositoryRequest
	IncludeFiles bool `json:"include_files"`
}

// GenerateDockerfileRequest contains the request data for Dockerfile generation
type GenerateDockerfileRequest struct {
	RepositoryRequest
	Language string `json:"language"`
}

// CodeCommentsRequest contains the request data for code commenting
type CodeCommentsRequest struct {
	RepositoryRequest
	FilePath string `json:"file_path" binding:"required"`
}

// CodeRefactorRequest contains the request data for code refactoring
type CodeRefactorRequest struct {
	RepositoryRequest
	FilePath     string `json:"file_path" binding:"required"`
	Instructions string `json:"instructions" binding:"required"`
}

// CodeSearchRequest contains the request data for code searching
type CodeSearchRequest struct {
	RepositoryRequest
	Query string `json:"query" binding:"required"`
}

// LLMOperationRequest contains the request data for a generic LLM operation
type LLMOperationRequest struct {
	Operation llm.Operation `json:"operation" binding:"required"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// GenerateResponse represents the response for generation operations
type GenerateResponse struct {
	Content string `json:"content"`
}

// RepositoryInfoResponse represents the response for repository info
type RepositoryInfoResponse struct {
	Owner         string   `json:"owner"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	DefaultBranch string   `json:"default_branch"`
	URL           string   `json:"url"`
	Language      string   `json:"language"`
	Languages     []string `json:"languages"`
	Stars         int      `json:"stars"`
	Forks         int      `json:"forks"`
	HasReadme     bool     `json:"has_readme"`
}

// FileContentResponse represents the response for file content
type FileContentResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// CodeSearchResponse represents the response for code search
type CodeSearchResponse struct {
	Query    string       `json:"query"`
	Results  []SearchItem `json:"results"`
	Analysis string       `json:"analysis,omitempty"`
}

// SearchItem represents a single search result
type SearchItem struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	HTMLLink string `json:"html_link"`
}