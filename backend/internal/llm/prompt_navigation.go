// internal/llm/prompt_navigation.go
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// buildCodeWalkthroughPrompt builds a prompt for generating a code walkthrough
func buildCodeWalkthroughPrompt(repoInfo map[string]interface{}, codebase map[string]string, entryPoints []string) string {
	repoInfoStr, _ := json.MarshalIndent(repoInfo, "", "  ")
	entryPointsStr := strings.Join(entryPoints, "\n")
	
	return fmt.Sprintf(`
You are an expert software architect and developer. Your task is to create a comprehensive walkthrough of a codebase to help newcomers understand it.

Here is the repository information:
%s

The main entry points of the application are:
%s

Please create a code walkthrough that:
1. Identifies the key components and their purposes
2. Maps the flow of control through the application
3. Explains how different parts of the codebase interact
4. Highlights important design patterns or architectural decisions
5. Suggests a logical order for exploring the code

For each important file or component, provide:
- A brief description of its purpose
- Its role in the overall architecture
- Key functions or methods to understand
- Dependencies and relationships with other components

Format your response as a structured walkthrough with clear sections and a logical progression.
`, repoInfoStr, entryPointsStr)
}

// buildFunctionExplainerPrompt builds a prompt for explaining a function
func buildFunctionExplainerPrompt(functionCode string, language string, fileName string) string {
	return fmt.Sprintf(`
You are an expert developer in %s. Your task is to explain the following function from file '%s' in detail.

Here is the function code:

CODE START
%s
CODE END

Please provide:
1. A clear description of what this function does
2. An explanation of each parameter:
   - Name
   - Type
   - Purpose
3. An explanation of return values:
   - Type
   - Meaning
   - Possible error conditions
4. A usage example showing how to call this function
5. Any notable algorithms, patterns, or techniques used
6. Potential edge cases or limitations
7. Performance characteristics if relevant

Format your response in a clear, structured manner that would help a newcomer understand this function.
`, language, fileName, functionCode)
}

// buildArchitectureVisualizerPrompt builds a prompt for visualizing architecture
func buildArchitectureVisualizerPrompt(repoInfo map[string]interface{}, fileStructure string, importMap map[string][]string) string {
	repoInfoStr, _ := json.MarshalIndent(repoInfo, "", "  ")
	importMapStr, _ := json.MarshalIndent(importMap, "", "  ")
	
	return fmt.Sprintf(`
You are an expert software architect. Your task is to analyze a codebase and create an architecture visualization.

Here is the repository information:
%s

Here is the file structure:
%s

Here is the import relationship map between files:
%s

Please provide:
1. A high-level overview of the architecture
2. A description of the major components and their responsibilities
3. The relationships and dependencies between components
4. Data flow through the system
5. Key architectural patterns used

Your response should include a structured representation of the architecture that could be used to generate a diagram, including:
- Nodes (components, modules, services)
- Edges (dependencies, calls, imports)
- Component descriptions

Format your response in a clear, structured manner suitable for generating an architecture diagram.
`, repoInfoStr, fileStructure, importMapStr)
}

// buildCodebaseQAPrompt builds a prompt for answering questions about a codebase
func buildCodebaseQAPrompt(question string, relevantCode map[string]string) string {
	var codeStr strings.Builder
	for file, content := range relevantCode {
		codeStr.WriteString(fmt.Sprintf("FILE: %s\n\n", file))
		codeStr.WriteString(content)
		codeStr.WriteString("\n\n---\n\n")
	}
	
	return fmt.Sprintf(`
You are an expert code analyst. Your task is to answer a question about a codebase using the available code snippets.

Question: %s

Here are the relevant code snippets:

%s

Please provide:
1. A comprehensive answer to the question
2. References to specific parts of the code that support your answer
3. Any additional context that would help understand the answer
4. If relevant, suggest 2-3 follow-up questions that might be helpful

Format your response in a clear, structured manner, and be sure to cite specific files and line numbers when referring to code.
`, question, codeStr.String())
}

