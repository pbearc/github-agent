// internal/services/codenavigation.go
package services

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// CodeNavigationService handles code navigation features
type CodeNavigationService struct {
	githubClient *github.Client
	llmClient    *llm.GeminiClient
	logger       *common.Logger
}

// NewCodeNavigationService creates a new CodeNavigationService instance
func NewCodeNavigationService(githubClient *github.Client, llmClient *llm.GeminiClient) *CodeNavigationService {
	return &CodeNavigationService{
		githubClient: githubClient,
		llmClient:    llmClient,
		logger:       common.NewLogger(),
	}
}

// GenerateCodeWalkthrough generates a code walkthrough for a repository
func (s *CodeNavigationService) GenerateCodeWalkthrough(ctx context.Context, owner, repo, branch string, depth int, focusPath string, entryPoints []string) (*models.CodeWalkthroughResponse, error) {
	// Get repository info
	repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to get repository info")
	}

	// If branch is not specified, use the default branch
	if branch == "" {
		branch = repoInfo.DefaultBranch
	}

	// Get some basic file structure
	_, err = s.githubClient.GetRepositoryStructure(ctx, owner, repo, branch)
	if err != nil {
		return nil, common.WrapError(err, "failed to get repository structure")
	}

	// Auto-detect entry points if not provided
	if len(entryPoints) == 0 {
		entryPoints = s.detectEntryPoints(ctx, owner, repo, branch, repoInfo.Language)
	}

	// Collect sample code for LLM context
	codebase := make(map[string]string)
	
	// Add entry points
	for _, entryPoint := range entryPoints {
		content, err := s.githubClient.GetFileContentText(ctx, owner, repo, entryPoint, branch)
		if err != nil {
			s.logger.WithField("error", err).Warning("Failed to get content for entry point: " + entryPoint)
			continue
		}
		codebase[entryPoint] = content.Content
	}

	// Update the focus path handling in internal/services/codenavigation.go
	if focusPath != "" {
		// Try to get file content first
		focusContent, err := s.githubClient.GetFileContentText(ctx, owner, repo, focusPath, branch)
		if err == nil {
			// It's a file, add to codebase
			codebase[focusPath] = focusContent.Content
		} else {
			// It might be a directory, try to list files
			s.logger.WithField("error", err).Warning("Failed to get content for focus path, trying as directory: " + focusPath)
			files, dirErr := s.githubClient.ListFiles(ctx, owner, repo, focusPath, branch)
			if dirErr == nil && len(files) > 0 {
				// It's a directory, get content for some files
				for i, file := range files {
					if i >= 3 || file == nil || file.Path == nil {
						continue // Limit to 3 files
					}
					
					// Only process source code files
					filePath := *file.Path
					if !github.IsSourceFile(filePath) {
						continue
					}
					
					contentFile, contentErr := s.githubClient.GetFileContentText(ctx, owner, repo, filePath, branch)
					if contentErr == nil {
						codebase[filePath] = contentFile.Content
					}
				}
			}
		}
	}

	// Convert repository info to a map for the LLM
	repoInfoMap := map[string]interface{}{
		"owner":          repoInfo.Owner,
		"name":           repoInfo.Name,
		"description":    repoInfo.Description,
		"default_branch": repoInfo.DefaultBranch,
		"url":            repoInfo.URL,
		"language":       repoInfo.Language,
		"stars":          repoInfo.Stars,
		"forks":          repoInfo.Forks,
	}

	// Generate walkthrough using LLM
	walkthroughJSON, err := s.llmClient.GenerateCodeWalkthrough(ctx, repoInfoMap, codebase, entryPoints)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate code walkthrough")
	}

	// Parse the JSON response
	var walkthrough models.CodeWalkthroughResponse
	err = json.Unmarshal([]byte(walkthroughJSON), &walkthrough)
	if err != nil {
		// If JSON parsing fails, try to structure the text response
		walkthrough = s.structureWalkthroughText(walkthroughJSON, entryPoints)
	}

	return &walkthrough, nil
}

