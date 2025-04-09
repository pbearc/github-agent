package llm

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"github.com/pbearc/github-agent/backend/pkg/common"
	"google.golang.org/api/option"
)

// GeminiClient represents a Gemini API client
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
	logger *common.Logger
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, common.NewError("Gemini API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, common.WrapError(err, "failed to create Gemini client")
	}

	// Use Gemini-1.5-pro model
	model := client.GenerativeModel("gemini-1.5-pro")

	return &GeminiClient{
		client: client,
		model:  model,
		logger: common.NewLogger(),
	}, nil
}

// GenerateText generates a text response based on the provided prompt
func (c *GeminiClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", common.NewError("prompt cannot be empty")
	}

	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", common.WrapError(err, "failed to generate content")
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", common.NewError("no response generated")
	}

	// Get the text from the response
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", common.NewError("unexpected response format")
	}

	return string(text), nil
}

// GenerateReadme generates a README.md file based on repository information
func (c *GeminiClient) GenerateReadme(ctx context.Context, repoInfo map[string]interface{}, files []string) (string, error) {
	// Create a prompt for generating a README
	prompt := buildReadmePrompt(repoInfo, files)
	
	return c.GenerateText(ctx, prompt)
}

// GenerateDockerfile generates a Dockerfile based on repository information
func (c *GeminiClient) GenerateDockerfile(ctx context.Context, repoInfo map[string]interface{}, mainLanguage string) (string, error) {
	// Create a prompt for generating a Dockerfile
	prompt := buildDockerfilePrompt(repoInfo, mainLanguage)
	
	return c.GenerateText(ctx, prompt)
}

// GenerateCodeComments generates comments for a code file
func (c *GeminiClient) GenerateCodeComments(ctx context.Context, code string, language string) (string, error) {
	// Create a prompt for generating code comments
	prompt := buildCodeCommentsPrompt(code, language)
	
	return c.GenerateText(ctx, prompt)
}

// GenerateCodeRefactor suggests refactoring for a code file
func (c *GeminiClient) GenerateCodeRefactor(ctx context.Context, code string, language string, instructions string) (string, error) {
	// Create a prompt for code refactoring
	prompt := buildCodeRefactorPrompt(code, language, instructions)
	
	return c.GenerateText(ctx, prompt)
}

// Close closes the client
func (c *GeminiClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// CreateEmbedding generates an embedding for a text using the Gemini API
// CreateEmbedding generates an embedding for a text using the Gemini API
func (c *GeminiClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, common.NewError("text cannot be empty")
	}

	// Use the Gemini embedding model
	model := c.client.EmbeddingModel("models/embedding-001")
	
	// Create the embedding - directly use the text as a part
	resp, err := model.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, common.WrapError(err, "failed to create embedding")
	}

	// Return embeddings
	return resp.Embedding.Values, nil
}
// GetEmbeddingDimension returns the dimension of the Gemini embeddings
func (c *GeminiClient) GetEmbeddingDimension() int {
	return 768 // Gemini embeddings are 768 dimensions
}

// GenerateCodeWalkthrough generates a code walkthrough
func (c *GeminiClient) GenerateCodeWalkthrough(ctx context.Context, repoInfo map[string]interface{}, codebase map[string]string, entryPoints []string) (string, error) {
	prompt := buildCodeWalkthroughPrompt(repoInfo, codebase, entryPoints)
	return c.GenerateText(ctx, prompt)
}

// ExplainFunction generates an explanation for a function
func (c *GeminiClient) ExplainFunction(ctx context.Context, functionCode string, language string, fileName string) (string, error) {
	prompt := buildFunctionExplainerPrompt(functionCode, language, fileName)
	return c.GenerateText(ctx, prompt)
}

// VisualizeArchitecture generates an architecture visualization
func (c *GeminiClient) VisualizeArchitecture(ctx context.Context, repoInfo map[string]interface{}, fileStructure string, importMap map[string][]string) (string, error) {
	prompt := buildArchitectureVisualizerPrompt(repoInfo, fileStructure, importMap)
	return c.GenerateText(ctx, prompt)
}

// AnswerCodebaseQuestion answers a question about the codebase
func (c *GeminiClient) AnswerCodebaseQuestion(ctx context.Context, question string, relevantCode map[string]string) (string, error) {
	prompt := buildCodebaseQAPrompt(question, relevantCode)
	return c.GenerateText(ctx, prompt)
}

// GenerateBestPracticesGuide generates a best practices guide
func (c *GeminiClient) GenerateBestPracticesGuide(ctx context.Context, repoInfo map[string]interface{}, codebase map[string]string) (string, error) {
	prompt := buildBestPracticesPrompt(repoInfo, codebase)
	return c.GenerateText(ctx, prompt)
}