// internal/models/codenavigation.go
package models

// CodeWalkthroughRequest contains the request data for code walkthrough generation
type CodeWalkthroughRequest struct {
	RepositoryRequest
	Depth       int      `json:"depth"`
	FocusPath   string   `json:"focus_path"`
	EntryPoints []string `json:"entry_points"`
}

// FunctionExplainerRequest contains the request data for function explanation
type FunctionExplainerRequest struct {
	RepositoryRequest
	FilePath     string `json:"file_path" binding:"required"`
	FunctionName string `json:"function_name" binding:"required"`
	LineStart    int    `json:"line_start"`
	LineEnd      int    `json:"line_end"`
}

// ArchitectureVisualizerRequest contains the request data for architecture visualization
type ArchitectureVisualizerRequest struct {
	RepositoryRequest
	Detail     string   `json:"detail"` // "high", "medium", "low"
	FocusPaths []string `json:"focus_paths"`
}

// CodebaseQARequest contains the request data for codebase Q&A
type CodebaseQARequest struct {
	RepositoryRequest
	Question string `json:"question" binding:"required"`
}

// BestPracticesRequest contains the request data for best practices guide
type BestPracticesRequest struct {
	RepositoryRequest
	Scope string `json:"scope"` // "full", "directory", "file"
	Path  string `json:"path"`
}

// CodeWalkthroughResponse represents the response for code walkthrough generation
type CodeWalkthroughResponse struct {
	Overview     string                   `json:"overview"`
	EntryPoints  []string                 `json:"entry_points"`
	Walkthrough  []CodeWalkthroughStep    `json:"walkthrough"`
	Dependencies map[string][]string      `json:"dependencies"`
}

// CodeWalkthroughStep represents a single step in a code walkthrough
type CodeWalkthroughStep struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Importance  int    `json:"importance"` // 1-10 scale
}

// FunctionExplainerResponse represents the response for function explanation
type FunctionExplainerResponse struct {
	FunctionName   string   `json:"function_name"`
	Description    string   `json:"description"`
	Parameters     []Param  `json:"parameters"`
	ReturnValues   []Param  `json:"return_values"`
	Usage          []string `json:"usage_examples"`
	Complexity     string   `json:"complexity"`
	RelatedFunctions []string `json:"related_functions"`
}

// Param represents a parameter or return value
type Param struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ArchitectureVisualizerResponse represents the response for architecture visualization
type ArchitectureVisualizerResponse struct {
	Overview           string            `json:"overview"`
	DiagramData        DiagramData       `json:"diagram_data"`
	ComponentDescriptions map[string]string `json:"component_descriptions"`
}

// DiagramData contains the data required to render an architecture diagram
type DiagramData struct {
	Nodes []DiagramNode `json:"nodes"`
	Edges []DiagramEdge `json:"edges"`
}

// DiagramNode represents a node in an architecture diagram
type DiagramNode struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"` // "file", "directory", "component"
	Size     int    `json:"size"`
	Category string `json:"category"` // For grouping/coloring
}

// DiagramEdge represents an edge in an architecture diagram
type DiagramEdge struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Type     string `json:"type"` // "imports", "calls", "extends"
	Weight   int    `json:"weight"`
}

// CodebaseQAResponse represents the response for codebase Q&A
type CodebaseQAResponse struct {
	Answer          string        `json:"answer"`
	RelevantFiles   []RelevantFile `json:"relevant_files"`
	FollowupQuestions []string    `json:"followup_questions"`
}

// RelevantFile represents a file relevant to a Q&A response
type RelevantFile struct {
	Path     string `json:"path"`
	Snippet  string `json:"snippet"`
	Relevance float64 `json:"relevance"`
}

// BestPracticesResponse represents the response for best practices guide
type BestPracticesResponse struct {
	Practices []BestPractice `json:"practices"`
	StyleGuide string        `json:"style_guide"`
	Issues     []CodeIssue   `json:"issues"`
}

// BestPractice represents a single best practice
type BestPractice struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

// CodeIssue represents a potential code issue
type CodeIssue struct {
	Path        string `json:"path"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // "low", "medium", "high"
	Suggestion  string `json:"suggestion"`
}