// ExplainFunction explains a function
func (s *CodeNavigationService) ExplainFunction(ctx context.Context, owner, repo, branch, filePath, functionName string, lineStart, lineEnd int) (*models.FunctionExplainerResponse, error) {
	// Get repository info
	repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to get repository info")
	}

	// If branch is not specified, use the default branch
	if branch == "" {
		branch = repoInfo.DefaultBranch
	}

	var functionCode string
	
	// If line numbers are specified, extract the function using those
	if lineStart > 0 && lineEnd > 0 {
		content, err := s.githubClient.GetFileContentText(ctx, owner, repo, filePath, branch)
		if err != nil {
			return nil, common.WrapError(err, "failed to get file content")
		}
		
		lines := strings.Split(content.Content, "\n")
		if lineStart <= len(lines) && lineEnd <= len(lines) && lineStart <= lineEnd {
			// internal/services/codenavigation.go (continued)
            functionLines := lines[lineStart-1:lineEnd]
            functionCode = strings.Join(functionLines, "\n")
        } else {
            return nil, common.NewError("invalid line numbers")
        }
    } else {
        // Try to detect the function automatically
        var startLine, endLine int
        functionCode, startLine, endLine, err = s.githubClient.GetFunctionCode(ctx, owner, repo, filePath, functionName, branch)
        if err != nil {
            return nil, common.WrapError(err, "failed to extract function code")
        }
        lineStart = startLine
        lineEnd = endLine
    }

    // Get the language from the file path
    language := github.GetLanguageFromPath(filePath)

    // Generate explanation using LLM
    explanationJSON, err := s.llmClient.ExplainFunction(ctx, functionCode, language, filePath)
    if err != nil {
        return nil, common.WrapError(err, "failed to explain function")
    }

    // Parse the JSON response
    var explanation models.FunctionExplainerResponse
    err = json.Unmarshal([]byte(explanationJSON), &explanation)
    if err != nil {
        // If JSON parsing fails, try to structure the text response
        explanation = s.structureFunctionExplanation(explanationJSON, functionName)
    }

    return &explanation, nil
}

// VisualizeArchitecture generates an architecture visualization
func (s *CodeNavigationService) VisualizeArchitecture(ctx context.Context, owner, repo, branch, detail string, focusPaths []string) (*models.ArchitectureVisualizerResponse, error) {
    // Get repository info
    repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
    if err != nil {
        return nil, common.WrapError(err, "failed to get repository info")
    }

    // If branch is not specified, use the default branch
    if branch == "" {
        branch = repoInfo.DefaultBranch
    }

    // Get the file structure
    fileStructure, err := s.githubClient.GetRepositoryStructure(ctx, owner, repo, branch)
    if err != nil {
        return nil, common.WrapError(err, "failed to get repository structure")
    }

    // Get import relationships
    importMap, err := s.githubClient.GetImportMap(ctx, owner, repo, branch)
    if err != nil {
        s.logger.WithField("error", err).Warning("Failed to get import map")
        importMap = make(map[string][]string)
    }

    // Convert repository info to a map for the LLM
    repoInfoMap := map[string]interface{}{
        "owner":          repoInfo.Owner,
        "name":           repoInfo.Name,
        "description":    repoInfo.Description,
        "default_branch": repoInfo.DefaultBranch,
        "url":            repoInfo.URL,
        "language":       repoInfo.Language,
        "stars":          repoInfo.Stars,
        "forks":          repoInfo.Forks,
    }

    // Generate architecture visualization using LLM
    architectureJSON, err := s.llmClient.VisualizeArchitecture(ctx, repoInfoMap, fileStructure, importMap)
    if err != nil {
        return nil, common.WrapError(err, "failed to visualize architecture")
    }

    // Parse the JSON response
    var architecture models.ArchitectureVisualizerResponse
    err = json.Unmarshal([]byte(architectureJSON), &architecture)
    if err != nil {
        // If JSON parsing fails, try to structure the text response
        architecture = s.structureArchitectureVisualization(architectureJSON)
    }

    return &architecture, nil
}

