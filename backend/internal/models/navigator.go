// internal/models/navigator.go (update)
package models

// CodebaseNavigatorRequest contains the request data for codebase Q&A
type CodebaseNavigatorRequest struct {
    RepositoryRequest
    Question string `json:"question" binding:"required"`
    TopK     int    `json:"top_k"`
}

// CodebaseNavigatorResponse represents the response for codebase Q&A
type CodebaseNavigatorResponse struct {
    Answer        string         `json:"answer"`
    RelevantFiles []RelevantFile `json:"relevant_files"` // Using the struct from codenavigation.go
}

// CodebaseIndexRequest contains the request data for indexing a codebase
type CodebaseIndexRequest struct {
    RepositoryRequest
    Force bool `json:"force"` // Force re-indexing even if already indexed
}

// CodebaseIndexResponse represents the response for codebase indexing
type CodebaseIndexResponse struct {
    Message    string `json:"message"`
    Status     string `json:"status"`
    Namespace  string `json:"namespace"`
    FileCount  int    `json:"file_count,omitempty"`
    ChunkCount int    `json:"chunk_count,omitempty"`
}