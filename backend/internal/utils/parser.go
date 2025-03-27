package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

// LanguageInfo contains information about a programming language
type LanguageInfo struct {
	Name             string
	Extensions       []string
	CommentPrefix    string
	MultiLineStart   string
	MultiLineEnd     string
	PackagePattern   *regexp.Regexp
	ImportPattern    *regexp.Regexp
	FunctionPattern  *regexp.Regexp
	ClassPattern     *regexp.Regexp
}

// Languages defines supported programming languages
var Languages = map[string]LanguageInfo{
	"go": {
		Name:             "Go",
		Extensions:       []string{".go"},
		CommentPrefix:    "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		PackagePattern:   regexp.MustCompile(`package\s+(\w+)`),
		ImportPattern:    regexp.MustCompile(`import\s+(?:"([^"]+)"|(\w+)\s+"([^"]+)")`),
		FunctionPattern:  regexp.MustCompile(`func\s+(\w+)`),
		ClassPattern:     nil, // Go doesn't have classes in the traditional sense
	},
	"python": {
		Name:             "Python",
		Extensions:       []string{".py"},
		CommentPrefix:    "#",
		MultiLineStart:   `"""`,
		MultiLineEnd:     `"""`,
		PackagePattern:   nil,
		ImportPattern:    regexp.MustCompile(`(?:from\s+(\w+(?:\.\w+)*)\s+)?import\s+(.+)`),
		FunctionPattern:  regexp.MustCompile(`def\s+(\w+)\s*\(`),
		ClassPattern:     regexp.MustCompile(`class\s+(\w+)`),
	},
	"javascript": {
		Name:             "JavaScript",
		Extensions:       []string{".js", ".jsx"},
		CommentPrefix:    "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		PackagePattern:   nil,
		ImportPattern:    regexp.MustCompile(`(?:import|require)\s+.*?(?:from\s+)?['"](.+?)['"]\)?`),
		FunctionPattern:  regexp.MustCompile(`(?:function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:function|\([^)]*\)\s*=>))`),
		ClassPattern:     regexp.MustCompile(`class\s+(\w+)`),
	},
	"typescript": {
		Name:             "TypeScript",
		Extensions:       []string{".ts", ".tsx"},
		CommentPrefix:    "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		PackagePattern:   nil,
		ImportPattern:    regexp.MustCompile(`import\s+.*?from\s+['"](.+?)['"]`),
		FunctionPattern:  regexp.MustCompile(`(?:function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:function|\([^)]*\)\s*=>)|(?:async\s+)?function\s*(\w+))`),
		ClassPattern:     regexp.MustCompile(`class\s+(\w+)`),
	},
	"java": {
		Name:             "Java",
		Extensions:       []string{".java"},
		CommentPrefix:    "//",
		MultiLineStart:   "/*",
		MultiLineEnd:     "*/",
		PackagePattern:   regexp.MustCompile(`package\s+([a-z0-9.]+)`),
		ImportPattern:    regexp.MustCompile(`import\s+([a-z0-9.]+)`),
		FunctionPattern:  regexp.MustCompile(`(?:public|private|protected|static|\s) +[\w<>\[\]]+\s+(\w+) *\([^\)]*\) *(?:\{?|[^;])`),
		ClassPattern:     regexp.MustCompile(`class\s+(\w+)`),
	},
}

// GetLanguageFromExtension returns the language information based on file extension
func GetLanguageFromExtension(filename string) *LanguageInfo {
	ext := filepath.Ext(filename)
	for _, lang := range Languages {
		for _, langExt := range lang.Extensions {
			if langExt == ext {
				// Create a copy of the language info to return its address
				langCopy := lang
				return &langCopy
			}
		}
	}
	return nil
}

// GetLanguageByName returns the language information based on name
func GetLanguageByName(name string) *LanguageInfo {
	name = strings.ToLower(name)
	if lang, ok := Languages[name]; ok {
		// Create a copy of the language info to return its address
		langCopy := lang
		return &langCopy
	}
	return nil
}

// CodeInfo contains extracted information from code
type CodeInfo struct {
	Language    string   `json:"language"`
	Package     string   `json:"package,omitempty"`
	Imports     []string `json:"imports,omitempty"`
	Functions   []string `json:"functions,omitempty"`
	Classes     []string `json:"classes,omitempty"`
	LineCount   int      `json:"line_count"`
	CommentLines int     `json:"comment_lines"`
	CodeLines   int      `json:"code_lines"`
}