// AnswerCodebaseQuestion answers a question about the codebase
func (s *CodeNavigationService) AnswerCodebaseQuestion(ctx context.Context, owner, repo, branch, question string) (*models.CodebaseQAResponse, error) {
    // Get repository info
    repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
    if err != nil {
        return nil, common.WrapError(err, "failed to get repository info")
    }

    // If branch is not specified, use the default branch
    if branch == "" {
        branch = repoInfo.DefaultBranch
    }

    // Search for relevant code based on the question
    searchResults, err := s.githubClient.SearchCode(ctx, owner, repo, question)
    if err != nil {
        s.logger.WithField("error", err).Warning("Failed to search code")
    }

    // Collect relevant code snippets
    relevantCode := make(map[string]string)
    for _, result := range searchResults {
        if result == nil || result.Path == nil {
            continue
        }
        
        content, err := s.githubClient.GetFileContentText(ctx, owner, repo, *result.Path, branch)
        if err != nil {
            s.logger.WithField("error", err).Warning("Failed to get content for file: " + *result.Path)
            continue
        }
        
        relevantCode[*result.Path] = content.Content
    }

    // If we couldn't find enough relevant code, get some key files
    if len(relevantCode) < 3 {
        entryPoints := s.detectEntryPoints(ctx, owner, repo, branch, repoInfo.Language)
        for _, entryPoint := range entryPoints {
            if _, exists := relevantCode[entryPoint]; !exists {
                content, err := s.githubClient.GetFileContentText(ctx, owner, repo, entryPoint, branch)
                if err != nil {
                    continue
                }
                relevantCode[entryPoint] = content.Content
            }
        }
    }

    // Generate answer using LLM
    answerJSON, err := s.llmClient.AnswerCodebaseQuestion(ctx, question, relevantCode)
    if err != nil {
        return nil, common.WrapError(err, "failed to answer question")
    }

    // Parse the JSON response
    var answer models.CodebaseQAResponse
    err = json.Unmarshal([]byte(answerJSON), &answer)
    if err != nil {
        // If JSON parsing fails, try to structure the text response
        answer = s.structureQAResponse(answerJSON, question, relevantCode)
    }

    return &answer, nil
}

// GenerateBestPracticesGuide generates a best practices guide
func (s *CodeNavigationService) GenerateBestPracticesGuide(ctx context.Context, owner, repo, branch, scope, path string) (*models.BestPracticesResponse, error) {
    // Get repository info
    repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
    if err != nil {
        return nil, common.WrapError(err, "failed to get repository info")
    }

    // If branch is not specified, use the default branch
    if branch == "" {
        branch = repoInfo.DefaultBranch
    }

    // Determine the scope of files to analyze
    var filesToAnalyze []string
    if scope == "file" && path != "" {
        filesToAnalyze = []string{path}
    } else if scope == "directory" && path != "" {
        files, err := s.githubClient.ListFiles(ctx, owner, repo, path, branch)
        if err != nil {
            return nil, common.WrapError(err, "failed to list files in directory")
        }
        
        for _, file := range files {
            if file == nil || file.Path == nil || file.Type == nil {
                continue
            }
            
            if *file.Type == "file" && github.IsSourceFile(*file.Path) {
                filesToAnalyze = append(filesToAnalyze, *file.Path)
            }
        }
    } else {
        // Full repository scope - get a sample of source files
        allFiles, err := s.githubClient.GetAllFiles(ctx, owner, repo, branch)
        if err != nil {
            return nil, common.WrapError(err, "failed to get all files")
        }
        
        var sourceFiles []string
        for _, file := range allFiles {
            if file.Type == "file" && github.IsSourceFile(file.Path) {
                sourceFiles = append(sourceFiles, file.Path)
            }
        }
        
        // Limit to a reasonable number of files for analysis
        maxFiles := 10
        if len(sourceFiles) > maxFiles {
            filesToAnalyze = sourceFiles[:maxFiles]
        } else {
            filesToAnalyze = sourceFiles
        }
    }

    // Collect code from files to analyze
    codebase := make(map[string]string)
    for _, filePath := range filesToAnalyze {
        content, err := s.githubClient.GetFileContentText(ctx, owner, repo, filePath, branch)
        if err != nil {
            s.logger.WithField("error", err).Warning("Failed to get content for file: " + filePath)
            continue
        }
        
        codebase[filePath] = content.Content
    }

    // Convert repository info to a map for the LLM
    repoInfoMap := map[string]interface{}{
        "owner":          repoInfo.Owner,
        "name":           repoInfo.Name,
        "description":    repoInfo.Description,
        "default_branch": repoInfo.DefaultBranch,
        "url":            repoInfo.URL,
        "language":       repoInfo.Language,
        "stars":          repoInfo.Stars,
        "forks":          repoInfo.Forks,
    }

    // Generate best practices guide using LLM
	bestPracticesJSON, err := s.llmClient.GenerateBestPracticesGuide(ctx, repoInfoMap, codebase)
	if err != nil {
		return nil, common.WrapError(err, "failed to generate best practices guide")
	}

	s.logger.WithField("llm_response", bestPracticesJSON).Info("LLM response for best practices")

	// If JSON parsing fails, try to extract some data from text
	var bestPractices models.BestPracticesResponse
	err = json.Unmarshal([]byte(bestPracticesJSON), &bestPractices)
	if err != nil {
		s.logger.WithField("error", err).Warning("Failed to parse LLM response as JSON, using text extraction")
		bestPractices = s.structureBestPracticesGuide(bestPracticesJSON)
		
		// If the structured response is still empty, create a minimal response
		if len(bestPractices.Practices) == 0 && bestPractices.StyleGuide == "" {
			// Add at least one practice to avoid empty response
			bestPractices.StyleGuide = "Style guide could not be automatically generated. Please review the codebase manually."
			bestPractices.Practices = append(bestPractices.Practices, models.BestPractice{
				Title: "Code Organization",
				Description: "The repository structure appears to follow standard conventions. Further manual analysis is recommended.",
				Examples: []string{"Review the directory structure to understand the organization."},
			})
		}
	}

	return &bestPractices, nil
}

