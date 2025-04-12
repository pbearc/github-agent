// internal/models/codenavigation.go
package models

// RepositoryInfo holds combined information about a GitHub repository.
type RepositoryInfo struct {
	Owner         string         `json:"owner"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	DefaultBranch string         `json:"default_branch"`
	URL           string         `json:"url"`           // HTML URL
	Language      string         `json:"language"`      // Primary language
	Languages     map[string]int `json:"languages"`     // Map of language -> bytes
	Stars         int            `json:"stars"`         // Stargazers count
	Forks         int            `json:"forks"`         // Forks count
	HasReadme     bool           `json:"has_readme"`    // ADDED: Indicates if a README file likely exists
}

// FileContent represents the content and metadata of a single file.
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"` // Decoded content
	SHA     string `json:"sha"`
}

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

// Update the DiagramNode and DiagramEdge types to support more detailed visualization
type DiagramNode struct {
	ID          string            `json:"id"`
	Label       string            `json:"label"`
	Type        string            `json:"type"`     // "file", "directory", "component", "service", "module"
	Size        int               `json:"size"`     // Relative importance (1-10)
	Category    string            `json:"category"` // For grouping/coloring
	Layer       string            `json:"layer"`    // "frontend", "backend", "database", etc.
	Technology  string            `json:"technology"` // Technology used
	Metadata    map[string]string `json:"metadata"`  // Additional information
}

// DiagramEdge represents an edge in an architecture diagram
type DiagramEdge struct {
	Source      string            `json:"source"`
	Target      string            `json:"target"`
	Type        string            `json:"type"`     // "imports", "calls", "extends", "uses", "depends", etc.
	Weight      int               `json:"weight"`   // Strength of relationship
	Label       string            `json:"label"`    // Description of relationship
	Bidirectional bool            `json:"bidirectional"` // True if relationship goes both ways
	Metadata    map[string]string `json:"metadata"`  // Additional information
}

// CodebaseQAResponse represents the response for codebase Q&A
type CodebaseQAResponse struct {
	Answer          string        `json:"answer"`
	RelevantFiles   []RelevantFile `json:"relevant_files"`
	FollowupQuestions []string    `json:"followup_questions"`
}

// RelevantFile represents a file relevant to a Q&A response
type RelevantFile struct {
    Path      string `json:"path"`
    Snippet   string `json:"snippet"`
    Relevance float64    `json:"relevance"`
    StartLine int    `json:"start_line,omitempty"`
    EndLine   int    `json:"end_line,omitempty"`
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

// ArchitectureVisualizerMermaidResponse represents the response for architecture visualization with Mermaid
type ArchitectureVisualizerMermaidResponse struct {
	Overview             string            `json:"overview"`
	MermaidDiagram       string            `json:"mermaid_diagram"`
	ComponentDescriptions map[string]string `json:"component_descriptions"`
}

