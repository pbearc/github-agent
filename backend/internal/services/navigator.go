package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/internal/pinecone"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// NavigatorService handles codebase navigation and Q&A
type NavigatorService struct {
	githubClient   *github.Client
	pineconeClient *pinecone.Client
	llmClient      *llm.GeminiClient
	indexerService *IndexerService
	logger         *common.Logger
}

// NewNavigatorService creates a new NavigatorService instance
func NewNavigatorService(
	githubClient *github.Client,
	pineconeClient *pinecone.Client,
	llmClient *llm.GeminiClient,
) *NavigatorService {
	indexerService := NewIndexerService(githubClient, pineconeClient, llmClient)
	
	return &NavigatorService{
		githubClient:   githubClient,
		pineconeClient: pineconeClient,
		llmClient:      llmClient,
		indexerService: indexerService,
		logger:         common.NewLogger(),
	}
}

// EnsureRepositoryIndexed ensures that a repository is indexed
func (s *NavigatorService) EnsureRepositoryIndexed(ctx context.Context, owner, repo, branch string) (string, error) {
	// Generate the namespace
	namespace := fmt.Sprintf("%s-%s-%s", owner, repo, branch)
	
	// Check if the namespace exists and has vectors
	stats, err := s.pineconeClient.DescribeIndexStats(ctx)
	if err != nil {
		return "", common.WrapError(err, "failed to describe index stats")
	}
	
	// Check if the namespace has any vectors
	namespaceExists := false
	if stats.Namespaces != nil {
		if _, ok := stats.Namespaces[namespace]; ok {
			namespaceExists = true
		}
	}
	
	// If namespace doesn't exist or is empty, index the repository
	if !namespaceExists {
		s.logger.Info(fmt.Sprintf("Namespace %s does not exist, indexing repository", namespace))
		return s.indexerService.IndexRepository(ctx, owner, repo, branch)
	}
	
	return namespace, nil
}

// AnswerQuestion answers a question about a codebase
func (s *NavigatorService) AnswerQuestion(ctx context.Context, owner, repo, branch, question string, topK int) (*models.CodebaseNavigatorResponse, error) {
	// Ensure the repository is indexed
	namespace, err := s.EnsureRepositoryIndexed(ctx, owner, repo, branch)
	if err != nil {
		return nil, common.WrapError(err, "failed to ensure repository is indexed")
	}
	
	// Generate embedding for the question
	questionEmbedding, err := s.llmClient.CreateEmbedding(ctx, question)
	if err != nil {
		return nil, common.WrapError(err, "failed to create embedding for question")
	}
	
	// Query Pinecone for similar vectors
	if topK <= 0 {
		topK = 5 // Default to 5 results
	}
	
	queryResp, err := s.pineconeClient.Query(ctx, pinecone.QueryRequest{
		Vector:          questionEmbedding,
		TopK:            topK,
		Namespace:       namespace,
		IncludeMetadata: true,
	})
	if err != nil {
		return nil, common.WrapError(err, "failed to query Pinecone")
	}
	
	if len(queryResp.Matches) == 0 {
		return &models.CodebaseNavigatorResponse{
			Answer: "I couldn't find any relevant code to answer your question.",
			RelevantFiles: []models.RelevantFile{},
		}, nil
	}
	
	// Get content for matched chunks
	var relevantFiles []models.RelevantFile
	var contextBuilder strings.Builder
	
	for _, match := range queryResp.Matches {
		metadata := match.Metadata
		filePath, _ := metadata["filePath"].(string)
		
		// Get file content
		fileContent, err := s.githubClient.GetFileContentText(ctx, owner, repo, filePath, branch)
		if err != nil {
			s.logger.WithField("error", err).WithField("path", filePath).Warning("Failed to get file content")
			continue
		}
		
		// Extract the chunk using line numbers from metadata
		startLineFloat, _ := metadata["startLine"].(float64)
		endLineFloat, _ := metadata["endLine"].(float64)
		
		startLine := int(startLineFloat)
		endLine := int(endLineFloat)
		
		// Extract just the chunk (we have the lines already)
		lines := strings.Split(fileContent.Content, "\n")
		start := startLine - 1 // 0-based index
		end := endLine
		if end > len(lines) {
			end = len(lines)
		}
		
		chunkContent := strings.Join(lines[start:end], "\n")
		
		// Add to context for LLM
		contextBuilder.WriteString(fmt.Sprintf("File: %s (lines %d-%d)\n", filePath, startLine, endLine))
		contextBuilder.WriteString("```\n")
		contextBuilder.WriteString(chunkContent)
		contextBuilder.WriteString("\n```\n\n")
		
		// Add to relevantFiles for response
		relevantFiles = append(relevantFiles, models.RelevantFile{
			Path:      filePath,
			Snippet:   chunkContent,
			Relevance: float64(match.Score),
		})
	}
	
	// Sort relevantFiles by relevance
	sort.Slice(relevantFiles, func(i, j int) bool {
		return relevantFiles[i].Relevance > relevantFiles[j].Relevance
	})
	
	// Build the prompt for the LLM
	prompt := fmt.Sprintf(`
You are an expert codebase navigator. Your task is to answer questions about a GitHub repository.

You're looking at the codebase for the repository %s/%s (branch: %s).

Here are the most relevant code sections related to the question:

%s

User Question: %s

Please provide:
1. A clear and comprehensive answer to the question
2. References to specific files and line numbers from the provided code
3. Explanations of how the relevant code works
4. Context about how this fits into the broader codebase if applicable

Format your response in markdown with appropriate headings, code blocks, and organization.
`, owner, repo, branch, contextBuilder.String(), question)
	
	// Generate answer with LLM
	answer, err := s.llmClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate answer")
	}
	
	return &models.CodebaseNavigatorResponse{
		Answer: answer,
		RelevantFiles: relevantFiles,
	}, nil
}