// Helper functions for detecting entry points and structuring LLM responses

// detectEntryPoints attempts to identify the main entry points of a repository
func (s *CodeNavigationService) detectEntryPoints(ctx context.Context, owner, repo, branch, language string) []string {
    var entryPoints []string
    
    // Look for common entry point patterns based on language
    switch language {
    case "Go":
        // Check for main.go files
        files, err := s.githubClient.GetAllFiles(ctx, owner, repo, branch)
        if err != nil {
            s.logger.WithField("error", err).Warning("Failed to get all files when detecting entry points")
            return entryPoints
        }
        
        for _, file := range files {
            if file.Type == "file" && (file.Name == "main.go" || strings.HasSuffix(file.Path, "/main.go")) {
                entryPoints = append(entryPoints, file.Path)
            }
        }
        
    case "JavaScript", "TypeScript":
        // Check for index.js, app.js, server.js, etc.
        commonFiles := []string{
            "index.js", "app.js", "server.js", "main.js",
            "index.ts", "app.ts", "server.ts", "main.ts",
        }
        
        files, err := s.githubClient.GetAllFiles(ctx, owner, repo, branch)
        if err != nil {
            s.logger.WithField("error", err).Warning("Failed to get all files when detecting entry points")
            return entryPoints
        }
        
        for _, file := range files {
            if file.Type == "file" {
                for _, commonFile := range commonFiles {
                    if file.Name == commonFile || strings.HasSuffix(file.Path, "/"+commonFile) {
                        entryPoints = append(entryPoints, file.Path)
                        break
                    }
                }
            }
        }
        
    case "Python":
        // Check for __main__.py, app.py, main.py, etc.
        commonFiles := []string{"__main__.py", "app.py", "main.py", "run.py"}
        
        files, err := s.githubClient.GetAllFiles(ctx, owner, repo, branch)
        if err != nil {
            s.logger.WithField("error", err).Warning("Failed to get all files when detecting entry points")
            return entryPoints
        }
        
        for _, file := range files {
            if file.Type == "file" {
                for _, commonFile := range commonFiles {
                    if file.Name == commonFile || strings.HasSuffix(file.Path, "/"+commonFile) {
                        entryPoints = append(entryPoints, file.Path)
                        break
                    }
                }
            }
        }
    }
    
    return entryPoints
}

// structureWalkthroughText structures raw text into a CodeWalkthroughResponse
func (s *CodeNavigationService) structureWalkthroughText(text string, entryPoints []string) models.CodeWalkthroughResponse {
    walkthrough := models.CodeWalkthroughResponse{
        EntryPoints:  entryPoints,
        Dependencies: make(map[string][]string),
    }
    
    // Extract overview (assume it's at the beginning)
    lines := strings.Split(text, "\n")
    var overviewBuilder strings.Builder
    inOverview := true
    
    for _, line := range lines {
        if strings.Contains(strings.ToLower(line), "walkthrough") || 
           strings.Contains(strings.ToLower(line), "step") {
            inOverview = false
        }
        
        if inOverview {
            overviewBuilder.WriteString(line)
            overviewBuilder.WriteString("\n")
        } else {
            // Start building walkthrough steps
            break
        }
    }
    
    walkthrough.Overview = strings.TrimSpace(overviewBuilder.String())
    
    // Try to extract walkthrough steps
    var steps []models.CodeWalkthroughStep
    var currentStep *models.CodeWalkthroughStep
    
    for _, line := range lines {
        // Look for step headers or file mentions
        if (strings.Contains(strings.ToLower(line), "step") || 
            strings.Contains(line, ".go") || 
            strings.Contains(line, ".js") || 
            strings.Contains(line, ".py")) && 
            (strings.Contains(line, ":") || strings.Contains(line, "-")) {
            
            // Save previous step if exists
            if currentStep != nil {
                steps = append(steps, *currentStep)
            }
            
            // Start new step
            currentStep = &models.CodeWalkthroughStep{
                Name:        line,
                Importance:  5, // Default mid-importance
            }
            
            // Try to extract file path
            for _, entryPoint := range entryPoints {
                if strings.Contains(line, entryPoint) {
                    currentStep.Path = entryPoint
                    break
                }
            }
            
            continue
        }
        
        // Add description to current step
        if currentStep != nil {
            currentStep.Description += line + "\n"
        }
    }
    
    // Add the final step
    if currentStep != nil {
        steps = append(steps, *currentStep)
    }
    
    walkthrough.Walkthrough = steps
    
    return walkthrough
}

