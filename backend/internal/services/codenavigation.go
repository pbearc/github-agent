// internal/services/codenavigation.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pbearc/github-agent/backend/internal/github"
	"github.com/pbearc/github-agent/backend/internal/graph"
	"github.com/pbearc/github-agent/backend/internal/llm"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/pkg/common"
	"github.com/sirupsen/logrus"
)

// CodeNavigationService handles code navigation features
type CodeNavigationService struct {
    githubClient *github.Client
    llmClient    *llm.GeminiClient
    neo4jClient  *graph.Neo4jClient
    logger       *common.Logger
}

// NewCodeNavigationService creates a new CodeNavigationService instance
func NewCodeNavigationService(githubClient *github.Client, llmClient *llm.GeminiClient, neo4jClient *graph.Neo4jClient) *CodeNavigationService {
    return &CodeNavigationService{
        githubClient: githubClient,
        llmClient:    llmClient,
        neo4jClient:  neo4jClient,
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

// Fix for StoreCodebaseInNeo4j method in services/codenavigation.go

// StoreCodebaseInNeo4j stores the codebase structure in Neo4j
func (s *CodeNavigationService) StoreCodebaseInNeo4j(ctx context.Context, owner, repo, branch string) error {
    s.logger.Info(fmt.Sprintf("Starting StoreCodebaseInNeo4j for %s/%s @ %s", owner, repo, branch))
    
    // If Neo4j client is not available, just return
    if s.neo4jClient == nil {
        return nil
    }
    
    // Get repository info
    repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
    if err != nil {
        return common.WrapError(err, "failed to get repository info")
    }

    // If branch is not specified, use the default branch
    if branch == "" {
        branch = repoInfo.DefaultBranch
        s.logger.Info(fmt.Sprintf("Using default branch: %s", branch))
    }
    
    // Step 1: Get all files from the repository
    files, err := s.githubClient.GetAllFiles(ctx, owner, repo, branch)
    if err != nil {
        return common.WrapError(err, "failed to get all files")
    }
    
    s.logger.Info(fmt.Sprintf("Found %d files/directories in repository", len(files)))
    
    // Step 2: Get import relationships between files
    importMap, err := s.githubClient.GetImportMap(ctx, owner, repo, branch)
    if err != nil {
        s.logger.WithError(err).Warning("Failed to get import map, continuing with file structure only")
        // Continue with empty import map rather than failing completely
        importMap = make(map[string][]string)
    }
    
    s.logger.Info(fmt.Sprintf("Found %d files with import relationships", len(importMap)))
    
    // Step 3: Store in Neo4j
    // Check the method signature in Neo4jClient
    err = s.neo4jClient.StoreCodebaseStructure(ctx, owner, repo, branch, files, importMap)
    if err != nil {
        return common.WrapError(err, "failed to store codebase structure in Neo4j")
    }
    
    s.logger.Info(fmt.Sprintf("Successfully stored codebase structure for %s/%s@%s in Neo4j", owner, repo, branch))
    return nil
}

// VisualizeArchitecture generates an architecture visualization
func (s *CodeNavigationService) VisualizeArchitecture(
    ctx context.Context,
    owner, repo, branch string,
    detail string,
    focusPaths []string,
) (*models.ArchitectureVisualizerResponse, error) {
    // Get codebase graph from Neo4j
    graphData, err := s.neo4jClient.GetCodebaseGraph(ctx, owner, repo, branch)
    if err != nil {
        // If Neo4j fails, try to generate a basic visualization without it
        s.logger.WithError(err).Warning("Failed to get codebase graph from Neo4j, attempting fallback visualization")
        return s.generateFallbackVisualization(ctx, owner, repo, branch, detail, focusPaths)
    }
    
    // Convert graph data to diagram data
    diagramData := s.convertToDiagramData(graphData, detail, focusPaths)
    
    // Generate component descriptions
    componentDescriptions := make(map[string]string)
    
    // For components that appear to be important (based on connections), generate descriptions
    importantNodes := s.findImportantNodes(diagramData)
    
    for _, node := range importantNodes {
        if node.Type == "file" {
            // Get file content
            fileContent, err := s.githubClient.GetFileContentText(ctx, owner, repo, node.ID, branch)
            if err != nil {
                s.logger.WithError(err).Warning(fmt.Sprintf("Failed to get content for %s", node.ID))
                continue
            }
            
            // Generate description using LLM
            description, err := s.generateComponentDescription(ctx, node.ID, fileContent.Content)
            if err != nil {
                s.logger.WithError(err).Warning(fmt.Sprintf("Failed to generate description for %s", node.ID))
                continue
            }
            
            componentDescriptions[node.ID] = description
        }
    }
    
    // Generate overview
    overview, err := s.generateArchitectureOverview(ctx, owner, repo, branch, diagramData)
    if err != nil {
        s.logger.WithError(err).Warning("Failed to generate architecture overview")
        overview = "This is a visualization of the codebase architecture showing key components and their relationships."
    }
    
    return &models.ArchitectureVisualizerResponse{
        Overview:               overview,
        DiagramData:            diagramData,
        ComponentDescriptions:  componentDescriptions,
    }, nil
}

// Helper function to convert graph data to diagram data
func (s *CodeNavigationService) convertToDiagramData(graphData map[string]interface{}, detail string, focusPaths []string) models.DiagramData {
    nodes := []models.DiagramNode{}
    edges := []models.DiagramEdge{}
    
    // Create a map of node IDs to indices
    nodeMap := make(map[string]int)
    
    // Process nodes
    graphNodes, ok := graphData["nodes"].([]map[string]interface{})
    if ok {
        for i, node := range graphNodes {
            // Skip if not a focus path when focus paths are specified
            if len(focusPaths) > 0 {
                matched := false
                for _, focusPath := range focusPaths {
                    if strings.HasPrefix(node["path"].(string), focusPath) {
                        matched = true
                        break
                    }
                }
                if !matched {
                    continue
                }
            }
            
            // Determine node type and category
            nodeType := node["type"].(string)
            category := "other"
            layer := "unknown"
            technology := "unknown"
            
            path := node["path"].(string)
            ext := filepath.Ext(path)
            
            // Set technology based on file extension
            switch ext {
            case ".go":
                technology = "Go"
                category = "backend"
                layer = "backend"
            case ".js", ".jsx", ".ts", ".tsx":
                technology = "JavaScript/TypeScript"
                category = "frontend"
                layer = "frontend"
            case ".css", ".scss", ".sass", ".less":
                technology = "CSS"
                category = "frontend"
                layer = "frontend"
            case ".html":
                technology = "HTML"
                category = "frontend"
                layer = "frontend"
            case ".sql":
                technology = "SQL"
                category = "database"
                layer = "database"
            case ".md":
                technology = "Markdown"
                category = "documentation"
                layer = "documentation"
            }
            
            // Special cases
            if nodeType == "directory" {
                category = "module"
                layer = "module"
                
                // Try to infer directory purpose
                dirName := filepath.Base(path)
                switch strings.ToLower(dirName) {
                case "model", "models":
                    category = "data"
                    layer = "data"
                case "controller", "controllers", "handlers":
                    category = "controller"
                    layer = "controller"
                case "view", "views", "templates":
                    category = "view"
                    layer = "frontend"
                case "config", "conf":
                    category = "configuration"
                    layer = "configuration"
                case "middleware", "middlewares":
                    category = "middleware"
                    layer = "middleware"
                case "service", "services":
                    category = "service"
                    layer = "service"
                case "util", "utils", "helper", "helpers":
                    category = "utility"
                    layer = "utility"
                case "test", "tests":
                    category = "test"
                    layer = "test"
                }
            }
            
            // Set size based on importance
            size := 5
            if nodeType == "directory" {
                size = 7
            }
            
            // Add the node
            diagramNode := models.DiagramNode{
                ID:        path,
                Label:     node["label"].(string),
                Type:      nodeType,
                Size:      size,
                Category:  category,
                Layer:     layer,
                Technology: technology,
                Metadata: make(map[string]string),
            }
            
            nodes = append(nodes, diagramNode)
            nodeMap[path] = i
        }
    }
    
    // Process relationships
    graphRelationships, ok := graphData["relationships"].([]map[string]interface{})
    if ok {
        for _, rel := range graphRelationships {
            source := rel["source"].(string)
            targets, ok := rel["targets"].([]map[string]interface{})
            
            if !ok || len(targets) == 0 {
                continue
            }
            
            // Skip source nodes that weren't included due to focus path filtering
            if _, exists := nodeMap[source]; !exists {
                continue
            }
            
            for _, target := range targets {
                // Skip null targets
                targetPath, ok := target["target"].(string)
                if !ok || targetPath == "" {
                    continue
                }
                
                // Skip target nodes that weren't included due to focus path filtering
                if _, exists := nodeMap[targetPath]; !exists {
                    continue
                }
                
                relType, ok := target["type"].(string)
                if !ok || relType == "" {
                    relType = "related_to"
                }
                
                // Add the edge
                edge := models.DiagramEdge{
                    Source:     source,
                    Target:     targetPath,
                    Type:       relType,
                    Weight:     1,
                    Label:      relType,
                    Bidirectional: false,
                    Metadata:   make(map[string]string),
                }
                
                edges = append(edges, edge)
            }
        }
    }
    
    // Filter based on detail level
    if detail != "high" {
        // For medium or low detail, limit the number of nodes and edges
        maxNodes := 100
        if detail == "low" {
            maxNodes = 30
        }
        
        // If we have too many nodes, keep only the most important ones
        if len(nodes) > maxNodes {
            // Sort nodes by importance (size, connections, etc.)
            // This is a simplified version - you may want to use a more sophisticated algorithm
            nodeImportance := make(map[string]int)
            
            // Count edges for each node
            for _, edge := range edges {
                nodeImportance[edge.Source]++
                nodeImportance[edge.Target]++
            }
            
            // Adjust importance based on node type
            for i, node := range nodes {
                baseImportance := nodeImportance[node.ID]
                
                // Directories and important files get a boost
                if node.Type == "directory" {
                    baseImportance += 10
                }
                
                // Specific important files get a boost
                if strings.Contains(node.ID, "main.go") || 
                   strings.Contains(node.ID, "index.js") ||
                   strings.Contains(node.ID, "app.js") ||
                   strings.Contains(node.ID, "server.js") {
                    baseImportance += 5
                }
                
                nodes[i].Size = baseImportance + 1 // Ensure at least size 1
            }
            
            // Sort nodes by importance
            sort.Slice(nodes, func(i, j int) bool {
                return nodes[i].Size > nodes[j].Size
            })
            
            // Keep only the most important nodes
            keptNodeIDs := make(map[string]bool)
            if len(nodes) > maxNodes {
                nodes = nodes[:maxNodes]
            }
            
            // Update the kept node IDs map
            for _, node := range nodes {
                keptNodeIDs[node.ID] = true
            }
            
            // Filter edges to only include connections between kept nodes
            filteredEdges := []models.DiagramEdge{}
            for _, edge := range edges {
                if keptNodeIDs[edge.Source] && keptNodeIDs[edge.Target] {
                    filteredEdges = append(filteredEdges, edge)
                }
            }
            edges = filteredEdges
        }
    }
    
    return models.DiagramData{
        Nodes: nodes,
        Edges: edges,
    }
}

// findImportantNodes identifies important nodes based on connections
func (s *CodeNavigationService) findImportantNodes(diagramData models.DiagramData) []models.DiagramNode {
    // Count connections for each node
    connections := make(map[string]int)
    
    for _, edge := range diagramData.Edges {
        connections[edge.Source]++
        connections[edge.Target]++
    }
    
    // Find nodes with many connections
    var importantNodes []models.DiagramNode
    
    for _, node := range diagramData.Nodes {
        // If it has more than 2 connections or is a directory, consider it important
        if connections[node.ID] > 2 || node.Type == "directory" {
            importantNodes = append(importantNodes, node)
        }
    }
    
    return importantNodes
}

// generateComponentDescription generates a description for a component using LLM
func (s *CodeNavigationService) generateComponentDescription(ctx context.Context, path, content string) (string, error) {
    // Use a simple prompt for now
    prompt := fmt.Sprintf("Provide a brief (2-3 sentences max) description of this file's purpose in the codebase: %s\n\nFile content:\n%s", path, content)
    
    // Truncate content if too long
    if len(prompt) > 2000 {
        prompt = prompt[:2000] + "...[content truncated]"
    }
    
    // Generate description using LLM
    description, err := s.llmClient.GenerateCompletion(ctx, prompt, 0.7, 100)
    if err != nil {
        return "", err
    }
    
    return description, nil
}

// generateArchitectureOverview generates an overview of the architecture using LLM
func (s *CodeNavigationService) generateArchitectureOverview(ctx context.Context, owner, repo, branch string, diagramData models.DiagramData) (string, error) {
    // Create a summary of the architecture
    var sb strings.Builder
    
    sb.WriteString(fmt.Sprintf("Generate a brief overview (3-5 sentences) of the architecture for the repository %s/%s.\n\n", owner, repo))
    
    // Add information about key components
    sb.WriteString("Key components:\n")
    
    // Focus on directories first
    for _, node := range diagramData.Nodes {
        if node.Type == "directory" {
            sb.WriteString(fmt.Sprintf("- %s (%s)\n", node.Label, node.Layer))
        }
    }
    
    // Add important files
    sb.WriteString("\nKey files:\n")
    for _, node := range diagramData.Nodes {
        if node.Type == "file" && node.Size > 5 {
            sb.WriteString(fmt.Sprintf("- %s (%s)\n", node.ID, node.Technology))
        }
    }
    
    // Add information about relationships
    sb.WriteString("\nRelationships:\n")
    sb.WriteString(fmt.Sprintf("- Total files: %d\n", countNodesByType(diagramData.Nodes, "file")))
    sb.WriteString(fmt.Sprintf("- Total directories: %d\n", countNodesByType(diagramData.Nodes, "directory")))
    sb.WriteString(fmt.Sprintf("- Total connections: %d\n", len(diagramData.Edges)))
    
    // Generate description using LLM
    prompt := sb.String()
    
    // Truncate if too long
    if len(prompt) > 3000 {
        prompt = prompt[:3000] + "...[truncated]"
    }
    
    overview, err := s.llmClient.GenerateCompletion(ctx, prompt, 0.7, 200)
    if err != nil {
        return "", err
    }
    
    return overview, nil
}

// Helper function to count nodes by type
func countNodesByType(nodes []models.DiagramNode, nodeType string) int {
    count := 0
    for _, node := range nodes {
        if node.Type == nodeType {
            count++
        }
    }
    return count
}

// generateFallbackVisualization creates a basic visualization when Neo4j is unavailable
func (s *CodeNavigationService) generateFallbackVisualization(
    ctx context.Context, 
    owner, repo, branch string,
    detail string,
    focusPaths []string,
) (*models.ArchitectureVisualizerResponse, error) {
    // Get all files
    files, err := s.githubClient.GetAllFiles(ctx, owner, repo, branch)
    if err != nil {
        return nil, err
    }
    
    // Create nodes for each file and directory
    nodes := []models.DiagramNode{}
    
    // Track unique directories
    directories := make(map[string]bool)
    
    for _, file := range files {
        // Skip files that don't match focus paths if specified
        if len(focusPaths) > 0 {
            matched := false
            for _, focusPath := range focusPaths {
                if strings.HasPrefix(file.Path, focusPath) {
                    matched = true
                    break
                }
            }
            if !matched {
                continue
            }
        }
        
        // Add file node
        if file.Type == "file" {
            ext := filepath.Ext(file.Path)
            
            category := "other"
            layer := "unknown"
            technology := "unknown"
            
            // Set technology based on file extension
            switch ext {
            case ".go":
                technology = "Go"
                category = "backend"
                layer = "backend"
            case ".js", ".jsx", ".ts", ".tsx":
                technology = "JavaScript/TypeScript"
                category = "frontend"
                layer = "frontend"
            case ".css", ".scss", ".sass", ".less":
                technology = "CSS"
                category = "frontend"
                layer = "frontend"
            case ".html":
                technology = "HTML"
                category = "frontend"
                layer = "frontend"
            case ".sql":
                technology = "SQL"
                category = "database"
                layer = "database"
            case ".md":
                technology = "Markdown"
                category = "documentation"
                layer = "documentation"
            }
            
            nodes = append(nodes, models.DiagramNode{
                ID:         file.Path,
                Label:      file.Name,
                Type:       "file",
                Size:       3,
                Category:   category,
                Layer:      layer,
                Technology: technology,
                Metadata:   make(map[string]string),
            })
        }
        
        // Track directories
        dirPath := filepath.Dir(file.Path)
        if dirPath != "." && dirPath != "" {
            directories[dirPath] = true
        }
    }
    
    // Add directory nodes
    for dirPath := range directories {
        dirName := filepath.Base(dirPath)
        
        category := "module"
        layer := "module"
        
        // Try to infer directory purpose
        switch strings.ToLower(dirName) {
        case "model", "models":
            category = "data"
            layer = "data"
        case "controller", "controllers", "handlers":
            category = "controller"
            layer = "controller"
        case "view", "views", "templates":
            category = "view"
            layer = "frontend"
        case "config", "conf":
            category = "configuration"
            layer = "configuration"
        case "middleware", "middlewares":
            category = "middleware"
            layer = "middleware"
        case "service", "services":
            category = "service"
            layer = "service"
        case "util", "utils", "helper", "helpers":
            category = "utility"
            layer = "utility"
        case "test", "tests":
            category = "test"
            layer = "test"
        }
        
        nodes = append(nodes, models.DiagramNode{
            ID:         dirPath,
            Label:      dirName,
            Type:       "directory",
            Size:       6,
            Category:   category,
            Layer:      layer,
            Technology: "N/A",
            Metadata:   make(map[string]string),
        })
    }
    
    // Create edges based on directory structure
    edges := []models.DiagramEdge{}
    
    // Connect files to their parent directories
    for _, node := range nodes {
        if node.Type == "file" {
            dirPath := filepath.Dir(node.ID)
            if dirPath != "." && dirPath != "" {
                edges = append(edges, models.DiagramEdge{
                    Source:        dirPath,
                    Target:        node.ID,
                    Type:          "contains",
                    Weight:        1,
                    Label:         "contains",
                    Bidirectional: false,
                    Metadata:      make(map[string]string),
                })
            }
        }
    }
    
    // Connect nested directories
    for dir1 := range directories {
        for dir2 := range directories {
            if dir1 != dir2 && filepath.Dir(dir2) == dir1 {
                edges = append(edges, models.DiagramEdge{
                    Source:        dir1,
                    Target:        dir2,
                    Type:          "contains",
                    Weight:        2,
                    Label:         "contains",
                    Bidirectional: false,
                    Metadata:      make(map[string]string),
                })
            }
        }
    }
    
    // Try to infer some additional relationships based on naming patterns
    for i, node1 := range nodes {
        if node1.Type != "file" {
            continue
        }
        
        baseName1 := strings.TrimSuffix(filepath.Base(node1.ID), filepath.Ext(node1.ID))
        
        for j, node2 := range nodes {
            if i == j || node2.Type != "file" {
                continue
            }
            
            baseName2 := strings.TrimSuffix(filepath.Base(node2.ID), filepath.Ext(node2.ID))
            
            // Connect test files to their implementation
            if strings.HasSuffix(baseName1, "_test") && strings.TrimSuffix(baseName1, "_test") == baseName2 {
                edges = append(edges, models.DiagramEdge{
                    Source:        node1.ID,
                    Target:        node2.ID,
                    Type:          "tests",
                    Weight:        3,
                    Label:         "tests",
                    Bidirectional: false,
                    Metadata:      make(map[string]string),
                })
            }
            
            // Try to infer model-controller relationships
            if strings.Contains(node1.ID, "/model/") && strings.Contains(node2.ID, "/controller/") {
                // If model name appears in controller name
                if strings.Contains(strings.ToLower(baseName2), strings.ToLower(baseName1)) {
                    edges = append(edges, models.DiagramEdge{
                        Source:        node2.ID,
                        Target:        node1.ID,
                        Type:          "uses",
                        Weight:        2,
                        Label:         "uses",
                        Bidirectional: false,
                        Metadata:      make(map[string]string),
                    })
                }
            }
        }
    }
    
    // Generate description for the fallback visualization
    overview := fmt.Sprintf("This is a basic visualization of the file structure for %s/%s. It shows the main directories and files, with inferred relationships based on naming patterns and directory structure.", owner, repo)
    
    return &models.ArchitectureVisualizerResponse{
        Overview:               overview,
        DiagramData:            models.DiagramData{Nodes: nodes, Edges: edges},
        ComponentDescriptions:  make(map[string]string),
    }, nil
}

// AnswerCodebaseQuestion answers a question about the codebase
func (s *CodeNavigationService) AnswerCodebaseQuestion(ctx context.Context, owner, repo, branch, question string, keywords []string) (*models.CodebaseQAResponse, error) {
    // Get repository info
    repoInfo, err := s.githubClient.GetRepositoryInfo(ctx, owner, repo)
    if err != nil {
        return nil, common.WrapError(err, "failed to get repository info")
    }

    // If branch is not specified, use the default branch
    if branch == "" {
        branch = repoInfo.DefaultBranch
    }

    // If no keywords provided, extract them
    if len(keywords) == 0 {
        var err error
        keywords, err = s.extractSearchKeywords(ctx, question, repoInfo.Language)
        if err != nil {
            s.logger.WithField("error", err).Warning("Failed to extract keywords, using question as keyword")
            keywords = []string{question}
        }
    }

    // Collect relevant code from all keywords
    relevantCode := make(map[string]string)
    
    // Search with each keyword and combine results
    for _, keyword := range keywords {
        searchResults, err := s.githubClient.SearchCode(ctx, owner, repo, keyword)
        if err != nil {
            s.logger.WithFields(logrus.Fields{
                "error": err,
                "keyword": keyword,
            }).Warning("Failed to search code with keyword")
            continue
        }

        for _, result := range searchResults {
            if result == nil || result.Path == nil {
                continue
            }
            
            // Skip if we already have this file
            if _, exists := relevantCode[*result.Path]; exists {
                continue
            }
            
            content, err := s.githubClient.GetFileContentText(ctx, owner, repo, *result.Path, branch)
            if err != nil {
                s.logger.WithField("error", err).Warning("Failed to get content for file: " + *result.Path)
                continue
            }
            
            relevantCode[*result.Path] = content.Content
        }
    }

    // If couldn't find enough relevant code, get some key files
    if len(relevantCode) < 2 {
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
        answer = s.structureQAResponse(answerJSON, question, relevantCode, keywords)
    }

    return &answer, nil
}

// extractSearchKeywords uses the LLM to extract relevant search keywords from a natural language question
func (s *CodeNavigationService) extractSearchKeywords(ctx context.Context, question, language string) ([]string, error) {
    prompt := fmt.Sprintf(`
You are an expert code search assistant. Your task is to extract specific technical keywords from a user's question that would be effective for searching in a code repository.

User Question: %s
Repository Language: %s

Instructions:
1. Identify technical terms, library names, function names, or concepts that would likely appear in code files
2. Focus on specific technical terms rather than general concepts
3. Return 2-5 of the most relevant search terms as a JSON array of strings
4. Each term should be short (1-3 words) and highly specific to technical implementations
5. If the question is about a specific library (like SQLAlchemy), include the library name as one of the terms

Response Format:
{"keywords": ["term1", "term2", "term3"]}
`, question, language)

    response, err := s.llmClient.GenerateText(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // Parse the JSON response
    var result struct {
        Keywords []string `json:"keywords"`
    }
    
    err = json.Unmarshal([]byte(response), &result)
    if err != nil {
        // If parsing fails, try to extract keywords from text response
        return s.fallbackKeywordExtraction(response), nil
    }
    
    return result.Keywords, nil
}

// fallbackKeywordExtraction attempts to extract keywords from a text response
// when JSON parsing fails
func (s *CodeNavigationService) fallbackKeywordExtraction(response string) []string {
    // Simple extraction by looking for keywords in quotes or after colons
    keywords := []string{}
    
    // Look for JSON-like patterns even in malformed responses
    re := regexp.MustCompile(`["']([^"']+)["']`)
    matches := re.FindAllStringSubmatch(response, -1)
    
    for _, match := range matches {
        if len(match) > 1 && len(match[1]) > 0 && len(match[1]) < 50 {
            keywords = append(keywords, match[1])
        }
    }
    
    // If we still don't have keywords, split by common separators
    if len(keywords) == 0 {
        splitWords := strings.FieldsFunc(response, func(r rune) bool {
            return r == ',' || r == ';' || r == '\n'
        })
        
        for _, word := range splitWords {
            word = strings.TrimSpace(word)
            if len(word) > 0 && len(word) < 50 {
                keywords = append(keywords, word)
            }
        }
    }
    
    // Deduplicate and limit results
    seen := make(map[string]bool)
    unique := []string{}
    
    for _, keyword := range keywords {
        keyword = strings.TrimSpace(strings.ToLower(keyword))
        if !seen[keyword] && keyword != "" {
            seen[keyword] = true
            unique = append(unique, keyword)
        }
    }
    
    // Limit to top 5 keywords
    if len(unique) > 5 {
        unique = unique[:5]
    }
    
    return unique
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

// calculateKeywordRelevanceScore computes relevance based on extracted keywords
// Returns a score from 1-100 representing the percentage of relevance
func (s *CodeNavigationService) calculateKeywordRelevanceScore(snippet string, keywords []string) int {
    // If there are no keywords, return minimum score
    if len(keywords) == 0 {
        return 1
    }
    
    // Convert snippet to lowercase for case-insensitive matching
    lowerSnippet := strings.ToLower(snippet)
    
    // Count how many keywords appear in the snippet
    matchCount := 0
    for _, keyword := range keywords {
        if strings.Contains(lowerSnippet, strings.ToLower(keyword)) {
            matchCount++
        }
    }
    
    // Calculate percentage of keywords matched (1-100 scale)
    // Minimum score is 1, maximum is 100
    if matchCount == 0 {
        return 1 // Minimum score
    }
    
    score := int((float64(matchCount) / float64(len(keywords))) * 100)
    
    // Ensure score is within bounds
    if score < 1 {
        score = 1
    } else if score > 100 {
        score = 100
    }
    
    return score
}

// structureQAResponse structures raw text into a CodebaseQAResponse
func (s *CodeNavigationService) structureQAResponse(text string, question string, relevantCode map[string]string, keywords []string) models.CodebaseQAResponse {
    qa := models.CodebaseQAResponse{
        RelevantFiles:      []models.RelevantFile{},
        FollowupQuestions:  []string{},
    }
    
    // Use the whole text as the answer
    qa.Answer = text
    
    // Track which files we've already processed to avoid duplicates
    processedFiles := make(map[string]bool)
    
    // Extract relevant files from the answer text
    lines := strings.Split(text, "\n")
    for _, line := range lines {
        for filePath := range relevantCode {
            if strings.Contains(line, filePath) && !processedFiles[filePath] {
                processedFiles[filePath] = true
                
                // Extract a more meaningful snippet around the reference
                snippet := s.extractImprovedSnippet(relevantCode[filePath], filePath, line)
                
                // Calculate relevance based on keywords (1-100 scale)
                relevanceScore := s.calculateKeywordRelevanceScore(snippet, keywords)
                
                // Extract line range from header comment if available
                startLine, endLine := 0, 0
                headerRegex := regexp.MustCompile(`// .+ \(lines (\d+)-(\d+) of \d+\)`)
                match := headerRegex.FindStringSubmatch(snippet)
                if len(match) >= 3 {
                    startLine, _ = strconv.Atoi(match[1])
                    endLine, _ = strconv.Atoi(match[2])
                }
                
                qa.RelevantFiles = append(qa.RelevantFiles, models.RelevantFile{
                    Path:      filePath,
                    Snippet:   snippet,
                    Relevance: float64(relevanceScore),
                    StartLine: startLine,
                    EndLine:   endLine,
                })
                
                break
            }
        }
    }
    
    // If we didn't find file references in the text, include all relevant files
    if len(qa.RelevantFiles) == 0 {
        for filePath, content := range relevantCode {
            snippet := s.extractImprovedSnippet(content, filePath, "")
            relevanceScore := s.calculateKeywordRelevanceScore(snippet, keywords)
            
            // Extract line range if available
            startLine, endLine := 0, 0
            headerRegex := regexp.MustCompile(`// .+ \(lines (\d+)-(\d+) of \d+\)`)
            match := headerRegex.FindStringSubmatch(snippet)
            if len(match) >= 3 {
                startLine, _ = strconv.Atoi(match[1])
                endLine, _ = strconv.Atoi(match[2])
            }
            
            qa.RelevantFiles = append(qa.RelevantFiles, models.RelevantFile{
                Path:      filePath,
                Snippet:   snippet,
                Relevance: float64(relevanceScore),
                StartLine: startLine,
                EndLine:   endLine,
            })
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

// extractImprovedSnippet extracts a more meaningful snippet from the file content
// with better parsing of standardized line number references
func (s *CodeNavigationService) extractImprovedSnippet(content string, filePath string, referenceLine string) string {
    lines := strings.Split(content, "\n")
    
    // If the file is small enough, return the entire file
    if len(lines) <= 50 {
        // Add a header comment but no line numbers in the actual code
        return fmt.Sprintf("// Complete file: %s (%d lines)\n\n%s", 
            filePath, len(lines), content)
    }

    // First, look for the standardized line references we asked for in the prompt
    // Format: "lines X-Y" or "line X" or "lines X, Y, Z"
    standardLineRegex := regexp.MustCompile(`lines?\s+(\d+)(?:\s*-\s*(\d+))?`)
    standardMatches := standardLineRegex.FindStringSubmatch(referenceLine)
    
    if len(standardMatches) > 1 {
        startLine, err := strconv.Atoi(standardMatches[1])
        if err != nil {
            startLine = 1 // Default to beginning if parsing fails
        }
        
        // Adjust to 0-based index
        startLine = startLine - 1
        if startLine < 0 {
            startLine = 0
        }
        
        endLine := startLine + 20 // Default to showing ~20 lines
        
        // If a range was specified
        if len(standardMatches) > 2 && standardMatches[2] != "" {
            parsedEndLine, err := strconv.Atoi(standardMatches[2])
            if err == nil {
                endLine = parsedEndLine - 1 // Convert to 0-based
            }
        }
        
        // Add context before and after
        contextBefore := 3
        contextAfter := 3
        
        // Adjust start and end to include context
        adjustedStart := max(0, startLine - contextBefore)
        adjustedEnd := min(len(lines) - 1, endLine + contextAfter)
        
        // Extract the relevant portion without line numbers in the code
        var snippetBuilder strings.Builder
        
        // Add a header comment with line range info
        snippetBuilder.WriteString(fmt.Sprintf("// %s (lines %d-%d of %d)\n\n", 
            filePath, adjustedStart+1, adjustedEnd+1, len(lines)))
        
        // Add the code without line numbers
        for i := adjustedStart; i <= adjustedEnd; i++ {
            snippetBuilder.WriteString(lines[i] + "\n")
        }
        
        return snippetBuilder.String()
    }
    
    // Also check for comma-separated line format: "lines X, Y, Z"
    commaLineRegex := regexp.MustCompile(`lines\s+(\d+)(?:\s*,\s*\d+)*(?:\s*(?:and|&)\s*\d+)?`)
    if commaLineRegex.MatchString(referenceLine) {
        // Extract all numbers from the reference
        numberRegex := regexp.MustCompile(`\b(\d+)\b`)
        allNumbers := numberRegex.FindAllString(referenceLine, -1)
        
        var lineNumbers []int
        for _, numStr := range allNumbers {
            num, err := strconv.Atoi(numStr)
            if err == nil && num > 0 && num <= len(lines) {
                lineNumbers = append(lineNumbers, num-1) // Convert to 0-based
            }
        }
        
        if len(lineNumbers) > 0 {
            sort.Ints(lineNumbers)
            
            // Find a good range to show based on the referenced lines
            startLine := lineNumbers[0]
            endLine := lineNumbers[len(lineNumbers)-1]
            
            // Ensure we show enough context
            if endLine - startLine < 10 {
                endLine = min(len(lines)-1, startLine + 10)
            }
            
            // Add a bit more context
            adjustedStart := max(0, startLine - 3)
            adjustedEnd := min(len(lines) - 1, endLine + 3)
            
            var snippetBuilder strings.Builder
            
            // Add a header comment with line range info
            snippetBuilder.WriteString(fmt.Sprintf("// %s (lines %d-%d of %d)\n\n", 
                filePath, adjustedStart+1, adjustedEnd+1, len(lines)))
            
            // Add the code without line numbers
            for i := adjustedStart; i <= adjustedEnd; i++ {
                snippetBuilder.WriteString(lines[i] + "\n")
            }
            
            return snippetBuilder.String()
        }
    }
    
    // If we still can't find line references, try other methods
    
    // Look for code snippets in backticks
    codeReferenceRegex := regexp.MustCompile("`([^`]+)`")
    codeReferences := codeReferenceRegex.FindAllStringSubmatch(referenceLine, -1)
    
    for _, match := range codeReferences {
        if len(match) > 1 {
            codeRef := match[1]
            
            // Search for the code reference in the file
            for i, line := range lines {
                if strings.Contains(line, codeRef) {
                    // Found the reference, extract surrounding context
                    startLine := max(0, i-5)
                    endLine := min(len(lines)-1, i+15)
                    
                    var snippetBuilder strings.Builder
                    snippetBuilder.WriteString(fmt.Sprintf("// %s (lines %d-%d of %d, around '%s')\n\n", 
                        filePath, startLine+1, endLine+1, len(lines), codeRef))
                    
                    // Add the code without line numbers
                    for j := startLine; j <= endLine; j++ {
                        snippetBuilder.WriteString(lines[j] + "\n")
                    }
                    
                    return snippetBuilder.String()
                }
            }
        }
    }
    
    // Look for class/function definitions mentioned in the text
    classOrMethodRegex := regexp.MustCompile(`\b(class|def|function)\s+(\w+)`)
    classMatches := classOrMethodRegex.FindAllStringSubmatch(referenceLine, -1)
    
    for _, match := range classMatches {
        if len(match) > 2 {
            defType := match[1]   // "class" or "def" or "function"
            defName := match[2]   // The name itself
            
            // Search for the definition in the file
            for i, line := range lines {
                if strings.Contains(line, defType+" "+defName) {
                    // Found the definition, extract the definition and implementation
                    startLine := max(0, i-2)
                    endLine := min(len(lines)-1, i+20)
                    
                    var snippetBuilder strings.Builder
                    snippetBuilder.WriteString(fmt.Sprintf("// %s (lines %d-%d of %d, definition of %s %s)\n\n", 
                        filePath, startLine+1, endLine+1, len(lines), defType, defName))
                    
                    // Add the code without line numbers
                    for j := startLine; j <= endLine; j++ {
                        snippetBuilder.WriteString(lines[j] + "\n")
                    }
                    
                    return snippetBuilder.String()
                }
            }
        }
    }
    
    // If the file path contains a significant part (e.g., "models.py"), 
    // look for key sections in the file
    parts := strings.Split(filePath, "/")
    if len(parts) > 0 {
        filename := parts[len(parts)-1]
        baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
        
        // If the base name is something meaningful (not just "env" or "init")
        if len(baseName) > 3 && !stringInSlice(baseName, []string{"env", "init", "config"}) {
            // Look for key definitions related to the file name
            for i, line := range lines {
                if (strings.Contains(line, "class "+baseName) || 
                    strings.Contains(line, "def "+baseName) || 
                    strings.Contains(line, "function "+baseName) ||
                    strings.Contains(line, "interface "+baseName) ||
                    strings.Contains(line, "struct "+baseName)) {
                    
                    // Found a definition related to the file name
                    startLine := max(0, i-2)
                    endLine := min(len(lines)-1, i+20)
                    
                    var snippetBuilder strings.Builder
                    snippetBuilder.WriteString(fmt.Sprintf("// %s (lines %d-%d of %d, key definition)\n\n", 
                        filePath, startLine+1, endLine+1, len(lines)))
                    
                    // Add the code without line numbers
                    for j := startLine; j <= endLine; j++ {
                        snippetBuilder.WriteString(lines[j] + "\n")
                    }
                    
                    return snippetBuilder.String()
                }
            }
        }
    }
    
    // Fallback: look for any key structural element
    for i, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "class ") || 
           strings.HasPrefix(line, "def ") || 
           strings.HasPrefix(line, "function ") || 
           strings.HasPrefix(line, "interface ") || 
           strings.HasPrefix(line, "struct ") {
            
            startLine := max(0, i-2)
            endLine := min(len(lines)-1, i+20)
            
            var snippetBuilder strings.Builder
            snippetBuilder.WriteString(fmt.Sprintf("// %s (lines %d-%d of %d, key section)\n\n", 
                filePath, startLine+1, endLine+1, len(lines)))
            
            // Add the code without line numbers
            for j := startLine; j <= endLine; j++ {
                snippetBuilder.WriteString(lines[j] + "\n")
            }
            
            return snippetBuilder.String()
        }
    }
    
    // Final fallback: return first section with file info
    var snippetBuilder strings.Builder
    endLine := min(30, len(lines) - 1)
    
    snippetBuilder.WriteString(fmt.Sprintf("// %s (first %d lines of %d total)\n\n", 
        filePath, endLine+1, len(lines)))
    
    // Add the code without line numbers
    for i := 0; i <= endLine; i++ {
        snippetBuilder.WriteString(lines[i] + "\n")
    }
    
    // If the file is much longer, add an indication that it's truncated
    if len(lines) > endLine + 1 {
        snippetBuilder.WriteString(fmt.Sprintf("\n// ... %d more lines not shown ...\n", len(lines) - endLine - 1))
    }
    
    return snippetBuilder.String()
}

// Helper function to check if a string is in a slice
func stringInSlice(str string, list []string) bool {
    for _, v := range list {
        if v == str {
            return true
        }
    }
    return false
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}



// Helper function for min of two integers (Go < 1.21 doesn't have this in std lib)
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
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
