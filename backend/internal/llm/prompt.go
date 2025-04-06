package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// buildReadmePrompt builds a prompt for generating a README.md file
func buildReadmePrompt(repoInfo map[string]interface{}, files []string) string {
	repoInfoStr, _ := json.MarshalIndent(repoInfo, "", "  ")
	filesList := strings.Join(files, "\n")
	
	return fmt.Sprintf(`
You are an expert technical writer. Your task is to create a comprehensive README.md file for a GitHub repository.

Here is the repository information:
%s

Here is a list of important files in the repository:
%s

Create a professional README.md that includes:
1. A clear title and description of the project
2. Installation instructions
3. Usage examples
4. Features list
5. Technology stack information
6. How to contribute (if applicable)
7. License information (if available)
8. Proper Markdown formatting with headings, code blocks, and lists
9. Badges if applicable (e.g., build status, version)

Your response will be directly copy and paste in the Github README.md file. Do not wrap your response in backticks or any other formatting. Just provide the content of the README.md. 
`, repoInfoStr, filesList)
}

// buildDockerfilePrompt builds a prompt for generating a Dockerfile
func buildDockerfilePrompt(repoInfo map[string]interface{}, mainLanguage string) string {
	repoInfoStr, _ := json.MarshalIndent(repoInfo, "", "  ")
	
	return fmt.Sprintf(`
You are an expert DevOps engineer. Your task is to create an optimized Dockerfile for a GitHub repository.

Here is the repository information:
%s

The main programming language of this repository is: %s

Create a production-ready Dockerfile that:
1. Uses the appropriate base image for the language/framework
2. Follows best practices (multi-stage builds if appropriate)
3. Minimizes image size
4. Sets up proper working directories
5. Handles dependencies efficiently
6. Exposes any necessary ports
7. Includes appropriate labels
8. Provides clear comments explaining each step

Format your response as a complete Dockerfile with comments, ready to be used in the repository. Do not wrap your response in backticks or any other formatting. Just provide the content of the Dockerfile.
`, repoInfoStr, mainLanguage)
}

// buildCodeCommentsPrompt builds a prompt for generating code comments
func buildCodeCommentsPrompt(code string, language string) string {
	// Avoid triple backticks in the prompt
	return fmt.Sprintf(`
You are an expert developer in %s. Your task is to add comprehensive comments to the following code:

CODE START
%s
CODE END

Please add:
1. File-level documentation explaining the overall purpose
2. Function/method-level documentation explaining:
   - Purpose
   - Parameters
   - Return values
   - Any exceptions/errors thrown
3. Comments for complex logic sections
4. Do NOT change the actual code, only add comments

Return the commented code in the same language and formatting as the original.
`, language, code)
}

// buildCodeRefactorPrompt builds a prompt for code refactoring
func buildCodeRefactorPrompt(code string, language string, instructions string) string {
	// Avoid triple backticks in the prompt
	return fmt.Sprintf(`
You are an expert developer in %s. Your task is to refactor the following code according to these instructions:

%s

Here is the code to refactor:

CODE START
%s
CODE END

Please provide:
1. The refactored code
2. A brief explanation of the changes made
3. Benefits of the refactoring

Return the refactored code in the same language as the original.
`, language, instructions, code)
}

// buildCodeSearchPrompt builds a prompt for analyzing code search results
func buildCodeSearchPrompt(query string, searchResults string) string {
	return fmt.Sprintf(`
You are an expert code analyzer. Your task is to analyze the following search results for the query "%s" and provide insights.

Here are the search results:
%s

Please provide:
1. A summary of what the search results reveal
2. Key code patterns or issues found
3. Suggestions for improvements if applicable
4. Any potential security or performance concerns based on these results

Format your response in a clear, structured manner with headings and bullet points where appropriate.
`, query, searchResults)
}