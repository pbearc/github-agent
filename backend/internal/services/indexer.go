package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/pinecone"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// CodeChunk represents a chunk of code from a file
type CodeChunk struct {
	FilePath    string
	Content     string
	StartLine   int
	EndLine     int
	ChunkNumber int
}

// IndexerService handles codebase indexing
type IndexerService struct {
	githubClient   *github.Client
	pineconeClient *pinecone.Client
	llmClient      *llm.GeminiClient
	logger         *common.Logger
}

// NewIndexerService creates a new IndexerService instance
func NewIndexerService(
	githubClient *github.Client,
	pineconeClient *pinecone.Client,
	llmClient *llm.GeminiClient,
) *IndexerService {
	return &IndexerService{
		githubClient:   githubClient,
		pineconeClient: pineconeClient,
		llmClient:      llmClient,
		logger:         common.NewLogger(),
	}
}

// IndexRepository indexes a GitHub repository
func (s *IndexerService) IndexRepository(ctx context.Context, owner, repo, branch string) (string, error) {
	// Generate a namespace for this repo+branch
	namespace := fmt.Sprintf("%s-%s-%s", owner, repo, branch)
	
	// Get repository info for metadata
	_, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		return "", common.WrapError(err, "failed to get repository info")
	}

	// Get repository structure
	repoStructure, err := s.githubClient.GetRepositoryStructure(ctx, owner, repo, branch)
	if err != nil {
		return "", common.WrapError(err, "failed to get repository structure")
	}
	
	// Parse file paths from repository structure
	filePaths := splitLines(repoStructure)
	
	// Filter to only include code files
	var codeFilePaths []string
	for _, path := range filePaths {
		if isCodeFile(path) {
			codeFilePaths = append(codeFilePaths, path)
		}
	}
	
	s.logger.Info(fmt.Sprintf("Found %d code files to index", len(codeFilePaths)))
	
	// Delete existing vectors for this namespace if they exist
	err = s.pineconeClient.Delete(ctx, pinecone.DeleteRequest{
		Namespace: namespace,
		DeleteAll: true,
	})
	if err != nil {
		s.logger.WithField("error", err).Warning("Failed to delete existing vectors, continuing...")
	}
	
	// Process each file
	var vectors []pinecone.Vector
	totalChunks := 0
	
	for _, path := range codeFilePaths {
		// Get file content
		fileContent, err := s.githubClient.GetFileContentText(ctx, owner, repo, path, branch)
		if err != nil {
			s.logger.WithField("error", err).WithField("path", path).Warning("Failed to get file content, skipping")
			continue
		}
		
		// Split file into chunks
		chunks := s.splitFileIntoChunks(fileContent.Content, path)
		totalChunks += len(chunks)
		
		// Process each chunk
		for _, chunk := range chunks {
			// Generate embedding for chunk
			embedding, err := s.llmClient.CreateEmbedding(ctx, chunk.Content)
			if err != nil {
				s.logger.WithField("error", err).WithField("path", chunk.FilePath).Warning("Failed to create embedding, skipping")
				continue
			}
			
			// Create Pinecone vector
			id := uuid.New().String()
			
			// Create metadata
			metadata := map[string]interface{}{
				"owner":       owner,
				"repo":        repo,
				"branch":      branch,
				"filePath":    chunk.FilePath,
				"startLine":   chunk.StartLine,
				"endLine":     chunk.EndLine,
				"chunkNumber": chunk.ChunkNumber,
				"language":    detectLanguageFromPath(chunk.FilePath),
				"timestamp":   time.Now().Unix(),
			}
			
			// Add vector to batch
			vector := pinecone.Vector{
				ID:       id,
				Values:   embedding,
				Metadata: metadata,
			}
			
			vectors = append(vectors, vector)
			
			// Upsert in batches of 100
			if len(vectors) >= 100 {
				count, err := s.pineconeClient.Upsert(ctx, vectors, namespace)
				if err != nil {
					return "", common.WrapError(err, "failed to upsert vectors")
				}
				
				s.logger.Info(fmt.Sprintf("Indexed batch of %d vectors", count))
				vectors = []pinecone.Vector{} // Reset batch
			}
		}
	}
	
	// Upsert any remaining vectors
	if len(vectors) > 0 {
		count, err := s.pineconeClient.Upsert(ctx, vectors, namespace)
		if err != nil {
			return "", common.WrapError(err, "failed to upsert remaining vectors")
		}
		
		s.logger.Info(fmt.Sprintf("Indexed final batch of %d vectors", count))
	}
	
	s.logger.Info(fmt.Sprintf("Indexed %d chunks from %d files", totalChunks, len(codeFilePaths)))
	
	return namespace, nil
}

// splitFileIntoChunks splits a file into chunks for embedding
func (s *IndexerService) splitFileIntoChunks(content, filePath string) []CodeChunk {
	// Split content into lines
	lines := strings.Split(content, "\n")
	
	// Initialize result
	var chunks []CodeChunk
	
	// Determine chunk size based on file size
	// We'll use smaller chunks for larger files
	chunkSize := 100 // Default chunk size in lines
	if len(lines) > 1000 {
		chunkSize = 50
	}
	
	// Split into chunks with some overlap
	overlap := 10 // Lines of overlap between chunks
	
	for i := 0; i < len(lines); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(lines) {
			end = len(lines)
		}
		
		// Skip if we're at the end and this would be a very small chunk
		if i > 0 && end-i < 20 && end == len(lines) {
			break
		}
		
		chunkContent := strings.Join(lines[i:end], "\n")
		
		chunks = append(chunks, CodeChunk{
			FilePath:    filePath,
			Content:     chunkContent,
			StartLine:   i + 1, // 1-based line numbers
			EndLine:     end,
			ChunkNumber: len(chunks) + 1,
		})
		
		// If we've reached the end, break
		if end == len(lines) {
			break
		}
	}
	
	return chunks
}

// isCodeFile determines if a file is a code file worth indexing
func isCodeFile(path string) bool {
	// List of file extensions to index
	codeExtensions := []string{
		".go", ".py", ".js", ".ts", ".jsx", ".tsx", ".java", ".c", ".cpp", ".cs", 
		".rb", ".php", ".swift", ".kt", ".rs", ".sh", ".html", ".css", ".md",
		".yml", ".yaml", ".json", ".xml", ".sql", ".h", ".proto", ".scala",
	}
	
	ext := filepath.Ext(path)
	for _, codeExt := range codeExtensions {
		if ext == codeExt {
			return true
		}
	}
	
	return false
}