// structureFunctionExplanation structures raw text into a FunctionExplainerResponse
func (s *CodeNavigationService) structureFunctionExplanation(text string, functionName string) models.FunctionExplainerResponse {
    explanation := models.FunctionExplainerResponse{
        FunctionName: functionName,
    }
    
    // Extract description (assume it's at the beginning)
    lines := strings.Split(text, "\n")
    var descriptionBuilder strings.Builder
    inDescription := true
    
    for i, line := range lines {
        lowercase := strings.ToLower(line)
        
        if strings.Contains(lowercase, "parameter") || 
           strings.Contains(lowercase, "argument") || 
           strings.Contains(lowercase, "return") {
            inDescription = false
            
            // Extract parameters section
            if strings.Contains(lowercase, "parameter") || strings.Contains(lowercase, "argument") {
                var params []models.Param
                for j := i + 1; j < len(lines); j++ {
                    paramLine := lines[j]
                    if strings.TrimSpace(paramLine) == "" {
                        continue
                    }
                    
                    if strings.Contains(strings.ToLower(paramLine), "return") {
                        break
                    }
                    
                    // Try to parse parameter
                    param := models.Param{}
                    
                    // Look for name and type
                    if strings.Contains(paramLine, ":") {
                        parts := strings.SplitN(paramLine, ":", 2)
                        param.Name = strings.TrimSpace(parts[0])
                        param.Description = strings.TrimSpace(parts[1])
                    } else if strings.Contains(paramLine, "-") {
                        parts := strings.SplitN(paramLine, "-", 2)
                        param.Name = strings.TrimSpace(parts[0])
                        param.Description = strings.TrimSpace(parts[1])
                    } else {
                        continue
                    }
                    
                    // Try to extract type from description
                    if strings.Contains(param.Description, "(") && strings.Contains(param.Description, ")") {
                        typeStart := strings.Index(param.Description, "(")
                        typeEnd := strings.Index(param.Description, ")")
                        if typeStart != -1 && typeEnd != -1 && typeEnd > typeStart {
                            param.Type = strings.TrimSpace(param.Description[typeStart+1:typeEnd])
                            param.Description = strings.TrimSpace(param.Description[:typeStart] + param.Description[typeEnd+1:])
                        }
                    }
                    
                    params = append(params, param)
                }
                
                explanation.Parameters = params
            }
            
            // Extract return values section
            if strings.Contains(lowercase, "return") {
                var returnValues []models.Param
                for j := i + 1; j < len(lines); j++ {
                    retLine := lines[j]
                    if strings.TrimSpace(retLine) == "" {
                        continue
                    }
                    
                    if strings.Contains(strings.ToLower(retLine), "usage") || 
                       strings.Contains(strings.ToLower(retLine), "example") {
                        break
                    }
                    
                    // Try to parse return value
                    retVal := models.Param{}
                    
                    // Look for type and description
                    if strings.Contains(retLine, ":") {
                        parts := strings.SplitN(retLine, ":", 2)
                        retVal.Name = strings.TrimSpace(parts[0])
                        retVal.Description = strings.TrimSpace(parts[1])
                    } else if strings.Contains(retLine, "-") {
                        parts := strings.SplitN(retLine, "-", 2)
                        retVal.Name = strings.TrimSpace(parts[0])
                        retVal.Description = strings.TrimSpace(parts[1])
                    } else {
                        continue
                    }
                    
                    // Try to extract type
                    if strings.Contains(retVal.Description, "(") && strings.Contains(retVal.Description, ")") {
                        typeStart := strings.Index(retVal.Description, "(")
                        typeEnd := strings.Index(retVal.Description, ")")
                        if typeStart != -1 && typeEnd != -1 && typeEnd > typeStart {
                            retVal.Type = strings.TrimSpace(retVal.Description[typeStart+1:typeEnd])
                            retVal.Description = strings.TrimSpace(retVal.Description[:typeStart] + retVal.Description[typeEnd+1:])
                        }
                    }
                    
                    returnValues = append(returnValues, retVal)
                }
                
                explanation.ReturnValues = returnValues
            }
            
            break
        }
        
        if inDescription {
            descriptionBuilder.WriteString(line)
            descriptionBuilder.WriteString("\n")
        }
    }
    
    explanation.Description = strings.TrimSpace(descriptionBuilder.String())
    
    // Try to extract usage examples
    var usageExamples []string
    inUsage := false
    var usageBuilder strings.Builder
    
    for _, line := range lines {
        lowercase := strings.ToLower(line)
        
        if strings.Contains(lowercase, "usage") || strings.Contains(lowercase, "example") {
            inUsage = true
            continue
        }
        
        if inUsage {
            if strings.Contains(lowercase, "complexity") || strings.Contains(lowercase, "related") {
                // End of usage section
                usageExamples = append(usageExamples, strings.TrimSpace(usageBuilder.String()))
                usageBuilder.Reset()
                inUsage = false
            } else {
                usageBuilder.WriteString(line)
                usageBuilder.WriteString("\n")
            }
        }
    }
    
    // Add the final usage example if still building one
    if inUsage && usageBuilder.Len() > 0 {
        usageExamples = append(usageExamples, strings.TrimSpace(usageBuilder.String()))
    }
    
    explanation.Usage = usageExamples
    
    return explanation
}