// buildBestPracticesPrompt builds a prompt for generating best practices guide
// Update the buildBestPracticesPrompt function in internal/llm/prompt_navigation.go
func buildBestPracticesPrompt(repoInfo map[string]interface{}, codebase map[string]string) string {
    repoInfoStr, _ := json.MarshalIndent(repoInfo, "", "  ")
    
    var codeStr strings.Builder
    for file, content := range codebase {
        codeStr.WriteString(fmt.Sprintf("FILE: %s\n\n", file))
        codeStr.WriteString(content)
        codeStr.WriteString("\n\n---\n\n")
    }
    
    return fmt.Sprintf(`
You are an expert code quality analyst. Your task is to identify specific coding patterns, conventions, and best practices in a codebase.

Here is the repository information:
%s

Here are code samples from the repository:

%s

Please provide a DETAILED analysis including:

1. A comprehensive style guide with specific examples from the code, including:
   - Naming conventions (variables, functions, classes)
   - Code organization (file structure, module organization)
   - Formatting standards (indentation, line length, whitespace)
   - Comment style and documentation practices

2. Best practices that are consistently followed, with specific examples from the code:
   - Design patterns used
   - Error handling approaches
   - Testing methodologies
   - Performance optimization techniques

3. Any unusual patterns or conventions that newcomers should be aware of, citing specific examples

4. Potential issues or inconsistencies in coding style, with line references

5. Concrete recommendations for maintaining consistent code quality with examples of how to apply them

Your response should be extremely specific and detailed, based directly on the provided code samples, not generic advice. Each observation should cite specific code examples.

Do not use placeholder text like "appears to follow standard conventions" without explaining exactly what those conventions are.
`, repoInfoStr, codeStr.String())
}

// GenerateCompletion generates a simple text completion
func (c *GeminiClient) GenerateCompletion(ctx context.Context, prompt string, temperature float32, maxTokens int) (string, error) {
    c.logger.Info("Generating completion with Gemini")
    
    // Validate parameters and set defaults if needed
    if prompt == "" {
        return "", common.NewError("prompt cannot be empty")
    }
    
    if temperature <= 0 {
        temperature = 0.7 // default temperature
    }
    
    if maxTokens <= 0 {
        maxTokens = 1024 // default max tokens
    }
    
    // Configure generation parameters
    c.model.SetTemperature(temperature)
    c.model.SetMaxOutputTokens(int32(maxTokens))
    c.model.SetTopP(0.95)
    c.model.SetTopK(40)
    
    // Set safety settings to be more permissive for code-related content
    c.model.SafetySettings = []*genai.SafetySetting{
        {
            Category:  genai.HarmCategoryHarassment,
            Threshold: genai.HarmBlockNone,
        },
        {
            Category:  genai.HarmCategoryHateSpeech,
            Threshold: genai.HarmBlockNone,
        },
        {
            Category:  genai.HarmCategoryDangerousContent,
            Threshold: genai.HarmBlockNone,
        },
        {
            Category:  genai.HarmCategorySexuallyExplicit,
            Threshold: genai.HarmBlockNone,
        },
    }
    
    // Generate content using the Gemini model
    resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        return "", common.WrapError(err, "failed to generate completion")
    }
    
    // Validate response
    if len(resp.Candidates) == 0 {
        return "", common.NewError("no candidates in response")
    }
    
    // Check if the response was blocked for safety reasons
    if resp.Candidates[0].FinishReason == genai.FinishReasonSafety {
        return "", common.NewError("content filtered due to safety concerns")
    }
    
    // Extract the text from the response
    var result string
    for _, part := range resp.Candidates[0].Content.Parts {
        if text, ok := part.(genai.Text); ok {
            result += string(text)
        }
    }
    
    if result == "" {
        return "", common.NewError("empty text in response")
    }
    
    c.logger.Info("Completion generated successfully")
    return result, nil
}