// ParseCode extracts information from code
func ParseCode(code, language string) (*CodeInfo, error) {
	lang := GetLanguageByName(language)
	if lang == nil {
		// Try to guess the language
		lang = guessLanguage(code)
		if lang == nil {
			// Default to a simple parser
			return simpleCodeParse(code, language), nil
		}
	}

	// Initialize the code info
	info := &CodeInfo{
		Language:  lang.Name,
		Imports:   []string{},
		Functions: []string{},
		Classes:   []string{},
	}

	// Split the code into lines
	lines := strings.Split(code, "\n")
	info.LineCount = len(lines)

	// Count comment and code lines
	inMultiLineComment := false
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		if inMultiLineComment {
			info.CommentLines++
			if strings.Contains(trimmedLine, lang.MultiLineEnd) {
				inMultiLineComment = false
			}
		} else if strings.HasPrefix(trimmedLine, lang.CommentPrefix) {
			info.CommentLines++
		} else if strings.HasPrefix(trimmedLine, lang.MultiLineStart) {
			info.CommentLines++
			if !strings.Contains(trimmedLine, lang.MultiLineEnd) {
				inMultiLineComment = true
			}
		} else if trimmedLine != "" {
			info.CodeLines++
		}
	}

	// Extract package name
	if lang.PackagePattern != nil {
		matches := lang.PackagePattern.FindStringSubmatch(code)
		if len(matches) > 1 {
			info.Package = matches[1]
		}
	}

	// Extract imports
	if lang.ImportPattern != nil {
		matches := lang.ImportPattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			for i := 1; i < len(match); i++ {
				if match[i] != "" {
					info.Imports = append(info.Imports, match[i])
					break
				}
			}
		}
	}

	// Extract functions
	if lang.FunctionPattern != nil {
		matches := lang.FunctionPattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				info.Functions = append(info.Functions, match[1])
			} else if len(match) > 2 && match[2] != "" {
				info.Functions = append(info.Functions, match[2])
			} else if len(match) > 3 && match[3] != "" {
				info.Functions = append(info.Functions, match[3])
			}
		}
	}

	// Extract classes
	if lang.ClassPattern != nil {
		matches := lang.ClassPattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			if len(match) > 1 {
				info.Classes = append(info.Classes, match[1])
			}
		}
	}

	return info, nil
}

// guessLanguage attempts to guess the language from the code content
func guessLanguage(code string) *LanguageInfo {
	// Simple heuristics for language detection
	
	// Check for Go
	if strings.Contains(code, "package main") || strings.Contains(code, "import (") {
		// Create a copy of the language info to return its address
		langCopy := Languages["go"]
		return &langCopy
	}
	
	// Check for Python
	if strings.Contains(code, "def ") && strings.Contains(code, "import ") {
		langCopy := Languages["python"]
		return &langCopy
	}
	
	// Check for JavaScript
	if strings.Contains(code, "const ") && strings.Contains(code, "function ") {
		langCopy := Languages["javascript"]
		return &langCopy
	}
	
	// Check for TypeScript
	if strings.Contains(code, "interface ") || strings.Contains(code, ": string") {
		langCopy := Languages["typescript"]
		return &langCopy
	}
	
	// Check for Java
	if strings.Contains(code, "public class ") || strings.Contains(code, "import java.") {
		langCopy := Languages["java"]
		return &langCopy
	}
	
	return nil
}

// simpleCodeParse performs a simple parsing of code without language-specific rules
func simpleCodeParse(code, language string) *CodeInfo {
	lines := strings.Split(code, "\n")
	
	return &CodeInfo{
		Language:    language,
		LineCount:   len(lines),
		CommentLines: 0, // We don't know the comment syntax
		CodeLines:    len(lines), // Assuming all lines are code
	}
}

// ExtractImports extracts import statements from code
func ExtractImports(code, language string) []string {
	info, err := ParseCode(code, language)
	if err != nil || info == nil {
		return []string{}
	}
	return info.Imports
}

// AddImport adds an import statement to code
func AddImport(code, importStmt, language string) string {
	lang := GetLanguageByName(language)
	if lang == nil {
		// If language is unknown, just prepend the import
		return importStmt + "\n" + code
	}

	// Different strategies based on language
	switch language {
	case "go":
		// Find existing import block
		importBlockRegex := regexp.MustCompile(`import\s+\(\s*((?:.|\n)*?)\s*\)`)
		matches := importBlockRegex.FindStringSubmatchIndex(code)
		
		if len(matches) > 0 {
			// Add to existing import block
			importBlock := code[matches[2]:matches[3]]
			newImportBlock := importBlock + "\n\t" + importStmt
			return code[:matches[2]] + newImportBlock + code[matches[3]:]
		} else {
			// Find package declaration
			packageRegex := regexp.MustCompile(`package\s+\w+`)
			packageMatch := packageRegex.FindStringIndex(code)
			
			if len(packageMatch) > 0 {
				// Add after package declaration
				return code[:packageMatch[1]] + "\n\nimport (\n\t" + importStmt + "\n)" + code[packageMatch[1]:]
			}
		}
	case "python":
		// Find the last import statement
		importRegex := regexp.MustCompile(`(?m)^(?:from|import)\s+.*$`)
		matches := importRegex.FindAllStringIndex(code, -1)
		
		if len(matches) > 0 {
			lastImport := matches[len(matches)-1]
			return code[:lastImport[1]] + "\n" + importStmt + code[lastImport[1]:]
		} else {
			// Add at the beginning
			return importStmt + "\n\n" + code
		}
	case "javascript", "typescript":
		// Find the last import statement
		importRegex := regexp.MustCompile(`(?m)^import\s+.*$`)
		matches := importRegex.FindAllStringIndex(code, -1)
		
		if len(matches) > 0 {
			lastImport := matches[len(matches)-1]
			return code[:lastImport[1]] + "\n" + importStmt + code[lastImport[1]:]
		} else {
			// Add at the beginning
			return importStmt + "\n\n" + code
		}
	}

	// Default: add at the beginning
	return importStmt + "\n" + code
}