// structureArchitectureVisualization structures raw text into an ArchitectureVisualizerResponse
func (s *CodeNavigationService) structureArchitectureVisualization(text string) models.ArchitectureVisualizerResponse {
    architecture := models.ArchitectureVisualizerResponse{
        ComponentDescriptions: make(map[string]string),
        DiagramData: models.DiagramData{
            Nodes: []models.DiagramNode{},
            Edges: []models.DiagramEdge{},
        },
    }
    
    // Extract overview
    sections := strings.Split(text, "##")
    if len(sections) > 1 {
        architecture.Overview = strings.TrimSpace(sections[0])
    } else {
        // Try different section marker
        sections = strings.Split(text, "#")
        if len(sections) > 1 {
            architecture.Overview = strings.TrimSpace(sections[0])
        } else {
            // Just take the first paragraph
            paragraphs := strings.Split(text, "\n\n")
            if len(paragraphs) > 0 {
                architecture.Overview = strings.TrimSpace(paragraphs[0])
            }
        }
    }
    
    // Look for component descriptions
    for _, section := range sections {
        if strings.Contains(strings.ToLower(section), "component") || 
           strings.Contains(strings.ToLower(section), "module") {
            
            lines := strings.Split(section, "\n")
            if len(lines) > 0 {
                title := strings.TrimSpace(lines[0])
                
                // Extract component name from title
                var componentName string
                if strings.Contains(title, ":") {
                    parts := strings.SplitN(title, ":", 2)
                    componentName = strings.TrimSpace(parts[0])
                } else {
                    componentName = title
                }
                
                // Extract description
                var descBuilder strings.Builder
                for i := 1; i < len(lines); i++ {
                    descBuilder.WriteString(lines[i])
                    descBuilder.WriteString("\n")
                }
                
                architecture.ComponentDescriptions[componentName] = strings.TrimSpace(descBuilder.String())
                
                // Create a node for this component
                architecture.DiagramData.Nodes = append(architecture.DiagramData.Nodes, models.DiagramNode{
                    ID:       componentName,
                    Label:    componentName,
                    Type:     "component",
                    Size:     10,
                    Category: "component",
                })
            }
        }
    }
    
    // Try to extract relationships
    for _, section := range sections {
        if strings.Contains(strings.ToLower(section), "relationship") || 
           strings.Contains(strings.ToLower(section), "dependenc") {
            
            lines := strings.Split(section, "\n")
            for _, line := range lines {
                line = strings.TrimSpace(line)
                
                // Look for "A -> B" or "A depends on B" patterns
                if strings.Contains(line, "->") {
                    parts := strings.Split(line, "->")
                    if len(parts) >= 2 {
                        source := strings.TrimSpace(parts[0])
                        target := strings.TrimSpace(parts[1])
                        
                        architecture.DiagramData.Edges = append(architecture.DiagramData.Edges, models.DiagramEdge{
                            Source: source,
                            Target: target,
                            Type:   "depends",
                            Weight: 1,
                        })
                    }
                } else if strings.Contains(strings.ToLower(line), "depend") {
                    // Try to extract components from text
                    for src, _ := range architecture.ComponentDescriptions {
                        if strings.Contains(line, src) {
                            for tgt, _ := range architecture.ComponentDescriptions {
                                if src != tgt && strings.Contains(line, tgt) {
                                    architecture.DiagramData.Edges = append(architecture.DiagramData.Edges, models.DiagramEdge{
                                        Source: src,
                                        Target: tgt,
                                        Type:   "depends",
                                        Weight: 1,
                                    })
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    
    return architecture
}

// structureQAResponse structures raw text into a CodebaseQAResponse
func (s *CodeNavigationService) structureQAResponse(text string, question string, relevantCode map[string]string) models.CodebaseQAResponse {
    qa := models.CodebaseQAResponse{
        RelevantFiles:      []models.RelevantFile{},
        FollowupQuestions:  []string{},
    }
    
    // Use the whole text as the answer
    qa.Answer = text
    
    // Extract relevant files from the answer text
    lines := strings.Split(text, "\n")
    for _, line := range lines {
        for filePath := range relevantCode {
            if strings.Contains(line, filePath) {
                // Extract the snippet around the reference
                snippet := s.extractSnippetAroundReference(relevantCode[filePath], filePath)
                
                qa.RelevantFiles = append(qa.RelevantFiles, models.RelevantFile{
                    Path:     filePath,
                    Snippet:  snippet,
                    Relevance: 1.0,
                })
                
                break
            }
        }
    }
    
    // Try to extract follow-up questions
    for i, line := range lines {
        lowercase := strings.ToLower(line)
        if strings.Contains(lowercase, "follow") && strings.Contains(lowercase, "question") {
            for j := i + 1; j < len(lines) && j < i + 5; j++ {
                followupLine := strings.TrimSpace(lines[j])
                if followupLine == "" {
                    continue
                }
                
                // Look for question patterns
                if strings.HasPrefix(followupLine, "-") || 
                   strings.HasPrefix(followupLine, "*") || 
                   strings.HasPrefix(followupLine, "1.") ||
                   strings.HasPrefix(followupLine, "2.") ||
                   strings.HasPrefix(followupLine, "3.") ||
                   strings.Contains(followupLine, "?") {
                    
                    // Remove bullet points
                    followupLine = strings.TrimPrefix(followupLine, "-")
                    followupLine = strings.TrimPrefix(followupLine, "*")
                    followupLine = strings.TrimPrefix(followupLine, "1.")
                    followupLine = strings.TrimPrefix(followupLine, "2.")
                    followupLine = strings.TrimPrefix(followupLine, "3.")
                    
                    qa.FollowupQuestions = append(qa.FollowupQuestions, strings.TrimSpace(followupLine))
                }
            }
            break
        }
    }
    
    return qa
}

// extractSnippetAroundReference extracts a snippet of code around a reference
func (s *CodeNavigationService) extractSnippetAroundReference(content, reference string) string {
    lines := strings.Split(content, "\n")
    var startLine, endLine int
    
    // Look for the reference in the content
    for i, line := range lines {
        if strings.Contains(line, reference) {
            startLine = i - 5
            if startLine < 0 {
                startLine = 0
            }
            
            endLine = i + 5
            if endLine >= len(lines) {
                endLine = len(lines) - 1
            }
            
            break
        }
    }
    
    // If reference not found, just take the first few lines
    if startLine == 0 && endLine == 0 {
        endLine = 10
        if endLine >= len(lines) {
            endLine = len(lines) - 1
        }
    }
    
    // Extract the snippet
    var snippetBuilder strings.Builder
    for i := startLine; i <= endLine; i++ {
        snippetBuilder.WriteString(lines[i])
        snippetBuilder.WriteString("\n")
    }
    
    return strings.TrimSpace(snippetBuilder.String())
}

// structureBestPracticesGuide structures raw text into a BestPracticesResponse
func (s *CodeNavigationService) structureBestPracticesGuide(text string) models.BestPracticesResponse {
    bestPractices := models.BestPracticesResponse{
        Practices: []models.BestPractice{},
        Issues:    []models.CodeIssue{},
    }
    
    // Extract style guide and practices
    sections := strings.Split(text, "##")
    
    var styleGuideBuilder strings.Builder
    for _, section := range sections {
        section = strings.TrimSpace(section)
        if section == "" {
            continue
        }
        
        lines := strings.Split(section, "\n")
        if len(lines) == 0 {
            continue
        }
        
        sectionTitle := strings.ToLower(lines[0])
        
        if strings.Contains(sectionTitle, "style guide") {
            // This is the style guide section
            for i := 1; i < len(lines); i++ {
                styleGuideBuilder.WriteString(lines[i])
                styleGuideBuilder.WriteString("\n")
            }
        } else if strings.Contains(sectionTitle, "best practice") || strings.Contains(sectionTitle, "convention") {
            // This might be a practice
            if len(lines) > 1 {
                practice := models.BestPractice{
                    Title: lines[0],
                    Examples: []string{},
                }
                
                var descBuilder strings.Builder
                // internal/services/codenavigation.go (continued)
                for i := 1; i < len(lines); i++ {
                    line := lines[i]
                    
                    // Look for examples
                    if strings.Contains(strings.ToLower(line), "example") {
                        // Add description so far
                        practice.Description = strings.TrimSpace(descBuilder.String())
                        
                        // Start collecting examples
                        var exampleBuilder strings.Builder
                        for j := i + 1; j < len(lines); j++ {
                            exampleLine := lines[j]
                            exampleBuilder.WriteString(exampleLine)
                            exampleBuilder.WriteString("\n")
                        }
                        
                        // Add the example
                        if exampleBuilder.Len() > 0 {
                            practice.Examples = append(practice.Examples, strings.TrimSpace(exampleBuilder.String()))
                        }
                        
                        break
                    } else {
                        descBuilder.WriteString(line)
                        descBuilder.WriteString("\n")
                    }
                }
                
                // If no examples section was found, use the description as is
                if practice.Description == "" {
                    practice.Description = strings.TrimSpace(descBuilder.String())
                }
                
                bestPractices.Practices = append(bestPractices.Practices, practice)
            }
        } else if strings.Contains(sectionTitle, "issue") || strings.Contains(sectionTitle, "concern") {
            // This might be listing issues
            for i := 1; i < len(lines); i++ {
                line := strings.TrimSpace(lines[i])
                if line == "" {
                    continue
                }
                
                // Look for issue patterns
                if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || 
                   strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") || strings.HasPrefix(line, "3.") {
                    
                    // Remove bullet points
                    line = strings.TrimPrefix(line, "-")
                    line = strings.TrimPrefix(line, "*")
                    line = strings.TrimPrefix(line, "1.")
                    line = strings.TrimPrefix(line, "2.")
                    line = strings.TrimPrefix(line, "3.")
                    line = strings.TrimSpace(line)
                    
                    issue := models.CodeIssue{
                        Description: line,
                        Severity:    "medium", // Default severity
                    }
                    
                    // Try to extract path and line number if mentioned
                    if strings.Contains(line, ":line") {
                        parts := strings.Split(line, ":line")
                        if len(parts) >= 2 {
                            issue.Path = strings.TrimSpace(parts[0])
                            
                            // Try to parse line number
                            lineNumStr := strings.Split(parts[1], " ")[0]
                            lineNum, err := strconv.Atoi(lineNumStr)
                            if err == nil {
                                issue.Line = lineNum
                            }
                            
                            // Update description to remove file reference
                            issue.Description = strings.TrimSpace(strings.Join(parts[1:], ""))
                        }
                    }
                    
                    // Try to determine severity
                    lowercase := strings.ToLower(line)
                    if strings.Contains(lowercase, "critical") || strings.Contains(lowercase, "severe") {
                        issue.Severity = "high"
                    } else if strings.Contains(lowercase, "minor") || strings.Contains(lowercase, "trivial") {
                        issue.Severity = "low"
                    }
                    
                    bestPractices.Issues = append(bestPractices.Issues, issue)
                }
            }
        }
    }
    
    bestPractices.StyleGuide = strings.TrimSpace(styleGuideBuilder.String())
    
    return bestPractices
}