// internal/github/api_extensions.go
package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// CommitInfo contains summarized commit information
type CommitInfo struct {
	SHA           string    `json:"sha"`
	Message       string    `json:"message"`
	Author        string    `json:"author"`
	AuthorEmail   string    `json:"author_email,omitempty"`
	CommitDate    time.Time `json:"commit_date"`
	URL           string    `json:"url"`
	FilesChanged  []string  `json:"files_changed,omitempty"`
	LinesAdded    int       `json:"lines_added"`
	LinesDeleted  int       `json:"lines_deleted"`
}

// PullRequestInfo contains summarized pull request information
type PullRequestInfo struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	ClosedAt    time.Time `json:"closed_at,omitempty"`
	MergedAt    time.Time `json:"merged_at,omitempty"`
	Author      string    `json:"author"`
	URL         string    `json:"url"`
	Labels      []string  `json:"labels,omitempty"`
	Reviewers   []string  `json:"reviewers,omitempty"`
	Description string    `json:"description,omitempty"`
	Files       []string  `json:"files,omitempty"`
}

// IssueInfo contains summarized issue information
type IssueInfo struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	ClosedAt    time.Time `json:"closed_at,omitempty"`
	Author      string    `json:"author"`
	URL         string    `json:"url"`
	Labels      []string  `json:"labels,omitempty"`
	Assignees   []string  `json:"assignees,omitempty"`
	Description string    `json:"description,omitempty"`
}

// ReleaseInfo contains summarized release information
type ReleaseInfo struct {
	ID           int64     `json:"id"`
	TagName      string    `json:"tag_name"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	PublishedAt  time.Time `json:"published_at,omitempty"`
	Author       string    `json:"author"`
	URL          string    `json:"url"`
	PreRelease   bool      `json:"pre_release"`
	Description  string    `json:"description,omitempty"`
	Assets       []string  `json:"assets,omitempty"`
	AssetsCount  int       `json:"assets_count"`
	CommitsCount int       `json:"commits_count,omitempty"`
}

// StatsInfo contains repository statistics
type StatsInfo struct {
	CommitFrequency map[string]int `json:"commit_frequency"` // Date -> count
	Contributors    []ContributorStats `json:"contributors"`
	CodeFrequency   map[string]CodeStats `json:"code_frequency"` // Date -> stats
	ParticipationStats map[string]int `json:"participation"`    // Weekly Stats
	PunchCardStats     [][]int        `json:"punch_card"`       // Day, Hour, Count
}

// ContributorStats contains contributor statistics
type ContributorStats struct {
	Author    string `json:"author"`
	Commits   int    `json:"commits"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Weeks     int    `json:"weeks"`
}

// CodeStats contains code frequency statistics
type CodeStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}

