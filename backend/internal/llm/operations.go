package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/pbearc/github-agent/backend/pkg/common"
)

// OperationType defines the type of LLM operation
type OperationType string

const (
	ReadmeGeneration    OperationType = "readme_generation"
	DockerfileGeneration OperationType = "dockerfile_generation"
	CodeComments        OperationType = "code_comments"
	CodeRefactor        OperationType = "code_refactor"
	CodeAnalysis        OperationType = "code_analysis"
)

// Operation represents an LLM operation request
type Operation struct {
	Type         OperationType      `json:"type"`
	RepoInfo     map[string]interface{} `json:"repo_info,omitempty"`
	Language     string             `json:"language,omitempty"`
	Code         string             `json:"code,omitempty"`
	Files        []string           `json:"files,omitempty"`
	Instructions string             `json:"instructions,omitempty"`
	Query        string             `json:"query,omitempty"`
	SearchResults string            `json:"search_results,omitempty"`
}

// ProcessOperation processes an LLM operation and returns the result
func (c *GeminiClient) ProcessOperation(ctx context.Context, op *Operation) (string, error) {
	if op == nil {
		return "", common.NewError("operation cannot be nil")
	}

	switch op.Type {
	case ReadmeGeneration:
		return c.GenerateReadme(ctx, op.RepoInfo, op.Files)
	case DockerfileGeneration:
		language := op.Language
		if language == "" && op.RepoInfo != nil {
			if langIface, ok := op.RepoInfo["language"]; ok {
				if lang, ok := langIface.(string); ok {
					language = lang
				}
			}
		}
		return c.GenerateDockerfile(ctx, op.RepoInfo, language)
	case CodeComments:
		if op.Code == "" {
			return "", common.NewError("code cannot be empty for code comments operation")
		}
		return c.GenerateCodeComments(ctx, op.Code, op.Language)
	case CodeRefactor:
		if op.Code == "" {
			return "", common.NewError("code cannot be empty for code refactor operation")
		}
		return c.GenerateCodeRefactor(ctx, op.Code, op.Language, op.Instructions)
	case CodeAnalysis:
		if op.SearchResults == "" {
			return "", common.NewError("search results cannot be empty for code analysis operation")
		}
		return c.GenerateText(ctx, buildCodeSearchPrompt(op.Query, op.SearchResults))
	default:
		return "", common.NewError(fmt.Sprintf("unsupported operation type: %s", op.Type))
	}
}

// ValidateOperation validates an operation request
func ValidateOperation(op *Operation) error {
	if op == nil {
		return common.NewError("operation cannot be nil")
	}

	switch op.Type {
	case ReadmeGeneration:
		if op.RepoInfo == nil {
			return common.NewError("repository information is required for README generation")
		}
	case DockerfileGeneration:
		if op.RepoInfo == nil {
			return common.NewError("repository information is required for Dockerfile generation")
		}
	case CodeComments:
		if op.Code == "" {
			return common.NewError("code is required for code comments operation")
		}
		if op.Language == "" {
			return common.NewError("language is required for code comments operation")
		}
	case CodeRefactor:
		if op.Code == "" {
			return common.NewError("code is required for code refactor operation")
		}
		if op.Language == "" {
			return common.NewError("language is required for code refactor operation")
		}
	case CodeAnalysis:
		if op.Query == "" {
			return common.NewError("query is required for code analysis operation")
		}
		if op.SearchResults == "" {
			return common.NewError("search results are required for code analysis operation")
		}
	default:
		return common.NewError(fmt.Sprintf("unsupported operation type: %s", op.Type))
	}

	return nil
}

// GetModel returns the currently configured model
func (c *GeminiClient) GetModel() string {
	return "gemini-1.5-pro" // This is hardcoded for now, but could be made configurable
}

// DetectLanguage attempts to detect the programming language from code
func DetectLanguage(code string) string {
	// This is a very basic implementation
	// A more robust implementation would use language detection libraries or heuristics
	
	if strings.Contains(code, "function") && strings.Contains(code, "var") {
		return "javascript"
	}
	if strings.Contains(code, "def") && strings.Contains(code, "import") {
		return "python"
	}
	if strings.Contains(code, "func") && strings.Contains(code, "package") {
		return "go"
	}
	if strings.Contains(code, "class") && strings.Contains(code, "public") {
		return "java"
	}
	
	return "unknown"
}