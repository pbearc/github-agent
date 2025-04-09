// internal/github/utils.go (new file)
package github

import (
	"path/filepath"
)

// IsSourceFile checks if a file is a source code file
func IsSourceFile(path string) bool {
    ext := filepath.Ext(path)
    sourceExts := map[string]bool{
        ".go":    true,
        ".java":  true,
        ".js":    true,
        ".jsx":   true,
        ".ts":    true,
        ".tsx":   true,
        ".py":    true,
        ".rb":    true,
        ".php":   true,
        ".c":     true,
        ".cpp":   true,
        ".h":     true,
        ".cs":    true,
        ".swift": true,
        ".kt":    true,
    }
    return sourceExts[ext]
}

// GetLanguageFromPath determines the programming language based on file extension
func GetLanguageFromPath(path string) string {
    ext := filepath.Ext(path)
    switch ext {
    case ".go":
        return "Go"
    case ".java":
        return "Java"
    case ".js", ".jsx":
        return "JavaScript"
    case ".ts", ".tsx":
        return "TypeScript"
    case ".py":
        return "Python"
    case ".rb":
        return "Ruby"
    case ".php":
        return "PHP"
    case ".c", ".h":
        return "C"
    case ".cpp":
        return "C++"
    case ".cs":
        return "C#"
    case ".swift":
        return "Swift"
    case ".kt":
        return "Kotlin"
    default:
        return "Unknown"
    }
}