// ContributorInfo contains contributor information
type ContributorInfo struct {
	Username      string `json:"username"`
	Contributions int    `json:"contributions"`
	URL           string `json:"url"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	IsBot         bool   `json:"is_bot"`
}

// LanguageInfo maps language to byte count
type LanguageInfo map[string]int

// GetCommits retrieves commits based on keywords
func (c *Client) GetCommits(ctx context.Context, owner, repo, branch string, keywords []string) ([]CommitInfo, error) {
	// Set up options for the commit list
	opts := &github.CommitsListOptions{
		SHA: branch, // Limit to specific branch if provided
		ListOptions: github.ListOptions{
			PerPage: 100, // Get a good number of commits
		},
	}

	// Get commits from the GitHub API
	commits, _, err := c.client.Repositories.ListCommits(ctx, owner, repo, opts)
	if err != nil {
		return nil, common.WrapError(err, "failed to list commits")
	}

	// Convert to our simplified commit info structure
	var commitInfos []CommitInfo
	var filteredCommits []CommitInfo

	for _, commit := range commits {
		if commit.SHA == nil || commit.Commit == nil || commit.Commit.Message == nil {
			continue // Skip invalid commits
		}

		// Create a CommitInfo object
		commitInfo := CommitInfo{
			SHA:     *commit.SHA,
			Message: *commit.Commit.Message,
		}

		// Extract author information if available
		if commit.Commit.Author != nil {
			if commit.Commit.Author.Name != nil {
				commitInfo.Author = *commit.Commit.Author.Name
			}
			if commit.Commit.Author.Email != nil {
				commitInfo.AuthorEmail = *commit.Commit.Author.Email
			}
			if commit.Commit.Author.Date != nil {
				commitInfo.CommitDate = *commit.Commit.Author.Date
			}
		}

		// Add URL if available
		if commit.HTMLURL != nil {
			commitInfo.URL = *commit.HTMLURL
		}

		// Get the commit details to find changed files
		if commit.SHA != nil {
			commitDetail, _, err := c.client.Repositories.GetCommit(ctx, owner, repo, *commit.SHA, nil)
			if err == nil && commitDetail.Files != nil {
				for _, file := range commitDetail.Files {
					if file.Filename != nil {
						commitInfo.FilesChanged = append(commitInfo.FilesChanged, *file.Filename)
					}
					if file.Additions != nil {
						commitInfo.LinesAdded += *file.Additions
					}
					if file.Deletions != nil {
						commitInfo.LinesDeleted += *file.Deletions
					}
				}
			}
		}

		commitInfos = append(commitInfos, commitInfo)
	}

	// If no keywords provided, return all commits
	if len(keywords) == 0 {
		// Limit to 50 most recent commits
		if len(commitInfos) > 50 {
			return commitInfos[:50], nil
		}
		return commitInfos, nil
	}

	// Filter commits by keywords
	for _, commit := range commitInfos {
		// Check if any keyword matches the commit message or files
		for _, keyword := range keywords {
			lowerKeyword := strings.ToLower(keyword)
			lowerMessage := strings.ToLower(commit.Message)
			
			if strings.Contains(lowerMessage, lowerKeyword) {
				filteredCommits = append(filteredCommits, commit)
				break // Once we've matched, no need to check other keywords
			}
			
			// Also check files changed
			for _, file := range commit.FilesChanged {
				lowerFile := strings.ToLower(file)
				if strings.Contains(lowerFile, lowerKeyword) {
					filteredCommits = append(filteredCommits, commit)
					break // Once we've matched, no need to check other files
				}
			}
		}
	}

	// If no commits matched the keywords, return the most recent commits
	if len(filteredCommits) == 0 {
		c.logger.Info("No commits matched the keywords, returning most recent commits")
		// Limit to 20 most recent commits
		if len(commitInfos) > 20 {
			return commitInfos[:20], nil
		}
		return commitInfos, nil
	}

	// Limit to 50 matching commits
	if len(filteredCommits) > 50 {
		return filteredCommits[:50], nil
	}

	return filteredCommits, nil
}

// GetPullRequests retrieves pull requests based on keywords
func (c *Client) GetPullRequests(ctx context.Context, owner, repo string, keywords []string) ([]PullRequestInfo, error) {
	// Set up options for the PR list
	opts := &github.PullRequestListOptions{
		State: "all", // Get both open and closed PRs
		ListOptions: github.ListOptions{
			PerPage: 100, // Get a good number of PRs
		},
	}

	// Get PRs from the GitHub API
	prs, _, err := c.client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, common.WrapError(err, "failed to list pull requests")
	}

	// Convert to our simplified PR info structure
	var prInfos []PullRequestInfo
	var filteredPRs []PullRequestInfo

	for _, pr := range prs {
		if pr.Number == nil || pr.Title == nil || pr.State == nil {
			continue // Skip invalid PRs
		}

		// Create a PullRequestInfo object
		prInfo := PullRequestInfo{
			Number: *pr.Number,
			Title:  *pr.Title,
			State:  *pr.State,
		}

		// Extract dates if available
		if pr.CreatedAt != nil {
			prInfo.CreatedAt = *pr.CreatedAt
		}
		if pr.ClosedAt != nil && !pr.ClosedAt.IsZero() {
			prInfo.ClosedAt = *pr.ClosedAt
		}
		if pr.MergedAt != nil && !pr.MergedAt.IsZero() {
			prInfo.MergedAt = *pr.MergedAt
		}

		// Extract author if available
		if pr.User != nil && pr.User.Login != nil {
			prInfo.Author = *pr.User.Login
		}

		// Add URL if available
		if pr.HTMLURL != nil {
			prInfo.URL = *pr.HTMLURL
		}

		// Extract labels if available
		if pr.Labels != nil {
			for _, label := range pr.Labels {
				if label.Name != nil {
					prInfo.Labels = append(prInfo.Labels, *label.Name)
				}
			}
		}

		// Extract description if available
		if pr.Body != nil {
			prInfo.Description = *pr.Body
		}

		// Get reviewers if available
		if pr.RequestedReviewers != nil {
			for _, reviewer := range pr.RequestedReviewers {
				if reviewer.Login != nil {
					prInfo.Reviewers = append(prInfo.Reviewers, *reviewer.Login)
				}
			}
		}

		// Get files changed if available
		if pr.Number != nil {
			files, _, err := c.client.PullRequests.ListFiles(ctx, owner, repo, *pr.Number, nil)
			if err == nil {
				for _, file := range files {
					if file.Filename != nil {
						prInfo.Files = append(prInfo.Files, *file.Filename)
					}
				}
			}
		}

		prInfos = append(prInfos, prInfo)
	}

	// If no keywords provided, return all PRs
	if len(keywords) == 0 {
		// Limit to 50 most recent PRs
		if len(prInfos) > 50 {
			return prInfos[:50], nil
		}
		return prInfos, nil
	}

	// Filter PRs by keywords
	for _, pr := range prInfos {
		// Check if any keyword matches the PR title, description, or files
		for _, keyword := range keywords {
			lowerKeyword := strings.ToLower(keyword)
			lowerTitle := strings.ToLower(pr.Title)
			lowerDesc := strings.ToLower(pr.Description)
			
			if strings.Contains(lowerTitle, lowerKeyword) || strings.Contains(lowerDesc, lowerKeyword) {
				filteredPRs = append(filteredPRs, pr)
				break // Once we've matched, no need to check other keywords
			}
			
			// Also check files changed
			for _, file := range pr.Files {
				lowerFile := strings.ToLower(file)
				if strings.Contains(lowerFile, lowerKeyword) {
					filteredPRs = append(filteredPRs, pr)
					break // Once we've matched, no need to check other files
				}
			}
			
			// Also check labels
			for _, label := range pr.Labels {
				lowerLabel := strings.ToLower(label)
				if strings.Contains(lowerLabel, lowerKeyword) {
					filteredPRs = append(filteredPRs, pr)
					break // Once we've matched, no need to check other labels
				}
			}
		}
	}

	// If no PRs matched the keywords, return the most recent PRs
	if len(filteredPRs) == 0 {
		c.logger.Info("No PRs matched the keywords, returning most recent PRs")
		// Limit to 20 most recent PRs
		if len(prInfos) > 20 {
			return prInfos[:20], nil
		}
		return prInfos, nil
	}

	// Limit to 50 matching PRs
	if len(filteredPRs) > 50 {
		return filteredPRs[:50], nil
	}

	return filteredPRs, nil
}

// GetIssues retrieves issues based on keywords
func (c *Client) GetIssues(ctx context.Context, owner, repo string, keywords []string) ([]IssueInfo, error) {
	// Set up options for the issue list
	opts := &github.IssueListByRepoOptions{
		State: "all", // Get both open and closed issues
		ListOptions: github.ListOptions{
			PerPage: 100, // Get a good number of issues
		},
	}

	// Get issues from the GitHub API
	issues, _, err := c.client.Issues.ListByRepo(ctx, owner, repo, opts)
	if err != nil {
		return nil, common.WrapError(err, "failed to list issues")
	}

	// Convert to our simplified issue info structure
	var issueInfos []IssueInfo
	var filteredIssues []IssueInfo

	for _, issue := range issues {
		// Skip pull requests (GitHub considers PRs as issues in the API)
		if issue.IsPullRequest() {
			continue
		}

		if issue.Number == nil || issue.Title == nil || issue.State == nil {
			continue // Skip invalid issues
		}

		// Create an IssueInfo object
		issueInfo := IssueInfo{
			Number: *issue.Number,
			Title:  *issue.Title,
			State:  *issue.State,
		}

		// Extract dates if available
		if issue.CreatedAt != nil {
			issueInfo.CreatedAt = *issue.CreatedAt
		}
		if issue.ClosedAt != nil && !issue.ClosedAt.IsZero() {
			issueInfo.ClosedAt = *issue.ClosedAt
		}

		// Extract author if available
		if issue.User != nil && issue.User.Login != nil {
			issueInfo.Author = *issue.User.Login
		}

		// Add URL if available
		if issue.HTMLURL != nil {
			issueInfo.URL = *issue.HTMLURL
		}

		// Extract labels if available
		if issue.Labels != nil {
			for _, label := range issue.Labels {
				if label.Name != nil {
					issueInfo.Labels = append(issueInfo.Labels, *label.Name)
				}
			}
		}

		// Extract assignees if available
		if issue.Assignees != nil {
			for _, assignee := range issue.Assignees {
				if assignee.Login != nil {
					issueInfo.Assignees = append(issueInfo.Assignees, *assignee.Login)
				}
			}
		}

		// Extract description if available
		if issue.Body != nil {
			issueInfo.Description = *issue.Body
		}

		issueInfos = append(issueInfos, issueInfo)
	}

	// If no keywords provided, return all issues
	if len(keywords) == 0 {
		// Limit to 50 most recent issues
		if len(issueInfos) > 50 {
			return issueInfos[:50], nil
		}
		return issueInfos, nil
	}

	// Filter issues by keywords
	for _, issue := range issueInfos {
		// Check if any keyword matches the issue title, description, or labels
		for _, keyword := range keywords {
			lowerKeyword := strings.ToLower(keyword)
			lowerTitle := strings.ToLower(issue.Title)
			lowerDesc := strings.ToLower(issue.Description)
			
			if strings.Contains(lowerTitle, lowerKeyword) || strings.Contains(lowerDesc, lowerKeyword) {
				filteredIssues = append(filteredIssues, issue)
				break // Once we've matched, no need to check other keywords
			}
			
			// Also check labels
			for _, label := range issue.Labels {
				lowerLabel := strings.ToLower(label)
				if strings.Contains(lowerLabel, lowerKeyword) {
					filteredIssues = append(filteredIssues, issue)
					break // Once we've matched, no need to check other labels
				}
			}
		}
	}

	// If no issues matched the keywords, return the most recent issues
	if len(filteredIssues) == 0 {
		c.logger.Info("No issues matched the keywords, returning most recent issues")
		// Limit to 20 most recent issues
		if len(issueInfos) > 20 {
			return issueInfos[:20], nil
		}
		return issueInfos, nil
	}

	// Limit to 50 matching issues
	if len(filteredIssues) > 50 {
		return filteredIssues[:50], nil
	}

	return filteredIssues, nil
}

// GetReleases retrieves repository releases
func (c *Client) GetReleases(ctx context.Context, owner, repo string) ([]ReleaseInfo, error) {
	// Set up options for the release list
	opts := &github.ListOptions{
		PerPage: 100, // Get a good number of releases
	}

	// Get releases from the GitHub API
	releases, _, err := c.client.Repositories.ListReleases(ctx, owner, repo, opts)
	if err != nil {
		return nil, common.WrapError(err, "failed to list releases")
	}

	// Convert to our simplified release info structure
	var releaseInfos []ReleaseInfo

	for _, release := range releases {
		if release.ID == nil || release.TagName == nil {
			continue // Skip invalid releases
		}

		// Create a ReleaseInfo object
		releaseInfo := ReleaseInfo{
			ID:         *release.ID,
			TagName:    *release.TagName,
			PreRelease: release.Prerelease != nil && *release.Prerelease,
		}

		// Extract name if available
		if release.Name != nil {
			releaseInfo.Name = *release.Name
		} else {
			releaseInfo.Name = releaseInfo.TagName // Use tag name as fallback
		}

		// Extract dates if available
		if release.CreatedAt != nil {
			releaseInfo.CreatedAt = release.CreatedAt.Time
		}
		if release.PublishedAt != nil && !release.PublishedAt.IsZero() {
			releaseInfo.PublishedAt = release.PublishedAt.Time
		}

		// Extract author if available
		if release.Author != nil && release.Author.Login != nil {
			releaseInfo.Author = *release.Author.Login
		}

		// Add URL if available
		if release.HTMLURL != nil {
			releaseInfo.URL = *release.HTMLURL
		}

		// Extract description if available
		if release.Body != nil {
			releaseInfo.Description = *release.Body
		}

		// Extract assets if available
		if release.Assets != nil {
			for _, asset := range release.Assets {
				if asset.Name != nil {
					releaseInfo.Assets = append(releaseInfo.Assets, *asset.Name)
				}
			}
			releaseInfo.AssetsCount = len(release.Assets)
		}

		// Try to get the number of commits between this release and the previous one
		// This is more complex and would require additional API calls

		releaseInfos = append(releaseInfos, releaseInfo)
	}

	return releaseInfos, nil
}

// GetRepositoryStats retrieves repository statistics
func (c *Client) GetRepositoryStats(ctx context.Context, owner, repo string) (*StatsInfo, error) {
	stats := &StatsInfo{
		CommitFrequency:   make(map[string]int),
		Contributors:      []ContributorStats{},
		CodeFrequency:     make(map[string]CodeStats),
		ParticipationStats: make(map[string]int),
		PunchCardStats:     [][]int{},
	}

	// Get contributor statistics
	contributorStats, _, err := c.client.Repositories.ListContributorsStats(ctx, owner, repo)
	if err != nil {
		c.logger.WithError(err).Warning("Failed to get contributor statistics")
	} else {
		for _, contributor := range contributorStats {
			if contributor.Author == nil || contributor.Author.Login == nil {
				continue
			}

			contribStat := ContributorStats{
				Author:  *contributor.Author.Login,
				Commits: 0,
			}

			// Sum up stats from all weeks
			if contributor.Weeks != nil {
				contribStat.Weeks = len(contributor.Weeks)
				for _, week := range contributor.Weeks {
					if week.Commits != nil {
						contribStat.Commits += *week.Commits
					}
					if week.Additions != nil {
						contribStat.Additions += *week.Additions
					}
					if week.Deletions != nil {
						contribStat.Deletions += *week.Deletions
					}
				}
			}

			stats.Contributors = append(stats.Contributors, contribStat)
		}
	}

	// Get commit activity
	commitActivity, _, err := c.client.Repositories.ListCommitActivity(ctx, owner, repo)
	if err != nil {
		c.logger.WithError(err).Warning("Failed to get commit activity")
	} else {
		for _, week := range commitActivity {
			if week.Week == nil || week.Total == nil {
				continue
			}

			// Convert Unix timestamp to date string
			date := week.Week.Time.Format("2006-01-02")
			stats.CommitFrequency[date] = *week.Total
		}
	}

	// Get code frequency
	codeFrequency, _, err := c.client.Repositories.ListCodeFrequency(ctx, owner, repo)
	if err != nil {
		c.logger.WithError(err).Warning("Failed to get code frequency")
	} else {
		for _, week := range codeFrequency {
			if week.Week != nil && week.Additions != nil && week.Deletions != nil {
				// Convert Unix timestamp to date string
				date := week.Week.Time.Format("2006-01-02")
				stats.CodeFrequency[date] = CodeStats{
					Additions: *week.Additions,
					Deletions: *week.Deletions,
				}
			}
		}
	}

	// Get participation statistics
	participation, _, err := c.client.Repositories.ListParticipation(ctx, owner, repo)
	if err != nil {
		c.logger.WithError(err).Warning("Failed to get participation statistics")
	} else {
		if participation.All != nil {
			for i, count := range participation.All {
				weekKey := fmt.Sprintf("week_%d", i)
				stats.ParticipationStats[weekKey] = count
			}
		}
	}

	// Get punch card statistics
	punchCard, _, err := c.client.Repositories.ListPunchCard(ctx, owner, repo)
	if err != nil {
		c.logger.WithError(err).Warning("Failed to get punch card statistics")
	} else {
		// Convert punchCard to [][]int
		for _, pc := range punchCard {
			if pc != nil {
				stats.PunchCardStats = append(stats.PunchCardStats, []int{*pc.Day, *pc.Hour, *pc.Commits})
			}
		}
	}

	return stats, nil
}

// GetContributors retrieves repository contributors
func (c *Client) GetContributors(ctx context.Context, owner, repo string) ([]ContributorInfo, error) {
	// Set up options for the contributor list
	opts := &github.ListContributorsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100, // Get a good number of contributors
		},
	}

	// Get contributors from the GitHub API
	contributors, _, err := c.client.Repositories.ListContributors(ctx, owner, repo, opts)
	if err != nil {
		return nil, common.WrapError(err, "failed to list contributors")
	}

	// Convert to our simplified contributor info structure
	var contributorInfos []ContributorInfo

	for _, contributor := range contributors {
		if contributor.Login == nil {
			continue // Skip invalid contributors
		}

		// Create a ContributorInfo object
		contributorInfo := ContributorInfo{
			Username:      *contributor.Login,
			Contributions: 0,
		}

		// Extract contributions count if available
		if contributor.Contributions != nil {
			contributorInfo.Contributions = *contributor.Contributions
		}

		// Add URLs if available
		if contributor.HTMLURL != nil {
			contributorInfo.URL = *contributor.HTMLURL
		}
		if contributor.AvatarURL != nil {
			contributorInfo.AvatarURL = *contributor.AvatarURL
		}

		// Check if it's a bot (simple heuristic)
		contributorInfo.IsBot = strings.HasSuffix(contributorInfo.Username, "[bot]") ||
			strings.HasSuffix(contributorInfo.Username, "-bot") ||
			strings.Contains(contributorInfo.Username, "bot")

		contributorInfos = append(contributorInfos, contributorInfo)
	}

	return contributorInfos, nil
}

// GetRepositoryLanguages retrieves the languages used in a repository
func (c *Client) GetRepositoryLanguages(ctx context.Context, owner, repo string) (LanguageInfo, error) {
	languages, _, err := c.client.Repositories.ListLanguages(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to list languages")
	}

	return LanguageInfo(languages), nil
}

// GetRepositoryTopics retrieves the topics/tags of a repository
func (c *Client) GetRepositoryTopics(ctx context.Context, owner, repo string) ([]string, error) {
	topics, _, err := c.client.Repositories.ListAllTopics(ctx, owner, repo)
	if err != nil {
		return nil, common.WrapError(err, "failed to list topics")
	}

	return topics, nil
}