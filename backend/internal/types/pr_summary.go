package types

// FileGroup represents a logical grouping of related files
type FileGroup struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
	Importance  int      `json:"importance"`
}

// PRSummary represents a summary of a pull request
type PRSummary struct {
	Title              string      `json:"title"`
	Description        string      `json:"description"`
	MainPoints         []string    `json:"main_points"`
	KeyChanges         []string    `json:"key_changes"`
	FileGroups         []FileGroup `json:"file_groups"`
	PotentialImpact    string      `json:"potential_impact"`
	SuggestedReviewers []string    `json:"suggested_reviewers"`
	TechnicalDetails   string      `json:"technical_details"`
	PRNumber           int         `json:"pr_number"`
	PRURL              string      `json:"pr_url"`
	Repository         string      `json:"repository"`
	Author             string      `json:"author"`
	ChangedFiles       int         `json:"changed_files"`
	Additions          int         `json:"additions"`
	Deletions          int         `json:"deletions"`
}