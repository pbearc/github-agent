package models

// GitHubRepo contains information about a GitHub repository
type GitHubRepo struct {
	Owner         string            `json:"owner"`
	Name          string            `json:"name"`
	FullName      string            `json:"full_name"`
	Description   string            `json:"description"`
	DefaultBranch string            `json:"default_branch"`
	HTMLURL       string            `json:"html_url"`
	Language      string            `json:"language"`
	Languages     map[string]int    `json:"languages"`
	Topics        []string          `json:"topics"`
	Stars         int               `json:"stars"`
	Forks         int               `json:"forks"`
	HasReadme     bool              `json:"has_readme"`
	HasDockerfile bool              `json:"has_dockerfile"`
	License       *GitHubLicense    `json:"license,omitempty"`
	Files         []GitHubFile      `json:"files,omitempty"`
}

// GitHubLicense contains information about a repository's license
type GitHubLicense struct {
	Name string `json:"name"`
	SPDX string `json:"spdx_id"`
	URL  string `json:"url"`
}

// GitHubFile contains information about a file in a repository
type GitHubFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int    `json:"size"`
	Type        string `json:"type"` // "file" or "dir"
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url,omitempty"`
}

// GitHubContent contains information about content in a repository
type GitHubContent struct {
	Name        string        `json:"name"`
	Path        string        `json:"path"`
	SHA         string        `json:"sha"`
	Size        int           `json:"size"`
	Type        string        `json:"type"` // "file" or "dir"
	Content     string        `json:"content,omitempty"`
	Encoding    string        `json:"encoding,omitempty"`
	HTMLURL     string        `json:"html_url"`
	DownloadURL string        `json:"download_url,omitempty"`
	SubContents []GitHubContent `json:"contents,omitempty"` // For directories
}

// GitHubBranch contains information about a branch in a repository
type GitHubBranch struct {
	Name      string        `json:"name"`
	Protected bool          `json:"protected"`
	Commit    *GitHubCommit `json:"commit,omitempty"`
}

// GitHubCommit contains information about a commit
type GitHubCommit struct {
	SHA     string `json:"sha"`
	HTMLURL string `json:"html_url"`
	Message string `json:"message"`
}

// GitHubSearchResult contains information about a code search result
type GitHubSearchResult struct {
	TotalCount int                 `json:"total_count"`
	Items      []GitHubSearchItem  `json:"items"`
}

// GitHubSearchItem contains information about a single search result
type GitHubSearchItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	HTMLURL     string `json:"html_url"`
	Repository  *GitHubRepo `json:"repository,omitempty"`
	Score       float64 `json:"score"`
	TextMatches []GitHubTextMatch `json:"text_matches,omitempty"`
}

// GitHubTextMatch contains information about text matches in a search result
type GitHubTextMatch struct {
	Fragment string `json:"fragment"`
	Matches  []GitHubMatch `json:"matches"`
}

// GitHubMatch contains information about a single match in a text match
type GitHubMatch struct {
	Text  string `json:"text"`
	Index int    `json:"indices"`
}

// GitHubPushResponse contains information about a push operation
type GitHubPushResponse struct {
	CommitURL string `json:"commit_url"`
	SHA       string `json:"sha"`
	Message   string `json:"message"`
}