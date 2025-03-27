package services

import (
	"context"
	"path/filepath"

	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// CommenterService handles code commenting
type CommenterService struct {
	githubClient *github.Client
	llmClient    *llm.GeminiClient
	logger       *common.Logger
}

// NewCommenterService creates a new CommenterService instance
func NewCommenterService(githubClient *github.Client, llmClient *llm.GeminiClient) *CommenterService {
	return &CommenterService{
		githubClient: githubClient,
		llmClient:    llmClient,
		logger:       common.NewLogger(),
	}
}

// GenerateComments generates comments for a code file
func (s *CommenterService) GenerateComments(ctx context.Context, owner, repo, branch, path string) (string, error) {
	// Get file content
	fileContent, err := s.githubClient.GetFileContentText(ctx, owner, repo, path, branch)
	if err != nil {
		return "", common.WrapError(err, "failed to get file content")
	}

	// Detect language from file extension and actually use it in the LLM call
	language := detectLanguageFromPath(path)

	// Generate comments
	commentedCode, err := s.llmClient.GenerateCodeComments(ctx, fileContent.Content, language)
	if err != nil {
		return "", common.WrapError(err, "failed to generate comments")
	}

	return commentedCode, nil
}

// PushCommentedFile generates comments for a file and pushes it to the repository
func (s *CommenterService) PushCommentedFile(ctx context.Context, owner, repo, branch, path string) error {
	// Generate comments
	commentedCode, err := s.GenerateComments(ctx, owner, repo, branch, path)
	if err != nil {
		return err
	}

	// If branch is not specified, get the default branch
	if branch == "" {
		repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
		if err != nil {
			return common.WrapError(err, "failed to get repository info")
		}
		branch = repoInfo.DefaultBranch
	}

	// Push the commented file
	commitMessage := "Add comments to " + path + " via GitHub Agent"
	err = s.githubClient.CreateOrUpdateFile(ctx, owner, repo, path, commitMessage, commentedCode, branch)
	if err != nil {
		return common.WrapError(err, "failed to push commented file")
	}

	return nil
}

// AddSummaryCommentToFile adds a summary comment to a file without modifying the rest of the file
func (s *CommenterService) AddSummaryCommentToFile(ctx context.Context, owner, repo, branch, path string) error {
	// Get file content
	fileContent, err := s.githubClient.GetFileContentText(ctx, owner, repo, path, branch)
	if err != nil {
		return common.WrapError(err, "failed to get file content")
	}

	// Detect language from file extension - this is used by AddCommentToFile
	language := detectLanguageFromPath(path)
	s.logger.WithField("language", language).Debug("Detected language for comment")

	// Generate a summary comment
	summaryPrompt := `
You are an expert developer. Your task is to write a brief but comprehensive summary comment for this code file.
The summary should explain the overall purpose of the file, key functionality, and any important considerations.
Do NOT include any code, just the summary text that would go in a file header comment.
Keep the summary concise (5-7 lines maximum).

Here is the code:
` + fileContent.Content

	summary, err := s.llmClient.GenerateText(ctx, summaryPrompt)
	if err != nil {
		return common.WrapError(err, "failed to generate summary comment")
	}

	// Add the summary comment to the file
	err = s.githubClient.AddCommentToFile(ctx, owner, repo, path, summary, branch)
	if err != nil {
		return common.WrapError(err, "failed to add summary comment")
	}

	return nil
}

// detectLanguageFromPath detects the programming language from a file path
func detectLanguageFromPath(path string) string {
	ext := filepath.Ext(path)
	
	switch ext {
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".go":
		return "go"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cpp", ".cc", ".cxx":
		return "c++"
	case ".cs":
		return "c#"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".rs":
		return "rust"
	case ".sh":
		return "shell"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".md":
		return "markdown"
	default:
		return "unknown"
	}
}